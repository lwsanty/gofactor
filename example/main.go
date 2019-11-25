package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lwsanty/gofactor"
	"github.com/opentracing/opentracing-go/log"
)

const (
	dir           = "../fixtures/multiple-match2/"
	beforeDefault = dir + "before"
	afterDefault  = dir + "after"
	testDefault   = dir + "example.go"
)

func main() {
	var (
		before string
		after  string
		test   string
	)
	flag.StringVar(&before, "before", beforeDefault, "a string var")
	flag.StringVar(&after, "after", afterDefault, "a string var")
	flag.StringVar(&test, "test", testDefault, "a string var")

	flag.Parse()

	handleErr := func(err error) {
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}
	readFile := func(filePath string) []byte {
		data, err := ioutil.ReadFile(filePath)
		handleErr(err)
		return data
	}

	beforeData := readFile(before)
	afterData := readFile(after)
	testData := readFile(test)

	refactor, err := gofactor.NewRefactor(string(beforeData), string(afterData))
	handleErr(err)

	code, err := refactor.Apply(string(testData))
	handleErr(err)

	fmt.Println(code)
}
