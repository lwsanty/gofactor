package main

import "fmt"

func main() {
	var (
		i int
		X int
	)

	if i%2 == 0 {
		i = 5
	}

	if X%2 == 0 {
		X = 5
	}

	fmt.Println(i)
}
