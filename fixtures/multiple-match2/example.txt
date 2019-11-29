package main

import "fmt"

func main() {
	var (
		i int
		X int
		j int
	)

	if i%2 == 0 {
		i = 5
	}

	if X%2 == 0 {
		X = 5
	}

	fmt.Println(i)

	if i%2 == 0 {
		i = 5
	}

	if j%2 == 0 {
		j = 5
	}
}

func a(i, X int) {
	if i%2 == 0 {
		i = 5
	}

	if X%2 == 0 {
		X = 5
	}

	fmt.Println(i)

	if i%2 == 0 {
		i = 5
	}

	if X%2 == 0 {
		X = 5
	}
}
