package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	chromaHTML "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/quick"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdHTML "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/spf13/afero"
)

type renderer struct {
	directory string
	fs        afero.Fs
	layout    func([]byte) string
	server    http.Handler
}

var _ = formatters.Register(
	"html-custom",
	chromaHTML.New(
		chromaHTML.Standalone(false),
		chromaHTML.WithClasses(false),
	),
)

func NewRenderer(
	directory string,
	layout func([]byte) string,
) (*renderer, error) {
	absDirectory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate directory %q: %w", directory, err)
	}

	chroma.StandardTypes[chroma.PreWrapper] = "highlight"

	fs := afero.NewMemMapFs()

	r := &renderer{
		directory: absDirectory,
		fs:        fs,
		layout:    layout,
		server:    http.FileServer(afero.NewHttpFs(fs)),
	}

	err = r.process()
	if err != nil {
		return nil, fmt.Errorf("could not process markdown files: %w", err)
	}

	return r, nil
}

func (r *renderer) process() error {
	opts := mdHTML.RendererOptions{
		Flags:          mdHTML.CommonFlags | mdHTML.FootnoteReturnLinks,
		RenderNodeHook: r.renderMarkdownHook,
	}
	mdRenderer := mdHTML.NewRenderer(opts)

	markdownFiles, err := doublestar.Glob(os.DirFS(r.directory), "**/*.md")
	if err != nil {
		return fmt.Errorf("could not find markdown files in %q: %w", r.directory, err)
	}

	for _, markdownFile := range markdownFiles {
		contents, err := os.ReadFile(markdownFile)
		if err != nil {
			return fmt.Errorf("could not read markdown file %q: %w", markdownFile, err)
		}

		strippedPath := r.trimPath(markdownFile, r.directory)

		html := markdown.ToHTML(
			contents,
			parser.NewWithExtensions(math.MaxInt32),
			mdRenderer,
		)
		err = afero.WriteFile(
			r.fs,
			strippedPath,
			[]byte(r.layout(html)),
			os.ModePerm,
		)
		if err != nil {
			return fmt.Errorf("could not setup HTML for %q: %w", strippedPath, err)
		}
	}

	return nil
}

func (n *renderer) renderMarkdownHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
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

func (r *renderer) trimPath(path string, trim string) string {
	s := strings.TrimPrefix(path, trim)
	s = strings.TrimSuffix(s, filepath.Ext(s))
	return "/" + s
}

func (r *renderer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	r.server.ServeHTTP(response, request)
}

var _ http.Handler = &renderer{}
