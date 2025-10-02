package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"premodernonsdagar/internal/aggregation"
	"slices"
)

var TemplateFuncs = map[string]interface{}{
	"slice":    Slice,
	"add":      func(a, b int) int { return a + b },
	"contains": func(slice []string, item string) bool { return slices.Contains(slice, item) },
	"cardtype": func(t string) string {
		switch t {
		case "creature":
			return "Creatures"
		case "other":
			return "Other Spells"
		case "land":
			return "Lands"
		default:
			return "Other"
		}
	},
}

func Slice(args ...interface{}) []interface{} {
	return args
}

func ColorScheme() TailwindClasses {
	primary := "blue-600"
	primaryDark := "blue-400"
	primaryHover := "blue-700"
	primaryDeep := "blue-900"

	return TailwindClasses{
		Link:          fmt.Sprintf("text-%s hover:underline dark:text-%s", primary, primaryDark),
		ButtonPrimary: fmt.Sprintf("block rounded bg-%s px-4 py-3 text-center font-medium text-white transition-colors hover:bg-%s", primary, primaryHover),
		ButtonBack:    fmt.Sprintf("rounded border border-%s bg-transparent px-4 py-2 font-semibold text-%s hover:border-transparent hover:bg-%s hover:text-white dark:border-%s dark:text-%s", primary, primary, primary, primaryDark, primaryDark),
		Navbar:        fmt.Sprintf("mb-3 bg-%s text-white", primary),
		SymbolPrimary: fmt.Sprintf("material-symbols-outlined text-%s dark:text-%s", primary, primaryDark),
		Primary:       primary,
		PrimaryDark:   primaryDark,
		PrimaryHover:  primaryHover,
		PrimaryActive: fmt.Sprintf("%s active:scale-95 active:opacity-80", primary),
		PrimaryDeep:   primaryDeep,
		TableRowHover: "hover:bg-gray-100 dark:hover:bg-gray-700",
		LinkInternal:  "text-gray-900 hover:underline dark:text-white",
		TableHeader:   "px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-300",
	}
}

func RenderAllTemplates() error {
	// Define mock data to prevent templates from crashing
	mockData := map[string]interface{}{
		"ActivePage":          "home",
		"NextEventDate":       "today",
		"NextEventWeekNumber": 42,
		"Scheme":              ColorScheme(),
		"Player":              aggregation.Player{},
	}

	htmlOutputDir := "pages/html"
	if err := os.MkdirAll(htmlOutputDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return err
	}

	templateDir := "templates"
	templateFiles := []string{}
	fs.WalkDir(os.DirFS(templateDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".tmpl" {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})

	for _, tmpl := range templateFiles {
		outputFile := fmt.Sprintf("%s/%s.html", htmlOutputDir, tmpl[:len(tmpl)-5])

		if err := renderTemplateToFile(tmpl, mockData, outputFile); err != nil {
			fmt.Printf("Failed to render template %s: %v\n", tmpl, err)
		}
	}
	return nil
}

func renderTemplateToFile(tmpl string, data interface{}, outputPath string) error {
	funcMap := template.FuncMap(TemplateFuncs)
	t, err := template.New("base.tmpl").Funcs(funcMap).ParseFiles(
		filepath.Join("templates", "base.tmpl"),
		filepath.Join("templates", tmpl),
	)
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	if err := t.ExecuteTemplate(file, "base", data); err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}

	return nil
}
