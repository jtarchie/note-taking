package main

import (
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/formatters/html"
)

var _ = formatters.Register(
	"html-custom",
	html.New(
		html.Standalone(false),
		html.WithClasses(false),
	),
)

func init() {
	chroma.StandardTypes[chroma.PreWrapper] = "highlight"
}