package fribidi

import (
	"fmt"
	"reflect"
	"testing"
)

// charset CapRTL used only for testing purposes
var capRTLCharTypes = [...]CharType{
	ON, ON, ON, ON, LTR, RTL, ON, ON, ON, ON, ON, ON, ON, BS, RLO, RLE, /* 00-0f */
	LRO, LRE, PDF, WS, LRI, RLI, FSI, PDI, ON, ON, ON, ON, ON, ON, ON, ON, /* 10-1f */
	WS, ON, ON, ON, ET, ON, ON, ON, ON, ON, ON, ET, CS, ON, ES, ES, /* 20-2f */
	EN, EN, EN, EN, EN, EN, AN, AN, AN, AN, CS, ON, ON, ON, ON, ON, /* 30-3f */
	RTL, AL, AL, AL, AL, AL, AL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, /* 40-4f */
	RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, ON, BS, ON, BN, ON, /* 50-5f */
	NSM, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, /* 60-6f */
	LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, ON, SS, ON, WS, ON, /* 70-7f */
}

var caprtl_to_unicode []rune

/* We do not support surrogates yet */
const FRIBIDI_UNICODE_CHARS = 0x110000

const numTypes = 23

func init() {
	caprtl_to_unicode = make([]rune, len(capRTLCharTypes))

	var (
		mark             [len(capRTLCharTypes)]bool
		num_types, count int
		to_type          [numTypes]CharType
		request          [numTypes]int
	)

	for i, ct := range capRTLCharTypes {
		if ct == GetBidiType(rune(i)) {
			caprtl_to_unicode[i] = rune(i)
			mark[i] = true
		} else {

			caprtl_to_unicode[i] = FRIBIDI_UNICODE_CHARS
			mark[i] = false
			if _, ok := getMirrorChar(rune(i)); ok {
				fmt.Println("warning: I could not map mirroring character map to itself in CapRTL")
			}

			var j int
			for j = 0; j < num_types; j++ {
				if to_type[j] == ct {
					break
				}
			}
			if j == num_types {
				num_types++
				to_type[j] = ct
				request[j] = 0
			}
			request[j]++
			count++
		}
	}
	for i := 0; i < 0x10000 && count != 0; i++ { /* Assign BMP chars to CapRTL entries */
		if _, ok := getMirrorChar(rune(i)); !ok && !(i < len(capRTLCharTypes) && mark[i]) {
			var j, k int
			t := GetBidiType(rune(i))
			for j = 0; j < num_types; j++ {
				if to_type[j] == t {
					break
				}
			}
			if j >= num_types || request[j] == 0 { /* Do not need this type */
				continue
			}
			for k = 0; k < len(capRTLCharTypes); k++ {
				if caprtl_to_unicode[k] == FRIBIDI_UNICODE_CHARS && to_type[j] == capRTLCharTypes[k] {
					request[j]--
					count--
					caprtl_to_unicode[k] = rune(i)
					break
				}
			}
		}
	}
	if count != 0 {
		var j int

		fmt.Println("warning: could not find a mapping for CapRTL to Unicode:")
		for j = 0; j < num_types; j++ {
			if request[j] != 0 {
				fmt.Printf("  need this type: %d\n", to_type[j])
			}
		}
	}
}

type capRTLCharset struct{}

// Decode
func (capRTLCharset) decode(s []byte) []rune {
	var us []rune
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '_' {
			i++
			switch s[i] {
			case '>':
				us = append(us, charLRM)
			case '<':
				us = append(us, charRLM)
			case 'l':
				us = append(us, charLRE)
			case 'r':
				us = append(us, charRLE)
			case 'o':
				us = append(us, charPDF)
			case 'L':
				us = append(us, charLRO)
			case 'R':
				us = append(us, charRLO)
			case 'i':
				us = append(us, charLRI)
			case 'y':
				us = append(us, charRLI)
			case 'f':
				us = append(us, charFSI)
			case 'I':
				us = append(us, charPDI)
			case '_':
				us = append(us, '_')
			default:
				us = append(us, '_')
				i--
			}
		} else {
			us = append(us, caprtl_to_unicode[s[i]])
		}
	}
	return us
}

func fribidi_unicode_to_cap_rtl_c(uch rune) byte {
	for i := 0; i < len(capRTLCharTypes); i++ {
		if uch == caprtl_to_unicode[i] {
			return byte(i)
		}
	}
	return '?'
}

func (capRTLCharset) encode(str []rune) []byte {
	var s []byte
	for _, ch := range str {
		if bd := GetBidiType(ch); !bd.isExplicit() && !bd.IsIsolate() &&
			ch != '_' && ch != charLRM && ch != charRLM {
			s = append(s, fribidi_unicode_to_cap_rtl_c(ch))
		} else {
			s = append(s, '_')
			switch ch {
			case charLRM:
				s = append(s, '>')
			case charRLM:
				s = append(s, '<')
			case charLRE:
				s = append(s, 'l')
			case charRLE:
				s = append(s, 'r')
			case charPDF:
				s = append(s, 'o')
			case charLRO:
				s = append(s, 'L')
			case charRLO:
				s = append(s, 'R')
			case charLRI:
				s = append(s, 'i')
			case charRLI:
				s = append(s, 'y')
			case charFSI:
				s = append(s, 'f')
			case charPDI:
				s = append(s, 'I')
			case '_':
				s = append(s, '_')
			default:
				if ch < 256 {
					s[len(s)-1] = fribidi_unicode_to_cap_rtl_c(ch)
				} else {
					s[len(s)-1] = '?'
				}
			}
		}
	}
	return s
}

func TestCharsetCapRTL(t *testing.T) {
	cs := capRTLCharset{}
	in := []rune("simple english words")
	b := cs.encode(in)
	runes := cs.decode(b)
	if !reflect.DeepEqual(runes, in) {
		t.Errorf("expected %v, got %v", in, runes)
	}
}
