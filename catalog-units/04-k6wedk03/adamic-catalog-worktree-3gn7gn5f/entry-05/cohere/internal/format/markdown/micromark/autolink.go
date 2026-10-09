package micromark

// autolink is micromark-core-commonmark/lib/autolink.js: the CommonMark `<https://example.com>` and
// `<user@example.com>` autolinks. The GFM literal autolinks are micromark-extension-gfm-autolink-literal's.
var autolink = &Construct{Name: "autolink", Tokenize: tokenizeAutolink}

func tokenizeAutolink(self *Self, effects *Effects, ok State, nok State) State {
	size := 0

	var start, open, schemeOrEmailAtext, schemeInsideOrEmailAtext, urlInside, emailAtext, emailAtSignOrDot,
		emailLabel, emailValue State

	// Start of an autolink.
	//
	//	> | a<https://example.com>b
	//	     ^
	//	> | a<user@example.com>b
	//	     ^
	start = func(code Code) State {
		effects.Enter(TypeAutolink, nil)
		effects.Enter(TypeAutolinkMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeAutolinkMarker)
		effects.Enter(TypeAutolinkProtocol, nil)
		return open
	}

	// After `<`, at protocol or atext.
	//
	//	> | a<https://example.com>b
	//	      ^
	//	> | a<user@example.com>b
	//	      ^
	open = func(code Code) State {
		if asciiAlpha(code) {
			effects.Consume(code)
			return schemeOrEmailAtext
		}
		if code == CodeAtSign {
			return nok(code)
		}
		return emailAtext(code)
	}

	// At second byte of protocol or atext.
	//
	//	> | a<https://example.com>b
	//	       ^
	//	> | a<user@example.com>b
	//	       ^
	schemeOrEmailAtext = func(code Code) State {
		// ASCII alphanumeric and `+`, `-`, and `.`.
		if code == CodePlusSign || code == CodeDash || code == CodeDot || asciiAlphanumeric(code) {
			// Count the previous alphabetical from `open` too.
			size = 1
			return schemeInsideOrEmailAtext(code)
		}
		return emailAtext(code)
	}

	// In ambiguous protocol or atext.
	//
	//	> | a<https://example.com>b
	//	       ^
	//	> | a<user@example.com>b
	//	       ^
	schemeInsideOrEmailAtext = func(code Code) State {
		if code == CodeColon {
			effects.Consume(code)
			size = 0
			return urlInside
		}

		// ASCII alphanumeric and `+`, `-`, and `.`. The `size++ < autolinkSchemeSizeMax` only runs, and so
		// only increments, when the code is one of them: `&&` short-circuits.
		if code == CodePlusSign || code == CodeDash || code == CodeDot || asciiAlphanumeric(code) {
			below := size < autolinkSchemeSizeMax
			size++
			if below {
				effects.Consume(code)
				return schemeInsideOrEmailAtext
			}
		}
		size = 0
		return emailAtext(code)
	}

	// After protocol, in URL.
	//
	//	> | a<https://example.com>b
	//	            ^
	urlInside = func(code Code) State {
		if code == CodeGreaterThan {
			effects.Exit(TypeAutolinkProtocol)
			effects.Enter(TypeAutolinkMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeAutolinkMarker)
			effects.Exit(TypeAutolink)
			return ok
		}

		// ASCII control, space, or `<`.
		if code == CodeEof || code == CodeSpace || code == CodeLessThan || asciiControl(code) {
			return nok(code)
		}
		effects.Consume(code)
		return urlInside
	}

	// In email atext.
	//
	//	> | a<user.name@example.com>b
	//	             ^
	emailAtext = func(code Code) State {
		if code == CodeAtSign {
			effects.Consume(code)
			return emailAtSignOrDot
		}
		if asciiAtext(code) {
			effects.Consume(code)
			return emailAtext
		}
		return nok(code)
	}

	// In label, after at-sign or dot.
	//
	//	> | a<user.name@example.com>b
	//	                ^       ^
	emailAtSignOrDot = func(code Code) State {
		if asciiAlphanumeric(code) {
			return emailLabel(code)
		}
		return nok(code)
	}

	// In label, where `.` and `>` are allowed.
	//
	//	> | a<user.name@example.com>b
	//	                  ^
	emailLabel = func(code Code) State {
		if code == CodeDot {
			effects.Consume(code)
			size = 0
			return emailAtSignOrDot
		}
		if code == CodeGreaterThan {
			// Exit, then change the token type.
			effects.Exit(TypeAutolinkProtocol).Type = TypeAutolinkEmail
			effects.Enter(TypeAutolinkMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeAutolinkMarker)
			effects.Exit(TypeAutolink)
			return ok
		}
		return emailValue(code)
	}

	// In label, where `.` and `>` are *not* allowed.
	//
	// Though, this is also used in `emailLabel` to parse other values.
	//
	//	> | a<user.name@ex-ample.com>b
	//	                   ^
	emailValue = func(code Code) State {
		// ASCII alphanumeric or `-`. As above, `size++ < autolinkDomainSizeMax` only runs, and increments,
		// when the code is one of them.
		if code == CodeDash || asciiAlphanumeric(code) {
			below := size < autolinkDomainSizeMax
			size++
			if below {
				next := emailLabel
				if code == CodeDash {
					next = emailValue
				}
				effects.Consume(code)
				return next
			}
		}
		return nok(code)
	}

	return start
}
