package fontconfig

import (
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

// FcCharset is a compact rune set.
//
// Its internal representation is composed of a variable
// number of 'pages', where each page is a boolean
// set of size 256, encoding the last byte of a rune.
// Each rune is then mapped to a page number, defined by
// it second and third bytes.
type FcCharset struct {
	// sorted list of the pages
	pageNumbers []uint16
	// same length as pageNumbers;
	// pages[pos] is the page for the number pageNumbers[pos]
	pages []charPage
}

// func  parseCharRange (str []byte) ([]byte, uint32, uint32, bool) {
// 	//  char *s = (char *) *string;
// 	//  char *t;
// 	//  long first, last;

// 	str = bytes.TrimLeftFunc(str, unicode.IsSpace)

// 	 t = s;
// 	 errno = 0;
// 	 first = last = strtol (s, &s, 16);
// 	strconv.ParseInt (string())

// 	 if (errno)
// 		 return false;
// 	 for (isspace(*s))
// 		 s++;
// 	 if (*s == '-')
// 	 {
// 		 s++;
// 		 errno = 0;
// 		 last = strtol (s, &s, 16);
// 		 if (errno)
// 		 return false;
// 	 }

// 	 if (s == t || first < 0 || last < 0 || last < first || last > 0x10ffff)
// 		  return false;

// 	 *string = (FcChar8 *) s;
// 	 *pfirst = first;
// 	 *plast = last;
// 	 return true;
//  }

func FcNameParseCharSet(str string) (FcCharset, error) {
	fields := strings.Fields(str)

	var out FcCharset
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
			out.addChar(u)
		}
	}
	return out, nil
}

// Search for the leaf containing with the specified num
// (binary search on fcs.numbers[low:])
// Return its index if it exists, otherwise return negative of
// the (position + 1) where it should be inserted
func (fcs FcCharset) findLeafForward(low int, num uint16) int {
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
func (fcs FcCharset) findLeafPos(page uint16) int {
	return fcs.findLeafForward(0, page)
}

func popCount(c1 uint32) int { return bits.OnesCount32(c1) }

// Returns the number of chars that are in `a` but not in `b`.
func charsetSubtractCount(a, b FcCharset) uint32 {
	var count int
	ai, bi := newCharsetIter(a), newCharsetIter(b)
	for ai.leaf != nil {
		aiPage := ai.page()
		if aiPage <= bi.page() {
			am := ai.leaf
			if aiPage == bi.page() {
				bm := bi.leaf
				for i := range am {
					count += popCount(am[i] & ^bm[i]) // *am++ & ~*bm++
				}
			} else {
				for i := range am {
					count += popCount(am[i])
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
func FcCharsetEqual(a, b FcCharset) bool {
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
func (a FcCharset) isSubset(b FcCharset) bool {
	ai, bi := 0, 0
	for ai < len(a.pageNumbers) && bi < len(b.pageNumbers) {
		an := a.pageNumbers[ai]
		bn := b.pageNumbers[ai]
		// Check matching pages
		if an == bn {
			am := a.pages[ai]
			bm := b.pages[ai]

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
func (fcs *FcCharset) findLeafCreate(page uint16) *charPage {
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
func (fcs *FcCharset) putLeaf(page uint16, leaf charPage, pos int) {
	// insert in slice
	fcs.pageNumbers = append(fcs.pageNumbers, 0)
	fcs.pages = append(fcs.pages, charPage{})
	copy(fcs.pageNumbers[pos+1:], fcs.pageNumbers[pos:])
	copy(fcs.pages[pos+1:], fcs.pages[pos:])
	fcs.pageNumbers[pos] = page
	fcs.pages[pos] = leaf
}

func (fcs *FcCharset) addLeaf(pageNumber uint16, leaf charPage) {
	new := fcs.findLeafCreate(pageNumber)
	*new = leaf
}

func (fcs *FcCharset) addChar(ucs4 rune) {
	leaf := fcs.findLeafCreate(uint16(ucs4 >> 8))
	b := &leaf[(ucs4&0xff)>>5]
	*b |= (1 << (ucs4 & 0x1f))
}

// Adds all chars in `b` to `a`.
// In other words, this is an in-place version of FcCharsetUnion.
// It returns whether any new chars from `b` were added to `a`.
func (a *FcCharset) merge(b FcCharset) bool {
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

func (fcs *FcCharset) findLeaf(pageNumber uint16) *charPage {
	pos := fcs.findLeafPos(pageNumber)
	if pos >= 0 {
		return &fcs.pages[pos]
	}
	return nil
}

func (fcs *FcCharset) hasChar(ucs4 rune) bool {
	leaf := fcs.findLeaf(uint16(ucs4 >> 8))
	if leaf == nil {
		return false
	}
	return leaf[(ucs4&0xff)>>5]&(1<<(ucs4&0x1f)) != 0
}

func (a FcCharset) count() int {
	count := 0
	ai := newCharsetIter(a)
	for ; ai.leaf != nil; ai.next() {
		for _, am := range ai.leaf {
			count += popCount(am)
		}
	}
	return count
}

func charsetUnion(a, b FcCharset) FcCharset {
	return operate(a, b, (*charPage).unionLeaf, true)
}

func operate(a, b FcCharset, overlap func(result *charPage, al, bl charPage) bool, bonly bool) FcCharset {
	var fcs FcCharset
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
func charsetSubtract(a, b FcCharset) FcCharset {
	return operate(a, b, (*charPage).subtractLeaf, false)
}

// charsetIter is an iterator for the leaves of a charset
type charsetIter struct {
	charset FcCharset
	pos     int
	// the value &charset.leaves[pos],
	// cached for convenience
	leaf *charPage
}

func newCharsetIter(fcs FcCharset) *charsetIter {
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

//  FcCharset *
//  FcCharsetCreate (void)
//  {
// 	 FcCharset	*fcs;

// 	 fcs = (FcCharset *) malloc (sizeof (FcCharset));
// 	 if (!fcs)
// 	 return 0;
// 	 FcRefInit (&fcs.ref, 1);
// 	 len(fcs.numbers) = 0;
// 	 fcs.leaves_offset = 0;
// 	 fcs.numbers_offset = 0;
// 	 return fcs;
//  }

//  FcCharset *
//  FcCharsetNew (void)
//  {
// 	 return FcCharsetCreate ();
//  }

//  void
//  FcCharsetDestroy (fcs *FcCharset)
//  {
// 	 int i;

// 	 if (fcs)
// 	 {
// 	 if (FcRefIsConst (&fcs.ref))
// 	 {
// 		 FcCacheObjectDereference (fcs);
// 		 return;
// 	 }
// 	 if (FcRefDec (&fcs.ref) != 1)
// 		 return;
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 		 free (FcCharsetLeaf (fcs, i));
// 	 if (len(fcs.numbers))
// 	 {
// 		 free (FcCharsetLeaves (fcs));
// 		 free (FcCharsetNumbers (fcs));
// 	 }
// 	 free (fcs);
// 	 }
//  }

//  static FcBool
//  FcCharsetInsertLeaf (fcs *FcCharset, ucs4 rune , FcCharLeaf *leaf)
//  {
// 	 int		    pos;

// 	 pos = findLeafPos (fcs, ucs4);
// 	 if (pos >= 0)
// 	 {
// 	 free (FcCharsetLeaf (fcs, pos));
// 	 FcCharsetLeaves(fcs)[pos] = FcPtrToOffset (FcCharsetLeaves(fcs),
// 							leaf);
// 	 return true;
// 	 }
// 	 pos = -pos - 1;
// 	 return putLeaf (fcs, ucs4, leaf, pos);
//  }

//  FcBool
//  FcCharsetDelChar (fcs *FcCharset, ucs4 rune )
//  {
// 	 leaf *FcCharLeaf;
// 	 uint32	*b;

// 	 if (fcs == NULL || FcRefIsConst (&fcs.ref))
// 	 return false;
// 	 leaf = FcCharsetFindLeaf (fcs, ucs4);
// 	 if (!leaf)
// 	 return true;
// 	 b = &leaf.map_[(ucs4 & 0xff) >> 5];
// 	 *b &= ~(1U << (ucs4 & 0x1f));
// 	 /* We don't bother removing the leaf if it's empty */
// 	 return true;
//  }

//  FcCharset *
//  FcCharsetCopy (FcCharset *src)
//  {
// 	 if (src)
// 	 {
// 	 if (!FcRefIsConst (&src.ref))
// 		 FcRefInc (&src.ref);
// 	 else
// 		 FcCacheObjectReference (src);
// 	 }
// 	 return src;
//  }

//  static FcBool
//  FcCharsetIntersectLeaf (FcCharLeaf *result,
// 			 const FcCharLeaf *al,
// 			 const FcCharLeaf *bl)
//  {
// 	 int	    i;
// 	 FcBool  nonempty = false;

// 	 for (i = 0; i < 256/32; i++)
// 	 if ((result.map_[i] = al.map_[i] & bl.map_[i]))
// 		 nonempty = true;
// 	 return nonempty;
//  }

//  FcCharset *
//  FcCharsetIntersect (const FcCharset *a, const FcCharset *b)
//  {
// 	 return operate (a, b, FcCharsetIntersectLeaf, false, false);
//  }

//  uint32
//  FcCharsetIntersectCount (const FcCharset *a, const FcCharset *b)
//  {
// 	 FcCharsetIter   ai, bi;
// 	 uint32	    count = 0;

// 	 if (a && b)
// 	 {
// 	 start (a, &ai);
// 	 start (b, &bi);
// 	 for (ai.leaf && bi.leaf)
// 	 {
// 		 if (ai.ucs4 == bi.ucs4)
// 		 {
// 		 uint32	*am = ai.leaf.map_;
// 		 uint32	*bm = bi.leaf.map_;
// 		 int		i = 256/32;
// 		 for (i--)
// 			 count += popCount (*am++ & *bm++);
// 		 next (a, &ai);
// 		 }
// 		 else if (ai.ucs4 < bi.ucs4)
// 		 {
// 		 ai.ucs4 = bi.ucs4;
// 		 set (a, &ai);
// 		 }
// 		 if (bi.ucs4 < ai.ucs4)
// 		 {
// 		 bi.ucs4 = ai.ucs4;
// 		 set (b, &bi);
// 		 }
// 	 }
// 	 }
// 	 return count;
//  }

//  /*
//   * These two functions efficiently walk the entire charmap for
//   * other software (like pango) that want their own copy
//   */

//  uint32
//  FcCharsetNextPage (const FcCharset  *a,
// 			uint32	    map_[FC_CHARSET_MAP_SIZE],
// 			uint32	    *next)
//  {
// 	 FcCharsetIter   ai;
// 	 uint32	    page;

// 	 if (!a)
// 	 return FC_CHARSET_DONE;
// 	 ai.ucs4 = *next;
// 	 set (a, &ai);
// 	 if (!ai.leaf)
// 	 return FC_CHARSET_DONE;

// 	 /*
// 	  * Save current information
// 	  */
// 	 page = ai.ucs4;
// 	 memcpy (map_, ai.leaf.map_, sizeof (ai.leaf.map_));
// 	 /*
// 	  * Step to next page
// 	  */
// 	 next (a, &ai);
// 	 *next = ai.ucs4;

// 	 return page;
//  }

//  uint32
//  FcCharsetFirstPage (const FcCharset *a,
// 			 uint32	    map_[FC_CHARSET_MAP_SIZE],
// 			 uint32	    *next)
//  {
// 	 *next = 0;
// 	 return FcCharsetNextPage (a, map_, next);
//  }

//  /*
//   * old coverage API, rather hard to use correctly
//   */

//  uint32
//  FcCharsetCoverage (const FcCharset *a, uint32 page, uint32 *result)
//  {
// 	 FcCharsetIter   ai;

// 	 ai.ucs4 = page;
// 	 set (a, &ai);
// 	 if (!ai.leaf)
// 	 {
// 	 memset (result, '\0', 256 / 8);
// 	 page = 0;
// 	 }
// 	 else
// 	 {
// 	 memcpy (result, ai.leaf.map_, sizeof (ai.leaf.map_));
// 	 next (a, &ai);
// 	 page = ai.ucs4;
// 	 }
// 	 return page;
//  }

//  static void
//  FcNameUnparseUnicode (FcStrBuf *buf, uint32 u)
//  {
// 	 FcChar8	    buf_static[64];
// 	 snprintf ((char *) buf_static, sizeof (buf_static), "%x", u);
// 	 FcStrBufString (buf, buf_static);
//  }

//  FcBool
//  FcNameUnparseCharSet (FcStrBuf *buf, const FcCharset *c)
//  {
// 	 FcCharsetIter   ci;
// 	 uint32	    first, last;
// 	 int		    i;
//  #ifdef CHECK
// 	 int		    len = buf.len;
//  #endif

// 	 first = last = 0x7FFFFFFF;

// 	 for (start (c, &ci);
// 	  ci.leaf;
// 	  next (c, &ci))
// 	 {
// 	 for (i = 0; i < 256/32; i++)
// 	 {
// 		 uint32 bits = ci.leaf.map_[i];
// 		 uint32 u = ci.ucs4 + i * 32;

// 		 for (bits)
// 		 {
// 		 if (bits & 1)
// 		 {
// 			 if (u != last + 1)
// 			 {
// 				 if (last != first)
// 				 {
// 				 FcStrBufChar (buf, '-');
// 				 FcNameUnparseUnicode (buf, last);
// 				 }
// 				 if (last != 0x7FFFFFFF)
// 				 FcStrBufChar (buf, ' ');
// 				 /* Start new range. */
// 				 first = u;
// 				 FcNameUnparseUnicode (buf, u);
// 			 }
// 			 last = u;
// 		 }
// 		 bits >>= 1;
// 		 u++;
// 		 }
// 	 }
// 	 }
// 	 if (last != first)
// 	 {
// 	 FcStrBufChar (buf, '-');
// 	 FcNameUnparseUnicode (buf, last);
// 	 }
//  #ifdef CHECK
// 	 {
// 	 FcCharset	*check;
// 	 uint32	missing;
// 	 FcCharsetIter	ci, checki;

// 	 /* null terminate for parser */
// 	 FcStrBufChar (buf, '\0');
// 	 /* step back over null for life after test */
// 	 buf.len--;
// 	 check = FcNameParseCharSet (buf.buf + len);
// 	 start (c, &ci);
// 	 start (check, &checki);
// 	 for (ci.leaf || checki.leaf)
// 	 {
// 		 if (ci.ucs4 < checki.ucs4)
// 		 {
// 		 printf ("Missing leaf node at 0x%x\n", ci.ucs4);
// 		 next (c, &ci);
// 		 }
// 		 else if (checki.ucs4 < ci.ucs4)
// 		 {
// 		 printf ("Extra leaf node at 0x%x\n", checki.ucs4);
// 		 next (check, &checki);
// 		 }
// 		 else
// 		 {
// 		 int	    i = 256/32;
// 		 uint32    *cm = ci.leaf.map_;
// 		 uint32    *checkm = checki.leaf.map_;

// 		 for (i = 0; i < 256; i += 32)
// 		 {
// 			 if (*cm != *checkm)
// 			 printf ("Mismatching sets at 0x%08x: 0x%08x != 0x%08x\n",
// 				 ci.ucs4 + i, *cm, *checkm);
// 			 cm++;
// 			 checkm++;
// 		 }
// 		 next (c, &ci);
// 		 next (check, &checki);
// 		 }
// 	 }
// 	 if ((missing = FcCharsetSubtractCount (c, check)))
// 		 printf ("%d missing in reparsed result\n", missing);
// 	 if ((missing = FcCharsetSubtractCount (check, c)))
// 		 printf ("%d extra in reparsed result\n", missing);
// 	 FcCharsetDestroy (check);
// 	 }
//  #endif

// 	 return true;
//  }

//  typedef struct _FcCharLeafEnt FcCharLeafEnt;

//  struct _FcCharLeafEnt {
// 	 FcCharLeafEnt   *next;
// 	 uint32	    hash;
// 	 FcCharLeaf	    leaf;
//  };

//  #define FC_CHAR_LEAF_BLOCK	(4096 / sizeof (FcCharLeafEnt))
//  #define FC_CHAR_LEAF_HASH_SIZE	257

//  typedef struct _FcCharsetEnt FcCharsetEnt;

//  struct _FcCharsetEnt {
// 	 FcCharsetEnt	*next;
// 	 uint32		hash;
// 	 FcCharset		set;
//  };

//  typedef struct _FcCharsetOrigEnt FcCharsetOrigEnt;

//  struct _FcCharsetOrigEnt {
// 	 FcCharsetOrigEnt	*next;
// 	 const FcCharset    	*orig;
// 	 const FcCharset    	*frozen;
//  };

//  #define FC_CHAR_SET_HASH_SIZE    67

//  struct _FcCharsetFreezer {
// 	 FcCharLeafEnt   *leaf_hash_table[FC_CHAR_LEAF_HASH_SIZE];
// 	 FcCharLeafEnt   **leaf_blocks;
// 	 int		    leaf_block_count;
// 	 FcCharsetEnt    *set_hash_table[FC_CHAR_SET_HASH_SIZE];
// 	 FcCharsetOrigEnt	*orig_hash_table[FC_CHAR_SET_HASH_SIZE];
// 	 FcCharLeafEnt   *current_block;
// 	 int		    leaf_remain;
// 	 int		    leaves_seen;
// 	 int		    charsets_seen;
// 	 int		    leaves_allocated;
// 	 int		    charsets_allocated;
//  };

//  static FcCharLeafEnt *
//  FcCharLeafEntCreate (FcCharsetFreezer *freezer)
//  {
// 	 if (!freezer.leaf_remain)
// 	 {
// 	 FcCharLeafEnt **newBlocks;

// 	 freezer.leaf_block_count++;
// 	 newBlocks = realloc (freezer.leaf_blocks, freezer.leaf_block_count * sizeof (FcCharLeafEnt *));
// 	 if (!newBlocks)
// 		 return 0;
// 	 freezer.leaf_blocks = newBlocks;
// 	 freezer.current_block = freezer.leaf_blocks[freezer.leaf_block_count-1] = malloc (FC_CHAR_LEAF_BLOCK * sizeof (FcCharLeafEnt));
// 	 if (!freezer.current_block)
// 		 return 0;
// 	 freezer.leaf_remain = FC_CHAR_LEAF_BLOCK;
// 	 }
// 	 freezer.leaf_remain--;
// 	 freezer.leaves_allocated++;
// 	 return freezer.current_block++;
//  }

//  static uint32
//  FcCharLeafHash (FcCharLeaf *leaf)
//  {
// 	 uint32	hash = 0;
// 	 int		i;

// 	 for (i = 0; i < 256/32; i++)
// 	 hash = ((hash << 1) | (hash >> 31)) ^ leaf.map_[i];
// 	 return hash;
//  }

//  static FcCharLeaf *
//  FcCharsetFreezeLeaf (FcCharsetFreezer *freezer, FcCharLeaf *leaf)
//  {
// 	 uint32			hash = FcCharLeafHash (leaf);
// 	 FcCharLeafEnt		**bucket = &freezer.leaf_hash_table[hash % FC_CHAR_LEAF_HASH_SIZE];
// 	 FcCharLeafEnt		*ent;

// 	 for (ent = *bucket; ent; ent = ent.next)
// 	 {
// 	 if (ent.hash == hash && !memcmp (&ent.leaf, leaf, sizeof (FcCharLeaf)))
// 		 return &ent.leaf;
// 	 }

// 	 ent = FcCharLeafEntCreate(freezer);
// 	 if (!ent)
// 	 return 0;
// 	 ent.leaf = *leaf;
// 	 ent.hash = hash;
// 	 ent.next = *bucket;
// 	 *bucket = ent;
// 	 return &ent.leaf;
//  }

//  static uint32
//  FcCharsetHash (fcs *FcCharset)
//  {
// 	 uint32	hash = 0;
// 	 int		i;

// 	 /* hash in leaves */
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 	 hash = ((hash << 1) | (hash >> 31)) ^ FcCharLeafHash (FcCharsetLeaf(fcs,i));
// 	 /* hash in numbers */
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 	 hash = ((hash << 1) | (hash >> 31)) ^ fcs.numbers[i];
// 	 return hash;
//  }

//  static FcBool
//  FcCharsetFreezeOrig (FcCharsetFreezer *freezer, const FcCharset *orig, const FcCharset *frozen)
//  {
// 	 FcCharsetOrigEnt	**bucket = &freezer.orig_hash_table[((uintptr_t) orig) % FC_CHAR_SET_HASH_SIZE];
// 	 FcCharsetOrigEnt	*ent;

// 	 ent = malloc (sizeof (FcCharsetOrigEnt));
// 	 if (!ent)
// 	 return false;
// 	 ent.orig = orig;
// 	 ent.frozen = frozen;
// 	 ent.next = *bucket;
// 	 *bucket = ent;
// 	 return true;
//  }

//  static FcCharset *
//  FcCharsetFreezeBase (FcCharsetFreezer *freezer, fcs *FcCharset)
//  {
// 	 uint32		hash = FcCharsetHash (fcs);
// 	 FcCharsetEnt	**bucket = &freezer.set_hash_table[hash % FC_CHAR_SET_HASH_SIZE];
// 	 FcCharsetEnt	*ent;
// 	 int			size;
// 	 int			i;

// 	 for (ent = *bucket; ent; ent = ent.next)
// 	 {
// 	 if (ent.hash == hash &&
// 		 ent.set.num == len(fcs.numbers) &&
// 		 !memcmp (FcCharsetNumbers(&ent.set),
// 			  fcs.numbers,
// 			  len(fcs.numbers) * sizeof (FcChar16)))
// 	 {
// 		 FcBool ok = true;
// 		 int i;

// 		 for (i = 0; i < len(fcs.numbers); i++)
// 		 if (FcCharsetLeaf(&ent.set, i) != FcCharsetLeaf(fcs, i))
// 			 ok = false;
// 		 if (ok)
// 		 return &ent.set;
// 	 }
// 	 }

// 	 size = (sizeof (FcCharsetEnt) +
// 		 len(fcs.numbers) * sizeof (FcCharLeaf *) +
// 		 len(fcs.numbers) * sizeof (FcChar16));
// 	 ent = malloc (size);
// 	 if (!ent)
// 	 return 0;

// 	 freezer.charsets_allocated++;

// 	 FcRefSetConst (&ent.set.ref);
// 	 ent.set.num = len(fcs.numbers);
// 	 if (len(fcs.numbers))
// 	 {
// 	 intptr_t    *ent_leaves;

// 	 ent.set.leaves_offset = sizeof (ent.set);
// 	 ent.set.numbers_offset = (ent.set.leaves_offset +
// 					len(fcs.numbers) * sizeof (intptr_t));

// 	 ent_leaves = FcCharsetLeaves (&ent.set);
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 		 ent_leaves[i] = FcPtrToOffset (ent_leaves,
// 						FcCharsetLeaf (fcs, i));
// 	 memcpy (FcCharsetNumbers (&ent.set),
// 		 FcCharsetNumbers (fcs),
// 		 len(fcs.numbers) * sizeof (FcChar16));
// 	 }
// 	 else
// 	 {
// 	 ent.set.leaves_offset = 0;
// 	 ent.set.numbers_offset = 0;
// 	 }

// 	 ent.hash = hash;
// 	 ent.next = *bucket;
// 	 *bucket = ent;

// 	 return &ent.set;
//  }

//  static const FcCharset *
//  FcCharsetFindFrozen (FcCharsetFreezer *freezer, const FcCharset *orig)
//  {
// 	 FcCharsetOrigEnt    **bucket = &freezer.orig_hash_table[((uintptr_t) orig) % FC_CHAR_SET_HASH_SIZE];
// 	 FcCharsetOrigEnt	*ent;

// 	 for (ent = *bucket; ent; ent = ent.next)
// 	 if (ent.orig == orig)
// 		 return ent.frozen;
// 	 return NULL;
//  }

//  const FcCharset *
//  FcCharsetFreeze (FcCharsetFreezer *freezer, fcs *FcCharset)
//  {
// 	 FcCharset	    *b;
// 	 const FcCharset *n = 0;
// 	 FcCharLeaf	    *l;
// 	 int		    i;

// 	 b = FcCharsetCreate ();
// 	 if (!b)
// 	 goto bail0;
// 	 for (i = 0; i < len(fcs.numbers); i++)
// 	 {
// 	 l = FcCharsetFreezeLeaf (freezer, FcCharsetLeaf(fcs, i));
// 	 if (!l)
// 		 goto bail1;
// 	 if (!FcCharsetInsertLeaf (b, fcs.numbers[i] << 8, l))
// 		 goto bail1;
// 	 }
// 	 n = FcCharsetFreezeBase (freezer, b);
// 	 if (!FcCharsetFreezeOrig (freezer, fcs, n))
// 	 {
// 	 n = NULL;
// 	 goto bail1;
// 	 }
// 	 freezer.charsets_seen++;
// 	 freezer.leaves_seen += len(fcs.numbers);
//  bail1:
// 	 if (b.num)
// 	 free (FcCharsetLeaves (b));
// 	 if (b.num)
// 	 free (FcCharsetNumbers (b));
// 	 free (b);
//  bail0:
// 	 return n;
//  }

//  FcCharsetFreezer *
//  FcCharsetFreezerCreate (void)
//  {
// 	 FcCharsetFreezer	*freezer;

// 	 freezer = calloc (1, sizeof (FcCharsetFreezer));
// 	 return freezer;
//  }

//  void
//  FcCharsetFreezerDestroy (FcCharsetFreezer *freezer)
//  {
// 	 int i;

// 	 if (FcDebug() & FC_DBG_CACHE)
// 	 {
// 	 printf ("\ncharsets %d . %d leaves %d . %d\n",
// 		 freezer.charsets_seen, freezer.charsets_allocated,
// 		 freezer.leaves_seen, freezer.leaves_allocated);
// 	 }
// 	 for (i = 0; i < FC_CHAR_SET_HASH_SIZE; i++)
// 	 {
// 	 FcCharsetEnt	*ent, *next;
// 	 for (ent = freezer.set_hash_table[i]; ent; ent = next)
// 	 {
// 		 next = ent.next;
// 		 free (ent);
// 	 }
// 	 }

// 	 for (i = 0; i < FC_CHAR_SET_HASH_SIZE; i++)
// 	 {
// 	 FcCharsetOrigEnt	*ent, *next;
// 	 for (ent = freezer.orig_hash_table[i]; ent; ent = next)
// 	 {
// 		 next = ent.next;
// 		 free (ent);
// 	 }
// 	 }

// 	 for (i = 0; i < freezer.leaf_block_count; i++)
// 	 free (freezer.leaf_blocks[i]);

// 	 free (freezer.leaf_blocks);
// 	 free (freezer);
//  }

//  FcBool
//  FcCharsetSerializeAlloc (FcSerialize *serialize, const FcCharset *cs)
//  {
// 	 intptr_t	    *leaves;
// 	 FcChar16	    *numbers;
// 	 int		    i;

// 	 if (!FcRefIsConst (&cs.ref))
// 	 {
// 	 if (!serialize.cs_freezer)
// 	 {
// 		 serialize.cs_freezer = FcCharsetFreezerCreate ();
// 		 if (!serialize.cs_freezer)
// 		 return false;
// 	 }
// 	 if (FcCharsetFindFrozen (serialize.cs_freezer, cs))
// 		 return true;

// 		 cs = FcCharsetFreeze (serialize.cs_freezer, cs);
// 	 }

// 	 leaves = FcCharsetLeaves (cs);
// 	 numbers = FcCharsetNumbers (cs);

// 	 if (!FcSerializeAlloc (serialize, cs, sizeof (FcCharset)))
// 	 return false;
// 	 if (!FcSerializeAlloc (serialize, leaves, cs.num * sizeof (intptr_t)))
// 	 return false;
// 	 if (!FcSerializeAlloc (serialize, numbers, cs.num * sizeof (FcChar16)))
// 	 return false;
// 	 for (i = 0; i < cs.num; i++)
// 	 if (!FcSerializeAlloc (serialize, FcCharsetLeaf(cs, i),
// 					sizeof (FcCharLeaf)))
// 		 return false;
// 	 return true;
//  }

//  FcCharset *
//  FcCharsetSerialize(FcSerialize *serialize, const FcCharset *cs)
//  {
// 	 FcCharset	*cs_serialized;
// 	 intptr_t	*leaves, *leaves_serialized;
// 	 FcChar16	*numbers, *numbers_serialized;
// 	 leaf *FcCharLeaf, *leaf_serialized;
// 	 int		i;

// 	 if (!FcRefIsConst (&cs.ref) && serialize.cs_freezer)
// 	 {
// 	 cs = FcCharsetFindFrozen (serialize.cs_freezer, cs);
// 	 if (!cs)
// 		 return NULL;
// 	 }

// 	 cs_serialized = FcSerializePtr (serialize, cs);
// 	 if (!cs_serialized)
// 	 return NULL;

// 	 FcRefSetConst (&cs_serialized.ref);
// 	 cs_serialized.num = cs.num;

// 	 if (cs.num)
// 	 {
// 	 leaves = FcCharsetLeaves (cs);
// 	 leaves_serialized = FcSerializePtr (serialize, leaves);
// 	 if (!leaves_serialized)
// 		 return NULL;

// 	 cs_serialized.leaves_offset = FcPtrToOffset (cs_serialized,
// 							   leaves_serialized);

// 	 numbers = FcCharsetNumbers (cs);
// 	 numbers_serialized = FcSerializePtr (serialize, numbers);
// 	 if (!numbers)
// 		 return NULL;

// 	 cs_serialized.numbers_offset = FcPtrToOffset (cs_serialized,
// 								numbers_serialized);

// 	 for (i = 0; i < cs.num; i++)
// 	 {
// 		 leaf = FcCharsetLeaf (cs, i);
// 		 leaf_serialized = FcSerializePtr (serialize, leaf);
// 		 if (!leaf_serialized)
// 		 return NULL;
// 		 *leaf_serialized = *leaf;
// 		 leaves_serialized[i] = FcPtrToOffset (leaves_serialized,
// 						   leaf_serialized);
// 		 numbers_serialized[i] = numbers[i];
// 	 }
// 	 }
// 	 else
// 	 {
// 	 cs_serialized.leaves_offset = 0;
// 	 cs_serialized.numbers_offset = 0;
// 	 }

// 	 return cs_serialized;
//  }
