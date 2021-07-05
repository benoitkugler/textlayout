package fontconfig

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// ported from fontconfig/src/FcCharset.c   Copyright © 2001 Keith Packard

// charPage is the base storage for a compact rune set.
// A rune is first reduced to its lower byte 'b'. Then the index
// of 'b' in the leaf is given by the 3 high bits (from 0 to 7)
// and the position in the resulting uint32 is given by the 5 lower bits (from 0 to 31)
type charPage [8]uint32

// By construction, only the 3 last bytes of a rune are taken
// into account meaning that only rune <= 0x00FFFFFF are valid,
// which is large enough for all valid Unicode points.
const maxCharsetRune = 0xFFFFFF

// Charset is a compact rune set.
//
// Its internal representation is composed of a variable
// number of 'pages', where each page is a boolean
// set of size 256, encoding the last byte of a rune.
// Each rune is then mapped to a page number, defined by
// it second and third bytes.
type Charset struct {
	// sorted list of the pages
	pageNumbers []uint16
	// same length as pageNumbers;
	// pages[pos] is the page for the number pageNumbers[pos]
	pages []charPage
}

var _ hasher = Charset{}

// String returns a represensation of the internal storage
// of the charset.
func (s Charset) String() string {
	return fmt.Sprintf("%v\n%v", s.pageNumbers, s.pages)
}

func (s Charset) hash() []byte {
	const size = 2 + 4*8
	out := make([]byte, size*len(s.pageNumbers))
	for i, p := range s.pageNumbers {
		binary.BigEndian.PutUint16(out[size*i:], p)
		for j, u := range s.pages[i] {
			binary.BigEndian.PutUint32(out[size*i+2+4*j:], u)
		}
	}
	return out
}

func parseCharSet(str string) (Charset, error) {
	fields := strings.Fields(str)

	var out Charset
	for len(fields) != 0 {
		// either the separator is surrounded by space or not
		chunks := strings.Split(fields[0], "-")
		firstS, lastS := fields[0], ""
		if len(chunks) == 2 {
			firstS, lastS = chunks[0], chunks[1]
			fields = fields[1:]
		} else if strings.HasSuffix(fields[0], "-") && len(fields) >= 2 {
			firstS = strings.TrimSuffix(firstS, "-")
			lastS = fields[1]
			fields = fields[2:]
		} else if len(fields) >= 3 && fields[1] == "-" {
			lastS = fields[2]
			fields = fields[3:]
		}

		first, err := strconv.ParseInt(firstS, 16, 32)
		if err != nil {
			return out, fmt.Errorf("fontconfig: invalid charset: invalid first element %s: %s", firstS, err)
		}
		last := first
		if lastS != "" {
			last, err = strconv.ParseInt(lastS, 16, 32)
			if err != nil {
				return out, fmt.Errorf("fontconfig: invalid charset: invalid last element %s: %s", lastS, err)
			}
		}

		if first < 0 || last < 0 || last < first || last > 0x10ffff {
			return out, fmt.Errorf("invalid charset range %s %s", firstS, lastS)
		}

		for u := rune(first); u < rune(last+1); u++ {
			out.AddChar(u)
		}
	}
	return out, nil
}

// Search for the leaf containing with the specified num
// (binary search on fcs.numbers[low:])
// Return its index if it exists, otherwise return negative of
// the (position + 1) where it should be inserted
func (fcs Charset) findLeafForward(low int, num uint16) int {
	numbers := fcs.pageNumbers
	high := len(numbers) - 1
	for low <= high {
		mid := (low + high) >> 1
		page := numbers[mid]
		if page == num {
			return mid
		}
		if page < num {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if high < 0 || (high < len(numbers) && numbers[high] < num) {
		high++
	}
	return -(high + 1)
}

// Locate the leaf containing the specified char, return
// its index if it exists, otherwise return negative of
// the (position + 1) where it should be inserted
// page is obtained from a rune by uint16(r>>8)
func (fcs Charset) findLeafPos(page uint16) int {
	return fcs.findLeafForward(0, page)
}

// Returns the number of chars that are in `a` but not in `b`.
func charsetSubtractCount(a, b Charset) uint32 {
	var count int
	ai, bi := newCharsetIter(a), newCharsetIter(b)
	for ai.leaf != nil {
		aiPage := ai.page()
		if aiPage <= bi.page() {
			am := ai.leaf
			if aiPage == bi.page() {
				bm := bi.leaf
				for i := range am {
					count += bits.OnesCount32(am[i] & ^bm[i]) // *am++ & ~*bm++
				}
			} else {
				for i := range am {
					count += bits.OnesCount32(am[i])
				}
			}
			ai.next()
		} else if bi.leaf != nil {
			bi.set(aiPage)
		}
	}
	return uint32(count)
}

// Returns whether `a` and `b` contain the same set of Unicode chars.
func charsetEqual(a, b Charset) bool {
	ai, bi := newCharsetIter(a), newCharsetIter(b)
	for ai.leaf != nil && bi.leaf != nil {
		if ai.page() != bi.page() {
			return false
		}
		if *ai.leaf != *bi.leaf {
			return false
		}
		ai.next()
		bi.next()
	}
	return ai.leaf == bi.leaf
}

// return true iff a is a subset of b
func (a Charset) isSubset(b Charset) bool {
	ai, bi := 0, 0
	for ai < len(a.pageNumbers) && bi < len(b.pageNumbers) {
		an := a.pageNumbers[ai]
		bn := b.pageNumbers[bi]
		// Check matching pages
		if an == bn {
			am := a.pages[ai]
			bm := b.pages[bi]

			if am != bm {
				//  Does am have any bits not in bm?
				for j, av := range am {
					if av & ^bm[j] != 0 {
						return false
					}
				}
			}
			ai++
			bi++
		} else if an < bn { // Does a have any pages not in b?
			return false
		} else {
			bi = b.findLeafForward(bi+1, an)
			if bi < 0 {
				bi = -bi - 1
			}
		}
	}
	//  did we look at every page?
	return ai >= len(a.pageNumbers)
}

// Locate the leaf containing the specified char, creating it if desired
// we assume the rune is valid and the return page is always non nil
func (fcs *Charset) findLeafCreate(page uint16) *charPage {
	pos := fcs.findLeafPos(page)
	if pos >= 0 {
		return &fcs.pages[pos]
	}

	pos = -pos - 1
	fcs.putLeaf(page, charPage{}, pos)
	return &fcs.pages[pos]
}

// insert at pos, meaning the resulting leaf can be accessed via &fcs.leaves[pos]
// assume the rune is valid
func (fcs *Charset) putLeaf(page uint16, leaf charPage, pos int) {
	// insert in slice
	fcs.pageNumbers = append(fcs.pageNumbers, 0)
	fcs.pages = append(fcs.pages, charPage{})
	copy(fcs.pageNumbers[pos+1:], fcs.pageNumbers[pos:])
	copy(fcs.pages[pos+1:], fcs.pages[pos:])
	fcs.pageNumbers[pos] = page
	fcs.pages[pos] = leaf
}

func (fcs *Charset) addLeaf(pageNumber uint16, leaf charPage) {
	new := fcs.findLeafCreate(pageNumber)
	*new = leaf
}

// AddChar add `r` to the set.
func (fcs *Charset) AddChar(r rune) {
	leaf := fcs.findLeafCreate(uint16(r >> 8))
	b := &leaf[(r&0xff)>>5]
	*b |= (1 << (r & 0x1f))
}

// DelChar remove the rune from the set.
func (fcs Charset) DelChar(r rune) {
	leaf := fcs.findLeaf(uint16(r >> 8))
	if leaf == nil {
		return
	}
	b := &leaf[(r&0xff)>>5]
	*b &= ^(1 << (r & 0x1f))
	// we don't bother removing the leaf if it's empty
}

// Copy returns a deep copy of the charset.
func (fcs Charset) Copy() Charset {
	var out Charset
	out.pageNumbers = append([]uint16(nil), fcs.pageNumbers...)
	out.pages = append([]charPage(nil), fcs.pages...)
	return out
}

// Adds all chars in `b` to `a`.
// In other words, this is an in-place version of FcCharsetUnion.
// It returns whether any new chars from `b` were added to `a`.
func (a *Charset) merge(b Charset) bool {
	if a == nil {
		return false
	}

	changed := !b.isSubset(*a)
	if !changed {
		return changed
	}

	for ai, bi := 0, 0; bi < len(b.pageNumbers); {
		an := ^uint16(0)
		if ai < len(a.pageNumbers) {
			an = a.pageNumbers[ai]
		}
		bn := b.pageNumbers[bi]

		if an < bn {
			ai = a.findLeafForward(ai+1, bn)
			if ai < 0 {
				ai = -ai - 1
			}
		} else {
			bl := b.pages[bi]
			if bn < an {
				a.addLeaf(bn, bl)
			} else {
				al := &a.pages[ai]
				al.unionLeaf(*al, bl)
			}

			ai++
			bi++
		}
	}

	return changed
}

func (fcs *Charset) findLeaf(pageNumber uint16) *charPage {
	pos := fcs.findLeafPos(pageNumber)
	if pos >= 0 {
		return &fcs.pages[pos]
	}
	return nil
}

// HasChar returns `true` if `r` is in the set.
func (fcs *Charset) HasChar(r rune) bool {
	leaf := fcs.findLeaf(uint16(r >> 8))
	if leaf == nil {
		return false
	}
	return leaf[(r&0xff)>>5]&(1<<(r&0x1f)) != 0
}

// Len returns the number of runes in the set.
func (a Charset) Len() int {
	count := 0
	ai := newCharsetIter(a)
	for ; ai.leaf != nil; ai.next() {
		for _, am := range ai.leaf {
			count += bits.OnesCount32(am)
		}
	}
	return count
}

func charsetUnion(a, b Charset) Charset {
	return operate(a, b, (*charPage).unionLeaf, true)
}

func operate(a, b Charset, overlap func(result *charPage, al, bl charPage) bool, bonly bool) Charset {
	var fcs Charset
	ai, bi := newCharsetIter(a), newCharsetIter(b)
	for ai.leaf != nil || (bonly && bi.leaf != nil) {
		aiPage, biPage := ai.page(), bi.page()
		if aiPage < biPage {
			if ai.leaf != nil {
				fcs.addLeaf(aiPage, *ai.leaf)
			}
			ai.next()
		} else if biPage < aiPage {
			if bonly {
				if bi.leaf != nil {
					fcs.addLeaf(biPage, *bi.leaf)
				}
				bi.next()
			} else {
				bi.set(aiPage)
			}
		} else {
			var leaf charPage
			if overlap(&leaf, *ai.leaf, *bi.leaf) {
				fcs.addLeaf(aiPage, leaf)
			}
			ai.next()
			bi.next()
		}
	}
	return fcs
}

// store in `result` the union of `a` and `b`
func (result *charPage) unionLeaf(a, b charPage) bool {
	for i := range result {
		result[i] = a[i] | b[i]
	}
	return true
}

func (result *charPage) subtractLeaf(al, bl charPage) bool {
	nonempty := false
	for i := range result {
		v := al[i] & ^bl[i]
		result[i] = v
		if v != 0 {
			nonempty = true
		}
	}
	return nonempty
}

// Returns a set including only those chars found in `a` but not `b`.
func charsetSubtract(a, b Charset) Charset {
	return operate(a, b, (*charPage).subtractLeaf, false)
}

// charsetIter is an iterator for the leaves of a charset
type charsetIter struct {
	// the value &charset.leaves[pos],
	// cached for convenience
	leaf    *charPage
	charset Charset
	pos     int
}

func newCharsetIter(fcs Charset) *charsetIter {
	out := &charsetIter{charset: fcs}
	out.updateLeaf()
	return out
}

// return the maximum page when out of pos
func (iter *charsetIter) page() uint16 {
	if iter.pos < len(iter.charset.pageNumbers) {
		return iter.charset.pageNumbers[iter.pos]
	}
	return ^uint16(0) // this is an invalid page
}

func (iter *charsetIter) updateLeaf() {
	if iter.pos < len(iter.charset.pageNumbers) {
		iter.leaf = &iter.charset.pages[iter.pos]
	} else {
		iter.leaf = nil
	}
}

// Set `iter` to given page.
// If the page is not in the charset, the next position is selected
func (iter *charsetIter) set(pageNumber uint16) {
	pos := iter.charset.findLeafPos(pageNumber)
	if pos < 0 {
		pos = -pos - 1
	}
	iter.pos = pos
	iter.updateLeaf()
}

func (iter *charsetIter) next() {
	iter.pos += 1
	iter.updateLeaf()
}
