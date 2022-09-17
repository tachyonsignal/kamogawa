package view

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"

	"kamogawa/config"

	"github.com/gin-contrib/multitemplate"
)

var (
	baseLayoutPath      = "theme/" + config.Theme + "/layout/*.tmpl"
	unauthedLayoutsPath = "theme/" + config.Theme + "/layout/unauthed/*.tmpl"
	authedLayoutsPath   = "theme/" + config.Theme + "/layout/authed/*.tmpl"
	unauthedViewsGlob   = "theme/" + config.Theme + "/unauthed/*.tmpl"
	authedViewsGlob     = "theme/" + config.Theme + "/authed/*.tmpl"
)

//go:embed theme/*
var views embed.FS

// https://www.josephspurrier.com/how-to-embed-assets-in-go-1-16
/** Middleware to render HTML using templates. */
func HTMLRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	unauthedViews, err := fs.Glob(views, unauthedViewsGlob)
	if err != nil {
		panic(err.Error())
	}
	for _, unauthedView := range unauthedViews {
		t, err := template.ParseFS(views, baseLayoutPath, unauthedLayoutsPath, unauthedView)
		if err != nil {
			log.Fatal(err)
		}
		r.Add(filepath.Base(unauthedView), t)
	}

	authedViews, err := fs.Glob(views, authedViewsGlob)
	if err != nil {
		panic(err.Error())
	}
	for _, authedView := range authedViews {
		t, err := template.ParseFS(views, baseLayoutPath, authedLayoutsPath, authedView)
		if err != nil {
			log.Fatal(err)
		}
		r.Add(filepath.Base(authedView), t)
	}

	return r
}