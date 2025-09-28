package templates

import "fmt"

var TemplateFuncs = map[string]interface{}{
	"slice": Slice,
	"add":   func(a, b int) int { return a + b },
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
		PriimaryDark:  primaryDark,
		PrimaryHover:  primaryHover,
		PrimaryActive: fmt.Sprintf("%s active:scale-95 active:opacity-80", primary),
		PrimaryDeep:   primaryDeep,
		TableRowHover: "hover:bg-gray-100 dark:hover:bg-gray-700",
		LinkInternal:  "text-gray-900 hover:underline dark:text-white",
	}
}
