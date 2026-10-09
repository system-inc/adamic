package cst

// JavaScript's string methods over UTF-16 code units, with their clamping rules, so the lexer can call
// them where upstream does without each call site re-deriving what happens out of range.

// characterAt is `text[index]`: the code unit, or -1 for undefined past either end.
func characterAt(text []uint16, index int) int {
	if index < 0 || index >= len(text) {
		return -1
	}
	return int(text[index])
}

// indexOf is `text.indexOf(character, from)`. A from past the end finds nothing; a negative from is 0.
func indexOf(text []uint16, character uint16, from int) int {
	if from < 0 {
		from = 0
	}
	for index := from; index < len(text); index++ {
		if text[index] == character {
			return index
		}
	}
	return -1
}

// substring is `text.substring(start, end)`: both clamped to [0, length], swapped when start > end.
func substring(text []uint16, start int, end int) []uint16 {
	start = clamp(start, 0, len(text))
	end = clamp(end, 0, len(text))
	if start > end {
		start, end = end, start
	}
	return text[start:end:end]
}

// substr is `text.substr(start, length)`: a negative start counts from the end, the length is clamped.
func substr(text []uint16, start int, length int) []uint16 {
	if start < 0 {
		start = max(len(text)+start, 0)
	}
	start = min(start, len(text))
	end := clamp(start+length, start, len(text))
	return text[start:end:end]
}

// slice is `text.slice(start, end)`: negatives count from the end, an end before the start is empty.
func slice(text []uint16, start int, end int) []uint16 {
	if start < 0 {
		start = max(len(text)+start, 0)
	}
	if end < 0 {
		end = max(len(text)+end, 0)
	}
	start = min(start, len(text))
	end = min(end, len(text))
	if end < start {
		return text[start:start:start]
	}
	return text[start:end:end]
}

// startsWith is `text.startsWith(prefix)` for an ASCII prefix.
func startsWith(text []uint16, prefix string) bool {
	if len(text) < len(prefix) {
		return false
	}
	for index := 0; index < len(prefix); index++ {
		if text[index] != uint16(prefix[index]) {
			return false
		}
	}
	return true
}

// equals is `text === literal` for an ASCII literal.
func equals(text []uint16, literal string) bool {
	return len(text) == len(literal) && startsWith(text, literal)
}

func clamp(value int, low int, high int) int {
	return min(max(value, low), high)
}
