package language

import "unicode"

// Script identifies different writing systems.
// It is represented as the binary encoding of a script tag of 4 letters,
// as specified by ISO 15924.
// Note that the default value is usually the Unknown script, not the 0 value (which is invalid)
type Script uint32

// LookupScript looks up the script for a particular character (as defined by
// Unicode Standard Annex #24), and returns Unknown if not found.
func LookupScript(r rune) Script {
	for name, table := range unicode.Scripts {
		if unicode.Is(table, r) {
			return scriptToTag[name]
		}
	}
	return Unknown
}

func (s Script) String() string {
	for k, v := range scriptToTag {
		if v == s {
			return k
		}
	}
	return "<script unknown>"
}

// IsRealScript return `true` if `s` if valid,
// and neither common or inherited.
func (s Script) IsRealScript() bool {
	switch s {
	case 0, Unknown, Common, Inherited:
		return false
	default:
		return true
	}
}

// IsSameScript compares two scripts: if one them
// is not 'real' (see IsRealScript), they are compared equal.
func (s1 Script) IsSameScript(s2 Script) bool {
	return s1 == s2 || !s1.IsRealScript() || !s2.IsRealScript()
}
