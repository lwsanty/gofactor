package gofactor

import (
	"bytes"
	"fmt"
	"go/format"
	"go/printer"
	"go/token"
	"io/ioutil"
	"strings"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
	"github.com/bblfsh/sdk/v3/uast/uastyaml"
	"github.com/lwsanty/gofactor/golang"
	"github.com/lwsanty/gofactor/transform/matroshka"
	"github.com/lwsanty/gofactor/transform/vartransform"
)

const (
	debug = false

	mainTemplate = `package main

func main() {
	%s
}`
)

type Refactor struct {
	before string
	after  string
	m      transformer.Transformer
}

func NewRefactor(before, after string) (*Refactor, error) {
	r := &Refactor{
		before: before,
		after:  after,
	}

	if err := r.prepare(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Refactor) prepare() error {
	in, err := parseNodeHack(r.before)
	if err != nil {
		return err
	}

	out, err := parseNodeHack(r.after)
	if err != nil {
		return err
	}

	inE, isE1 := asExpr(in)
	outE, isE2 := asExpr(out)
	if isE1 && isE2 {
		in, out = inE, outE
	}

	if debug {
		dump(in, "before.yml")
		dump(out, "after.yml")
	}

	var inOp, outOp transformer.Op
	if isE1 && isE2 {
		// both sides are expression pattern - omit array ops and match directly on objects
		inOp = nodeToOp(in)
		outOp = nodeToOp(out)
	} else {
		// we need both: left side always does Check and the right side performs Construct
		inOp = &matroshka.MatroshkaArray{Op: nodeToOp(in).(transformer.ArrayOp)}
		outOp = &matroshka.MatroshkaArray{Op: nodeToOp(out).(transformer.ArrayOp)}
	}

	r.m = transformer.Mappings(transformer.Map(inOp, outOp))
	return nil
}

func (r *Refactor) Apply(code string) (string, error) {
	test, err := golang.Parse(code)
	if err != nil {
		return "", err
	}

	test, err = trimPositions(test)
	if err != nil {
		return "", err
	}

	if debug {
		dump(test, "test.yml")
	}

	res, err := r.m.Do(test)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := printer.Fprint(buf, token.NewFileSet(), golang.NodeToAST(res)); err != nil {
		return "", err
	}

	fdata, err := format.Source(buf.Bytes())
	if err != nil {
		return "", err
	}

	return string(fdata), nil
}

func trimPositions(n nodes.Node) (nodes.Node, error) {
	return transformer.Mappings(transformer.Map(
		transformer.Part("_", transformer.Obj{uast.KeyPos: transformer.Any()}),
		transformer.Part("_", transformer.Obj{}),
	)).Do(n)
}

func nodeToOp(n nodes.Node) transformer.Op {
	switch o := n.(type) {
	case nil:
		return transformer.Is(o)
	case nodes.Value:
		return transformer.Is(o)
	case nodes.Object:
		if uast.TypeOf(o) == "Ident" {
			name := o["Name"]
			str := name.(nodes.String)
			if strings.HasPrefix(string(str), "X") {
				return vartransform.Var(string(str))
			}
		}
		// transformer.Fields is extended version of transformer.Obj
		var res transformer.Fields
		for k, v := range o {
			// conditions to drop pos
			var field = transformer.Field{Name: k}
			if k == uast.KeyPos {
				field.Drop = true
				field.Op = transformer.Any()
			} else {
				field.Op = nodeToOp(v)
			}
			res = append(res, field)
		}
		return res
	case nodes.Array:
		var res []transformer.Op
		for _, node := range o {
			res = append(res, nodeToOp(node))
		}
		return transformer.Arr(res...)
	default:
		panic("not supported type " + o.Kind().String())
	}
}

// TODO functions
func parseNodeHack(snippet string) (nodes.Node, error) {
	wrapped, err := golang.Parse(wrapInMain(snippet))
	if err != nil {
		return nil, err
	}

	list := wrapped.(nodes.Object)["Decls"].(nodes.Array)[0].(nodes.Object)["Body"].(nodes.Object)["List"].(nodes.Array)
	return trimPositions(list)
}

// asExpr tries to convert a given AST node to an expression.
func asExpr(n nodes.Node) (nodes.Node, bool) {
	x := n
	if arr, ok := n.(nodes.Array); ok {
		if len(arr) != 1 {
			// set of statements, not a single expression
			return n, false
		}
		x = arr[0]
	}
	if uast.TypeOf(x) != "ExprStmt" {
		// another statement, probably
		return n, false
	}
	// get underlying expression
	return x.(nodes.Object)["X"], true
}

// TODO gofmt
func wrapInMain(code string) string {
	return fmt.Sprintf(mainTemplate, code)
}

func dump(n nodes.Node, filePath string) {
	data, err := uastyaml.Marshal(n)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		panic(err)
	}
}
