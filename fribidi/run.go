package fribidi

import (
	"fmt"
	"log"
)

const levelSentinel = -1

type oneRun struct {
	prev *oneRun
	next *oneRun

	pos, len            int
	type_               CharType
	level, isolateLevel Level
	bracketType         BracketType

	/* Additional links for connecting the isolate tree */
	prevIsolate, nextIsolate *oneRun
}

func newRunList() *oneRun {
	var run oneRun

	run.type_ = maskSENTINEL
	run.level = levelSentinel
	run.pos = levelSentinel
	run.len = levelSentinel
	run.next = &run
	run.prev = &run

	return &run
}

func (x *oneRun) deleteNode() {
	x.prev.next = x.next
	x.next.prev = x.prev
}

func (list *oneRun) insertNodeBefore(x *oneRun) {
	x.prev = list.prev
	list.prev.next = x
	x.next = list
	list.prev = x
}

func (list *oneRun) moveNodeBefore(x *oneRun) {
	if x.prev != nil {
		x.deleteNode()
	}
	list.insertNodeBefore(x)
}

/* Return the type of previous run or the SOR, if already at the start of
   a level run. */
func (pp *oneRun) prevTypeOrSOR() CharType {
	if pp.prev.level == pp.level {
		return pp.prev.type_
	}
	return levelToDir(maxL(pp.prev.level, pp.level))
}

/* "Within this scope, bidirectional types EN and AN are treated as R" */
func (list *oneRun) typeAnEnAsRTL() CharType {
	if list.type_ == AN || list.type_ == EN {
		return RTL
	}
	return list.type_
}

/* Return the embedding direction of a link. */
func (link *oneRun) embeddingDirection() CharType {
	return levelToDir(link.level)
}

// bracketTypes is either empty or with same length as `bidiTypes`
func encodeBidiTypes(bidiTypes []CharType, bracketTypes []BracketType) *oneRun {
	/* Create the list sentinel */
	list := newRunList()
	last := list
	hasBrackets := len(bracketTypes) != 0

	/* Scan over the character types */
	for i, charType := range bidiTypes {
		bracketType := NoBracket
		if hasBrackets {
			bracketType = bracketTypes[i]
		}

		if charType != last.type_ || bracketType != NoBracket || // Always separate bracket into single char runs!
			last.bracketType != NoBracket || charType.IsIsolate() {
			run := &oneRun{}
			run.type_ = charType
			run.pos = i
			last.len = run.pos - last.pos
			last.next = run
			run.prev = last
			run.bracketType = bracketType
			last = run
		}
	}

	/* Close the circle */
	last.len = len(bidiTypes) - last.pos
	last.next = list
	list.prev = last

	list.validate()

	return list
}

/* override the run list 'base', with the runs in the list 'over', to
reinsert the previously-removed explicit codes (at X9) from
'explicits_list' back into 'type_rl_list' for example. This is used at the
end of I2 to restore the explicit marks, and also to reset the character
types of characters at L1.

it is assumed that the 'pos' of the first element in 'base' list is not
more than the 'pos' of the first element of the 'over' list, and the
'pos' of the last element of the 'base' list is not less than the 'pos'
of the last element of the 'over' list. these two conditions are always
satisfied for the two usages mentioned above.

Note:
  frees the over list.

Todo:
  use some explanatory names instead of p, q, ...
  rewrite comment above to remove references to special usage.
*/
func shadowRunList(base, over *oneRun, preserveLength bool) {
	var (
		r, t      *oneRun
		p         = base
		pos, pos2 int
	)

	base.validate()
	over.validate()
	//    for_run_list (q, over)
	for q := over.next; q.type_ != maskSENTINEL; q = q.next {
		if q.len == 0 || q.pos < pos {
			continue
		}
		pos = q.pos
		for p.next.type_ != maskSENTINEL && p.next.pos <= pos {
			p = p.next
		}
		/* now p is the element that q must be inserted 'in'. */
		pos2 = pos + q.len
		r = p
		for r.next.type_ != maskSENTINEL && r.next.pos < pos2 {
			r = r.next
		}
		if preserveLength {
			r.len += q.len
		}
		/* now r is the last element that q affects. */
		if p == r {
			/* split p into at most 3 intervals, and insert q in the place of
			the second interval, set r to be the third part. */
			/* third part needed? */
			if p.pos+p.len > pos2 {
				r = &oneRun{}
				p.next.prev = r
				r.next = p.next
				r.level = p.level
				r.isolateLevel = p.isolateLevel
				r.type_ = p.type_
				r.len = p.pos + p.len - pos2
				r.pos = pos2
			} else {
				r = r.next
			}

			if p.pos+p.len >= pos {
				/* first part needed? */
				if p.pos < pos {
					/* cut the end of p. */
					p.len = pos - p.pos
				} else {
					t = p
					p = p.prev
				}
			}
		} else {
			if p.pos+p.len >= pos {
				/* p needed? */
				if p.pos < pos {
					/* cut the end of p. */
					p.len = pos - p.pos
				} else {
					p = p.prev
				}
			}

			/* r needed? */
			if r.pos+r.len > pos2 {
				/* cut the beginning of r. */
				r.len = r.pos + r.len - pos2
				r.pos = pos2
			} else {
				r = r.next
			}

			/* remove the elements between p and r. */
			for s := p.next; s != r; {
				t = s
				s = s.next
			}
		}
		/* before updating the next and prev runs to point to the inserted q,
		we must remember the next element of q in the 'over' list.
		*/
		t = q
		q = q.prev
		t.deleteNode()
		p.next = t
		t.prev = p
		t.next = r
		r.prev = t
	}

	base.validate()
}

func (second *oneRun) mergeWithPrev() *oneRun {
	first := second.prev
	first.next = second.next
	first.next.prev = first
	first.len += second.len
	if second.nextIsolate != nil {
		second.nextIsolate.prevIsolate = second.prevIsolate
		/* The following edge case typically shouldn't happen, but fuzz
		   testing shows it does, and the assignment protects against
		   a dangling pointer. */
	} else if second.next.prevIsolate == second {
		second.next.prevIsolate = second.prevIsolate
	}
	if second.prevIsolate != nil {
		second.prevIsolate.nextIsolate = second.nextIsolate
	}
	first.nextIsolate = second.nextIsolate

	return first
}

func (list *oneRun) compact() {
	if list.next == nil {
		return
	}
	for list = list.next; list.type_ != maskSENTINEL; list = list.next {
		/* Don't join brackets! */
		if list.prev.type_ == list.type_ && list.prev.level == list.level &&
			list.bracketType == NoBracket && list.prev.bracketType == NoBracket {
			list = list.mergeWithPrev()
		}
	}
}

func (list *oneRun) compactNeutrals() {
	if list.next == nil {
		return
	}
	for list = list.next; list.type_ != maskSENTINEL; list = list.next {
		if list.prev.level == list.level &&
			(list.prev.type_ == list.type_ ||
				(list.prev.type_.isNeutral() && list.type_.isNeutral())) &&
			list.bracketType == NoBracket /* Don't join brackets! */ &&
			list.prev.bracketType == NoBracket {
			list = list.mergeWithPrev()
		}
	}
}

// The static sentinel is used to signal the end of an isolating sequence
var sentinel = oneRun{type_: maskSENTINEL, level: -1, isolateLevel: -1}

func (list *oneRun) getAdjacentRun(forward, skipNeutral bool) *oneRun {
	ppp := list.prevIsolate
	if forward {
		ppp = list.nextIsolate
	}

	if ppp == nil {
		return &sentinel
	}

	for ppp != nil {
		pppType := ppp.type_

		if pppType == maskSENTINEL {
			break
		}

		/* Note that when sweeping forward we continue one run
		   beyond the PDI to see what lies behind. When looking
		   backwards, this is not necessary as the leading isolate
		   run has already been assigned the resolved level. */
		if ppp.isolateLevel > list.isolateLevel /* <- How can this be true? */ ||
			(forward && pppType == PDI) || (skipNeutral && !pppType.IsStrong()) {
			if forward {
				ppp = ppp.nextIsolate
			} else {
				ppp = ppp.prevIsolate

			}
			if ppp == nil {
				ppp = &sentinel
			}

			continue
		}
		break
	}

	return ppp
}

// debug helpers

func assertT(b bool) {
	if !b {
		log.Fatal("assertion error")
	}
}

func (runList *oneRun) validate() {
	if debugMode {
		assertT(runList != nil)
		assertT(runList.next != nil)
		assertT(runList.next.prev == runList)
		assertT(runList.type_ == maskSENTINEL)
		q := runList
		for ; q.type_ != maskSENTINEL; q = q.next {
			assertT(q.next != nil)
			assertT(q.next.prev == q)
		}
		assertT(q == runList)
	}
}

func (r oneRun) printTypesRe() {
	fmt.Print("  Run types  : ")
	for pp := r.next; pp.type_ != maskSENTINEL; pp = pp.next {
		fmt.Printf("%d:%d(%s)[%d,%d] ", pp.pos, pp.len, pp.type_, pp.level, pp.isolateLevel)
	}
	fmt.Println()
}

func (r oneRun) printResolvedTypes() {
	fmt.Print("  Res. types: ")
	for pp := r.next; pp.type_ != maskSENTINEL; pp = pp.next {
		for i := pp.len; i != 0; i-- {
			fmt.Printf("%s ", pp.type_)
		}
	}
	fmt.Println()
}

func (r oneRun) printResolvedLevels() {
	fmt.Print("  Res. levels: ")
	for pp := r.next; pp.type_ != maskSENTINEL; pp = pp.next {
		for i := pp.len; i != 0; i-- {
			fmt.Printf("%d ", pp.level)
		}
	}
	fmt.Println()
}
