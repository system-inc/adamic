package compose

// Ported from eemeli/yaml 2.9.0, dist/compose/resolve-props.js and resolve-end.js.

import "github.com/system-inc/cohere/internal/format/yaml/cst"

type resolvePropsOptions struct {
	// flow is "flow map" or "flow sequence" inside a flow collection, "" (undefined) elsewhere.
	flow      string
	indicator string
	// next is the token after the props, nil for undefined.
	next           *cst.Token
	offset         int
	onError        onErrorFunc
	parentIndent   int
	startOnNewline bool
}

// props is resolveProps's result. A token field is nil for upstream's null.
type props struct {
	comma            *cst.Token
	found            *cst.Token
	spaceBefore      bool
	comment          string
	hasNewline       bool
	anchor           *cst.Token
	tag              *cst.Token
	newlineAfterProp *cst.Token
	end              int
	start            int
}

func resolveProps(tokens []*cst.Token, options resolvePropsOptions) props {
	flow, indicator, next, onError := options.flow, options.indicator, options.next, options.onError
	spaceBefore := false
	atNewline := options.startOnNewline
	hasSpace := options.startOnNewline
	comment := ""
	commentSep := ""
	hasNewline := false
	reqSpace := false
	var tab, anchor, tag, newlineAfterProp, comma, found *cst.Token
	start, startSet := 0, false
	for _, token := range tokens {
		if reqSpace {
			if token.Type != "space" && token.Type != "newline" && token.Type != "comma" {
				onError(token.Offset, "MISSING_CHAR",
					"Tags and anchors must be separated from the next token by white space", false)
			}
			reqSpace = false
		}
		if tab != nil {
			if atNewline && token.Type != "comment" && token.Type != "newline" {
				onError(tab, "TAB_AS_INDENT", "Tabs are not allowed as indentation", false)
			}
			tab = nil
		}
		switch token.Type {
		case "space":
			// At the doc level, tabs at line start may be parsed
			// as leading white space rather than indentation.
			// In a flow collection, only the parser handles indent.
			if flow == "" && (indicator != "doc-start" || next == nil || next.Type != "flow-collection") &&
				includesUnit(token.Source, '\t') {
				tab = token
			}
			hasSpace = true
		case "comment":
			if !hasSpace {
				onError(token, "MISSING_CHAR",
					"Comments must be separated from other tokens by white space characters", false)
			}
			// `token.source.substring(1) || ' '`
			cb := unitsToString(substring(token.Source, 1, len(token.Source)))
			if cb == "" {
				cb = " "
			}
			if comment == "" {
				comment = cb
			} else {
				comment += commentSep + cb
			}
			commentSep = ""
			atNewline = false
		case "newline":
			if atNewline {
				if comment != "" {
					comment += unitsToString(token.Source)
				} else if found == nil || indicator != "seq-item-ind" {
					spaceBefore = true
				}
			} else {
				commentSep += unitsToString(token.Source)
			}
			atNewline = true
			hasNewline = true
			if anchor != nil || tag != nil {
				newlineAfterProp = token
			}
			hasSpace = true
		case "anchor":
			if anchor != nil {
				onError(token, "MULTIPLE_ANCHORS", "A node can have at most one anchor", false)
			}
			if characterAt(token.Source, len(token.Source)-1) == ':' {
				onError(token.Offset+len(token.Source)-1, "BAD_ALIAS", "Anchor ending in : is ambiguous", true)
			}
			anchor = token
			if !startSet {
				start, startSet = token.Offset, true
			}
			atNewline = false
			hasSpace = false
			reqSpace = true
		case "tag":
			if tag != nil {
				onError(token, "MULTIPLE_TAGS", "A node can have at most one tag", false)
			}
			tag = token
			if !startSet {
				start, startSet = token.Offset, true
			}
			atNewline = false
			hasSpace = false
			reqSpace = true
		case indicator:
			// Could here handle preceding comments differently
			if anchor != nil || tag != nil {
				onError(token, "BAD_PROP_ORDER",
					"Anchors and tags must be after the "+unitsToString(token.Source)+" indicator", false)
			}
			if found != nil {
				// `flow ?? 'collection'`
				in := flow
				if in == "" {
					in = "collection"
				}
				onError(token, "UNEXPECTED_TOKEN", "Unexpected "+unitsToString(token.Source)+" in "+in, false)
			}
			found = token
			atNewline = indicator == "seq-item-ind" || indicator == "explicit-key-ind"
			hasSpace = false
		case "comma":
			if flow != "" {
				if comma != nil {
					onError(token, "UNEXPECTED_TOKEN", "Unexpected , in "+flow, false)
				}
				comma = token
				atNewline = false
				hasSpace = false
				break
			}
			// else fallthrough
			fallthrough
		default:
			onError(token, "UNEXPECTED_TOKEN", "Unexpected "+token.Type+" token", false)
			atNewline = false
			hasSpace = false
		}
	}
	end := options.offset
	if len(tokens) > 0 {
		last := tokens[len(tokens)-1]
		end = last.Offset + len(last.Source)
	}
	if reqSpace && next != nil && next.Type != "space" && next.Type != "newline" && next.Type != "comma" &&
		(next.Type != "scalar" || len(next.Source) != 0) {
		onError(next.Offset, "MISSING_CHAR", "Tags and anchors must be separated from the next token by white space", false)
	}
	if tab != nil && (atNewline && tab.Indent <= options.parentIndent ||
		next != nil && (next.Type == "block-map" || next.Type == "block-seq")) {
		onError(tab, "TAB_AS_INDENT", "Tabs are not allowed as indentation", false)
	}
	if !startSet {
		start = end
	}
	return props{
		comma:            comma,
		found:            found,
		spaceBefore:      spaceBefore,
		comment:          comment,
		hasNewline:       hasNewline,
		anchor:           anchor,
		tag:              tag,
		newlineAfterProp: newlineAfterProp,
		end:              end,
		start:            start,
	}
}

type resolvedEnd struct {
	comment string
	offset  int
}

// resolveEnd reads the tokens after a node to the end of its line: its comment, and where it ends. end
// is nil for upstream's undefined.
func resolveEnd(end []*cst.Token, offset int, reqSpace bool, onError onErrorFunc) resolvedEnd {
	comment := ""
	if end != nil {
		hasSpace := false
		sep := ""
		for _, token := range end {
			source, tokenType := token.Source, token.Type
			switch tokenType {
			case "space":
				hasSpace = true
			case "comment":
				if reqSpace && !hasSpace {
					onError(token, "MISSING_CHAR",
						"Comments must be separated from other tokens by white space characters", false)
				}
				// `source.substring(1) || ' '`
				cb := unitsToString(substring(source, 1, len(source)))
				if cb == "" {
					cb = " "
				}
				if comment == "" {
					comment = cb
				} else {
					comment += sep + cb
				}
				sep = ""
			case "newline":
				if comment != "" {
					sep += unitsToString(source)
				}
				hasSpace = true
			default:
				onError(token, "UNEXPECTED_TOKEN", "Unexpected "+tokenType+" at node end", false)
			}
			// `offset += source.length`: a token without a source (a collection) would make this NaN
			// upstream; the parser never ends a node with one.
			offset += len(source)
		}
	}
	return resolvedEnd{comment: comment, offset: offset}
}
