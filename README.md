# Text layout library for Golang

This module provides a chain of tools to layout text. It is mainly a port of the following C libraries : Pango, fribidi, fontconfig and harfbuzz.

## Overview

The package [fonts](fonts) provides the low level primitives to load and read font files. The selection of a font given some criterion (like language or style) is facilitated by the [fontconfig](fontconfig) package. Once a font is selected, [harfbuzz](harfbuzz) is responsible for laying out a line of text, that is transforming a sequence of unicode points (runes) to a sequence of positionned glyphs. [fribidi](fribidi) provides support for bidirectional text : it finds the alternates LTR and RTL sequences (embedding levels) in a paragraph. Finally, [pango](pango) wraps these tools to provide an higher level interface capable of laying out an entire text.

## Licensing

This module is a derivative work of severals libraries. As such, it is licensed under the less permissive license among the original implementations : the GNU Lesser GPL. The original licenses may be found in the reference repositories.
