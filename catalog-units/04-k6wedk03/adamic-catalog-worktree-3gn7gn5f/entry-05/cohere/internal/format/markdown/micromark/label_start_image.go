package micromark

// labelStartImage is micromark-core-commonmark/lib/label-start-image.js. Its resolveAll is labelEnd's,
// assigned in label_end.go's init.
var labelStartImage = &Construct{Name: "labelStartImage", Tokenize: tokenizeLabelStartImage}

func tokenizeLabelStartImage(self *Self, effects *Effects, ok State, nok State) State {
	var start, open, after State

	start = func(code Code) State {
		effects.Enter(TypeLabelImage, nil)
		effects.Enter(TypeLabelImageMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeLabelImageMarker)
		return open
	}

	open = func(code Code) State {
		if code == CodeLeftSquareBracket {
			effects.Enter(TypeLabelMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeLabelMarker)
			effects.Exit(TypeLabelImage)
			return after
		}

		return nok(code)
	}

	// Upstream refuses `^` here when `_hiddenFootnoteSupport` is in the constructs, which nothing in our
	// extension set puts there.
	after = func(code Code) State {
		return ok(code)
	}

	return start
}
