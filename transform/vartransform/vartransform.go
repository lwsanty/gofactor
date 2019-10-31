package vartransform

import (
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
)

func Var(name string) transformer.MappingOp {
	return opVar{name: name, kinds: nodes.KindsAny}
}

// original SDK's opVar does strict assertion to variable name, we need a softer check
type opVar struct {
	name  string
	kinds nodes.Kind
}

func (op opVar) Mapping() (src, dst transformer.Op) {
	return op, op
}

func (op opVar) Kinds() nodes.Kind {
	return op.kinds
}

func (op opVar) Check(st *transformer.State, n nodes.Node) (bool, error) {
	if err := st.SetVar(op.name, n); err != nil {
		if transformer.ErrVariableRedeclared.Is(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (op opVar) Construct(st *transformer.State, n nodes.Node) (nodes.Node, error) {
	if err := noNode(n); err != nil {
		return nil, err
	}
	val, err := st.MustGetVar(op.name)
	if err != nil {
		return nil, err
	}
	// TODO: should we clone it?
	return val, nil
}

func noNode(n nodes.Node) error {
	if n == nil {
		return nil
	}
	return transformer.ErrUnexpectedNode.New(n)
}
