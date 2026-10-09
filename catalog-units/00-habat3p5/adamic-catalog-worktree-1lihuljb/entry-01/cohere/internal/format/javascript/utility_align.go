package javascript

// document/builders/align.js: addAlignmentToDoc, the one builder there that is not already a doc
// constructor (align, dedent, dedentToRoot and markAsRoot live in builders.go).

// addAlignmentToDoc is upstream's addAlignmentToDoc.
func addAlignmentToDoc(document Doc, size int, tabWidth int) Doc {
	aligned := document
	if size > 0 {
		// Use indent to add tabs for all the levels of tabs we need
		for level := 0; level < size/tabWidth; level++ {
			aligned = indent(aligned)
		}
		// Use align for all the spaces that are needed
		aligned = align(size%tabWidth, aligned)
		// size is absolute from 0 and not relative to the current
		// indentation, so we use -Infinity to reset the indentation to 0
		aligned = dedentToRoot(aligned)
	}
	return aligned
}
