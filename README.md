# faillint [![](https://github.com/fatih/faillint/workflows/build/badge.svg)](https://github.com/fatih/faillint/actions)

Faillint is a simple Go linter that fails when a specific set of import paths
are used. It's meant to be used in CI/CD environments to catch rules you want
to enforce in your projects. 

As an example, you could enforce the usage of `github.com/pkg/errors` instead
of the `errors` package. To prevent the usage of the `errors` package, you can
configure `faillint` to fail whenever someone imports the `errors` package in
this case.

![faillint](https://user-images.githubusercontent.com/438920/74105764-8c644780-4b15-11ea-885e-3ad45b965da4.gif)


## Install

```bash
go get github.com/fatih/faillint
```

## Usage

`faillint` works on a a file, directory or a Go package:

```sh
$ faillint -paths "errors" foo.go # pass a file
$ faillint -paths "errors" ./...  # recursively analyze all files
$ faillint -paths "errors" github.com/fatih/gomodifytags # or pass a package
```

By default, `faillint` will not check any import paths. You need to explicitly
define it with the `-paths` flag, which is comma-separated list. Some examples are:

```
# fail if the errors package is used
-paths "errors"

# fail if the old context package is imported
-paths "golang.org/x/net/context"

# fail both on stdlib log and errors package to enforce other internal libraries
-paths "log,errors"
```


If you have a preferred import path to suggest, append the suggestion after a `=` charachter:

```
# fail if the errors package is used and suggest to use github.com/pkg/errors
-paths "errors=github.com/pkg/errors"

# fail for the old context import path and suggest to use the stdlib context
-paths "golang.org/x/net/context=context"

# fail both on stdlib log and errors package to enforce other libraries
-paths "log=go.uber.org/zap,errors=github.com/pkg/errors"
```

## Example

Assume we have the following file:

```go
package a

import (
        "errors"
)

func foo() error {
        return errors.New("bar!")
}
```

Let's run `faillint` to check if `errors` import is used and report it:

```
$ faillint -paths "errors=github.com/pkg/errors" a.go
a.go:4:2: package "errors" shouldn't be imported, suggested: "github.com/pkg/errors"
```

## The need for this tool?

Most of these checks should be probably detected during the review cycle. But
it's totally normal to accidently import them (we're all humans in the end). 

Second, tools like `goimports` favors certain packages. As an example going
forward if you decided to use `github.com/pkg/errors` in you project, and write
`errors.New()` in a new file, `goimports` will automatically import the
`errors` package (and not `github.com/pkg/errors`). The code will perfectly
compile. `faillint` would be able to detect and report it to you.

## Credits

This tool is built on top of the excellent `go/analysis` package that makes it
easy to write custom analyzers in Go. If you're interested in writing a tool,
check out my **[Using go/analysis to write a custom
linter](https://arslan.io/2019/06/13/using-go-analysis-to-write-a-custom-linter/)**
blog post.

Part of the code is modified and based on [astutil.UsesImport](https://pkg.go.dev/golang.org/x/tools/go/ast/astutil?tab=doc#UsesImport)
