
//line use_machine.rl:1
package opentype 

// ported from harfbuzz/src/hb-ot-shape-complex-use-machine.rl Copyright © 2015 Mozilla Foundation. Google, Inc. Jonathan Kew Behdad Esfahbod
 
 
// /* buffer var allocations */
// #define use_category() complex_var_u8_category()

// #define USE(Cat) use_syllable_machine_ex_##Cat

// enum use_syllable_type_t {
    const (
  use_independent_cluster = iota
  use_virama_terminated_cluster 
  use_sakot_terminated_cluster 
  use_standard_cluster 
  use_number_joiner_terminated_cluster 
  use_numeral_cluster 
  use_symbol_cluster 
  use_hieroglyph_cluster 
  use_broken_cluster 
  use_non_cluster 
    )


//line use_machine.go:29
const use_syllable_machine_ex_B = 1
const use_syllable_machine_ex_CMAbv = 31
const use_syllable_machine_ex_CMBlw = 32
const use_syllable_machine_ex_CS = 43
const use_syllable_machine_ex_FAbv = 24
const use_syllable_machine_ex_FBlw = 25
const use_syllable_machine_ex_FMAbv = 45
const use_syllable_machine_ex_FMBlw = 46
const use_syllable_machine_ex_FMPst = 47
const use_syllable_machine_ex_FPst = 26
const use_syllable_machine_ex_G = 49
const use_syllable_machine_ex_GB = 5
const use_syllable_machine_ex_H = 12
const use_syllable_machine_ex_HN = 13
const use_syllable_machine_ex_HVM = 44
const use_syllable_machine_ex_J = 50
const use_syllable_machine_ex_MAbv = 27
const use_syllable_machine_ex_MBlw = 28
const use_syllable_machine_ex_MPre = 30
const use_syllable_machine_ex_MPst = 29
const use_syllable_machine_ex_N = 4
const use_syllable_machine_ex_O = 0
const use_syllable_machine_ex_R = 18
const use_syllable_machine_ex_S = 19
const use_syllable_machine_ex_SB = 51
const use_syllable_machine_ex_SE = 52
const use_syllable_machine_ex_SMAbv = 41
const use_syllable_machine_ex_SMBlw = 42
const use_syllable_machine_ex_SUB = 11
const use_syllable_machine_ex_Sk = 48
const use_syllable_machine_ex_VAbv = 33
const use_syllable_machine_ex_VBlw = 34
const use_syllable_machine_ex_VMAbv = 37
const use_syllable_machine_ex_VMBlw = 38
const use_syllable_machine_ex_VMPre = 23
const use_syllable_machine_ex_VMPst = 39
const use_syllable_machine_ex_VPre = 22
const use_syllable_machine_ex_VPst = 35
const use_syllable_machine_ex_ZWNJ = 14


//line use_machine.go:71
var _use_syllable_machine_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3, 
	1, 4, 1, 5, 1, 6, 1, 7, 
	1, 8, 1, 9, 1, 10, 1, 11, 
	1, 12, 1, 13, 1, 14, 1, 15, 
	1, 16, 
}

var _use_syllable_machine_key_offsets []int16 = []int16{
	0, 1, 2, 38, 62, 86, 87, 103, 
	114, 120, 125, 129, 131, 132, 142, 151, 
	159, 160, 167, 182, 196, 209, 227, 244, 
	263, 286, 298, 299, 300, 326, 328, 329, 
	353, 369, 380, 386, 391, 395, 397, 398, 
	408, 417, 425, 432, 447, 461, 474, 492, 
	509, 528, 551, 563, 564, 565, 566, 595, 
	619, 621, 622, 624, 626, 629, 
}

var _use_syllable_machine_trans_keys []byte = []byte{
	1, 1, 0, 1, 4, 5, 11, 12, 
	13, 18, 19, 23, 24, 25, 26, 27, 
	28, 30, 31, 32, 33, 34, 35, 37, 
	38, 39, 41, 42, 43, 44, 45, 46, 
	47, 48, 49, 51, 22, 29, 11, 12, 
	23, 24, 25, 26, 27, 28, 30, 31, 
	32, 33, 34, 35, 37, 38, 39, 44, 
	45, 46, 47, 48, 22, 29, 11, 12, 
	23, 24, 25, 26, 27, 28, 30, 33, 
	34, 35, 37, 38, 39, 44, 45, 46, 
	47, 48, 22, 29, 31, 32, 1, 22, 
	23, 24, 25, 26, 33, 34, 35, 37, 
	38, 39, 44, 45, 46, 47, 48, 23, 
	24, 25, 26, 37, 38, 39, 45, 46, 
	47, 48, 24, 25, 26, 45, 46, 47, 
	25, 26, 45, 46, 47, 26, 45, 46, 
	47, 45, 46, 46, 24, 25, 26, 37, 
	38, 39, 45, 46, 47, 48, 24, 25, 
	26, 38, 39, 45, 46, 47, 48, 24, 
	25, 26, 39, 45, 46, 47, 48, 1, 
	24, 25, 26, 45, 46, 47, 48, 23, 
	24, 25, 26, 33, 34, 35, 37, 38, 
	39, 44, 45, 46, 47, 48, 23, 24, 
	25, 26, 34, 35, 37, 38, 39, 44, 
	45, 46, 47, 48, 23, 24, 25, 26, 
	35, 37, 38, 39, 44, 45, 46, 47, 
	48, 22, 23, 24, 25, 26, 28, 29, 
	33, 34, 35, 37, 38, 39, 44, 45, 
	46, 47, 48, 22, 23, 24, 25, 26, 
	29, 33, 34, 35, 37, 38, 39, 44, 
	45, 46, 47, 48, 23, 24, 25, 26, 
	27, 28, 33, 34, 35, 37, 38, 39, 
	44, 45, 46, 47, 48, 22, 29, 11, 
	12, 23, 24, 25, 26, 27, 28, 30, 
	32, 33, 34, 35, 37, 38, 39, 44, 
	45, 46, 47, 48, 22, 29, 1, 23, 
	24, 25, 26, 37, 38, 39, 45, 46, 
	47, 48, 13, 4, 11, 12, 23, 24, 
	25, 26, 27, 28, 30, 31, 32, 33, 
	34, 35, 37, 38, 39, 41, 42, 44, 
	45, 46, 47, 48, 22, 29, 41, 42, 
	42, 11, 12, 23, 24, 25, 26, 27, 
	28, 30, 33, 34, 35, 37, 38, 39, 
	44, 45, 46, 47, 48, 22, 29, 31, 
	32, 22, 23, 24, 25, 26, 33, 34, 
	35, 37, 38, 39, 44, 45, 46, 47, 
	48, 23, 24, 25, 26, 37, 38, 39, 
	45, 46, 47, 48, 24, 25, 26, 45, 
	46, 47, 25, 26, 45, 46, 47, 26, 
	45, 46, 47, 45, 46, 46, 24, 25, 
	26, 37, 38, 39, 45, 46, 47, 48, 
	24, 25, 26, 38, 39, 45, 46, 47, 
	48, 24, 25, 26, 39, 45, 46, 47, 
	48, 24, 25, 26, 45, 46, 47, 48, 
	23, 24, 25, 26, 33, 34, 35, 37, 
	38, 39, 44, 45, 46, 47, 48, 23, 
	24, 25, 26, 34, 35, 37, 38, 39, 
	44, 45, 46, 47, 48, 23, 24, 25, 
	26, 35, 37, 38, 39, 44, 45, 46, 
	47, 48, 22, 23, 24, 25, 26, 28, 
	29, 33, 34, 35, 37, 38, 39, 44, 
	45, 46, 47, 48, 22, 23, 24, 25, 
	26, 29, 33, 34, 35, 37, 38, 39, 
	44, 45, 46, 47, 48, 23, 24, 25, 
	26, 27, 28, 33, 34, 35, 37, 38, 
	39, 44, 45, 46, 47, 48, 22, 29, 
	11, 12, 23, 24, 25, 26, 27, 28, 
	30, 32, 33, 34, 35, 37, 38, 39, 
	44, 45, 46, 47, 48, 22, 29, 1, 
	23, 24, 25, 26, 37, 38, 39, 45, 
	46, 47, 48, 1, 4, 13, 1, 5, 
	11, 12, 13, 23, 24, 25, 26, 27, 
	28, 30, 31, 32, 33, 34, 35, 37, 
	38, 39, 41, 42, 44, 45, 46, 47, 
	48, 22, 29, 11, 12, 23, 24, 25, 
	26, 27, 28, 30, 31, 32, 33, 34, 
	35, 37, 38, 39, 44, 45, 46, 47, 
	48, 22, 29, 41, 42, 42, 1, 5, 
	50, 52, 49, 50, 52, 49, 51, 
}

var _use_syllable_machine_single_lengths []byte = []byte{
	1, 1, 34, 22, 20, 1, 16, 11, 
	6, 5, 4, 2, 1, 10, 9, 8, 
	1, 7, 15, 14, 13, 18, 17, 17, 
	21, 12, 1, 1, 24, 2, 1, 20, 
	16, 11, 6, 5, 4, 2, 1, 10, 
	9, 8, 7, 15, 14, 13, 18, 17, 
	17, 21, 12, 1, 1, 1, 27, 22, 
	2, 1, 2, 2, 3, 2, 
}

var _use_syllable_machine_range_lengths []byte = []byte{
	0, 0, 1, 1, 2, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 1, 
	1, 0, 0, 0, 1, 0, 0, 2, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	1, 1, 0, 0, 0, 0, 1, 1, 
	0, 0, 0, 0, 0, 0, 
}

var _use_syllable_machine_index_offsets []int16 = []int16{
	0, 2, 4, 40, 64, 87, 89, 106, 
	118, 125, 131, 136, 139, 141, 152, 162, 
	171, 173, 181, 197, 212, 226, 245, 263, 
	282, 305, 318, 320, 322, 348, 351, 353, 
	376, 393, 405, 412, 418, 423, 426, 428, 
	439, 449, 458, 466, 482, 497, 511, 530, 
	548, 567, 590, 603, 605, 607, 609, 638, 
	662, 665, 667, 670, 673, 677, 
}

var _use_syllable_machine_indicies []byte = []byte{
	1, 0, 2, 0, 3, 4, 6, 7, 
	1, 8, 9, 10, 11, 13, 14, 15, 
	16, 17, 18, 19, 20, 21, 22, 23, 
	24, 25, 26, 27, 28, 29, 30, 31, 
	32, 33, 34, 8, 35, 36, 12, 5, 
	38, 39, 41, 42, 43, 44, 45, 46, 
	47, 4, 48, 49, 50, 51, 52, 53, 
	54, 55, 56, 57, 58, 39, 40, 37, 
	38, 39, 41, 42, 43, 44, 45, 46, 
	47, 49, 50, 51, 52, 53, 54, 55, 
	56, 57, 58, 39, 40, 48, 37, 38, 
	59, 40, 41, 42, 43, 44, 49, 50, 
	51, 52, 53, 54, 41, 56, 57, 58, 
	60, 37, 41, 42, 43, 44, 52, 53, 
	54, 56, 57, 58, 60, 37, 42, 43, 
	44, 56, 57, 58, 37, 43, 44, 56, 
	57, 58, 37, 44, 56, 57, 58, 37, 
	56, 57, 37, 57, 37, 42, 43, 44, 
	52, 53, 54, 56, 57, 58, 60, 37, 
	42, 43, 44, 53, 54, 56, 57, 58, 
	60, 37, 42, 43, 44, 54, 56, 57, 
	58, 60, 37, 62, 61, 42, 43, 44, 
	56, 57, 58, 60, 37, 41, 42, 43, 
	44, 49, 50, 51, 52, 53, 54, 41, 
	56, 57, 58, 60, 37, 41, 42, 43, 
	44, 50, 51, 52, 53, 54, 41, 56, 
	57, 58, 60, 37, 41, 42, 43, 44, 
	51, 52, 53, 54, 41, 56, 57, 58, 
	60, 37, 40, 41, 42, 43, 44, 46, 
	40, 49, 50, 51, 52, 53, 54, 41, 
	56, 57, 58, 60, 37, 40, 41, 42, 
	43, 44, 40, 49, 50, 51, 52, 53, 
	54, 41, 56, 57, 58, 60, 37, 41, 
	42, 43, 44, 45, 46, 49, 50, 51, 
	52, 53, 54, 41, 56, 57, 58, 60, 
	40, 37, 38, 39, 41, 42, 43, 44, 
	45, 46, 47, 48, 49, 50, 51, 52, 
	53, 54, 55, 56, 57, 58, 39, 40, 
	37, 38, 41, 42, 43, 44, 52, 53, 
	54, 56, 57, 58, 60, 59, 64, 63, 
	6, 65, 38, 39, 41, 42, 43, 44, 
	45, 46, 47, 4, 48, 49, 50, 51, 
	52, 53, 54, 11, 66, 55, 56, 57, 
	58, 39, 40, 37, 11, 66, 67, 66, 
	67, 1, 69, 13, 14, 15, 16, 17, 
	18, 19, 22, 23, 24, 25, 26, 27, 
	31, 32, 33, 34, 69, 12, 21, 68, 
	12, 13, 14, 15, 16, 22, 23, 24, 
	25, 26, 27, 13, 32, 33, 34, 70, 
	68, 13, 14, 15, 16, 25, 26, 27, 
	32, 33, 34, 70, 68, 14, 15, 16, 
	32, 33, 34, 68, 15, 16, 32, 33, 
	34, 68, 16, 32, 33, 34, 68, 32, 
	33, 68, 33, 68, 14, 15, 16, 25, 
	26, 27, 32, 33, 34, 70, 68, 14, 
	15, 16, 26, 27, 32, 33, 34, 70, 
	68, 14, 15, 16, 27, 32, 33, 34, 
	70, 68, 14, 15, 16, 32, 33, 34, 
	70, 68, 13, 14, 15, 16, 22, 23, 
	24, 25, 26, 27, 13, 32, 33, 34, 
	70, 68, 13, 14, 15, 16, 23, 24, 
	25, 26, 27, 13, 32, 33, 34, 70, 
	68, 13, 14, 15, 16, 24, 25, 26, 
	27, 13, 32, 33, 34, 70, 68, 12, 
	13, 14, 15, 16, 18, 12, 22, 23, 
	24, 25, 26, 27, 13, 32, 33, 34, 
	70, 68, 12, 13, 14, 15, 16, 12, 
	22, 23, 24, 25, 26, 27, 13, 32, 
	33, 34, 70, 68, 13, 14, 15, 16, 
	17, 18, 22, 23, 24, 25, 26, 27, 
	13, 32, 33, 34, 70, 12, 68, 1, 
	69, 13, 14, 15, 16, 17, 18, 19, 
	21, 22, 23, 24, 25, 26, 27, 31, 
	32, 33, 34, 69, 12, 68, 1, 13, 
	14, 15, 16, 25, 26, 27, 32, 33, 
	34, 70, 68, 1, 71, 72, 68, 9, 
	68, 4, 4, 1, 69, 9, 13, 14, 
	15, 16, 17, 18, 19, 20, 21, 22, 
	23, 24, 25, 26, 27, 28, 29, 31, 
	32, 33, 34, 69, 12, 68, 1, 69, 
	13, 14, 15, 16, 17, 18, 19, 20, 
	21, 22, 23, 24, 25, 26, 27, 31, 
	32, 33, 34, 69, 12, 68, 28, 29, 
	68, 29, 68, 4, 4, 71, 74, 35, 
	73, 35, 74, 74, 73, 35, 36, 73, 
	
}

var _use_syllable_machine_trans_targs []byte = []byte{
	2, 31, 42, 2, 3, 2, 26, 28, 
	51, 52, 54, 29, 32, 33, 34, 35, 
	36, 46, 47, 48, 55, 49, 43, 44, 
	45, 39, 40, 41, 56, 57, 58, 50, 
	37, 38, 2, 59, 61, 2, 4, 5, 
	6, 7, 8, 9, 10, 21, 22, 23, 
	24, 18, 19, 20, 13, 14, 15, 25, 
	11, 12, 2, 2, 16, 2, 17, 2, 
	27, 2, 30, 2, 2, 0, 1, 2, 
	53, 2, 60, 
}

var _use_syllable_machine_trans_actions []byte = []byte{
	33, 5, 5, 7, 0, 13, 0, 0, 
	0, 0, 5, 0, 5, 5, 0, 0, 
	0, 5, 5, 5, 5, 5, 5, 5, 
	5, 5, 5, 5, 0, 0, 0, 5, 
	0, 0, 11, 0, 0, 19, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 9, 15, 0, 17, 0, 23, 
	0, 21, 0, 25, 29, 0, 0, 31, 
	0, 27, 0, 
}

var _use_syllable_machine_to_state_actions []byte = []byte{
	0, 0, 1, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 
}

var _use_syllable_machine_from_state_actions []byte = []byte{
	0, 0, 3, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 
}

var _use_syllable_machine_eof_trans []int16 = []int16{
	1, 1, 0, 38, 38, 60, 38, 38, 
	38, 38, 38, 38, 38, 38, 38, 38, 
	62, 38, 38, 38, 38, 38, 38, 38, 
	38, 60, 64, 66, 38, 68, 68, 69, 
	69, 69, 69, 69, 69, 69, 69, 69, 
	69, 69, 69, 69, 69, 69, 69, 69, 
	69, 69, 69, 72, 69, 69, 69, 69, 
	69, 69, 72, 74, 74, 74, 
}

const use_syllable_machine_start int = 2
const use_syllable_machine_first_final int = 2
const use_syllable_machine_error int = -1

const use_syllable_machine_en_main int = 2


//line use_machine.rl:30



//line use_machine.rl:147


// #define found_syllable(syllable_type) \
//   HB_STMT_START { \
//     if (0) fprintf (stderr, "syllable %d..%d %s\n", (*ts).second.first, (*te).second.first, #syllable_type); \
//     for (unsigned i = (*ts).second.first; i < (*te).second.first; ++i) \
//       info[i].syllable() = (syllable_serial << 4) | syllable_type; \
//     syllable_serial++; \
//     if (unlikely (syllable_serial == 16)) syllable_serial = 1; \
//   } HB_STMT_END


// template <typename Iter>
// struct machine_index_t :
//   hb_iter_with_fallback_t<machine_index_t<Iter>,
// 			  typename Iter::item_t>
// {
//   machine_index_t (const Iter& it) : it (it) {}
//   machine_index_t (const machine_index_t& o) : it (o.it) {}

//   static constexpr bool is_random_access_iterator = Iter::is_random_access_iterator;
//   static constexpr bool is_sorted_iterator = Iter::is_sorted_iterator;

//   typename Iter::item_t __item__ () const { return *it; }
//   typename Iter::item_t __item_at__ (unsigned i) const { return it[i]; }
//   unsigned __len__ () const { return it.len (); }
//   void __next__ () { ++it; }
//   void __forward__ (unsigned n) { it += n; }
//   void __prev__ () { --it; }
//   void __rewind__ (unsigned n) { it -= n; }
//   void operator = (unsigned n)
//   { unsigned index = (*it).first; if (index < n) it += n - index; else if (index > n) it -= index - n; }
//   void operator = (const machine_index_t& o) { *this = (*o.it).first; }
//   bool operator == (const machine_index_t& o) const { return (*it).first == (*o.it).first; }
//   bool operator != (const machine_index_t& o) const { return !(*this == o); }

//   private:
//   Iter it;
// };
// struct
// {
//   template <typename Iter,
// 	    hb_requires (hb_is_iterable (Iter))>
//   machine_index_t<hb_iter_type<Iter>>
//   operator () (Iter&& it) const
//   { return machine_index_t<hb_iter_type<Iter>> (hb_iter (it)); }
// }
// HB_FUNCOBJ (machine_index);

// static bool
// not_standard_default_ignorable (const hb_glyph_info_t &i)
// { return !(i.use_category() == USE(O) && _hb_glyph_info_is_default_ignorable (&i)); }

func find_syllables_use (buffer *cm.Buffer) {
//   hb_glyph_info_t *info = buffer.info;
//   auto p =
//     + hb_iter (info, buffer.len)
//     | hb_enumerate
//     | hb_filter ([] (const hb_glyph_info_t &i) { return not_standard_default_ignorable (i); },
// 		 hb_second)
//     | hb_filter ([&] (const hb_pair_t<unsigned, const hb_glyph_info_t &> p)
// 		 {
// 		   if (p.second.use_category() == USE(ZWNJ))
// 		     for (unsigned i = p.first + 1; i < buffer.len; ++i)
// 		       if (not_standard_default_ignorable (info[i]))
// 			 return !_hb_glyph_info_is_unicode_mark (&info[i]);
// 		   return true;
// 		 })
//     | hb_enumerate
//     | machine_index
//     ;
//   auto pe = p + p.len ();
//   auto eof = +pe;
//   auto ts = +p;
//   auto te = +p;
//   unsigned int act HB_UNUSED;
    p := 0
    pe := 20
  var cs, act int
  
//line use_machine.go:446
	{
	cs = use_syllable_machine_start
	ts = 0
	te = 0
	act = 0
	}

//line use_machine.rl:229


     var syllable_serial uint  = 1;
  
//line use_machine.go:459
	{
	var _klen int
	var _trans int
	var _acts int
	var _nacts uint
	var _keys int
	if p == pe {
		goto _test_eof
	}
_resume:
	_acts = int(_use_syllable_machine_from_state_actions[cs])
	_nacts = uint(_use_syllable_machine_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		 _acts++
		switch _use_syllable_machine_actions[_acts - 1] {
		case 1:
//line NONE:1
ts = p

//line use_machine.go:479
		}
	}

	_keys = int(_use_syllable_machine_key_offsets[cs])
	_trans = int(_use_syllable_machine_index_offsets[cs])

	_klen = int(_use_syllable_machine_single_lengths[cs])
	if _klen > 0 {
		_lower := int(_keys)
		var _mid int
		_upper := int(_keys + _klen - 1)
		for {
			if _upper < _lower {
				break
			}

			_mid = _lower + ((_upper - _lower) >> 1)
			switch {
			case ( (data[p]).second.second.use_category()) < _use_syllable_machine_trans_keys[_mid]:
				_upper = _mid - 1
			case ( (data[p]).second.second.use_category()) > _use_syllable_machine_trans_keys[_mid]:
				_lower = _mid + 1
			default:
				_trans += int(_mid - int(_keys))
				goto _match
			}
		}
		_keys += _klen
		_trans += _klen
	}

	_klen = int(_use_syllable_machine_range_lengths[cs])
	if _klen > 0 {
		_lower := int(_keys)
		var _mid int
		_upper := int(_keys + (_klen << 1) - 2)
		for {
			if _upper < _lower {
				break
			}

			_mid = _lower + (((_upper - _lower) >> 1) & ^1)
			switch {
			case ( (data[p]).second.second.use_category()) < _use_syllable_machine_trans_keys[_mid]:
				_upper = _mid - 2
			case ( (data[p]).second.second.use_category()) > _use_syllable_machine_trans_keys[_mid + 1]:
				_lower = _mid + 2
			default:
				_trans += int((_mid - int(_keys)) >> 1)
				goto _match
			}
		}
		_trans += _klen
	}

_match:
	_trans = int(_use_syllable_machine_indicies[_trans])
_eof_trans:
	cs = int(_use_syllable_machine_trans_targs[_trans])

	if _use_syllable_machine_trans_actions[_trans] == 0 {
		goto _again
	}

	_acts = int(_use_syllable_machine_trans_actions[_trans])
	_nacts = uint(_use_syllable_machine_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		_acts++
		switch _use_syllable_machine_actions[_acts-1] {
		case 2:
//line NONE:1
te = p+1

		case 3:
//line use_machine.rl:134
te = p+1
{ found_syllable (use_independent_cluster); }
		case 4:
//line use_machine.rl:137
te = p+1
{ found_syllable (use_standard_cluster); }
		case 5:
//line use_machine.rl:142
te = p+1
{ found_syllable (use_broken_cluster); }
		case 6:
//line use_machine.rl:143
te = p+1
{ found_syllable (use_non_cluster); }
		case 7:
//line use_machine.rl:135
te = p
p--
{ found_syllable (use_virama_terminated_cluster); }
		case 8:
//line use_machine.rl:136
te = p
p--
{ found_syllable (use_sakot_terminated_cluster); }
		case 9:
//line use_machine.rl:137
te = p
p--
{ found_syllable (use_standard_cluster); }
		case 10:
//line use_machine.rl:138
te = p
p--
{ found_syllable (use_number_joiner_terminated_cluster); }
		case 11:
//line use_machine.rl:139
te = p
p--
{ found_syllable (use_numeral_cluster); }
		case 12:
//line use_machine.rl:140
te = p
p--
{ found_syllable (use_symbol_cluster); }
		case 13:
//line use_machine.rl:141
te = p
p--
{ found_syllable (use_hieroglyph_cluster); }
		case 14:
//line use_machine.rl:142
te = p
p--
{ found_syllable (use_broken_cluster); }
		case 15:
//line use_machine.rl:143
te = p
p--
{ found_syllable (use_non_cluster); }
		case 16:
//line use_machine.rl:142
p = (te) - 1
{ found_syllable (use_broken_cluster); }
//line use_machine.go:618
		}
	}

_again:
	_acts = int(_use_syllable_machine_to_state_actions[cs])
	_nacts = uint(_use_syllable_machine_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		_acts++
		switch _use_syllable_machine_actions[_acts-1] {
		case 0:
//line NONE:1
ts = 0

//line use_machine.go:632
		}
	}

	p++
	if p != pe {
		goto _resume
	}
	_test_eof: {}
	if p == eof {
		if _use_syllable_machine_eof_trans[cs] > 0 {
			_trans = int(_use_syllable_machine_eof_trans[cs] - 1)
			goto _eof_trans
		}
	}

	}

//line use_machine.rl:234

}

 
