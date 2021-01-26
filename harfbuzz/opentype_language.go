package harfbuzz

type LangTag struct {
	language string
	tag      hb_tag_t
}

func bfindLanguage(lang string) int {
	low, high := 0, len(ot_languages)
	for low <= high {
		mid := (low + high) / 2
		p := ot_languages[mid]

		if lang < p.language {
			high = mid - 1
		} else if lang > p.language {
			low = mid + 1
		} else {
			return mid
		}
	}
	return -1
}
