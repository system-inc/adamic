package native

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var reactionFunction = regexp.MustCompile(`(?m)^[A-Za-z_][^;{}]*\([^;{}]*\)\s*\{`)
var reactionPromise = regexp.MustCompile(`\badamic_(?:async_promise|promise)\s*\*\s*([A-Za-z_][A-Za-z_0-9]*)`)
var reactionAllocation = regexp.MustCompile(`adamic_allocate\([^;]*\badamic_kind_async_reaction\b`)

// Discover functions and promise receivers from source, including inactive
// feature branches. Clearing or moving an existing list is not attachment.
func reactionWriterFailures(source string) (writers, failures []string) {
	for _, signature := range reactionFunction.FindAllStringIndex(source, -1) {
		start := signature[1] - 1
		depth, end := 0, start
		for _, token := range cStorageTokens.FindAllStringIndex(source[start:], -1) {
			text := source[start+token[0] : start+token[1]]
			if text == "{" {
				depth++
			} else if text == "}" {
				depth--
				if depth == 0 {
					end = start + token[1]
					break
				}
			}
		}
		body := source[signature[0]:end]
		allocation := reactionAllocation.FindStringIndex(body)
		for _, promise := range reactionPromise.FindAllStringSubmatch(body, -1) {
			name := promise[1]
			store := regexp.MustCompile(`\b` + name + `->(?:first|last)\s*=\s*([^=;][^;]*);`)
			first := len(body)
			for _, write := range store.FindAllStringSubmatchIndex(body, -1) {
				if !strings.Contains(body[write[2]:write[3]], "NULL") {
					first = min(first, write[0])
				}
			}
			if allocation != nil {
				first = min(first, allocation[0])
			}
			if first == len(body) {
				continue
			}
			label := strings.TrimSpace(source[signature[0]:start]) + " receiver " + name
			writers = append(writers, label)
			mark := regexp.MustCompile(`(?m)^\s*` + name + `->handled = true;`)
			if !mark.MatchString(body[:first]) {
				failures = append(failures, label)
			}
		}
	}
	return
}

func TestAsyncReactionWritersMarkHandled(t *testing.T) {
	t.Parallel()
	writers := 0
	// Reads: internal/native_runtime/c/*
	err := filepath.WalkDir("../native_runtime/c", func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || (filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h") {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		found, failures := reactionWriterFailures(string(source))
		writers += len(found)
		for _, writer := range found {
			t.Logf("%s: %s", path, writer)
		}
		for _, failure := range failures {
			t.Errorf("%s: reaction writer does not mark handled before attachment: %s", path, failure)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if writers == 0 {
		t.Fatal("reaction writer scan found no attachment")
	}
	// Scan emitted text too. A future emitter that writes reactions must keep
	// the handled write alongside its attachment rather than bypass this pin.
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			var text strings.Builder
			ast.Inspect(function, func(node ast.Node) bool {
				if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING {
					value, err := strconv.Unquote(literal.Value)
					if err != nil {
						t.Fatal(err)
					}
					text.WriteString(value)
					text.WriteByte('\n')
				}
				return true
			})
			emitted := text.String()
			links := regexp.MustCompile(`->(?:first|last)\s*=\s*([^=;][^;]*);`)
			attaches := strings.Contains(emitted, "adamic_kind_async_reaction") || strings.Contains(emitted, "adamic_async_reaction")
			for _, link := range links.FindAllStringSubmatch(emitted, -1) {
				attaches = attaches || !strings.Contains(link[1], "NULL")
			}
			if attaches {
				if !strings.Contains(emitted, "->handled = true;") {
					t.Errorf("%s: %s emits reactions without marking handled", path, function.Name.Name)
				}
			}
		}
	}
}

func TestAsyncReactionWriterPinFindsNewAttachment(t *testing.T) {
	t.Parallel()
	body := "void attach(adamic_async_promise *other, adamic_async_reaction *entry) {\n other->last = entry;\n}\n"
	writers, failures := reactionWriterFailures(body)
	if len(writers) != 1 || len(failures) != 1 {
		t.Fatalf("unmarked new writer escaped: writers %v failures %v", writers, failures)
	}
	body = strings.Replace(body, "other->last", "other->handled = true;\n other->last", 1)
	_, failures = reactionWriterFailures(body)
	if len(failures) != 0 {
		t.Fatalf("marked attachment rejected: %v", failures)
	}
}
