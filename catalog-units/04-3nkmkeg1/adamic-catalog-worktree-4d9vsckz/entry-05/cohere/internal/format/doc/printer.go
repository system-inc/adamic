package doc

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf16"
	"unsafe"
)

// Options are the layout settings the printer reads. Upstream's endOfLine is always "lf" in our
// configuration, so it is not an option here; a printer that needs CRLF is a different port.
type Options struct {
	PrintWidth int
	TabWidth   int
	UseTabs    bool

	// ExpectedLength is about how long the printed text will be, in bytes: the length of the text the doc
	// was printed from, for a formatter. The output starts with room for it rather than growing from empty,
	// which was 97 MB of a cold run on ahra (#wcgw0n4). Zero starts empty. It never changes the output.
	ExpectedLength int
}

type mode int

const (
	// modeUnset is what a groupModeMap lookup returns for a group not yet printed. Upstream gets
	// undefined there, and the printer and fits() treat undefined differently, so it is its own value.
	modeUnset mode = iota
	modeBreak
	modeFlat
)

// indentCommandKind mirrors upstream's INDENT_COMMAND_TYPE_*.
type indentCommandKind int

const (
	indentCommandIndent indentCommandKind = iota
	indentCommandDedent
	indentCommandWidth
	indentCommandString
)

type indentCommand struct {
	kind   indentCommandKind
	width  int
	string string
}

// indentation is upstream's Indent: the rendered prefix, its length, the queue it was built from, and
// the root that literal lines return to. Immutable once built, like upstream's spread copies.
type indentation struct {
	value  string
	length int
	queue  []indentCommand
	root   *indentation
}

// rootIndentation is upstream's ROOT_INDENT, whose root is itself.
var rootIndentation = func() *indentation {
	root := &indentation{}
	root.root = root
	return root
}()

// utf16Length is JavaScript's String.length, which upstream uses for indentation strings.
func utf16Length(text string) int {
	return len(utf16.Encode([]rune(text)))
}

// generateIndent is upstream's generateIndent, printer/indent.js.
func generateIndent(base *indentation, command indentCommand, options Options) *indentation {
	var queue []indentCommand
	if command.kind == indentCommandDedent {
		if len(base.queue) > 0 {
			queue = append([]indentCommand(nil), base.queue[:len(base.queue)-1]...)
		}
	} else {
		queue = append(append([]indentCommand(nil), base.queue...), command)
	}

	var value strings.Builder
	length := 0
	lastTabs := 0
	lastSpaces := 0

	addTabs := func(count int) {
		value.WriteString(strings.Repeat("\t", count))
		length += options.TabWidth * count
	}
	addSpaces := func(count int) {
		value.WriteString(strings.Repeat(" ", count))
		length += count
	}
	resetLast := func() {
		lastTabs = 0
		lastSpaces = 0
	}
	flushTabs := func() {
		if lastTabs > 0 {
			addTabs(lastTabs)
		}
		resetLast()
	}
	flushSpaces := func() {
		if lastSpaces > 0 {
			addSpaces(lastSpaces)
		}
		resetLast()
	}
	flush := func() {
		if options.UseTabs {
			flushTabs()
		} else {
			flushSpaces()
		}
	}

	for _, queued := range queue {
		switch queued.kind {
		case indentCommandIndent:
			flush()
			if options.UseTabs {
				addTabs(1)
			} else {
				addSpaces(options.TabWidth)
			}
		case indentCommandString:
			flush()
			value.WriteString(queued.string)
			length += utf16Length(queued.string)
		case indentCommandWidth:
			lastTabs++
			lastSpaces += queued.width
		default:
			panic(fmt.Sprintf("unexpected indent command %d", queued.kind))
		}
	}
	flushSpaces()

	return &indentation{value: value.String(), length: length, queue: queue, root: base.root}
}

// indentations is one Print's indentations, each built once from its base and command. An indentation is
// immutable, and one print's options are fixed, so the same base and command always build the same one;
// building it again for every indent and align the printer met was 3.0M of the JavaScript printer's
// allocations on ahra (#fyw36kf). Per print rather than global, because the options differ between prints.
type indentations map[indentKey]*indentation

type indentKey struct {
	base    *indentation
	command indentCommand
}

// generate is generateIndent, from the print's indentations when this base and command were built before.
func (built indentations) generate(base *indentation, command indentCommand, options Options) *indentation {
	key := indentKey{base: base, command: command}
	if known, present := built[key]; present {
		return known
	}
	generated := generateIndent(base, command, options)
	built[key] = generated
	return generated
}

// makeIndent is upstream's makeIndent.
func makeIndent(base *indentation, options Options, built indentations) *indentation {
	return built.generate(base, indentCommand{kind: indentCommandIndent}, options)
}

// makeAlign is upstream's makeAlign.
func makeAlign(base *indentation, align *Align, options Options, built indentations) *indentation {
	switch align.Kind {
	case AlignRoot:
		copied := *base
		copied.root = base
		return &copied
	case AlignDedentToRoot:
		return base.root
	case AlignString:
		if align.String == "" {
			return base
		}
		return built.generate(base, indentCommand{kind: indentCommandString, string: align.String}, options)
	default:
		if align.Width == 0 {
			return base
		}
		if align.Width < 0 {
			return built.generate(base, indentCommand{kind: indentCommandDedent}, options)
		}
		return built.generate(base, indentCommand{kind: indentCommandWidth, width: align.Width}, options)
	}
}

// fitsCommands holds fits' command stacks between calls.
var fitsCommands = sync.Pool{New: func() any { return new([]command) }}

// trailingIndentation is how many spaces and tabs text ends in once appended to text that ended in
// previous of them: all of text's when text is nothing else, added to previous, and otherwise text's own.
func trailingIndentation(text string, previous int) int {
	count := 0
	for index := len(text) - 1; index >= 0; index-- {
		if text[index] != ' ' && text[index] != '\t' {
			return count
		}
		count++
	}
	return previous + count
}

// printResult is upstream's PrintResult without cursor tracking.
//
// The settled and unsettled split is kept to read against upstream, not because output depends on it.
// Upstream needs it for cursor positions. For text it is unobservable: everything settled was settled
// by a trim, so it ends in a non-space, and a trim reaching past it would remove nothing. A mutation
// that stopped settling survived 5,000 differential docs, and that is why this comment says so.
//
// Both halves are one buffer, with the settled length marked, so settling copies nothing and the unsettled
// half keeps its memory from line to line: two builders, one reset on every settle, regrew it each line,
// and were 2.7M of the JavaScript printer's allocations on ahra (#fyw36kf).
type printResult struct {
	buffer  []byte
	settled int
}

func (result *printResult) write(text string) { result.buffer = append(result.buffer, text...) }

func (result *printResult) settle() { result.settled = len(result.buffer) }

// trim is upstream's trimIndentation over the unsettled half: its trailing spaces and tabs, and their count.
func (result *printResult) trim() int {
	count := 0
	for index := len(result.buffer) - 1; index >= result.settled; index-- {
		if result.buffer[index] != ' ' && result.buffer[index] != '\t' {
			break
		}
		count++
	}
	result.buffer = result.buffer[:len(result.buffer)-count]
	result.settle()
	return count
}

// finish is the printed text. It hands back the buffer itself rather than a copy: nothing writes to a
// printResult once it is finished, so the bytes can't change under the string.
func (result *printResult) finish() string {
	result.settle()
	if len(result.buffer) == 0 {
		return ""
	}
	return unsafe.String(&result.buffer[0], len(result.buffer))
}

// command is one entry of the printer's stack.
type command struct {
	indent *indentation
	mode   mode
	doc    Doc
}

// fillProgress is upstream's `{...doc, [DOC_FILL_PRINTED_LENGTH]: offset}`: a fill partly printed.
type fillProgress struct {
	fill   *Fill
	offset int
}

func (*fillProgress) isDoc() {}

// fits is upstream's fits, printer/printer.js.
//
// Upstream keeps the text it measured, for a trim to remove trailing whitespace from. Only the count of that
// trailing whitespace is ever read, so it is kept instead of the text: the builder was 1.7M of the
// JavaScript printer's allocations on ahra (#fyw36kf). The command stack comes from a pool for the same
// reason.
func fits(next command, rest []command, remainingWidth int, hasLineSuffix bool, groupModes map[*GroupID]mode, mustBeFlat bool) bool {
	restIndex := len(rest)
	hasPendingSpace := false
	pooled := fitsCommands.Get().(*[]command)
	commands := append((*pooled)[:0], next)
	defer func() {
		*pooled = commands[:0]
		fitsCommands.Put(pooled)
	}()
	// trailing is how many spaces and tabs the text measured so far ends in, which is what upstream's trimIndentation
	// would remove from it.
	trailing := 0

	for remainingWidth >= 0 {
		if len(commands) == 0 {
			if restIndex == 0 {
				return true
			}
			restIndex--
			commands = append(commands, rest[restIndex])
			continue
		}

		current := commands[len(commands)-1]
		commands = commands[:len(commands)-1]

		switch document := current.doc.(type) {
		case Text:
			if document != "" {
				if hasPendingSpace {
					trailing++
					remainingWidth--
					hasPendingSpace = false
				}
				trailing = trailingIndentation(string(document), trailing)
				remainingWidth -= StringWidth(string(document))
			}

		case *Fill:
			for index := len(document.Parts) - 1; index >= 0; index-- {
				commands = append(commands, command{mode: current.mode, doc: document.Parts[index]})
			}

		case *fillProgress:
			for index := len(document.fill.Parts) - 1; index >= document.offset; index-- {
				commands = append(commands, command{mode: current.mode, doc: document.fill.Parts[index]})
			}

		case *Indent:
			commands = append(commands, command{mode: current.mode, doc: document.Contents})
		case *Align:
			commands = append(commands, command{mode: current.mode, doc: document.Contents})
		case *IndentIfBreak:
			commands = append(commands, command{mode: current.mode, doc: document.Contents})
		case *Label:
			commands = append(commands, command{mode: current.mode, doc: document.Contents})

		case trimDoc:
			remainingWidth += trailing
			trailing = 0

		case *Group:
			if mustBeFlat && document.Break {
				return false
			}
			groupMode := current.mode
			if document.Break {
				groupMode = modeBreak
			}
			contents := document.Contents
			if document.ExpandedStates != nil && groupMode == modeBreak {
				contents = document.ExpandedStates[len(document.ExpandedStates)-1]
			}
			commands = append(commands, command{mode: groupMode, doc: contents})

		case *IfBreak:
			groupMode := current.mode
			if document.GroupID != nil {
				groupMode = groupModes[document.GroupID]
				if groupMode == modeUnset {
					groupMode = modeFlat
				}
			}
			contents := document.FlatContents
			if groupMode == modeBreak {
				contents = document.BreakContents
			}
			if !isEmpty(contents) {
				commands = append(commands, command{mode: current.mode, doc: contents})
			}

		case *Line:
			if current.mode == modeBreak || document.Hard {
				return true
			}
			if !document.Soft {
				hasPendingSpace = true
			}

		case *LineSuffix:
			hasLineSuffix = true

		case lineSuffixBoundaryDoc:
			if hasLineSuffix {
				return false
			}

		case breakParentDoc:

		default:
			if parts, isArray := Parts(document); isArray {
				for index := len(parts) - 1; index >= 0; index-- {
					commands = append(commands, command{mode: current.mode, doc: parts[index]})
				}
			}
		}
	}
	return false
}

// isEmpty is JavaScript truthiness for a doc: an empty string or a missing doc is falsy.
func isEmpty(document Doc) bool {
	if document == nil {
		return true
	}
	text, isText := document.(Text)
	return isText && text == ""
}

// Print is upstream's printDocToString, printer/printer.js.
func Print(document Doc, options Options) string {
	groupModes := map[*GroupID]mode{}
	width := options.PrintWidth
	position := 0
	commands := []command{{indent: rootIndentation, mode: modeBreak, doc: document}}
	shouldRemeasure := false
	var lineSuffix []command
	result := &printResult{buffer: make([]byte, 0, options.ExpectedLength)}
	built := indentations{}

	PropagateBreaks(document)

	for len(commands) > 0 {
		current := commands[len(commands)-1]
		commands = commands[:len(commands)-1]

		switch document := current.doc.(type) {
		case Text:
			if document != "" {
				result.write(string(document))
				if len(commands) > 0 {
					position += StringWidth(string(document))
				}
			}

		case *Indent:
			commands = append(commands, command{indent: makeIndent(current.indent, options, built), mode: current.mode, doc: document.Contents})

		case *Align:
			commands = append(commands, command{indent: makeAlign(current.indent, document, options, built), mode: current.mode, doc: document.Contents})

		case trimDoc:
			position -= result.trim()

		case *Group:
			next := printGroup(document, current, commands, width-position, len(lineSuffix) > 0, groupModes, &shouldRemeasure)
			commands = append(commands, next)
			if document.ID != nil {
				groupModes[document.ID] = next.mode
			}

		case *Fill:
			commands = printFill(&fillProgress{fill: document}, current, commands, width-position, len(lineSuffix) > 0, groupModes)
		case *fillProgress:
			commands = printFill(document, current, commands, width-position, len(lineSuffix) > 0, groupModes)

		case *IfBreak:
			groupMode := current.mode
			if document.GroupID != nil {
				groupMode = groupModes[document.GroupID]
			}
			if groupMode == modeBreak && !isEmpty(document.BreakContents) {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: document.BreakContents})
			}
			if groupMode == modeFlat && !isEmpty(document.FlatContents) {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: document.FlatContents})
			}

		case *IndentIfBreak:
			groupMode := current.mode
			if document.GroupID != nil {
				groupMode = groupModes[document.GroupID]
			}
			if groupMode == modeBreak {
				contents := document.Contents
				if !document.Negate {
					contents = NewIndent(document.Contents)
				}
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: contents})
			}
			if groupMode == modeFlat {
				contents := document.Contents
				if document.Negate {
					contents = NewIndent(document.Contents)
				}
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: contents})
			}

		case *LineSuffix:
			lineSuffix = append(lineSuffix, command{indent: current.indent, mode: current.mode, doc: document.Contents})

		case lineSuffixBoundaryDoc:
			if len(lineSuffix) > 0 {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: HardlineWithoutBreakParent})
			}

		case *Line:
			handled := false
			if current.mode == modeFlat {
				if !document.Hard {
					if !document.Soft {
						result.write(" ")
						position++
					}
					handled = true
				} else {
					// Forced into the output while flat, so the next group must remeasure.
					shouldRemeasure = true
				}
			}
			if handled {
				break
			}
			if len(lineSuffix) > 0 {
				commands = append(commands, current)
				for index := len(lineSuffix) - 1; index >= 0; index-- {
					commands = append(commands, lineSuffix[index])
				}
				lineSuffix = lineSuffix[:0]
				break
			}
			if document.Literal {
				result.write("\n")
				position = 0
				if current.indent.root != nil {
					if current.indent.root.value != "" {
						result.write(current.indent.root.value)
					}
					position = current.indent.root.length
				}
			} else {
				result.trim()
				result.write("\n")
				result.write(current.indent.value)
				position = current.indent.length
			}

		case *Label:
			commands = append(commands, command{indent: current.indent, mode: current.mode, doc: document.Contents})

		case breakParentDoc:

		default:
			parts, isArray := Parts(document)
			if !isArray {
				panic(fmt.Sprintf("invalid doc %T", current.doc))
			}
			for index := len(parts) - 1; index >= 0; index-- {
				commands = append(commands, command{indent: current.indent, mode: current.mode, doc: parts[index]})
			}
		}

		// Flush deferred line suffixes at the end, in case no newline follows them.
		if len(commands) == 0 && len(lineSuffix) > 0 {
			for index := len(lineSuffix) - 1; index >= 0; index-- {
				commands = append(commands, lineSuffix[index])
			}
			lineSuffix = lineSuffix[:0]
		}
	}

	return result.finish()
}

// printGroup is the immediately-invoked printGroup in upstream's DOC_TYPE_GROUP case.
func printGroup(group *Group, current command, rest []command, remainingWidth int, hasLineSuffix bool, groupModes map[*GroupID]mode, shouldRemeasure *bool) command {
	if current.mode == modeFlat && !*shouldRemeasure {
		groupMode := modeFlat
		if group.Break {
			groupMode = modeBreak
		}
		return command{indent: current.indent, mode: groupMode, doc: group.Contents}
	}

	*shouldRemeasure = false
	flat := command{indent: current.indent, mode: modeFlat, doc: group.Contents}
	if !group.Break && fits(flat, rest, remainingWidth, hasLineSuffix, groupModes, false) {
		return flat
	}

	if group.ExpandedStates == nil {
		return command{indent: current.indent, mode: modeBreak, doc: group.Contents}
	}

	if !group.Break {
		for index := 1; index < len(group.ExpandedStates)-1; index++ {
			state := command{indent: current.indent, mode: modeFlat, doc: group.ExpandedStates[index]}
			if fits(state, rest, remainingWidth, hasLineSuffix, groupModes, false) {
				return state
			}
		}
	}
	return command{indent: current.indent, mode: modeBreak, doc: group.ExpandedStates[len(group.ExpandedStates)-1]}
}

// printFill is upstream's DOC_TYPE_FILL case, returning the command stack it leaves.
func printFill(progress *fillProgress, current command, commands []command, remainingWidth int, hasLineSuffix bool, groupModes map[*GroupID]mode) []command {
	parts := progress.fill.Parts
	offset := progress.offset
	length := len(parts) - offset
	if length == 0 {
		return commands
	}

	content := parts[offset]
	contentFlat := command{indent: current.indent, mode: modeFlat, doc: content}
	contentBreak := command{indent: current.indent, mode: modeBreak, doc: content}
	contentFits := fits(contentFlat, nil, remainingWidth, hasLineSuffix, groupModes, true)

	if length == 1 {
		if contentFits {
			return append(commands, contentFlat)
		}
		return append(commands, contentBreak)
	}

	whitespace := parts[offset+1]
	whitespaceFlat := command{indent: current.indent, mode: modeFlat, doc: whitespace}
	whitespaceBreak := command{indent: current.indent, mode: modeBreak, doc: whitespace}

	if length == 2 {
		if contentFits {
			return append(commands, whitespaceFlat, contentFlat)
		}
		return append(commands, whitespaceBreak, contentBreak)
	}

	secondContent := parts[offset+2]
	remaining := command{indent: current.indent, mode: current.mode, doc: &fillProgress{fill: progress.fill, offset: offset + 2}}
	firstAndSecond := command{indent: current.indent, mode: modeFlat, doc: Concat{content, whitespace, secondContent}}
	firstAndSecondFit := fits(firstAndSecond, nil, remainingWidth, hasLineSuffix, groupModes, true)

	commands = append(commands, remaining)
	switch {
	case firstAndSecondFit:
		return append(commands, whitespaceFlat, contentFlat)
	case contentFits:
		return append(commands, whitespaceBreak, contentFlat)
	default:
		return append(commands, whitespaceBreak, contentBreak)
	}
}
