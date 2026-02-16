package transform

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"

	"github.com/jacoelho/gophershot/internal/doc"
)

type stripImportsTransform struct{}

func (stripImportsTransform) Name() string {
	return "stripimports"
}

func (stripImportsTransform) Apply(input doc.Document) (doc.Document, error) {
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

	remove := make(map[int]struct{})
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}

		startLine := filePos.Line(gen.Pos()) - 1
		endLine := filePos.Line(gen.End()) - 1
		if startLine < 0 || endLine < startLine {
			continue
		}
		if endLine >= len(input.Lines) {
			endLine = len(input.Lines) - 1
		}
		for i := startLine; i <= endLine; i++ {
			remove[i] = struct{}{}
		}
	}

	if len(remove) == 0 {
		return input.Clone(), nil
	}

	out := make([]doc.Line, 0, len(input.Lines)-len(remove))
	for i, line := range input.Lines {
		if _, ok := remove[i]; ok {
			continue
		}
		out = append(out, doc.Line{Text: line.Text, Origins: slices.Clone(line.Origins)})
	}

	return doc.Document{Lines: out}, nil
}
