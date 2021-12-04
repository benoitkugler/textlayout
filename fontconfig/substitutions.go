package fontconfig

import (
	"fmt"
	"strings"
)

const (
	qualAny uint8 = iota
	qualAll
	qualFirst
	qualNotFirst
)

type ruleTest struct {
	expr   *expression
	kind   matchKind
	qual   uint8
	object Object
	op     opKind
}

// String returns a human friendly representation of a Test
func (test ruleTest) String() string {
	out := fmt.Sprintf("<<%s ", test.kind)
	switch test.qual {
	case qualAny:
		out += "any "
	case qualAll:
		out += "all "
	case qualFirst:
		out += "first "
	case qualNotFirst:
		out += "not_first "
	}
	out += fmt.Sprintf("%s %s %s>>", test.object, test.op, test.expr)
	return out
}

func (test ruleTest) copyT() ruleTest {
	test.expr = test.expr.copyT()
	return test
}

// Check the tests to see if they all match the pattern
// if keep is false, the rule is not matched
func (r ruleTest) match(selectedKind matchKind, p, pPat Pattern, data familyTable,
	valuePos []int, targets []*valueList, tst []ruleTest) (table *familyTable, keep bool) {
	m := p
	table = &data
	if selectedKind == MatchResult && r.kind == MatchQuery {
		m = pPat
		table = nil
	}
	object := r.object
	e := m[object]
	// different 'kind' won't be the target of edit
	if targets[object] == nil && selectedKind == r.kind {
		targets[object] = e
		tst[object] = r
	}
	// If there's no such field in the font, then qualAll matches for qualAny does not
	if e == nil {
		if r.qual == qualAll {
			valuePos[object] = -1
			return table, true
		}
		return table, false
	}
	// Check to see if there is a match, mark the location to apply match-relative edits
	vlIndex := matchValueList(m, pPat, selectedKind, r, *e, table)
	// different 'kind' won't be the target of edit
	if valuePos[object] == -1 && selectedKind == r.kind && vlIndex != -1 {
		valuePos[object] = vlIndex
	}
	if vlIndex == -1 || (r.qual == qualFirst && vlIndex != 0) ||
		(r.qual == qualNotFirst && vlIndex == 0) {
		return table, false
	}
	return table, true
}

type ruleEdit struct {
	expr    *expression
	binding valueBinding
	object  Object
	op      opKind
}

func (edit ruleEdit) String() string {
	return fmt.Sprintf("<<%s %s %s>>", edit.object, edit.op, edit.expr)
}

func (edit ruleEdit) copyT() ruleEdit {
	edit.expr = edit.expr.copyT()
	return edit
}

func (r ruleEdit) edit(selectedKind matchKind, p, pPat Pattern, table *familyTable,
	valuePos []int, targets []*valueList, tst []ruleTest) {
	object := r.object

	// Evaluate the list of expressions
	l := r.expr.toValues(p, pPat, selectedKind, r.binding)

	if tst[object].kind == MatchResult || selectedKind == MatchQuery {
		targets[object] = p[object]
	}

	targetList := targets[object]

	switch r.op.getOp() {
	case opAssign:
		// If there was a test, then replace the matched value with the newList list of values
		if valuePos[object] != -1 {
			thisValue := valuePos[object]

			// Append the newList list of values after the current value
			targetList.insert(thisValue, true, l, object, table)

			//  Delete the marked value
			if thisValue != -1 {
				targetList.del(thisValue, object, table)
			}

			// Adjust a pointer into the value list to ensure future edits occur at the same place
			break
		}
		fallthrough
	case opAssignReplace:
		// Delete all of the values and insert the newList set
		p.delWithTable(object, table)
		p.addWithTable(object, l, true, table)
		// Adjust a pointer into the value list as they no longer point to anything valid
		valuePos[object] = -1
	case opPrepend:
		if valuePos[object] != -1 {
			targetList.insert(valuePos[object], false, l, object, table)
			break
		}
		fallthrough
	case opPrependFirst:
		p.addWithTable(object, l, false, table)
	case opAppend:
		if valuePos[object] != -1 {
			targetList.insert(valuePos[object], true, l, object, table)
			break
		}
		fallthrough
	case opAppendLast:
		p.addWithTable(object, l, true, table)
	case opDelete:
		if valuePos[object] != -1 {
			targetList.del(valuePos[object], object, table)
			break
		}
		fallthrough
	case opDeleteAll:
		p.delWithTable(object, table)
	}
	// Now go through the pattern and eliminate any properties without data
	p.canon(object)
}

// patterns which match all of the tests are subjected to all the edits
type directive struct {
	tests []ruleTest
	edits []ruleEdit
}

func (d directive) String() string {
	lines := []string{"\ttests:"}
	for _, r := range d.tests {
		lines = append(lines, fmt.Sprintf("\t\t%s", r))
	}
	lines = append(lines, "\tedits:")
	for _, r := range d.edits {
		lines = append(lines, fmt.Sprintf("\t\t%s", r))
	}
	return strings.Join(lines, "\n")
}

func (d directive) copy() directive {
	out := directive{}
	if d.tests != nil {
		out.tests = make([]ruleTest, len(d.tests))
	}
	if d.edits != nil {
		out.edits = make([]ruleEdit, len(d.edits))
	}
	for i, v := range d.tests {
		out.tests[i] = v.copyT()
	}
	for i, v := range d.edits {
		out.edits[i] = v.copyT()
	}
	return out
}

// group of rules from the same origin (typically, file)
type ruleSet struct {
	name        string
	description string
	domain      string
	subst       [matchKindEnd][]directive
}

func (rs *ruleSet) String() string {
	lines := []string{"RuleSet from " + rs.name}
	for i, v := range rs.subst {
		if len(v) == 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("\tkind '%s'", matchKind(i)))
		for _, l := range v {
			lines = append(lines, l.String())
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

func (rs ruleSet) copy() ruleSet {
	for i, v := range rs.subst {
		if v == nil {
			continue
		}
		cp := make([]directive, len(v))
		for j, k := range v {
			cp[j] = k.copy()
		}
		rs.subst[i] = cp
	}
	return rs
}

// returns the offset of the maximum custom object used
// (or zero for no custom objects)
func (rs *ruleSet) add(rule directive, kind matchKind) int {
	rs.subst[kind] = append(rs.subst[kind], rule)

	var maxObject Object
	for i, r := range rule.tests {
		if r.kind == matchDefault { // resolve the default value
			rule.tests[i].kind = kind
		}
		if maxObject < r.object {
			maxObject = r.object
		}
	}
	for _, r := range rule.edits {
		if maxObject < r.object {
			maxObject = r.object
		}
	}

	if debugMode {
		fmt.Printf("Add rule for %s (from %s)\n", kind, rs.name)
		fmt.Println(rule)
	}

	if maxObject < FirstCustomObject {
		return 0
	}
	return int(maxObject - FirstCustomObject)
}

/* The bulk of the time in substitute is spent walking
 * lists of family names. We speed this up with a hash table.
 * Since we need to take the ignore-blanks option into account,
 * we use two separate hash tables. */
type familyTable struct {
	familyBlankHash familyBlankMap
	familyHash      familyMap
}

func newFamilyTable(p Pattern) familyTable {
	table := familyTable{
		familyBlankHash: make(familyBlankMap),
		familyHash:      make(familyMap),
	}

	e := p.getVals(FAMILY)
	table.add(e)
	return table
}

func (table familyTable) lookup(op opKind, s String) bool {
	flags := op.getFlags()
	var has bool

	if (flags & opFlagIgnoreBlanks) != 0 {
		_, has = table.familyBlankHash.lookup(s)
	} else {
		_, has = table.familyHash.lookup(s)
	}

	return has
}

func (table familyTable) add(values valueList) {
	for _, ll := range values {
		s := ll.Value.(String)

		count, _ := table.familyHash.lookup(s)
		count++
		table.familyHash.add(s, count)

		count, _ = table.familyBlankHash.lookup(s)
		count++
		table.familyBlankHash.add(s, count)
	}
}

func (table familyTable) del(s String) {
	count, ok := table.familyHash.lookup(s)
	if ok {
		count--
		if count == 0 {
			table.familyHash.del(s)
		} else {
			table.familyHash.add(s, count)
		}
	}

	count, ok = table.familyBlankHash.lookup(s)
	if ok {
		count--
		if count == 0 {
			table.familyBlankHash.del(s)
		} else {
			table.familyBlankHash.add(s, count)
		}
	}
}

// return the index into values, or -1
func matchValueList(p, pPat Pattern, kind matchKind,
	t ruleTest, values valueList, table *familyTable) int {

	var (
		value Value
		e     = t.expr
		ret   = -1
	)

	for e != nil {
		// Compute the value of the match expression
		if e.op.getOp() == opComma {
			tree := e.u.(exprTree)
			value = tree.left.evaluate(p, pPat, kind)
			e = tree.right
		} else {
			value = e.evaluate(p, pPat, kind)
			e = nil
		}

		if t.object == FAMILY && table != nil {
			op := t.op.getOp()
			if op == opEqual || op == opListing {
				if !table.lookup(t.op, value.(String)) {
					ret = -1
					continue
				}
			}
			if op == opNotEqual && t.qual == qualAll {
				ret = -1
				if !table.lookup(t.op, value.(String)) {
					ret = 0
				}
				continue
			}
		}

		for i, v := range values {
			// Compare the pattern value to the match expression value
			if compareValue(v.Value, t.op, value) {
				if ret == -1 {
					ret = i
				}
				if t.qual != qualAll {
					break
				}
			} else {
				if t.qual == qualAll {
					ret = -1
					break
				}
			}
		}
	}
	return ret
}

func substituteLang(p Pattern) {
	strs := getDefaultLangs()
	var lsund Langset
	lsund.add("und")

	for lang := range strs {
		for _, ll := range p.getVals(LANG) {
			vvL := ll.Value

			if vv, ok := vvL.(Langset); ok {
				var ls Langset
				ls.add(lang)

				b := vv.includes(ls)
				if b {
					return
				}
				if vv.includes(lsund) {
					return
				}
			} else {
				vv, _ := vvL.(String)
				if cmpIgnoreCase(string(vv), lang) == 0 {
					return
				}
				if cmpIgnoreCase(string(vv), "und") == 0 {
					return
				}
			}
		}
		p.addWithBinding(LANG, String(lang), vbWeak, true)
	}
}

// Substitute performs the sequence of pattern modification operations on `p`.
// If `kind` is MatchQuery, then those tagged as pattern operations are applied, else
// if `kind` is MatchResult, those tagged as font operations are applied and
// `testPattern` is used for <test> elements with target=pattern.
func (config *Config) Substitute(p, testPattern Pattern, kind matchKind) {
	if kind == MatchQuery {
		substituteLang(p)

		if _, res := p.GetAt(PRGNAME, 0); res == ResultNoMatch {
			if prgname := getProgramName(); prgname != "" {
				p.Add(PRGNAME, String(prgname), true)
			}
		}
	}

	nobjs := int(FirstCustomObject) - 1 + config.maxObjects + 2
	valuePos := make([]int, nobjs)
	targets := make([]*valueList, nobjs)
	tst := make([]ruleTest, nobjs)

	if debugMode {
		fmt.Println()
		fmt.Printf("Substitute with pattern: %s", p)
		fmt.Println()
	}

	data := newFamilyTable(p)
	table := &data
	for _, rs := range config.subst {
		rulesList := rs.subst[kind]

		if debugMode {
			fmt.Printf("\tapplying the %d rule(s) from %s:\n\n", len(rulesList), rs.name)
		}

	subsLoop:
		for _, rule := range rulesList {
			for i := range valuePos { // reset the edits locations
				targets[i] = nil
				valuePos[i] = -1
				tst[i] = ruleTest{}
			}
			for _, test := range rule.tests {
				if debugMode {
					fmt.Println("\t\ttest for substitute", test)
				}
				var keep bool
				table, keep = test.match(kind, p, testPattern, data, valuePos, targets, tst)
				if !keep {
					if debugMode {
						fmt.Println("\t\t-> dont pass")
					}
					continue subsLoop
				}
			}
			for _, edit := range rule.edits {
				if debugMode {
					fmt.Println("\t\tsubstitute edit", edit)
				}
				edit.edit(kind, p, testPattern, table, valuePos, targets, tst)

				if debugMode {
					fmt.Println("\t\tafter edit", p.String())
				}
			}
		}
	}
	if debugMode {
		fmt.Println("substitute done --> ", p.String())
	}

	return
}
