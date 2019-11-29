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

const debug = false

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
		return fmt.Errorf("error parsing source sample: %v", err)
	}

	out, err := parseNodeHack(r.after)
	if err != nil {
		return fmt.Errorf("error parsing destination sample: %v", err)
	}

	if debug {
		dump(in, "before.yml")
		dump(out, "after.yml")
	}

	_, isArr1 := in.(nodes.Array)
	_, isArr2 := out.(nodes.Array)

	var inOp, outOp transformer.Op
	if isArr1 || isArr2 {
		// both should be arrays for out custom operator to work
		if !isArr1 {
			in = nodes.Array{in}
		}
		if !isArr2 {
			out = nodes.Array{out}
		}
		// make sure we can convert part of an array: we need a custom operation for it
		// left side (in) always does Check and the right side (out) performs Construct
		inOp = &matroshka.MatroshkaArray{Op: nodeToOp(in).(transformer.ArrayOp)}
		outOp = &matroshka.MatroshkaArray{Op: nodeToOp(out).(transformer.ArrayOp)}
	} else {
		inOp = nodeToOp(in)
		outOp = nodeToOp(out)
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
	snippet = strings.TrimSpace(snippet)
	root, err := parseAsType(snippet)
	if err == nil {
		return trimPositions(root)
	}
	root, err = parseAsExpr(snippet)
	if err == nil {
		return trimPositions(root)
	}
	root, err = parseAsDecls(snippet)
	if err == nil {
		return trimPositions(root)
	}
	root, err = parseAsStmts(snippet)
	if err == nil {
		return trimPositions(root)
	}
	return nil, err
}

func unwrapFile(root nodes.Node) nodes.Array {
	return root.(nodes.Object)["Decls"].(nodes.Array)
}

func parseAsDecls(code string) (nodes.Node, error) {
	wrapped, err := golang.Parse(fmt.Sprintf(`package main

%s
`, code))
	if err != nil {
		return nil, err
	}
	arr := unwrapFile(wrapped)
	if len(arr) == 1 {
		return arr[0], nil
	}
	return arr, nil
}

func parseAsStmts(code string) (nodes.Node, error) {
	wrapped, err := golang.Parse(fmt.Sprintf(`package main
func main() {
	%s
}
`, code))
	if err != nil {
		return nil, err
	}
	arr := unwrapFile(wrapped)[0].(nodes.Object)["Body"].(nodes.Object)["List"].(nodes.Array)
	if len(arr) == 1 {
		return arr[0], nil
	}
	return arr, nil
}

func parseAsExpr(code string) (nodes.Node, error) {
	wrapped, err := golang.Parse(fmt.Sprintf(`package main
var _ = %s
`, code))
	if err != nil {
		return nil, err
	}
	n := unwrapFile(wrapped)[0].(nodes.Object)["Specs"].(nodes.Array)[0].(nodes.Object)["Values"].(nodes.Array)[0]
	return n, nil
}

func parseAsType(code string) (nodes.Node, error) {
	wrapped, err := golang.Parse(fmt.Sprintf(`package main
var _ %s
`, code))
	if err != nil {
		return nil, err
	}
	n := unwrapFile(wrapped)[0].(nodes.Object)["Specs"].(nodes.Array)[0].(nodes.Object)["Type"]
	return n, nil
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

func dump(n nodes.Node, filePath string) {
	data, err := uastyaml.Marshal(n)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		panic(err)
	}
}
