package parser

import (
	"strings"
)

func ParseFields(s string) ([]string, error) {
	esc := '\\'
	fields := make([]string, 0)
	current := strings.Builder{}
	inQuotes := false
	quoteChar := rune(0)

	for i, r := range s {
		switch {
		case r == esc && i+1 < len(s): // Escape character
			i++
			current.WriteRune(rune(s[i]))

		case inQuotes && r == quoteChar: // Closing quote
			inQuotes = false

		case !inQuotes && (r == '"' || r == '\''): // Opening quote
			inQuotes = true
			quoteChar = r

		case !inQuotes && r == ',': // Separator
			if current.Len() > 0 {
				field, err := UnquoteString(current.String(), esc)
				if err != nil {
					field = current.String()
				}
				field = RemoveEscapes(field, esc)
				fields = append(fields, field)
				current.Reset()
			}

		default: // Regular character
			current.WriteRune(r)
		}
	}

	// Add the last field if non-empty
	if current.Len() > 0 {
		field, _ := UnquoteString(current.String(), esc)
		field = RemoveEscapes(field, esc)
		fields = append(fields, field)
	}

	return fields, nil
}
