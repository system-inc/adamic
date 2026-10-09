package cst

// Ported from eemeli/yaml 2.9.0, dist/parse/line-counter.js.

// LineCounter tracks newlines during parsing in order to provide an efficient API for determining the
// one-indexed { line, col } position for any offset within the input. Offsets and columns are UTF-16
// code units.
type LineCounter struct {
	LineStarts []int
}

// NewLineCounter is upstream's `new LineCounter()`.
func NewLineCounter() *LineCounter {
	return &LineCounter{LineStarts: []int{}}
}

// AddNewLine should be called in ascending order. Otherwise, sort LineStarts before calling LinePos.
// Upstream's addNewLine is an arrow function bound to the instance; pass the method value.
func (lineCounter *LineCounter) AddNewLine(offset int) {
	lineCounter.LineStarts = append(lineCounter.LineStarts, offset)
}

// LinePos performs a binary search and returns the 1-indexed { line, col } position of offset. If
// line == 0, AddNewLine has never been called or offset is before the first known newline.
func (lineCounter *LineCounter) LinePos(offset int) (line int, col int) {
	low := 0
	high := len(lineCounter.LineStarts)
	for low < high {
		mid := (low + high) >> 1 // Math.floor((low + high) / 2)
		if lineCounter.LineStarts[mid] < offset {
			low = mid + 1
		} else {
			high = mid
		}
	}
	// `this.lineStarts[low] === offset`: past the end it is undefined, never equal.
	if low < len(lineCounter.LineStarts) && lineCounter.LineStarts[low] == offset {
		return low + 1, 1
	}
	if low == 0 {
		return 0, offset
	}
	start := lineCounter.LineStarts[low-1]
	return low, offset - start + 1
}
