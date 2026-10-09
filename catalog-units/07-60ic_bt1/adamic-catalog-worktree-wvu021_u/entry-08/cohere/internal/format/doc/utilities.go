package doc

// InheritLabel is upstream's inheritLabel, document/utilities/index.js: apply fn to a doc's contents,
// keeping the label when there is one.
//
// It matters because printers branch on labels (the JavaScript printer's member chains are labelled
// "member-chain"), and wrapping a labelled doc in its comments without this would hide the label.
func InheritLabel(document Doc, fn func(Doc) Doc) Doc {
	if labelled, isLabel := document.(*Label); isLabel {
		return &Label{Label: labelled.Label, Contents: fn(labelled.Contents)}
	}
	return fn(document)
}
