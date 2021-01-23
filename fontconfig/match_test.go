package fontconfig

import (
	"fmt"
	"testing"
)

// ported from fontconfig/test/test-family-matching.c Copyright © 2020 Zoltan Vandrus

const fcTestResult = "testresult"

func matchPattern(test string, p Pattern) (bool, error) {
	xml := fmt.Sprintf(`
		 <fontconfig>
		   <match>
		   	%s
			<edit name="%s">
				<bool>true</bool>
			</edit>
		   </match>
		 </fontconfig>
		`, test, fcTestResult)

	pat := p.Duplicate()

	cfg := NewConfig()

	err := cfg.ParseAndLoadFromMemory([]byte(xml))
	if err != nil {
		return false, err
	}

	cfg.SubstituteWithPat(pat, nil, MatchQuery)

	// the parsing side effect registred TfcestResult
	o := getRegisterObjectType(fcTestResult).object
	if o < firstCustomObject {
		return false, fmt.Errorf("got invalid custom object %d", o)
	}
	_, result := pat.GetBool(o)
	return result, nil
}

func shouldMatchPattern(t *testing.T, test string, pat Pattern, negate bool) {
	res, err := matchPattern(test, pat)
	if err != nil {
		t.Errorf("unexpected error in test %s: %s", test, err)
	}
	if res && negate {
		t.Errorf("%s unexpectedly matched:\non\n%s", test, pat)
	} else if !res && !negate {
		t.Errorf("%s should have matched:\non\n%s", test, pat)
	}
}

func TestFamily(t *testing.T) {
	pat := BuildPattern(
		PatternElement{Object: FC_FAMILY, Value: String("family1")},
		PatternElement{Object: FC_FAMILY, Value: String("family2")},
		PatternElement{Object: FC_FAMILY, Value: String("family3")},
	)
	var test string

	test = `<test qual="all" name="family" compare="not_eq">
	    <string>foo</string>
	</test>
	`
	shouldMatchPattern(t, test, pat, false)

	test = `
	<test qual="all" name="family" compare="not_eq">
	    <string>family2</string>
	</test>
	`
	shouldMatchPattern(t, test, pat, true)

	test = `
	<test qual="any" name="family" compare="eq">
	    <string>family3</string>
	</test>
	`
	shouldMatchPattern(t, test, pat, false)

	test = `
	<test qual="any" name="family" compare="eq">
	    <string>foo</string>
	</test>
	`
	shouldMatchPattern(t, test, pat, true)
}
