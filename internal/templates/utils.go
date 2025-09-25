package templates

var TemplateFuncs = map[string]interface{}{
	"slice": Slice,
}

func Slice(args ...interface{}) []interface{} {
	return args
}
