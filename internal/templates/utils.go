package templates

var TemplateFuncs = map[string]interface{}{
	"slice": Slice,
	"add":   func(a, b int) int { return a + b },
}

func Slice(args ...interface{}) []interface{} {
	return args
}
