package provider

var refTypes = []string{"any", "pattern", "branch"}
var refTypesMap = map[string]string{
	"any":     "ANY_REF",
	"pattern": "PATTERN",
	"branch":  "BRANCH",
}
var refTypesReverseMap = map[string]string{
	"ANY_REF": "any",
	"PATTERN": "pattern",
	"BRANCH":  "branch",
}
