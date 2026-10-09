package native

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var runtimeStaticsDirectory = flag.String("runtime-statics-directory", "runtime", "runtime source directory for inventory review")

var listRuntimeStatics = flag.Bool("runtime-statics-list", false, "print the runtime storage inventory")

// Scan source rather than preprocessed C: inactive target and feature branches must
// be audited too. Comments, strings and directives cannot introduce declarations.
var cStorageTokens = regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|[A-Za-z_][A-Za-z_0-9]*|[0-9]+|[^\s]`)

type runtimeStatic struct {
	file, name  string
	line        int
	declaration string
}
type storageToken struct {
	text string
	line int
}

func runtimeStorage(file, source string) []runtimeStatic {
	var tokens []storageToken
	previous, line := 0, 1
	for _, match := range cStorageTokens.FindAllStringIndex(source, -1) {
		text := source[match[0]:match[1]]
		if strings.HasPrefix(text, "//") || strings.HasPrefix(text, "/*") {
			continue
		}
		line += strings.Count(source[previous:match[0]], "\n")
		previous = match[0]
		tokens = append(tokens, storageToken{text, line})
	}
	// Remove complete preprocessor lines, including continuation lines.
	lines := strings.Split(source, "\n")
	directives := map[int]bool{}
	continued := false
	for i, line := range lines {
		if continued || strings.HasPrefix(strings.TrimSpace(line), "#") {
			directives[i+1] = true
			continued = strings.HasSuffix(strings.TrimSpace(line), "\\")
		}
	}
	filtered := tokens[:0]
	for _, token := range tokens {
		if !directives[token.line] {
			filtered = append(filtered, token)
		}
	}
	tokens = filtered
	var found []runtimeStatic
	depth, start := 0, 0
	for i := 0; i < len(tokens); i++ {
		text := tokens[i].text
		if depth == 0 && i == start && text == "typedef" {
			end, nested := i, 0
			for ; end < len(tokens); end++ {
				if tokens[end].text == "{" {
					nested++
				}
				if tokens[end].text == "}" {
					nested--
				}
				if tokens[end].text == ";" && nested == 0 {
					break
				}
			}
			if text == "typedef" {
				i = end
				start = i + 1
				continue
			}
		}
		if text == "static" || (depth == 0 && i == start) {
			end, parens, brackets, braces := i, 0, 0, 0
			initialized, function := false, false
			aggregate, aggregateEnd := false, -1
			for ; end < len(tokens); end++ {
				next := tokens[end].text
				if next == "(" && parens == 0 && braces == 0 && end > i && tokens[end-1].text != "_Atomic" {
					aggregate = false
				}
				if !initialized && parens == 0 && (next == "struct" || next == "enum" || next == "union") {
					aggregate = true
				}
				if next == "=" && parens == 0 && brackets == 0 && braces == 0 {
					initialized = true
				}
				if next == "{" && parens == 0 && brackets == 0 && braces == 0 && !initialized && !aggregate {
					function = true
					break
				}
				if next == ";" && parens == 0 && brackets == 0 && braces == 0 {
					break
				}
				switch next {
				case "(":
					parens++
				case ")":
					parens--
				case "[":
					brackets++
				case "]":
					brackets--
				case "{":
					braces++
				case "}":
					braces--
					if aggregate && braces == 0 && !initialized {
						aggregateEnd = end
					}
				}
			}
			if end < len(tokens) && !function {
				declaration := tokens[i:end]
				if aggregateEnd >= 0 {
					if aggregateEnd+1 == end {
						i = end
						start = i + 1
						continue
					}
					declaration = append([]storageToken{{text: "int", line: tokens[i].line}}, tokens[aggregateEnd+1:end]...)
				}
				found = append(found, storageDeclarators(file, declaration)...)
				i = end
				start = i + 1
				continue
			}
		}
		switch text {
		case "{":
			depth++
		case "}":
			depth--
			if depth == 0 {
				start = i + 1
			}
		case ";":
			start = i + 1
		}
	}
	return found
}

var cStorageIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func storageDeclarators(file string, declaration []storageToken) []runtimeStatic {
	if len(declaration) == 0 {
		return nil
	}
	for _, token := range declaration {
		if token.text == "typedef" {
			return nil
		}
	}
	var parts [][]storageToken
	start, depth := 0, 0
	for i, token := range declaration {
		switch token.text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
		case ",":
			if depth == 0 {
				parts = append(parts, declaration[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, declaration[start:])
	var result []runtimeStatic
	baseConst, baseFunction := false, false
	for partIndex, part := range parts {
		end := len(part)
		for i, token := range part {
			if token.text == "=" {
				end = i
				break
			}
		}
		head := part[:end]
		if len(head) == 0 {
			continue
		}
		// These aliases in adamic.h denote function types, not pointer types.
		// A declaration through either alias is a prototype unless it has a *.
		functionAlias, pointer := baseFunction, false
		for _, token := range head {
			functionAlias = functionAlias || token.text == "adamic_code_function" || token.text == "adamic_counted_code_function"
			pointer = pointer || token.text == "*"
		}
		baseFunction = functionAlias
		if functionAlias && !pointer {
			continue
		}
		// Ordinary function prototypes have no storage. Function pointers do.
		for i, token := range head {
			if token.text == "(" && i > 0 && head[i-1].text != "_Atomic" && head[i-1].text != "__attribute__" {
				if i+1 >= len(head) || head[i+1].text != "*" {
					return nil
				}
				break
			}
		}
		functionPointer := false
		for i := 0; i+1 < len(head); i++ {
			if head[i].text == "(" && head[i+1].text == "*" {
				functionPointer = true
			}
		}
		if !functionPointer {
			for i, token := range head {
				if token.text == "(" && i > 0 && head[i-1].text != "_Atomic" && head[i-1].text != "__attribute__" {
					return nil
				}
			}
		}
		nameIndex, lastStar := -1, -1
		for i, token := range head {
			if token.text == "*" {
				lastStar = i
			}
			if token.text == "[" || (functionPointer && token.text == ")") {
				break
			}
			if cStorageIdentifier.MatchString(token.text) && token.text != "const" && token.text != "volatile" && token.text != "restrict" {
				nameIndex = i
			}
		}
		if nameIndex < 0 {
			continue
		}
		if partIndex == 0 {
			for _, token := range head[:nameIndex] {
				if token.text == "*" {
					break
				}
				if token.text == "const" {
					baseConst = true
				}
			}
		}
		immutable := baseConst
		if lastStar >= 0 {
			immutable = false
			for _, token := range head[lastStar+1 : nameIndex] {
				if token.text == "const" {
					immutable = true
				}
			}
		}
		if immutable {
			continue
		}
		var words []string
		for _, token := range declaration {
			words = append(words, token.text)
		}
		result = append(result, runtimeStatic{file, head[nameIndex].text, head[nameIndex].line, strings.Join(words, " ")})
	}
	return result
}

func TestRuntimeStaticsAreListed(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob(filepath.Join(*runtimeStaticsDirectory, "*"))
	if err != nil {
		t.Fatal(err)
	}
	table, err := os.ReadFile("../../docs/runtime-statics.md")
	if err != nil && !*listRuntimeStatics {
		t.Fatal(err)
	}
	for _, file := range files {
		if filepath.Ext(file) != ".c" && filepath.Ext(file) != ".h" {
			continue
		}
		if !*listRuntimeStatics && !strings.Contains(string(table), "`runtime-file:"+filepath.Base(file)+"`") {
			t.Errorf("%s:1: unaudited runtime file; review its storage in docs/runtime-statics.md", file)
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		occurrences := map[string]int{}
		for _, variable := range runtimeStorage(filepath.Base(file), string(source)) {
			occurrences[variable.name]++
			key := fmt.Sprintf("%s:%s:%d", variable.file, variable.name, occurrences[variable.name])
			if *listRuntimeStatics {
				fmt.Printf("%s\t%d\t%s\n", key, variable.line, variable.declaration)
				continue
			}
			if !strings.Contains(string(table), "`"+key+"`") {
				t.Errorf("%s:%d: unlisted mutable storage %s; audit it in docs/runtime-statics.md", file, variable.line, variable.name)
			}
		}
	}
}

func TestRuntimeStorageScanner(t *testing.T) {
	t.Parallel()
	source := `// static int ignored;
#define HIDDEN static int hidden;
static const char message[] = "static int ignored;";
static const char *pointer;
static char *const fixed = 0;
int global;
static int first, second;
void work(void) { static int local; int automatic; }
static int function(void);
static int (*callback)(void);
typedef struct fields { int field; } fields;
void prototype(void (*argument)(void));
static int array[2] = {1, 2};
struct pair { int field; } record;
void more(void) { static struct { int field; } local_record; }
static adamic_code_function declared_function;
static adamic_counted_code_function counted_function;
static adamic_code_function *function_pointer;
static adamic_code_function *mixed_pointer, mixed_prototype;
static adamic_code_function first_prototype, *second_pointer;
static int *const fixed_pointer = 0, scalar;
static struct pair *maker(void) { static int nested; return 0; }`
	var names []string
	for _, variable := range runtimeStorage("probe.c", source) {
		names = append(names, variable.name)
	}
	if got := strings.Join(names, ","); got != "pointer,global,first,second,local,callback,array,record,local_record,function_pointer,mixed_pointer,second_pointer,scalar,nested" {
		t.Fatalf("storage inventory = %s", got)
	}
}
