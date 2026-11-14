package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Probe only, never for landing (#g4arkr4 step 1): which construct of a rest parameter's body emits C that
// clang refuses. Each program prints Node's output, whether it lowers, and where native compile fails,
// clang's first error lines, the whole emitted C function clang names, every line naming the undeclared
// identifier, and the lowered IR of the program's own functions.
func TestG4aProbe(t *testing.T) {
	directory := "internal/oracle/testdata/g4a_probe"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	mainLine := regexp.MustCompile(`main\.c:([0-9]+):`)
	undeclared := regexp.MustCompile(`undeclared identifier '([A-Za-z0-9_]+)'`)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ts" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			name := strings.TrimSuffix(entry.Name(), ".ts")
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			t.Logf("PROBE %s node exit=%d stdout=%q stderr=%q", name, truth.exitCode, cut(string(truth.stdout), 200), cut(string(truth.stderr), 200))
			program, err := lowered(t, absolute)
			if err != nil {
				t.Logf("PROBE %s lower stops: %s", name, cut(strings.ReplaceAll(err.Error(), "\n", " / "), 400))
				return
			}
			t.Logf("PROBE %s lowers", name)
			code := native.C(program)
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				message := err.Error()
				clangErrors := 0
				for _, line := range strings.Split(message, "\n") {
					if strings.Contains(line, "error:") && clangErrors < 3 {
						clangErrors++
						t.Logf("PROBE %s clang: %s", name, cut(strings.TrimSpace(line), 400))
					}
				}
				if clangErrors == 0 {
					t.Logf("PROBE %s clang (no error: line): %s", name, cut(strings.ReplaceAll(message, "\n", " / "), 400))
				}
				lines := strings.Split(code, "\n")
				if match := mainLine.FindStringSubmatch(message); match != nil {
					number, err := strconv.Atoi(match[1])
					if err != nil {
						t.Fatal(err)
					}
					// The function holding clang's line: back to the nearest line at column 0 that opens a
					// brace, forward to the first line that is a closing brace alone.
					start := number - 1
					for start > 0 && !(len(lines[start]) > 0 && lines[start][0] != '\t' && lines[start][0] != ' ' && strings.HasSuffix(strings.TrimSpace(lines[start]), "{")) {
						start--
					}
					end := number - 1
					for end < len(lines)-1 && lines[end] != "}" {
						end++
					}
					if end-start > 200 {
						t.Logf("PROBE %s C function %d-%d is over 200 lines, printing its first 200", name, start+1, end+1)
						end = start + 199
					}
					for index := start; index <= end; index++ {
						t.Logf("PROBE %s C %d: %s", name, index+1, cut(lines[index], 400))
					}
				} else {
					t.Logf("PROBE %s clang names no main.c line", name)
				}
				if match := undeclared.FindStringSubmatch(message); match != nil {
					mentions := regexp.MustCompile(regexp.QuoteMeta(match[1]) + `\b`)
					prefixed := 0
					for index, line := range lines {
						if mentions.MatchString(line) {
							t.Logf("PROBE %s names %s at C %d: %s", name, match[1], index+1, cut(strings.TrimSpace(line), 400))
						} else if strings.Contains(line, match[1]) {
							prefixed++
						}
					}
					t.Logf("PROBE %s lines holding %s as a prefix only: %d", name, match[1], prefixed)
				}
				// The program's own functions, by their file; a function with no position is printed when it's a
				// closure or takes a rest, at most eight in all.
				printed := 0
				for index, function := range program.Functions {
					if !strings.HasSuffix(function.Position.File, entry.Name()) && !(function.Position.File == "" && (function.Closure || function.RestElement != 0)) {
						continue
					}
					if printed++; printed > 8 {
						t.Logf("PROBE %s IR more functions than eight, stopping", name)
						break
					}
					t.Logf("PROBE %s IR function %d %q at %d:%d closure=%v parameters=%v rest=%d returns=%d environment=%v mayThrow=%v", name, index, function.Name, function.Position.Line, function.Position.Column, function.Closure, function.Parameters, function.RestElement, function.Returns, function.Environment, function.MayThrow)
					for _, local := range function.Parameters {
						if local >= 0 && local < len(program.Locals) {
							t.Logf("PROBE %s IR function %d local %d: %s", name, index, local, cut(irText(reflect.ValueOf(program.Locals[local])), 400))
						}
					}
					for statement, value := range function.Body {
						t.Logf("PROBE %s IR function %d statement %d: %s", name, index, statement, cut(irText(reflect.ValueOf(value)), 1800))
					}
				}
				return
			}
			result := execute(t, binary)
			t.Logf("PROBE %s sanitized exit=%d stdout=%q stderr=%q", name, result.exitCode, cut(string(result.stdout), 200), cut(string(result.stderr), 200))
			script := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript exit=%d stdout=%q", name, script.exitCode, cut(string(script.stdout), 200))
		})
	}
}

// irText writes an IR value compactly: each struct by its type's name with only its non-zero fields,
// pointers and interfaces followed to what they hold.
func irText(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Invalid:
		return "nil"
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return "nil"
		}
		return irText(value.Elem())
	case reflect.Struct:
		var fields []string
		for index := range value.NumField() {
			field := value.Type().Field(index)
			if !field.IsExported() || value.Field(index).IsZero() {
				continue
			}
			fields = append(fields, field.Name+":"+irText(value.Field(index)))
		}
		return value.Type().Name() + "{" + strings.Join(fields, " ") + "}"
	case reflect.Slice, reflect.Array:
		var items []string
		for index := range value.Len() {
			items = append(items, irText(value.Index(index)))
		}
		return "[" + strings.Join(items, " ") + "]"
	case reflect.Map:
		return fmt.Sprintf("map(%d)", value.Len())
	case reflect.Func, reflect.Chan:
		return value.Type().String()
	}
	return fmt.Sprint(value.Interface())
}

// cut keeps a probe line compact.
func cut(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}
