package python

import (
	"fmt"
	"unicode"
)

const (
	PARSE_SEEK_KEY int = iota
	PARSE_KEY
	PARSE_SEEK_EQUALS
	PARSE_SEEK_VALUE
	PARSE_VALUE
)

func translateEscaped(r rune) rune {
	switch r {
	case 'r':
		return '\r'
	case 'n':
		return '\n'
	case 't':
		return '\t'
	default:
		return r
	}
}

// ParseEnvIntoSetStrings breaks and env string in K=V pairs into an
// array of K=V strings, for Environment inclusion in SystemD
// @TODO Handle escaping of complex values in Environment values
func ParseEnvIntoSetStrings(env string) ([]string, error) {
	a, err := ParseEnv(env)
	if nil!=err {
		return nil, err
	}
	out := make([]string, len(a))
	for i, p := range a {
		out[i] = fmt.Sprintf("%s=%s", p[0], p[1])
	}
	return out, nil
}

// ParseEnv parses an string of the form (Kn=Vn[,? ])* and returns an array of Kn=Vn, where
// Vn can be "' delimited
func ParseEnv(env string) ([][2]string, error) {
	pairs := [][2]string{}
	state := PARSE_SEEK_KEY
	var key []rune
	var value []rune
	var delimiter rune
	var instring bool
	var escaped bool
	for i, r := range env {
		switch state {
		case PARSE_SEEK_KEY:
			if (unicode.IsSpace(r) || ','==r || ';'==r) {
				continue
			}
			key = []rune{}
			if ('"'==r) || ('\''==r) {
				delimiter = r
				instring = true
				key = []rune{}
			} else {
				key = []rune{r}
			}
			state = PARSE_KEY
		case PARSE_KEY:
			switch {
			case escaped:
				key = append(key, translateEscaped(r))
				escaped = false
			case '\\'==r:
				escaped = true 
			case instring:
				if r==delimiter {
					instring = false
					state = PARSE_SEEK_EQUALS
					continue
				}
				key = append(key, r)
			case unicode.IsSpace(r):
				state = PARSE_SEEK_EQUALS
			case '='==r:
				state = PARSE_SEEK_VALUE
			case ';'==r  || ','==r:
				pairs = append(pairs, [2]string{ string(key), `` })
				state = PARSE_SEEK_KEY
			default:
				key = append(key, r)
			}
		case PARSE_SEEK_EQUALS:
			switch {
			case unicode.IsSpace(r):
				continue
			case '='==r:
				state = PARSE_SEEK_VALUE
			case ';'==r || ','==r:
				// EMPTY VALUE
				pairs = append(pairs, [2]string{ string(key), `` })
				state = PARSE_SEEK_KEY
			default:
				return nil, fmt.Errorf(`Expected = at col %d, but got '%c' in '%s'`, i+1, r, env)
			}
		case PARSE_SEEK_VALUE:
			if unicode.IsSpace(r) {
				continue
			}
			if ';'==r || ','==r {
				// EMPTY VALUE
				pairs = append(pairs, [2]string{ string(key),`` })				
				state = PARSE_SEEK_KEY
				continue
			}
			if ('"'==r) || ('\''==r) {
				delimiter = r
				instring = true
				value = []rune{}
			} else {
				value = []rune{r}
			}
			state = PARSE_VALUE
		case PARSE_VALUE:
			switch {
			case escaped:
				value = append(value, translateEscaped(r))
				escaped = false
			case '\\'==r:
				escaped = true
			case instring:
				if r==delimiter {
					instring = false
					pairs = append(pairs, [2]string{ string(key), string(value) })
					state = PARSE_SEEK_KEY
					continue
				}
				value = append(value, r)
			case unicode.IsSpace(r) || ';'==r || ','==r:
				pairs = append(pairs, [2]string { string(key), string(value) })
				state = PARSE_SEEK_KEY				
			default:
				value = append(value, r)
			}
		}
	}
	if instring {
		return nil, fmt.Errorf(`Unterminated string delimited with %c in %s`, delimiter, env)
	}
	switch state {
	case PARSE_SEEK_KEY:
		// Nothing to do here - we're good to finish in this state
	case PARSE_KEY:
		fallthrough
	case PARSE_SEEK_EQUALS:
		fallthrough
	case PARSE_SEEK_VALUE:
		pairs = append(pairs, [2]string  { string(key), ``})
	case PARSE_VALUE:
		pairs = append(pairs, [2]string{string(key), string(value)})
	}
	return pairs, nil
}