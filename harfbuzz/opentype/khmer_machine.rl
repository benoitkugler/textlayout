package opentype

// Code generated with ragel -Z -o khmer_machine.go khmer_machine.rl ; sed -i '/^\/\/line/ d' khmer_machine.go ; goimports -w khmer_machine.go  DO NOT EDIT.

// ported from harfbuzz/src/hb-ot-shape-complex-khmer-machine.rl Copyright © 2015 Google, Inc. Behdad Esfahbod


const (
  khmerConsonantSyllable = iota
  khmerBrokenCluster
  khmerNonKhmerCluster
)

%%{
  machine khmerSyllableMachine;
  alphtype byte;
  write exports;
  write data;
}%%

%%{

export C    = 1;
export V    = 2;
export ZWNJ = 5;
export ZWJ  = 6;
export PLACEHOLDER = 11;
export DOTTEDCIRCLE = 12;
export Coeng= 14;
export Ra   = 16;
export Robatic = 20;
export Xgroup  = 21;
export Ygroup  = 22;
export VAbv = 26;
export VBlw = 27;
export VPre = 28;
export VPst = 29;

c = (C | Ra | V);
cn = c.((ZWJ|ZWNJ)?.Robatic)?;
joiner = (ZWJ | ZWNJ);
xgroup = (joiner*.Xgroup)*;
ygroup = Ygroup*;

# This grammar was experimentally extracted from what Uniscribe allows.

matra_group = VPre? xgroup VBlw? xgroup (joiner?.VAbv)? xgroup VPst?;
syllable_tail = xgroup matra_group xgroup (Coeng.c)? ygroup;


broken_cluster =	(Coeng.cn)* (Coeng | syllable_tail);
consonant_syllable =	(cn|PLACEHOLDER|DOTTEDCIRCLE) broken_cluster;
other =			any;

main := |*
	consonant_syllable	=> { foundSyllableKhmer (khmerConsonantSyllable, ts, te, info, &syllableSerial); };
	broken_cluster		=> { foundSyllableKhmer (khmerBrokenCluster, ts, te, info, &syllableSerial); };
	other			=> { foundSyllableKhmer (khmerNonKhmerCluster, ts, te, info, &syllableSerial); };
*|;


}%%


func findSyllablesKhmer (buffer *cm.Buffer) {
    var p, ts, te, act, cs int 
    info := buffer.Info;
    %%{
        write init;
        getkey info[p].ComplexCategory;
    }%%

    pe := len(info)
    eof := pe

    var syllableSerial uint8 = 1;
    %%{
        write exec;
    }%%
    _ = act // needed by Ragel, but unused
}



