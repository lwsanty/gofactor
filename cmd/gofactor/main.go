package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lwsanty/gofactor"
)

var (
	fSrc = flag.String("before", "", "path to a source sample")
	fDst = flag.String("after", "", "path to a destination sample")
)

func main() {
	flag.Parse()
	if err := run(*fSrc, *fDst, flag.Args()...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(src, dst string, files ...string) error {
	if src == "" {
		return errors.New("path to a source sample not specified (--before)")
	} else if dst == "" {
		return errors.New("path to a destination sample not specified (--after)")
	} else if len(files) == 0 {
		return errors.New("specify at least one file to transform")
	}
	dsrc, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	ddst, err := ioutil.ReadFile(dst)
	if err != nil {
		return err
	}
	ref, err := gofactor.NewRefactor(string(dsrc), string(ddst))
	if err != nil {
		return err
	}
	for _, path := range files {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		out, err := ref.Apply(string(data))
		if err != nil {
			return fmt.Errorf("failed to transform %q: %v", path, err)
		}
		err = ioutil.WriteFile(path, []byte(out), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
