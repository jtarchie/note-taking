package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	chromaHTML "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/quick"
	"github.com/alexflint/go-arg"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/jtarchie/notes/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/afero"
)

//go:generate qtc -dir=templates

var args struct {
	Directory string `arg:"required"`
	Port      uint   `default:"8080"`
}

func main() {
	chroma.StandardTypes[chroma.PreWrapper] = "highlight"

	err := execute()
	if err != nil {
		log.Fatalf("could not execute: %s", err)
	}
}

func trimPath(path string, trim string) string {
	s := strings.TrimPrefix(path, trim)
	s = strings.TrimSuffix(s, filepath.Ext(s))
	return "/" + s
}

var _ = formatters.Register(
	"html",
	chromaHTML.New(
		chromaHTML.Standalone(false),
		chromaHTML.WithClasses(false),
	),
)

func renderHookDropCodeBlock(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if n, ok := node.(*ast.CodeBlock); ok {
		source := string(n.Literal)
		log.Printf("source=%q\n", source)
		log.Printf("source=%#v\n", n)

		_ = quick.Highlight(
			w,
			source,
			string(n.Info),
			"html",
			"solarized-dark256",
		)
		return ast.GoToNext, true
	}

	return ast.GoToNext, false
}

func execute() error {
	err := arg.Parse(&args)
	if err != nil {
		return fmt.Errorf("could not parse arguments: %w", err)
	}

	originalDirectory, err := filepath.Abs(args.Directory)
	if err != nil {
		return fmt.Errorf("could not evaluate directory %q: %w", args.Directory, err)
	}
	log.Printf("using base directory %q\n", originalDirectory)

	markdownFiles, err := doublestar.Glob(os.DirFS(originalDirectory), "**/*.md")
	if err != nil {
		return fmt.Errorf("could not find markdown files in %q: %w", originalDirectory, err)
	}
	log.Printf("found %d files with glob '**/*.md", len(markdownFiles))

	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		RenderNodeHook: renderHookDropCodeBlock,
	}
	renderer := html.NewRenderer(opts)
	memFS := afero.NewMemMapFs()

	for _, markdownFile := range markdownFiles {
		contents, err := os.ReadFile(markdownFile)
		if err != nil {
			return fmt.Errorf("could not read markdown file %q: %w", markdownFile, err)
		}

		strippedPath := trimPath(markdownFile, originalDirectory)
		log.Printf("creating HTML of file %q\n", strippedPath)

		html := markdown.ToHTML(contents, nil, renderer)
		err = afero.WriteFile(
			memFS,
			strippedPath,
			[]byte(templates.Render(html)),
			os.ModePerm,
		)
		if err != nil {
			return fmt.Errorf("could not setup HTML for %q: %w", strippedPath, err)
		}
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/*", echo.WrapHandler(http.FileServer(afero.NewHttpFs(memFS))))

	err = e.Start(fmt.Sprintf(":%d", args.Port))
	if err != nil {
		return fmt.Errorf("http server stopped: %w", err)
	}

	return nil
}
