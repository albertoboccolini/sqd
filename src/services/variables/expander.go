package variables

import (
	"os"
	"strings"
)

type Expander struct{}

func NewExpander() *Expander {
	return &Expander{}
}

func (expander *Expander) Expand(query string, variables map[string]string) string {
	var result strings.Builder
	result.Grow(len(query))

	for index := 0; index < len(query); {
		currentByte := query[index]
		if currentByte != '$' {
			result.WriteByte(currentByte)
			index++
			continue
		}

		nextIndex := index + 1
		if nextIndex >= len(query) {
			result.WriteByte('$')
			index++
			continue
		}

		if query[nextIndex] == '{' {
			name, consumed := readBracedIdentifier(query, nextIndex+1)
			if consumed > 0 {
				writeReplacement(&result, name, true, variables)
				index = nextIndex + consumed + 2
				continue
			}
		}

		name, consumed := readIdentifier(query, nextIndex)
		if consumed > 0 {
			writeReplacement(&result, name, false, variables)
			index = nextIndex + consumed
			continue
		}

		result.WriteByte('$')
		index++
	}

	return result.String()
}

func readBracedIdentifier(query string, start int) (string, int) {
	name, consumed := readIdentifier(query, start)
	if consumed == 0 {
		return "", 0
	}

	endAfterName := start + consumed
	if endAfterName >= len(query) || query[endAfterName] != '}' {
		return "", 0
	}
	return name, consumed
}

func readIdentifier(query string, start int) (string, int) {
	if start >= len(query) || !isIdentifierStart(query[start]) {
		return "", 0
	}

	end := start + 1
	for end < len(query) && isIdentifierContinue(query[end]) {
		end++
	}
	return query[start:end], end - start
}

func isIdentifierStart(char byte) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || char == '_'
}

func isIdentifierContinue(char byte) bool {
	return isIdentifierStart(char) || (char >= '0' && char <= '9')
}

func writeReplacement(result *strings.Builder, name string, braced bool, variables map[string]string) {
	if value, ok := variables[name]; ok {
		result.WriteString(value)
		return
	}

	if value, ok := os.LookupEnv(name); ok {
		result.WriteString(value)
		return
	}

	if braced {
		result.WriteString("${")
		result.WriteString(name)
		result.WriteString("}")
		return
	}

	result.WriteString("$")
	result.WriteString(name)
}
