// Package templates provides utilities for rendering HTML templates with custom functions.
package templates

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	funcMap := template.FuncMap(TemplateFuncs)
	t, err := template.New("base.tmpl").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "base.tmpl"),
		filepath.Join("templates", tmpl),
	)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
	}

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}
