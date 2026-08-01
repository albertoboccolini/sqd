package variables

import (
	"os"
	"regexp"
)

type Expander struct {
	pattern *regexp.Regexp
}

func NewExpander() *Expander {
	return &Expander{
		pattern: regexp.MustCompile(`\$(?:\{([A-Za-z_][A-Za-z0-9_]*)\}|([A-Za-z_][A-Za-z0-9_]*))`),
	}
}

func (expander *Expander) Expand(query string, variables map[string]string) string {
	return expander.pattern.ReplaceAllStringFunc(query, func(match string) string {
		submatches := expander.pattern.FindStringSubmatch(match)
		name := submatches[1]
		if name == "" {
			name = submatches[2]
		}

		if value, ok := variables[name]; ok {
			return value
		}

		if value, ok := os.LookupEnv(name); ok {
			return value
		}

		return match
	})
}
