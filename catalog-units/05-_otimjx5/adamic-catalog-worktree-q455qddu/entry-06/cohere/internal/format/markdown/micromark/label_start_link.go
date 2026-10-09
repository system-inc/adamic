package micromark

// labelStartLink is micromark-core-commonmark/lib/label-start-link.js. Its resolveAll is labelEnd's,
// assigned in label_end.go's init so the three constructs share one resolver, as upstream's do.
var labelStartLink = &Construct{Name: "labelStartLink", Tokenize: tokenizeLabelStartLink}

func tokenizeLabelStartLink(self *Self, effects *Effects, ok State, nok State) State {
	var start, after State

	start = func(code Code) State {
		effects.Enter(TypeLabelLink, nil)
		effects.Enter(TypeLabelMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeLabelMarker)
		effects.Exit(TypeLabelLink)
		return after
	}

	// Upstream refuses `^` here when `_hiddenFootnoteSupport` is in the constructs, which nothing in our
	// extension set puts there.
	after = func(code Code) State {
		return ok(code)
	}

	return start
}
