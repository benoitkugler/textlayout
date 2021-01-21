package truetype

import (
	"encoding/binary"
	"io"

	"github.com/benoitkugler/textlayout/fonts"
)

type otfHeader struct {
	ScalerType    Tag
	NumTables     uint16
	SearchRange   uint16
	EntrySelector uint16
	RangeShift    uint16
}

const otfHeaderLength = 12
const directoryEntryLength = 16

func (header *otfHeader) checkSum() uint32 {
	return uint32(header.ScalerType) +
		(uint32(header.NumTables)<<16 | uint32(header.SearchRange)) +
		(uint32(header.EntrySelector)<<16 + uint32(header.RangeShift))
}

// An Entry in an OpenType table.
type directoryEntry struct {
	Tag      Tag
	CheckSum uint32
	Offset   uint32
	Length   uint32
}

func (entry *directoryEntry) checkSum() uint32 {
	return uint32(entry.Tag) + entry.CheckSum + entry.Offset + entry.Length
}

func readOTFHeader(r io.Reader, header *otfHeader) error {
	var buf [12]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return err
	}

	header.ScalerType = newTag(buf[0:4])
	header.NumTables = binary.BigEndian.Uint16(buf[4:6])
	header.SearchRange = binary.BigEndian.Uint16(buf[6:8])
	header.EntrySelector = binary.BigEndian.Uint16(buf[8:10])
	header.RangeShift = binary.BigEndian.Uint16(buf[10:12])

	return nil
}

func readDirectoryEntry(r io.Reader, entry *directoryEntry) error {
	var buf [16]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return err
	}

	entry.Tag = newTag(buf[0:4])
	entry.CheckSum = binary.BigEndian.Uint32(buf[4:8])
	entry.Offset = binary.BigEndian.Uint32(buf[8:12])
	entry.Length = binary.BigEndian.Uint32(buf[12:16])

	return nil
}

// parseOTF reads an OpenTyp (.otf) or TrueType (.ttf) file and returns a Font.
// If parsing fails, then an error is returned and Font will be nil.
func parseOTF(file fonts.Ressource) (*Font, error) {
	var header otfHeader
	if err := readOTFHeader(file, &header); err != nil {
		return nil, err
	}

	font := &Font{
		file: file,

		Type:   header.ScalerType,
		tables: make(map[Tag]*tableSection, header.NumTables),
	}

	for i := 0; i < int(header.NumTables); i++ {
		var entry directoryEntry
		if err := readDirectoryEntry(file, &entry); err != nil {
			return nil, err
		}

		// TODO Check the checksum.

		if _, found := font.tables[entry.Tag]; found {
			// ignore duplicate tables – the first one wins
			continue
		}

		font.tables[entry.Tag] = &tableSection{
			offset: entry.Offset,
			length: entry.Length,
		}
	}

	if _, ok := font.tables[tagHead]; !ok {
		return nil, errMissingHead
	}

	return font, nil
}
