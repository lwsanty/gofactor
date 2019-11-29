package gofactor_test

import (
	"go/format"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/lwsanty/gofactor"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	dirs, err := filepath.Glob("fixtures/*")
	require.NoError(t, err)

	for _, d := range dirs {
		d := d
		t.Run(d, func(t *testing.T) {
			testCase(t, d)
		})
	}
}

func testCase(t *testing.T, d string) {
	getFileContent := func(name string) string {
		data, err := ioutil.ReadFile(filepath.Join(d, name))
		require.NoError(t, err)
		return string(data)
	}
	getFormattedContent := func(name string) string {
		data, err := ioutil.ReadFile(filepath.Join(d, name))
		require.NoError(t, err)

		fdata, err := format.Source(data)
		require.NoError(t, err)
		return string(fdata)
	}

	var (
		after    = getFileContent("after.txt")
		before   = getFileContent("before.txt")
		example  = getFormattedContent("example.txt")
		expected = getFormattedContent("expected.txt")
	)

	refactor, err := gofactor.NewRefactor(before, after)
	require.NoError(t, err)

	actual, err := refactor.Apply(example)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
