# goignore

A simple gitignore parser for `Go`

## Install

```shell
go get github.com/botondmester/goignore
```

## Usage

This is a simple example showing how to use the library:
```go
import (
    "os"
    "strings"

    "github.com/botondmester/goignore"
)

func main() {
    ignore := goignore.CompileIgnoreLines([]string{
		"/*",
		"!/foo",
		"/foo/*",
		"!/foo/bar",
	})

    // should print `foo/baz is ignored`
    if ignore.Match("foo/baz") {
        println("foo/baz is ignored")
    } else {
        println("foo/baz is not ignored")
    }
}
```

For more examples, refer to the [goignore_test.go](goignore_test.go) file.

## Tests

This package's tests were copied from the [go-gitignore](https://github.com/sabhiram/go-gitignore) package, and were modified where needed.