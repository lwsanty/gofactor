package matroshka

import (
	"fmt"

	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
)

type MatroshkaArray struct {
	Op transformer.ArrayOp
}

// Kinds defines nodes type/object/value to match to
func (*MatroshkaArray) Kinds() nodes.Kind {
	return nodes.KindArray
}

// Checks checks if node is array and returns transformation
func (m *MatroshkaArray) Check(st *transformer.State, n nodes.Node) (bool, error) {
	return m.checkMultipleMatches(st, n)
}

func (m *MatroshkaArray) checkSingleMatch(st *transformer.State, n nodes.Node) (bool, error) {
	arr, ok := n.(nodes.Array)
	if !ok {
		return false, nil
	}

	windowArr, err := m.Op.Arr(st)
	if err != nil {
		return false, err
	}
	// refactor window length
	windowLen := len(windowArr)
	if windowLen == 0 {
		return false, fmt.Errorf("this should not happen")
	}

	// if windows length is higher then array observed => did not match
	lenArr := len(arr)
	if windowLen > len(arr) {
		return false, nil
	}

	// iterate over arr and check each window for equality with windowArr
	// if equality is detected then we need to save left and right sides to the state so reconstruct could renew it
loop:
	for i := 0; i <= lenArr-windowLen; i++ {
		forkedSt := st.Clone()

		// currently works for one match
		for iter, n := range arr[i : i+windowLen] {
			res, err := windowArr[iter].Check(forkedSt, n)
			if err != nil {
				return false, err
			}
			if !res {
				continue loop
			}
		}
		// if this code reached then we've got a match, so merge states and return true
		st.ApplyFrom(forkedSt)

		// set left and right parts to the state
		if err := st.SetVar("left", arr[:i:i]); err != nil {
			return false, err
		}
		if err := st.SetVar("right", arr[i+windowLen:]); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (m *MatroshkaArray) checkMultipleMatches(st *transformer.State, n nodes.Node) (bool, error) {
	arr, ok := n.(nodes.Array)
	if !ok {
		return false, nil
	}

	windowArr, err := m.Op.Arr(st)
	if err != nil {
		return false, err
	}
	// refactor window length
	windowLen := len(windowArr)
	if windowLen == 0 {
		return false, fmt.Errorf("this should not happen")
	}

	// if windows length is higher then array observed => did not match
	lenArr := len(arr)
	if windowLen > len(arr) {
		return false, nil
	}

	// split logic
	var (
		// statesResult is array of matched states
		statesResult []*transformer.State
		// leftNodes is array of arrays of nodes by the left side from the matched node
		leftNodes nodes.Array
	)

	var lastMatch int
loop:
	for i := 0; i <= lenArr-windowLen; i++ {
		forkedSt := st.Clone()

		// currently works for one match
		for iter, n := range arr[i : i+windowLen] {
			res, err := windowArr[iter].Check(forkedSt, n)
			if err != nil {
				return false, err
			}
			if !res {
				continue loop
			}
		}

		// each match is independent
		statesResult = append(statesResult, forkedSt)
		leftNodes = append(leftNodes, arr[lastMatch:i:i])
		lastMatch = i + windowLen
	}
	leftNodes = append(leftNodes, arr[lastMatch:])

	if len(statesResult) == 0 {
		return false, nil
	}
	if err := st.SetVar("side", leftNodes); err != nil {
		return false, err
	}
	if err := st.SetStateVar("matched", statesResult); err != nil {
		return false, err
	}

	return true, nil
}

func (m *MatroshkaArray) Construct(st *transformer.State, n nodes.Node) (nodes.Node, error) {
	return m.constructMultipleMatch(st, n)
}

func (m *MatroshkaArray) constructSingleMatch(st *transformer.State, n nodes.Node) (nodes.Node, error) {
	// signature of this func is weird: n nodes.Node should be actually empty because it's used for construction
	// for construction we need only states information

	if n != nil {
		return nil, fmt.Errorf("received node should be nil")
	}

	var result nodes.Array
	// we can construct left and right parts implicitly
	lArr, err := getArr(st, "left")
	if err != nil {
		return nil, err
	}
	result = append(result, lArr...)

	// to construct matched part we call Construct in embedded operation
	centerNode, err := m.Op.Construct(st, nil)
	if err != nil {
		return nil, err
	}
	result = append(result, centerNode.(nodes.Array)...)

	rArr, err := getArr(st, "right")
	if err != nil {
		return nil, err
	}

	return append(result, rArr...), nil
}

func (m *MatroshkaArray) constructMultipleMatch(st *transformer.State, n nodes.Node) (nodes.Node, error) {
	// signature of this func is weird: sideNode nodes.Node should be actually empty because it's used for construction
	// for construction we need only states information

	if n != nil {
		return nil, fmt.Errorf("received node should be nil")
	}

	var result nodes.Array
	// we can construct parts by merging them
	side, err := getArr(st, "side")
	if err != nil {
		return nil, err
	}

	states, ok := st.GetStateVar("matched")
	if !ok {
		return nil, fmt.Errorf("GetStateVar(\"matched\"): false")
	}

	for i, sideNode := range side {
		result = append(result, sideNode.(nodes.Array)...)
		if len(states) > i {
			n, err := m.Op.Construct(states[i], nil)
			if err != nil {
				return nil, err
			}
			result = append(result, n.(nodes.Array)...)
		}
	}

	return result, nil
}

func getArr(st *transformer.State, val string) (nodes.Array, error) {
	n, ok := st.GetVar(val)
	if !ok {
		return nil, fmt.Errorf("getVar %v: false", val)
	}
	return n.(nodes.Array), nil
}
