# go-term-markdown

[![Build Status](https://travis-ci.com/MichaelMure/go-term-markdown.svg?branch=master)](https://travis-ci.com/MichaelMure/go-term-markdown)
[![GoDoc](https://godoc.org/github.com/MichaelMure/go-term-markdown?status.svg)](https://godoc.org/github.com/MichaelMure/go-term-markdown)
[![Go Report Card](https://goreportcard.com/badge/github.com/MichaelMure/go-term-markdown)](https://goreportcard.com/report/github.com/MichaelMure/go-term-markdown)
[![codecov](https://codecov.io/gh/MichaelMure/go-term-markdown/branch/master/graph/badge.svg)](https://codecov.io/gh/MichaelMure/go-term-markdown)
[![GitHub license](https://img.shields.io/github/license/MichaelMure/go-term-markdown.svg)](https://github.com/MichaelMure/go-term-markdown/blob/master/LICENSE)
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/the-git-bug/Lobby)

`go-term-markdown` is a go package implementing a Markdown renderer for the terminal.

Note: Markdown being originally designed to render as HTML, rendering in a terminal is occasionally challenging and some adaptation had to be made. 

Features:
- formatting
- lists
- tables
- images
- code blocks with syntax highlighting
- basic HTML support

Note: this renderer is packaged as a standalone terminal viewer at https://github.com/MichaelMure/mdr/

## Usage

```go
import (
	"fmt"
	"io/ioutil"

	markdown "github.com/MichaelMure/go-term-markdown"
)

func main() {
	path := "Readme.md"
	source, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	result := markdown.Render(string(source), 80, 6)

	fmt.Println(result)
}
```

## Example

Here is the [Readme](https://github.com/MichaelMure/go-term-text/blob/v0.2.4/Readme.md) of `go-term-text` rendered with `go-term-markdown`:

![rendering example](misc/result.png)

Here is an example of table rendering:

![table rendering](misc/table.png)

## Origin

This package has been extracted from the [git-bug](https://github.com/MichaelMure/git-bug) project. As such, its aim is to support this project and not to provide an all-in-one solution. Contributions or full-on takeover as welcome though.

## Contribute

PRs accepted.

## License

MIT
