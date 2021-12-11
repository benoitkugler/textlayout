package fontconfig

import (
	"fmt"
	"strings"
)

// this file implements the dump of a Config object to
// a corresponding Go source code.

func (e exprMatrix) asGoSource() string {
	return fmt.Sprintf("exprMatrix{xx: &%s, xy:  &%s, yx: &%s, yy: &%s}",
		e.xx.asGoSource(), e.xy.asGoSource(), e.yx.asGoSource(), e.yy.asGoSource())
}

func (e exprName) asGoSource() string {
	return fmt.Sprintf("exprName{object: %d /* %s */, kind:  %d /* %s */}", e.object, e.object, e.kind, e.kind)
}

func (e exprTree) asGoSource() string {
	return fmt.Sprintf("exprTree{&%s, &%s}",
		e.left.asGoSource(), e.right.asGoSource())
}

func (e ruleTest) asGoSourceOpt(withType bool) string {
	s := fmt.Sprintf(`{
		expr: &%s, 
		kind:  %d, 
		qual: %d, 
		object: %d /* %s */, 
		op: %d /* %s */,
	}`,
		e.expr.asGoSource(), e.kind, e.qual, e.object, e.object, e.op, e.op)
	if withType {
		return "ruleTest" + s
	}
	return s
}

func (e ruleEdit) asGoSourceOpt(withType bool) string {
	s := fmt.Sprintf(`{
		expr: &%s, 
		binding:  %d, 
		object: %d /* %s */, 
		op: %d /* %s */,
	}`,
		e.expr.asGoSource(), e.binding, e.object, e.object, e.op, e.op)
	if withType {
		return "ruleEdit" + s
	}
	return s
}

func (e ruleEdit) asGoSource() string { return e.asGoSourceOpt(true) }
func (e ruleTest) asGoSource() string { return e.asGoSourceOpt(true) }

func (e expression) asGoSource() string {
	return fmt.Sprintf("expression{u: %s, op:  %d /* %s */}",
		e.u.asGoSource(), e.op, e.op)
}

func (e Pattern) asGoSource() string {
	chunks := make([]string, len(e))
	for i, obj := range e.sortedKeys() {
		chunks[i] = fmt.Sprintf("%d /* %s */: &%s", obj, obj, e[obj].asGoSource())
	}
	return fmt.Sprintf("Pattern{%s}", strings.Join(chunks, ","))
}

func (v Int) asGoSource() string    { return fmt.Sprintf("Int(%v)", v) }
func (v Float) asGoSource() string  { return fmt.Sprintf("Float(%v)", v) }
func (v String) asGoSource() string { return fmt.Sprintf("String(%q)", v) }
func (v Bool) asGoSource() string   { return fmt.Sprintf("Bool(%d)", v) }
func (v Charset) asGoSource() string {
	pages := make([]string, len(v.pages))
	for i, p := range v.pages {
		pages[i] = fmt.Sprintf("%#v", [8]uint32(p))
	}
	return fmt.Sprintf("Charset{pageNumbers: %#v, pages: []charPage{%s}", v.pageNumbers, strings.Join(pages, ","))
}

func (v Langset) asGoSource() string {
	return fmt.Sprintf("Langset{extra:%#v, page: %#v}", v.extra, v.page)
}

func (v Matrix) asGoSource() string {
	return fmt.Sprintf("Matrix{Xx: %v, Xy:  %v, Yx: %v, Yy: %v}",
		v.Xx, v.Xy, v.Yx, v.Yy)
}

func (v Range) asGoSource() string {
	return fmt.Sprintf("Range{Begin: %v, End:  %v}",
		v.Begin, v.End)
}

func (d directive) asGoSource() string {
	tests, edits := make([]string, len(d.tests)), make([]string, len(d.edits))
	for i, t := range d.tests {
		tests[i] = t.asGoSourceOpt(false)
	}
	for i, e := range d.edits {
		edits[i] = e.asGoSourceOpt(false)
	}

	// preserve reflect.DeepEqual
	testsString := "nil"
	if d.tests != nil {
		testsString = fmt.Sprintf("[]ruleTest{%s}", strings.Join(tests, ",\n"))
	}
	editsString := "nil"
	if d.edits != nil {
		editsString = fmt.Sprintf("[]ruleEdit{%s}", strings.Join(edits, ",\n"))
	}

	return fmt.Sprintf(`{
		tests: %s, 
		edits: %s,
		}`, testsString, editsString)
}

func (r ruleSet) asGoSource() string {
	subst := make([]string, matchKindEnd)
	for i, s := range r.subst {
		// preserve reflect.DeepEqual
		if s == nil {
			subst[i] = "nil"
		} else {
			l := make([]string, len(s))
			for j, d := range s {
				l[j] = d.asGoSource()
			}
			subst[i] = fmt.Sprintf("{%s}", strings.Join(l, ","))
		}
	}
	return fmt.Sprintf("{name: %q, description: %q, domain: %q, subst: [matchKindEnd][]directive{%s}}",
		r.name, r.description, r.domain, strings.Join(subst, ","))
}

func (fs Fontset) asGoSource() string {
	patterns := make([]string, len(fs))
	for i, p := range fs {
		patterns[i] = p.asGoSource()
	}
	return fmt.Sprintf("Fontset{%s}", strings.Join(patterns, ",\n"))
}

func (c Config) asGoSource() string {
	rules := make([]string, len(c.subst))
	for i, r := range c.subst {
		rules[i] = r.asGoSource()
	}

	customs := fmt.Sprintf("%#v", c.customObjects)
	customs = strings.ReplaceAll(customs, "fontconfig.", "")

	return fmt.Sprintf(`Config{
		subst: []ruleSet{%s},  
		customObjects: %s,
		acceptGlobs: %#v, 
		rejectGlobs: %#v, 
		acceptPatterns: %s, 
		rejectPatterns: %s, 
		maxObjects: %d,
		}`,
		strings.Join(rules, ",\n"), customs, c.acceptGlobs, c.rejectGlobs,
		c.acceptPatterns.asGoSource(), c.rejectPatterns.asGoSource(), c.maxObjects)
}
