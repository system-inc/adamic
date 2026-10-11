package buildcache

import "strings"

// strace's lines, as Settle reads them, parsed by hand. Settle reads the trace on one thread while the run goes on, and
// as regular expressions these were 80% of its time, about 3 µs a line (#vt46geg step 3). Each parser takes exactly
// what the expression in its comment took, the same groups from the same lines and nothing from the rest, which
// TestTheStraceParsersMatchTheirExpressions and FuzzTheStraceParsers hold against the expressions themselves.

func isDigit(character byte) bool {
	return '0' <= character && character <= '9'
}

// isWord is \w: an ASCII letter, a digit or an underscore.
func isWord(character byte) bool {
	return isDigit(character) || 'a' <= character && character <= 'z' || 'A' <= character && character <= 'Z' || character == '_'
}

// isSpace is \s: a space, tab, newline, form feed or carriage return.
func isSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\f' || character == '\r'
}

// leading is how many of text's first bytes are in class.
func leading(text string, class func(byte) bool) int {
	count := 0
	for count < len(text) && class(text[count]) {
		count++
	}
	return count
}

// splitTraceLine is `^(\d+)\s+(.*)$`: the pid a line starts with and the rest of it.
func splitTraceLine(text string) (pid, rest string, ok bool) {
	digits := leading(text, isDigit)
	spaces := leading(text[digits:], isSpace)
	if digits == 0 || spaces == 0 || strings.IndexByte(text[digits+spaces:], '\n') >= 0 {
		return "", "", false
	}
	return text[:digits], text[digits+spaces:], true
}

// isExit is `^\+\+\+ (?:exited with \d+|killed by \w+(?: \(core dumped\))?) \+\+\+$`: a process's end.
func isExit(rest string) bool {
	middle, ok := strings.CutPrefix(rest, "+++ ")
	if !ok {
		return false
	}
	if middle, ok = strings.CutSuffix(middle, " +++"); !ok {
		return false
	}
	if code, ok := strings.CutPrefix(middle, "exited with "); ok {
		return code != "" && leading(code, isDigit) == len(code)
	}
	if signal, ok := strings.CutPrefix(middle, "killed by "); ok {
		signal = strings.TrimSuffix(signal, " (core dumped)")
		return signal != "" && leading(signal, isWord) == len(signal)
	}
	return false
}

// resumedTail is `^<\.\.\. (\w+) resumed>\s?(.*)$`: what follows a call's resumption, the second half of a call an
// earlier line left unfinished.
func resumedTail(rest string) (string, bool) {
	tail, ok := strings.CutPrefix(rest, "<... ")
	if !ok {
		return "", false
	}
	name := leading(tail, isWord)
	if name == 0 {
		return "", false
	}
	if tail, ok = strings.CutPrefix(tail[name:], " resumed>"); !ok {
		return "", false
	}
	if tail != "" && isSpace(tail[0]) {
		tail = tail[1:]
	}
	if strings.IndexByte(tail, '\n') >= 0 {
		return "", false
	}
	return tail, true
}

// A straceCall is a completed call's parts: its name, its arguments, what it returned, the path a returned descriptor
// decodes to, and the error it failed with.
type straceCall struct {
	call, arguments, value, opened, errno string
}

// parseCall is `^(\w+)\((.*)\)\s+=\s+(-?\d+|\?)(?:<([^>]*)>)?(?:\s+(E[A-Z0-9]+))?`. The arguments run to the last ")"
// a result follows, as the greedy (.*) takes them, so a path holding ") = 1" stays in the arguments.
func parseCall(rest string) (straceCall, bool) {
	name := leading(rest, isWord)
	if name == 0 || name == len(rest) || rest[name] != '(' {
		return straceCall{}, false
	}
	body := rest[name+1:]
	for end := strings.LastIndexByte(body, ')'); end >= 0; end = strings.LastIndexByte(body[:end], ')') {
		if strings.IndexByte(body[:end], '\n') >= 0 {
			continue
		}
		if value, opened, errno, ok := callResult(body[end+1:]); ok {
			return straceCall{call: rest[:name], arguments: body[:end], value: value, opened: opened, errno: errno}, true
		}
	}
	return straceCall{}, false
}

// callResult is `\s+=\s+(-?\d+|\?)(?:<([^>]*)>)?(?:\s+(E[A-Z0-9]+))?`, at the start of text.
func callResult(text string) (value, opened, errno string, ok bool) {
	spaces := leading(text, isSpace)
	if spaces == 0 || spaces == len(text) || text[spaces] != '=' {
		return "", "", "", false
	}
	text = text[spaces+1:]
	if spaces = leading(text, isSpace); spaces == 0 {
		return "", "", "", false
	}
	text = text[spaces:]
	if strings.HasPrefix(text, "?") {
		value, text = "?", text[1:]
	} else {
		sign := 0
		if strings.HasPrefix(text, "-") {
			sign = 1
		}
		digits := leading(text[sign:], isDigit)
		if digits == 0 {
			return "", "", "", false
		}
		value, text = text[:sign+digits], text[sign+digits:]
	}
	if strings.HasPrefix(text, "<") {
		if close := strings.IndexByte(text[1:], '>'); close >= 0 {
			opened, text = text[1:1+close], text[2+close:]
		}
	}
	if spaces = leading(text, isSpace); spaces > 0 && spaces+1 < len(text) && text[spaces] == 'E' {
		if letters := leading(text[spaces+1:], func(character byte) bool { return isDigit(character) || 'A' <= character && character <= 'Z' }); letters > 0 {
			errno = text[spaces : spaces+1+letters]
		}
	}
	return value, opened, errno, true
}

// directoryArgument is `^(AT_FDCWD|-?\d+)(?:<([^>]*)>)?,\s*`: an at-call's directory argument, the path its descriptor
// decodes to, and how long the argument is with its comma and the spaces after.
func directoryArgument(arguments string) (decoded string, length int, ok bool) {
	if strings.HasPrefix(arguments, "AT_FDCWD") {
		length = len("AT_FDCWD")
	} else {
		sign := 0
		if strings.HasPrefix(arguments, "-") {
			sign = 1
		}
		digits := leading(arguments[sign:], isDigit)
		if digits == 0 {
			return "", 0, false
		}
		length = sign + digits
	}
	if strings.HasPrefix(arguments[length:], "<") {
		if close := strings.IndexByte(arguments[length+1:], '>'); close >= 0 {
			decoded, length = arguments[length+1:length+1+close], length+2+close
		}
	}
	if !strings.HasPrefix(arguments[length:], ",") {
		return "", 0, false
	}
	length++
	return decoded, length + leading(arguments[length:], isSpace), true
}

// quotedArgument is `^"((?:[^"\\]|\\.)*)"`: a quoted argument's text, still escaped, and how long it is with its quotes.
func quotedArgument(text string) (string, int, bool) {
	if !strings.HasPrefix(text, `"`) {
		return "", 0, false
	}
	for index := 1; index < len(text); index++ {
		switch text[index] {
		case '"':
			return text[1:index], index + 1, true
		case '\\':
			if index+1 == len(text) || text[index+1] == '\n' {
				return "", 0, false
			}
			index++
		}
	}
	return "", 0, false
}

// descriptorPath is `^-?\d+<([^>]*)>`: the path a descriptor argument decodes to.
func descriptorPath(text string) (string, bool) {
	sign := 0
	if strings.HasPrefix(text, "-") {
		sign = 1
	}
	digits := leading(text[sign:], isDigit)
	if digits == 0 || !strings.HasPrefix(text[sign+digits:], "<") {
		return "", false
	}
	text = text[sign+digits+1:]
	close := strings.IndexByte(text, '>')
	if close < 0 {
		return "", false
	}
	return text[:close], true
}
