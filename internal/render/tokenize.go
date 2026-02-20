package render

import (
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"image/color"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jacoelho/gophershot/internal/lang"
)

type tokenClass int

const (
	tokenClassDefault tokenClass = iota
	tokenClassKeyword
	tokenClassTypeOrBuiltin
	tokenClassString
	tokenClassNumber
	tokenClassComment
	tokenClassOperator
)

var (
	goBuiltins = map[string]struct{}{
		"append": {}, "bool": {}, "byte": {}, "cap": {}, "close": {}, "complex": {}, "complex64": {},
		"complex128": {}, "copy": {}, "delete": {}, "error": {}, "false": {}, "float32": {}, "float64": {},
		"imag": {}, "int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "iota": {}, "len": {},
		"make": {}, "new": {}, "nil": {}, "panic": {}, "print": {}, "println": {}, "real": {}, "recover": {},
		"rune": {}, "string": {}, "true": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
		"uintptr": {},
	}
	terraformKeywords = map[string]struct{}{
		"data": {}, "false": {}, "for": {}, "if": {}, "in": {}, "locals": {}, "module": {}, "null": {},
		"output": {}, "provider": {}, "resource": {}, "terraform": {}, "true": {}, "variable": {},
	}
)

type segmentBuilder struct {
	lines [][]segment
}

func newSegmentBuilder() *segmentBuilder {
	return &segmentBuilder{lines: [][]segment{{}}}
}

func (b *segmentBuilder) appendText(text string, class tokenClass) {
	if text == "" {
		return
	}

	parts := strings.Split(normalizeRenderableText(text), "\n")
	for i, part := range parts {
		if part != "" {
			b.appendSegment(part, colorForClass(class))
		}
		if i < len(parts)-1 {
			b.lines = append(b.lines, []segment{})
		}
	}
}

func (b *segmentBuilder) appendSegment(text string, col color.RGBA) {
	line := b.lines[len(b.lines)-1]
	if len(line) > 0 {
		last := line[len(line)-1]
		if last.color == col {
			line[len(line)-1].text += text
			b.lines[len(b.lines)-1] = line
			return
		}
	}
	b.lines[len(b.lines)-1] = append(line, segment{text: text, color: col})
}

func (b *segmentBuilder) build() [][]segment {
	if len(b.lines) == 0 {
		return [][]segment{{}}
	}
	return b.lines
}

func tokenize(src string, language lang.Language) [][]segment {
	switch language {
	case lang.LanguageTerraform:
		return tokenizeTerraform(src)
	default:
		return tokenizeGo(src)
	}
}

func tokenizeGo(src string) [][]segment {
	builder := newSegmentBuilder()
	if src == "" {
		return builder.build()
	}

	highlighted := collectGoHighlightedIdentifierOffsets(src)

	fset := token.NewFileSet()
	file := fset.AddFile("source.go", fset.Base(), len(src))

	var s scanner.Scanner
	s.Init(file, []byte(src), nil, scanner.ScanComments)

	cursor := 0
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		start := file.Offset(pos)
		if start < cursor {
			start = cursor
		}
		if start > len(src) {
			start = len(src)
		}

		if start > cursor {
			builder.appendText(src[cursor:start], tokenClassDefault)
		}

		// The Go scanner synthesizes semicolons at line ends (lit == "\n").
		// Keep source output faithful by leaving those bytes in the default gap.
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}

		end := start + tokenTextLen(tok, lit)
		if end < start {
			end = start
		}
		if end > len(src) {
			end = len(src)
		}

		text := src[start:end]
		builder.appendText(text, goTokenClass(tok, src, text, start, end, highlighted))
		cursor = end
	}

	if cursor < len(src) {
		builder.appendText(src[cursor:], tokenClassDefault)
	}

	return builder.build()
}

func tokenizeTerraform(src string) [][]segment {
	builder := newSegmentBuilder()
	i := 0
	for i < len(src) {
		if end, ok := scanTerraformHeredoc(src, i); ok {
			builder.appendText(src[i:end], tokenClassString)
			i = end
			continue
		}

		switch {
		case strings.HasPrefix(src[i:], "//"):
			end := scanLineComment(src, i)
			builder.appendText(src[i:end], tokenClassComment)
			i = end
			continue
		case strings.HasPrefix(src[i:], "/*"):
			end := scanBlockComment(src, i)
			builder.appendText(src[i:end], tokenClassComment)
			i = end
			continue
		case src[i] == '#':
			end := scanLineComment(src, i)
			builder.appendText(src[i:end], tokenClassComment)
			i = end
			continue
		}

		switch src[i] {
		case '"', '\'', '`':
			end := scanStringLiteral(src, i, src[i])
			builder.appendText(src[i:end], tokenClassString)
			i = end
			continue
		}

		if isNumberStart(src, i) {
			end := scanNumber(src, i)
			builder.appendText(src[i:end], tokenClassNumber)
			i = end
			continue
		}

		r, size := utf8.DecodeRuneInString(src[i:])
		if isIdentifierStart(r) {
			end := scanIdentifier(src, i)
			ident := src[i:end]

			class := tokenClassDefault
			if _, ok := terraformKeywords[ident]; ok {
				class = tokenClassKeyword
			}
			builder.appendText(ident, class)
			i = end
			continue
		}

		if isOperatorRune(r) {
			builder.appendText(src[i:i+size], tokenClassOperator)
			i += size
			continue
		}

		builder.appendText(src[i:i+size], tokenClassDefault)
		i += size
	}

	return builder.build()
}

func colorForClass(class tokenClass) color.RGBA {
	switch class {
	case tokenClassKeyword:
		return color.RGBA{R: 0xCF, G: 0x22, B: 0x5B, A: 0xFF}
	case tokenClassTypeOrBuiltin:
		return color.RGBA{R: 0x82, G: 0x55, B: 0xD7, A: 0xFF}
	case tokenClassString:
		return color.RGBA{R: 0x0A, G: 0x7E, B: 0x39, A: 0xFF}
	case tokenClassNumber:
		return color.RGBA{R: 0x05, G: 0x5D, B: 0xB5, A: 0xFF}
	case tokenClassComment:
		return color.RGBA{R: 0x57, G: 0x6A, B: 0x73, A: 0xFF}
	case tokenClassOperator:
		return color.RGBA{R: 0x24, G: 0x2A, B: 0x2F, A: 0xFF}
	default:
		return defaultColor
	}
}

func scanLineComment(src string, start int) int {
	for i := start; i < len(src); i++ {
		if src[i] == '\n' {
			return i
		}
	}
	return len(src)
}

func scanBlockComment(src string, start int) int {
	for i := start + 2; i < len(src)-1; i++ {
		if src[i] == '*' && src[i+1] == '/' {
			return i + 2
		}
	}
	return len(src)
}

func scanStringLiteral(src string, start int, quote byte) int {
	if quote == '`' {
		for i := start + 1; i < len(src); i++ {
			if src[i] == '`' {
				return i + 1
			}
		}
		return len(src)
	}

	escaped := false
	for i := start + 1; i < len(src); i++ {
		switch {
		case escaped:
			escaped = false
		case src[i] == '\\':
			escaped = true
		case src[i] == quote:
			return i + 1
		case src[i] == '\n':
			return i
		}
	}
	return len(src)
}

func scanNumber(src string, start int) int {
	i := start
	hasDot := false
	if src[i] == '.' {
		hasDot = true
		i++
	}

	for i < len(src) {
		ch := src[i]
		switch {
		case isDigit(ch):
			i++
		case ch == '_' || isHexLetter(ch):
			i++
		case ch == '.' && !hasDot:
			hasDot = true
			i++
		case (ch == '+' || ch == '-') && i > start && isExponentMarker(src[i-1]):
			i++
		default:
			return i
		}
	}
	return i
}

func scanIdentifier(src string, start int) int {
	i := start
	_, size := utf8.DecodeRuneInString(src[i:])
	i += size
	for i < len(src) {
		r, nextSize := utf8.DecodeRuneInString(src[i:])
		if isIdentifierPart(r) {
			i += nextSize
			continue
		}
		break
	}
	return i
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	return r == '-'
}

func isNumberStart(src string, i int) bool {
	if isDigit(src[i]) {
		return true
	}
	return src[i] == '.' && i+1 < len(src) && isDigit(src[i+1])
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') || ch == 'x' || ch == 'X' || ch == 'o' || ch == 'O' || ch == 'b' || ch == 'B' || ch == 'e' || ch == 'E' || ch == 'p' || ch == 'P'
}

func isExponentMarker(ch byte) bool {
	return ch == 'e' || ch == 'E' || ch == 'p' || ch == 'P'
}

func isOperatorRune(r rune) bool {
	switch r {
	case '+', '-', '*', '/', '%', '=', '&', '|', '!', '<', '>', '^', '~', ':', '?', '.', ',', ';', '(', ')', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

func tokenTextLen(tok token.Token, lit string) int {
	if lit != "" {
		return len(lit)
	}

	if tok.IsKeyword() || tok.IsOperator() {
		return len(tok.String())
	}

	return 1
}

func goTokenClass(tok token.Token, src, text string, start, end int, highlighted map[int]struct{}) tokenClass {
	switch {
	case tok == token.COMMENT:
		return tokenClassComment
	case tok == token.STRING || tok == token.CHAR:
		return tokenClassString
	case tok == token.INT || tok == token.FLOAT || tok == token.IMAG:
		return tokenClassNumber
	case tok.IsKeyword():
		return tokenClassKeyword
	case tok == token.IDENT:
		if _, ok := goBuiltins[text]; ok {
			return tokenClassTypeOrBuiltin
		}
		if _, ok := highlighted[start]; ok {
			return tokenClassTypeOrBuiltin
		}
		if shouldColorIdentifierByContext(src, start, end) {
			return tokenClassTypeOrBuiltin
		}
		return tokenClassDefault
	case tok.IsOperator():
		return tokenClassOperator
	default:
		return tokenClassDefault
	}
}

func collectGoHighlightedIdentifierOffsets(src string) map[int]struct{} {
	if highlighted, ok := parseGoHighlightedIdentifierOffsets(src, 0); ok {
		return highlighted
	}

	const prefix = "package p\n"
	if highlighted, ok := parseGoHighlightedIdentifierOffsets(prefix+src, len(prefix)); ok {
		return highlighted
	}

	return map[int]struct{}{}
}

func parseGoHighlightedIdentifierOffsets(src string, offsetAdjust int) (map[int]struct{}, bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", src, parser.SkipObjectResolution|parser.ParseComments)
	if err != nil {
		return nil, false
	}

	out := make(map[int]struct{})
	ast.Inspect(file, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			markIdentOffset(fset, v.Name, offsetAdjust, out)
		case *ast.SelectorExpr:
			markIdentOffset(fset, v.Sel, offsetAdjust, out)
		case *ast.CallExpr:
			switch fn := v.Fun.(type) {
			case *ast.Ident:
				markIdentOffset(fset, fn, offsetAdjust, out)
			case *ast.SelectorExpr:
				markIdentOffset(fset, fn.Sel, offsetAdjust, out)
			}
		}
		return true
	})

	return out, true
}

func markIdentOffset(fset *token.FileSet, ident *ast.Ident, offsetAdjust int, out map[int]struct{}) {
	if ident == nil {
		return
	}
	pos := fset.PositionFor(ident.Pos(), false)
	offset := pos.Offset - offsetAdjust
	if offset < 0 {
		return
	}
	out[offset] = struct{}{}
}

func shouldColorIdentifierByContext(src string, start, end int) bool {
	prev, ok := previousNonSpaceByte(src, start)
	if ok && prev == '.' {
		return true
	}

	next, ok := nextNonSpaceByte(src, end)
	if ok && next == '(' {
		return true
	}

	return false
}

func previousNonSpaceByte(src string, before int) (byte, bool) {
	for i := before - 1; i >= 0; i-- {
		if isWhitespaceByte(src[i]) {
			continue
		}
		return src[i], true
	}
	return 0, false
}

func nextNonSpaceByte(src string, from int) (byte, bool) {
	for i := from; i < len(src); i++ {
		if isWhitespaceByte(src[i]) {
			continue
		}
		return src[i], true
	}
	return 0, false
}

func isWhitespaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func scanTerraformHeredoc(src string, start int) (int, bool) {
	if !strings.HasPrefix(src[start:], "<<") {
		return 0, false
	}

	i := start + 2
	allowIndent := false
	if i < len(src) && src[i] == '-' {
		allowIndent = true
		i++
	}

	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}

	markerStart := i
	for i < len(src) {
		r, size := utf8.DecodeRuneInString(src[i:])
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			i += size
			continue
		}
		break
	}
	if markerStart == i {
		return 0, false
	}
	marker := src[markerStart:i]

	lineEnd := i
	for lineEnd < len(src) && src[lineEnd] != '\n' {
		lineEnd++
	}
	if lineEnd < len(src) {
		lineEnd++
	}

	cursor := lineEnd
	for cursor < len(src) {
		next := cursor
		for next < len(src) && src[next] != '\n' {
			next++
		}
		line := src[cursor:next]
		check := line
		if allowIndent {
			check = strings.TrimLeft(line, " \t")
		}
		if check == marker {
			if next < len(src) {
				next++
			}
			return next, true
		}
		if next < len(src) {
			cursor = next + 1
			continue
		}
		return len(src), true
	}

	return len(src), true
}
