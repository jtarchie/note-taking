package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"

	"github.com/alecthomas/chroma/quick"
	"github.com/goccy/go-yaml"
)

type Metadata struct {
	Title      string
	Confluence struct {
		Section string
	}
}

type Doc struct {
	contents []byte
	filename string
	metadata *Metadata
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
	if d.markdown != nil {
		return d.markdown
	}

	index := bytes.Index(d.contents, []byte("---\n"))
	if index >= 0 {
		metadata := &Metadata{}

		possibleYAML := d.contents[:index]
		err := yaml.Unmarshal(possibleYAML, metadata)
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

func (d *Doc) Title() string {
	if d.metadata != nil {
		if title := d.metadata.Title; title != "" {
			return title
		}
	}

	findH1 := regexp.MustCompile(`^#[^#]\s*(.*)\s*\n`)
	groups := findH1.FindSubmatch(d.Markdown())
	if groups == nil {
		return ""
	} else {
		return string(groups[1])
	}
}
