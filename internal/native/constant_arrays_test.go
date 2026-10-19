package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strings"
	"testing"
	"time"
)

// constantArraysSource builds a flat table and a table of rows past constantNumbers, with the
// numbers a table must keep exactly (NaN, both infinities, -0, the smallest subnormal, the largest
// double), in main and in a function, each array fresh: one changed later changes no other.
func constantArraysSource() string {
	flat := []string{"0", "-0", "NaN", "Infinity", "-Infinity", "5e-324", "1.7976931348623157e308", "0.1", "-2.5"}
	for value := 0; len(flat) < constantNumbers+6; value++ {
		flat = append(flat, fmt.Sprint(value*7-30))
	}
	rows := []string{"[]"}
	for row := 1; len(rows) < 40; row++ {
		values := []string{}
		for column := 0; column < row%5; column++ {
			values = append(values, fmt.Sprint(row*10+column))
		}
		rows = append(rows, "["+strings.Join(values, ", ")+"]")
	}
	return fmt.Sprintf(`const flat = [%s];
const rows = [%s];
function table() {
	return [%s];
}
flat.push(1);
for (const row of rows) {
	if (row.length === 2) {
		row.push(3);
	}
}
const first = table();
const second = table();
const changed = first[1];
if (changed !== undefined) {
	changed.push(4);
}
console.log(`+"`${flat.length} ${rows.length} ${1 / (flat[1] ?? 1)} ${flat[2] ?? 0} ${flat[5] ?? 0} ${flat[6] ?? 0}`"+`);
console.log(flat.join(','));
console.log(rows.map((row) => row.join(' ')).join('|'));
console.log(`+"`${first[1]?.length ?? -1} ${second[1]?.length ?? -1} ${first.length}`"+`);
`, strings.Join(flat, ", "), strings.Join(rows, ", "), strings.Join(rows, ", "))
}

// Constant tables print what Node prints, under the sanitizers, and let go of every array they make.
// Three mutants of the loops must each be seen: the flat loop one short, a row read from the wrong
// place in the table, and a row the outer array retains instead of taking.
func TestConstantArraysAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	source := constantArraysSource()
	path := filepath.Join(directory, "tables.ts")
	script := filepath.Join(directory, "tables.mjs")
	for _, file := range []string{path, script} {
		if err := os.WriteFile(file, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want := runWithInput(t, "", "node", script)
	code := C(loopProgram(t, path))
	if strings.Count(code, "static const double adamic_numbers_") != 3 || strings.Count(code, "static const size_t adamic_lengths_") != 2 {
		t.Fatalf("the tables weren't emitted as data:\n%s", code)
	}
	if got, failed := runConstantArrays(t, code); failed != "" || got != want {
		t.Fatalf("native %q (%s); Node %q", got, failed, want)
	}
	flatBound := regexp.MustCompile(`(for \(size_t adamic_temporary_\d+ = 0; adamic_temporary_\d+ < )(\d+)(;)`)
	rowRead := regexp.MustCompile(`(\.number = adamic_numbers_\d+\[)(adamic_temporary_\d+)(\]\})`)
	rowTaken := regexp.MustCompile(`(\(adamic_value\)\{\.reference = )(adamic_temporary_\d+)(\}\);\n\t*\})`)
	for _, mutant := range []struct {
		name    string
		pattern *regexp.Regexp
		replace string
	}{
		{"flat loop one short", flatBound, "${1}${2} - 1${3}"},
		{"table read at index 0", rowRead, "${1}0${3}"},
		{"row retained", rowTaken, "${1}adamic_retain(${2})${3}"},
	} {
		if !mutant.pattern.MatchString(code) {
			t.Fatalf("mutant %q lost its anchor", mutant.name)
		}
		got, failed := runConstantArrays(t, mutant.pattern.ReplaceAllString(code, mutant.replace))
		if failed == "" && got == want {
			if mutant.name == "row retained" && goruntime.GOOS != "linux" {
				// The leak shows to LeakSanitizer, which only Linux's sanitized builds run.
				continue
			}
			t.Fatalf("mutant %q survived", mutant.name)
		}
	}
}

// runConstantArrays builds code under the sanitizers and runs it, returning what it printed and,
// when it failed or leaked, why.
func runConstantArrays(t *testing.T, code string) (string, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "tables")
	if err := Build(code, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := programCommand(ctx, "", binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() > 0 {
		return stdout.String(), fmt.Sprintf("%v: %s", err, stderr.String())
	}
	return stdout.String(), ""
}
