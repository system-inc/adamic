package native

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// A graph record's wrapper and table are merged into one region at birth (adamic_record_new_graph,
// record.c), which is safe only while the table is the wrapper's alone: if anything else retained the
// table or stored it, the merge would have joined the record's whole region to that holder's life. This
// holds the table there. Outside record.c no runtime file may call record.c's table() helper, which
// must stay static, and a read of a record's slots[0] (the table) must read through it in place, never
// hand it on: retained, stored, returned or passed. Every such read is listed below, by file and
// function, with how many it makes.
//
// A slot read is a record's when its receiver is named a record in the same function: the first
// argument of an adamic_record_ call, or declared adamic_record *. A receiver that's no plain name
// (a cast, a call) counts when the function names a record at all.

// recordTableReaders are the functions outside record.c that read a record's table, with how many
// reads each makes. The list may only shrink: a new reader fails until it's reviewed and listed here,
// and a listed one that's gone or changed fails until the list says so, as testdata/runtime-statics
// does for storage. None of them may hand the table on, listed or not.
var recordTableReaders = map[string]int{
	// Reads the table's reference_values, so a record of scalars never answers that it holds a
	// function, and keeps nothing.
	"union.c:adamic_object_own_function": 1,
}

// recordTableRetainers are the calls that keep or join what they're given: a table named in one of
// their statements is handed on.
var recordTableRetainers = []string{"adamic_retain", "adamic_graph_retain", "adamic_graph_hold", "adamic_graph_take", "adamic_graph_merge"}

type recordTableAccess struct {
	file, function string
	line           int
	// call is a call of table(); otherwise it's a read of a record's slots[0].
	call bool
	// handed is a read that doesn't read through the table in place.
	handed bool
}

// recordGuardTokens are a runtime file's tokens without comments and preprocessor lines, as
// runtimeStorage reads them, so braces count the way the compiler sees them. cStorageTokens reads
// punctuation a character at a time, so an arrow is joined back into one token here.
func recordGuardTokens(source string) []storageToken {
	lines := strings.Split(source, "\n")
	directives := map[int]bool{}
	continued := false
	for i, line := range lines {
		if continued || strings.HasPrefix(strings.TrimSpace(line), "#") {
			directives[i+1] = true
			continued = strings.HasSuffix(strings.TrimSpace(line), "\\")
		}
	}
	var tokens []storageToken
	previous, line, minus := 0, 1, -1
	for _, match := range cStorageTokens.FindAllStringIndex(source, -1) {
		text := source[match[0]:match[1]]
		line += strings.Count(source[previous:match[0]], "\n")
		previous = match[0]
		if strings.HasPrefix(text, "//") || strings.HasPrefix(text, "/*") || directives[line] {
			continue
		}
		if text == ">" && minus == match[0] && len(tokens) > 0 {
			tokens[len(tokens)-1].text = "->"
			minus = -1
			continue
		}
		minus = -1
		if text == "-" {
			minus = match[1]
		}
		tokens = append(tokens, storageToken{text, line})
	}
	return tokens
}

// recordTableAccesses finds every access to a record's table in one runtime file outside record.c.
func recordTableAccesses(file, source string) []recordTableAccess {
	tokens := recordGuardTokens(source)
	text := func(index int) string {
		if index < 0 || index >= len(tokens) {
			return ""
		}
		return tokens[index].text
	}
	// within is the function each token is in, by name, its parameters included, or "" at file scope.
	// A function is a brace at file scope after a parameter list's parenthesis, named by the last name
	// before a parenthesis there.
	within := make([]string, len(tokens))
	depth, parens, parameters := 0, 0, 0
	candidate, function := "", ""
	for i, token := range tokens {
		switch token.text {
		case "(":
			if depth == 0 && parens == 0 && cStorageIdentifier.MatchString(text(i-1)) && !storageQualifierGroups[text(i-1)] {
				candidate = text(i - 1)
				parameters = i
			}
			parens++
		case ")":
			parens--
		case "{":
			if depth == 0 {
				function = ""
				if text(i-1) == ")" && candidate != "" {
					function = candidate
					for parameter := parameters; parameter < i; parameter++ {
						within[parameter] = function
					}
				}
			}
			depth++
		case "}":
			depth--
			if depth == 0 {
				function, candidate = "", ""
			}
		case ";":
			if depth == 0 {
				candidate = ""
			}
		}
		if depth > 0 {
			within[i] = function
		}
	}
	// records are the names a function gives records, and mentions whether it names one at all.
	records := map[string]map[string]bool{}
	mentions := map[string]bool{}
	for i, token := range tokens {
		name := within[i]
		if records[name] == nil {
			records[name] = map[string]bool{}
		}
		if strings.HasPrefix(token.text, "adamic_record_") && text(i+1) == "(" {
			mentions[name] = true
			if cStorageIdentifier.MatchString(text(i+2)) && (text(i+3) == "," || text(i+3) == ")") {
				records[name][text(i+2)] = true
			}
		}
		if token.text == "adamic_record" && text(i+1) == "*" {
			mentions[name] = true
			next := i + 2
			if text(next) == "const" {
				next++
			}
			if cStorageIdentifier.MatchString(text(next)) {
				records[name][text(next)] = true
			}
		}
	}
	// handed says whether the read at slots (the index of "slots") hands the table on: it isn't read
	// through in place, as in ((adamic_map *)record->slots[0].reference)->count, or its statement calls
	// something that keeps or joins what it's given.
	handed := func(slots int) bool {
		start, end := slots, slots
		for start > 0 && text(start-1) != ";" && text(start-1) != "{" && text(start-1) != "}" {
			start--
		}
		for end < len(tokens) && text(end) != ";" {
			end++
		}
		for index := start; index < end; index++ {
			if slices.Contains(recordTableRetainers, text(index)) {
				return true
			}
		}
		after := slots + 4
		if text(after) != "." || text(after+1) != "reference" {
			return true
		}
		after += 2
		for text(after) == ")" {
			after++
		}
		return text(after) != "->"
	}
	var found []recordTableAccess
	for i, token := range tokens {
		name := within[i]
		if token.text == "table" && text(i+1) == "(" && text(i-1) != "." && text(i-1) != "->" {
			found = append(found, recordTableAccess{file, name, token.line, true, true})
			continue
		}
		if token.text != "slots" || text(i+1) != "[" || text(i+2) != "0" || text(i+3) != "]" || (text(i-1) != "->" && text(i-1) != ".") {
			continue
		}
		receiver := text(i - 2)
		named := cStorageIdentifier.MatchString(receiver)
		if (named && records[name][receiver]) || (!named && mentions[name]) {
			found = append(found, recordTableAccess{file, name, token.line, false, handed(i)})
		}
	}
	return found
}

// recordTableProblems is what the guard refuses in a runtime snapshot, file name to source.
func recordTableProblems(sources map[string]string) []string {
	var problems []string
	if !strings.Contains(sources["record.c"], "static adamic_map *table(const adamic_record *record) {") {
		problems = append(problems, "record.c: table() must stay static to record.c")
	}
	reads := map[string]int{}
	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, file := range names {
		if file == "record.c" || (!strings.HasSuffix(file, ".c") && !strings.HasSuffix(file, ".h")) {
			continue
		}
		for _, access := range recordTableAccesses(file, sources[file]) {
			key := access.file + ":" + access.function
			switch {
			case access.call:
				problems = append(problems, fmt.Sprintf("%s:%d: %s calls record.c's table(); read a record through adamic_record_ functions", access.file, access.line, key))
			case access.handed:
				problems = append(problems, fmt.Sprintf("%s:%d: %s hands a record's table on (retained, stored, returned or passed); a graph record's merge at birth needs the table to be the wrapper's alone", access.file, access.line, key))
			default:
				reads[key]++
			}
		}
	}
	for key, count := range reads {
		if listed, ok := recordTableReaders[key]; !ok {
			problems = append(problems, fmt.Sprintf("%s reads a record's table and isn't listed (reads: %d); review it and list it in recordTableReaders", key, count))
		} else if listed != count {
			problems = append(problems, fmt.Sprintf("%s reads a record's table a number of times the list doesn't say (reads: %d, listed: %d); review the change and the list", key, count, listed))
		}
	}
	for key := range recordTableReaders {
		if reads[key] == 0 {
			problems = append(problems, fmt.Sprintf("%s is listed but reads no record's table now; take it off recordTableReaders", key))
		}
	}
	slices.Sort(problems)
	return problems
}

func recordTableSources(t *testing.T) map[string]string {
	t.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, file := range files {
		sources[file.name] = string(file.contents)
	}
	return sources
}

func TestRecordTableStaysTheWrappers(t *testing.T) {
	t.Parallel()
	sources := recordTableSources(t)
	if problems := recordTableProblems(sources); len(problems) != 0 {
		t.Fatalf("a record's table reached outside record.c:\n%s", strings.Join(problems, "\n"))
	}
	for key, count := range recordTableReaders {
		t.Logf("%s reads a record's table %d times, in place", key, count)
	}
}

// Each mutant rewrites one runtime file's text in memory, by exact match, and the guard must refuse it.
func TestRecordTableGuardMutants(t *testing.T) {
	t.Parallel()
	const ownFunction = "if (!((const adamic_map *)object->slots[0].reference)->reference_values) return false;"
	const keyedClass = "// keyed_class is the emitter's row for the object's class, found by the class itself, or NULL.\n"
	for _, mutant := range []struct{ name, file, before, after, refused string }{
		// A new reader in another file that retains the table.
		{"new reader retains", "union.c", keyedClass, "void *adamic_record_table_kept(adamic_object *object) {\n\tif (!adamic_record_is(object)) { return NULL; }\n\treturn adamic_retain(object->slots[0].reference);\n}\n\n" + keyedClass, "union.c:adamic_record_table_kept hands a record's table on"},
		// The listed reader keeps the table it reads.
		{"listed reader retains", "union.c", ownFunction, "adamic_map *kept = adamic_retain(object->slots[0].reference);\n\t\tif (!kept->reference_values) return false;", "union.c:adamic_object_own_function hands a record's table on"},
		// A reader in another file through record.c's helper.
		{"table called outside record.c", "json_objects.c", "if (adamic_record_is(object)) stop(", "if (adamic_record_is(object) && table(object)->count == 0) stop(", "calls record.c's table()"},
		// A new reader that only reads in place still has to be listed.
		{"new reader unlisted", "timers_impl.h", "    if (adamic_record_is(object)) {\n        adamic_string key = ADAMIC_STRING(\"_onTimeout\");", "    if (adamic_record_is(object) && ((const adamic_map *)object->slots[0].reference)->count == 0) return false;\n    if (adamic_record_is(object)) {\n        adamic_string key = ADAMIC_STRING(\"_onTimeout\");", "timers_impl.h:timer_foreign_handle reads a record's table and isn't listed"},
		// The listed reader gone: the list must shrink with it.
		{"listed reader gone", "union.c", ownFunction, "if (adamic_record_values(object) != adamic_field_reference) return false;", "take it off recordTableReaders"},
		// The helper made visible to other files.
		{"table not static", "record.c", "static adamic_map *table(const adamic_record *record) {", "adamic_map *table(const adamic_record *record) {", "table() must stay static"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			sources := recordTableSources(t)
			if found := strings.Count(sources[mutant.file], mutant.before); found != 1 {
				t.Fatalf("the mutant's text is in %s %d times, want once", mutant.file, found)
			}
			sources[mutant.file] = strings.Replace(sources[mutant.file], mutant.before, mutant.after, 1)
			problems := recordTableProblems(sources)
			if !slices.ContainsFunc(problems, func(problem string) bool { return strings.Contains(problem, mutant.refused) }) {
				t.Fatalf("the guard didn't refuse it with %q; it said:\n%s", mutant.refused, strings.Join(problems, "\n"))
			}
		})
	}
}
