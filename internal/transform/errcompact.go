package transform

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"strings"

	"github.com/jacoelho/gophershot/internal/doc"
)

type errCompactTransform struct{}

func (errCompactTransform) Name() string {
	return "errcompact"
}

func (errCompactTransform) Apply(input doc.Document) (doc.Document, error) {
	src := input.Bytes()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.SkipObjectResolution|parser.ParseComments)
	if err != nil {
		return doc.Document{}, fmt.Errorf("parse source: %w", err)
	}

	filePos := fset.File(file.Pos())
	if filePos == nil {
		return doc.Document{}, fmt.Errorf("resolve source positions")
	}

	replacements := collectErrCompactReplacements(file, filePos, src, input.Lines)
	if len(replacements) == 0 {
		return input.Clone(), nil
	}

	outLines := cloneLines(input.Lines)
	for i := len(replacements) - 1; i >= 0; i-- {
		r := replacements[i]
		replacement := doc.Line{Text: r.text, Origins: slices.Clone(r.origins)}
		outLines = append(outLines[:r.startLine], append([]doc.Line{replacement}, outLines[r.endLine+1:]...)...)
	}

	return doc.Document{Lines: outLines}, nil
}

type lineReplacement struct {
	startLine int
	endLine   int
	text      string
	origins   []int
}

func collectErrCompactReplacements(file *ast.File, filePos *token.File, src []byte, lines []doc.Line) []lineReplacement {
	replacements := make([]lineReplacement, 0)

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if ifStmt.Else != nil || !isErrNotNil(ifStmt.Cond) {
			return true
		}

		startOffset := filePos.Offset(ifStmt.If)
		bodyStartOffset := filePos.Offset(ifStmt.Body.Lbrace)
		if startOffset < 0 || bodyStartOffset < startOffset || bodyStartOffset > len(src) {
			return true
		}
		header := normalizeInlineWhitespace(string(src[startOffset:bodyStartOffset]))
		if header == "" {
			return true
		}

		startLine := filePos.Line(ifStmt.If) - 1
		endLine := filePos.Line(ifStmt.End()) - 1
		if startLine < 0 || endLine < startLine || endLine >= len(lines) {
			return true
		}

		replacements = append(replacements, lineReplacement{
			startLine: startLine,
			endLine:   endLine,
			text:      leadingIndent(lines[startLine].Text) + header + " { /* ... */ }",
			origins:   mergeOrigins(lines[startLine : endLine+1]),
		})
		return true
	})

	if len(replacements) == 0 {
		return nil
	}

	slices.SortFunc(replacements, func(a, b lineReplacement) int {
		if a.startLine == b.startLine {
			return cmp.Compare(b.endLine, a.endLine)
		}
		return cmp.Compare(a.startLine, b.startLine)
	})

	filtered := make([]lineReplacement, 0, len(replacements))
	lastEnd := -1
	for _, r := range replacements {
		if r.startLine <= lastEnd {
			continue
		}
		filtered = append(filtered, r)
		lastEnd = r.endLine
	}

	return filtered
}

func mergeOrigins(lines []doc.Line) []int {
	if len(lines) == 0 {
		return nil
	}

	set := make(map[int]struct{})
	for _, line := range lines {
		for _, origin := range line.Origins {
			set[origin] = struct{}{}
		}
	}

	origins := slices.Collect(maps.Keys(set))
	slices.Sort(origins)
	return origins
}

func cloneLines(lines []doc.Line) []doc.Line {
	out := make([]doc.Line, len(lines))
	for i, line := range lines {
		out[i] = doc.Line{Text: line.Text, Origins: slices.Clone(line.Origins)}
	}
	return out
}

func normalizeInlineWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func leadingIndent(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}
	return s
}

func isErrNotNil(cond ast.Expr) bool {
	binary, ok := unwrapParens(cond).(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		return false
	}

	left := unwrapParens(binary.X)
	right := unwrapParens(binary.Y)
	if isIdent(left, "err") && isNil(right) {
		return true
	}
	if isNil(left) && isIdent(right, "err") {
		return true
	}
	return false
}

func isIdent(expr ast.Expr, name string) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == name
}

func isNil(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func unwrapParens(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}
