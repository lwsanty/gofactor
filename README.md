# gofactor
Advanced utility for golang refactor based on DSL transformations provided by [bblfsh/sdk](https://github.com/bblfsh/sdk)

## Requirements
- Go 1.13+

## Build

Build the CLI:

1) clone `gofactor` repo

```bash
git clone https://github.com/lwsanty/gofactor
```

2) build CLI

```bash
go install ./cmd/gofactor
```

## Usage example

Imagine you have a piece of code

```go
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
```

And you want to replace all code patterns like

```go
if i%2 == 0 {
    i = 5
}

if X%2 == 0 {
    X = 5
}

```

to

```go
if i%2 == 1 {
    i = 1
} else {
    X = 1
}
```

Here's where `gofactor` tool comes for help.

Save the "before" snippet to `before.txt`, "after" snippet to an `after.txt` file, and run:

```bash
gofactor --before before.txt --after after.txt some_file.go
```

**Result**
```go
package main

import "fmt"

func main() {
	var (
		i int
		X int
		j int
	)
	if i%2 == 1 {
		i = 1
	} else {
		X = 1
	}
	fmt.Println(i)
	if i%2 == 1 {
		i = 1
	} else {
		j = 1
	}
}
func a(i, X int) {
	if i%2 == 1 {
		i = 1
	} else {
		X = 1
	}
	fmt.Println(i)
	if i%2 == 1 {
		i = 1
	} else {
		X = 1
	}
}
```

## Usage as a library

It is also possible to use the tools as a library.

1) Init refactor object

```go
refactor, err := gofactor.NewRefactor(beforeSnippet, afterSnippet)
if err != nil {
    log.Error(err)
    os.Exit(1)
}
``` 

2) Apply generated transformations to the desired code

```go
code, err := refactor.Apply(desiredCode)
if err != nil {
    log.Error(err)
    os.Exit(1)
}
```

## Supported cases
See `fixtures`

## Under the hood
1) both input and output patterns are converted to go `AST` nodes
2) both input and output nodes converted to `bblfsh` `uast.Node`s
3) define mapping of transformation operations from input to output node
4) apply transformation mapping to the desired code: traverse over the `uast.Node`s tree and transform matching nodes
5) convert transformed tree back to golang `AST`
6) convert golang `AST` to string

## Roadmap
- currently library copy-pastes a part fo the `bblfsh/go-driver` because of the dependency [issue](https://github.com/bblfsh/go-driver/issues/67), fix this part 
- support functions refactor
- handle cases with cascade `if`s, `switch`es and tail recursions
- during the transformations we are forced to drop nodes positions, need to investigate the possibilities of preserving/reconstructing them(probably using DST nodes could help, related issue https://github.com/dave/dst/issues/38) 
