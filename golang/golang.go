package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"

	"github.com/bblfsh/sdk/v3/uast/nodes"
)

func ParseString(code string) (*ast.File, *token.FileSet, error) {
	fs := token.NewFileSet()
	tree, err := parser.ParseFile(fs, "input.go", code, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return tree, fs, nil
}

func Parse(code string) (nodes.Node, error) {
	f, fs, err := ParseString(code)
	if err != nil {
		return nil, err
	}
	return ValueToNode(reflect.ValueOf(f), fs)
}
