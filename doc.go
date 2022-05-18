package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"

	"github.com/alecthomas/chroma/quick"
	"github.com/goccy/go-yaml"
)

type Doc struct {
	contents []byte
	filename string
	metadata map[string]string
	markdown []byte
}

func renderMarkdownHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch n := node.(type) {
	case *ast.CodeBlock:
		source := string(n.Literal)
		_ = quick.Highlight(
			w,
			source,
			string(n.Info),
			"html-custom",
			"solarized-dark256",
		)
		return ast.GoToNext, true
	default:
		return ast.GoToNext, false
	}
}

func NewDoc(filename string) (*Doc, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read markdown file %q: %w", filename, err)
	}

	return &Doc{
		contents: contents,
		filename: filename,
	}, nil
}

func (d *Doc) Markdown() []byte {
	index := bytes.Index(d.contents, []byte("---\n"))
	if index >= 0 {
		var metadata map[string]string

		possibleYAML := d.contents[:index]
		err := yaml.Unmarshal(possibleYAML, &metadata)
		if err == nil {
			d.metadata = metadata
			d.markdown = d.contents[index+4:]
			return d.markdown
		}
	}

	d.markdown = d.contents
	return d.markdown
}

func (d *Doc) ToHTML(layout func([]byte) string) string {
	opts := html.RendererOptions{
		Flags:          html.CommonFlags | html.FootnoteReturnLinks,
		RenderNodeHook: renderMarkdownHook,
	}
	mdRenderer := html.NewRenderer(opts)

	html := markdown.ToHTML(
		d.Markdown(),
		parser.NewWithExtensions(math.MaxInt32),
		mdRenderer,
	)

	return layout(html)
}
