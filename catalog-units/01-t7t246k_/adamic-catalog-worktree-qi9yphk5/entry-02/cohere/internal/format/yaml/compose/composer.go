package compose

// Ported from eemeli/yaml 2.9.0, dist/compose/composer.js.

import (
	"iter"

	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// onErrorFunc is upstream's ComposeErrorHandler: source is an offset (int), a range ([]int, of which the
// first two are read), a position ([2]int), or a CST token.
type onErrorFunc func(source any, code string, message string, warning bool)

// getErrorPos is upstream's: an offset is one character, a token is its source (or one character for a
// token without one).
func getErrorPos(source any) [2]int {
	switch src := source.(type) {
	case int:
		return [2]int{src, src + 1}
	case [2]int:
		return src
	case []int:
		return [2]int{src[0], src[1]}
	case *cst.Token:
		if src.HasSource() {
			return [2]int{src.Offset, src.Offset + len(src.Source)}
		}
		return [2]int{src.Offset, src.Offset + 1}
	}
	panic("compose: no error position for this source")
}

// parsePrelude collects the comments of the prelude: the comment and newline tokens, and the directive
// lines, since the previous document.
func parsePrelude(prelude [][]uint16) (comment string, afterEmptyLine bool) {
	atComment := false
	for i := 0; i < len(prelude); i++ {
		source := prelude[i]
		switch characterAt(source, 0) {
		case '#':
			separator := ""
			if comment != "" {
				if afterEmptyLine {
					separator = "\n\n"
				} else {
					separator = "\n"
				}
			}
			// `source.substring(1) || ' '`
			text := unitsToString(substring(source, 1, len(source)))
			if text == "" {
				text = " "
			}
			comment += separator + text
			atComment = true
			afterEmptyLine = false
		case '%':
			// `prelude[i + 1]?.[0] !== '#'`
			if i+1 >= len(prelude) || characterAt(prelude[i+1], 0) != '#' {
				i += 1
			}
			atComment = false
		default:
			// This may be wrong after doc-end, but in that case it doesn't matter
			if !atComment {
				afterEmptyLine = true
			}
			atComment = false
		}
	}
	return comment, afterEmptyLine
}

// Composer composes a stream of CST nodes into a stream of YAML Documents.
type Composer struct {
	doc          *Document
	atDirectives bool
	prelude      [][]uint16
	errors       []*YAMLError
	warnings     []*YAMLError
	directives   *Directives
	options      Options
}

// NewComposer is `new Composer(options)`.
func NewComposer(options Options) *Composer {
	version := options.Version
	if version == "" {
		version = "1.2"
	}
	return &Composer{
		prelude:    [][]uint16{},
		errors:     []*YAMLError{},
		warnings:   []*YAMLError{},
		directives: newDirectives(DirectivesYAML{Version: version}, nil),
		options:    options,
	}
}

func (composer *Composer) onError(source any, code string, message string, warning bool) {
	pos := getErrorPos(source)
	if warning {
		composer.warnings = append(composer.warnings, newYAMLWarning(pos, code, message))
	} else {
		composer.errors = append(composer.errors, newYAMLParseError(pos, code, message))
	}
}

func (composer *Composer) decorate(doc *Document, afterDoc bool) {
	comment, afterEmptyLine := parsePrelude(composer.prelude)
	if comment != "" {
		dc := doc.Contents
		if afterDoc {
			if doc.Comment != "" {
				doc.Comment = doc.Comment + "\n" + comment
			} else {
				doc.Comment = comment
			}
		} else if afterEmptyLine || doc.Directives.DocStart || dc == nil {
			doc.CommentBefore = comment
		} else if IsCollection(dc) && !dc.Flow && len(dc.Items) > 0 {
			it := dc.Items[0]
			if IsPair(it) {
				it = it.Key
			}
			cb := it.CommentBefore
			if cb != "" {
				it.CommentBefore = comment + "\n" + cb
			} else {
				it.CommentBefore = comment
			}
		} else {
			cb := dc.CommentBefore
			if cb != "" {
				dc.CommentBefore = comment + "\n" + cb
			} else {
				dc.CommentBefore = comment
			}
		}
	}
	if afterDoc {
		doc.Errors = append(doc.Errors, composer.errors...)
		doc.Warnings = append(doc.Warnings, composer.warnings...)
	} else {
		doc.Errors = composer.errors
		doc.Warnings = composer.warnings
	}
	composer.prelude = [][]uint16{}
	composer.errors = []*YAMLError{}
	composer.warnings = []*YAMLError{}
}

// Compose composes tokens into documents. With forceDoc, a stream that contains no document still
// yields a final document including any comments and directives that would be applied to a subsequent
// document; endOffset should then be set, to set the document range end and to indicate errors
// correctly.
func (composer *Composer) Compose(tokens []*cst.Token, forceDoc bool, endOffset int) iter.Seq[*Document] {
	return func(yield func(*Document) bool) {
		for _, token := range tokens {
			if !composer.next(token, yield) {
				return
			}
		}
		composer.end(forceDoc, endOffset, yield)
	}
}

// next advances the composer by one CST token. It returns false when the consumer stopped.
func (composer *Composer) next(token *cst.Token, yield func(*Document) bool) bool {
	switch token.Type {
	case "directive":
		composer.directives.add(token.Source, func(offset int, message string, warning bool) {
			pos := getErrorPos(token)
			pos[0] += offset
			composer.onError(pos, "BAD_DIRECTIVE", message, warning)
		})
		composer.prelude = append(composer.prelude, token.Source)
		composer.atDirectives = true
	case "document":
		doc := composeDoc(composer.options, composer.directives, token, composer.onError)
		if composer.atDirectives && !doc.Directives.DocStart {
			composer.onError(token, "MISSING_CHAR", "Missing directives-end/doc-start indicator line", false)
		}
		composer.decorate(doc, false)
		if composer.doc != nil {
			if !yield(composer.doc) {
				return false
			}
		}
		composer.doc = doc
		composer.atDirectives = false
	case "byte-order-mark", "space":
	case "comment", "newline":
		composer.prelude = append(composer.prelude, token.Source)
	case "error":
		msg := token.Message
		if len(token.Source) > 0 {
			msg = token.Message + ": " + jsonStringify(token.Source)
		}
		yamlError := newYAMLParseError(getErrorPos(token), "UNEXPECTED_TOKEN", msg)
		if composer.atDirectives || composer.doc == nil {
			composer.errors = append(composer.errors, yamlError)
		} else {
			composer.doc.Errors = append(composer.doc.Errors, yamlError)
		}
	case "doc-end":
		if composer.doc == nil {
			msg := "Unexpected doc-end without preceding document"
			composer.errors = append(composer.errors, newYAMLParseError(getErrorPos(token), "UNEXPECTED_TOKEN", msg))
			break
		}
		composer.doc.Directives.DocEnd = true
		end := resolveEnd(token.End, token.Offset+len(token.Source), composer.doc.Options.Strict, composer.onError)
		composer.decorate(composer.doc, true)
		if end.comment != "" {
			dc := composer.doc.Comment
			if dc != "" {
				composer.doc.Comment = dc + "\n" + end.comment
			} else {
				composer.doc.Comment = end.comment
			}
		}
		composer.doc.Range[2] = end.offset
	default:
		composer.errors = append(composer.errors, newYAMLParseError(getErrorPos(token), "UNEXPECTED_TOKEN",
			"Unsupported token "+token.Type))
	}
	return true
}

// end is called at end of input to yield any remaining document.
func (composer *Composer) end(forceDoc bool, endOffset int, yield func(*Document) bool) {
	if composer.doc != nil {
		composer.decorate(composer.doc, true)
		doc := composer.doc
		composer.doc = nil
		yield(doc)
	} else if forceDoc {
		doc := newDocument(composer.directives, composer.options)
		if composer.atDirectives {
			composer.onError(endOffset, "MISSING_CHAR", "Missing directives-end indicator line", false)
		}
		doc.Range = []int{0, endOffset, endOffset}
		composer.decorate(doc, false)
		yield(doc)
	}
}
