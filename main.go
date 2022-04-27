package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"
	"github.com/jtarchie/notes/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:generate qtc -dir=templates

var args struct {
	Directory string `arg:"required"`
	Port      uint   `default:"8080"`
}

func main() {
	err := execute()
	if err != nil {
		log.Fatalf("could not execute: %s", err)
	}
}

func execute() error {
	err := arg.Parse(&args)
	if err != nil {
		return fmt.Errorf("could not parse arguments: %w", err)
	}

	renderer, err := NewRenderer(args.Directory, templates.Render)
	if err != nil {
		return fmt.Errorf("could not create renderer: %w", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/*", echo.WrapHandler(renderer))

	err = e.Start(fmt.Sprintf(":%d", args.Port))
	if err != nil {
		return fmt.Errorf("http server stopped: %w", err)
	}

	return nil
}
