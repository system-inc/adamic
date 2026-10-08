package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are the four rule rows and all six lead values added by c2e39b75.
// The changed hyphen configuration is also held to every valid UTF-8 lead value.
func warningJavaScriptRows(t *testing.T) []string {
	t.Helper()
	var rows []string
	add := func(name, source string, options any) {
		path := filepath.Join(t.TempDir(), name+".ts")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(options)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\tno-warning-comments\t\t\tfalse\t"+string(encoded))
	}
	add("upstream-nbsp", "//\u00a0TODO later", nil)
	add("upstream-bom", "//\ufeffTODO later", nil)
	decoration := map[string]any{"decoration": []string{"*", "-", "/"}}
	add("upstream-hyphen", "/*- todo */", decoration)
	add("upstream-plus", "/*+ todo */", decoration)
	configurations := []any{nil, map[string]any{"decoration": []string{"*", "/"}}, decoration,
		map[string]any{"decoration": []string{"k"}},
		map[string]any{"terms": []string{"*todo", " fixme", "kelvin", "ſtop", "日本"}, "decoration": []string{"*"}},
		map[string]any{"terms": []string{""}}, map[string]any{"location": "anywhere"}}
	added := []string{"\u00a0todo", "\ufeff todo", "\u2028todo", "\u3000 fixme", "- todo", "-+ todo"}
	for i, options := range configurations {
		for j, value := range added {
			add(fmt.Sprintf("upstream-lead-%d-%d", i, j), "/*"+value+"*/", options)
		}
	}
	previous := []string{"", " ", "\t\n", " TODO: x", "todo", "Todo later", "  fixme", "xxx", "XXX!", "something todo",
		"\v todo", " todo", "* todo", "** todo", "*/ todo", "+ todo", ", todo", "k todo", "K todo",
		"K todo", "Kelvin", "Kelvin", "ſtop", "Stop", "stop", "日本", "todos", "*todo", "* *todo", " fixme", "  fixme"}
	for j, value := range previous {
		source := "/*" + value + "*/"
		if strings.Contains(value, "*/") {
			source = "//" + value
		}
		add(fmt.Sprintf("upstream-changed-range-%d", j), source, decoration)
	}
	// All 25 JavaScript whitespace characters, plus Go-only NEL and boundary folds.
	for _, point := range []rune{9, 10, 11, 12, 13, 32, 160, 5760, 8192, 8193, 8194, 8195, 8196, 8197, 8198, 8199, 8200, 8201, 8202, 8232, 8233, 8239, 8287, 12288, 65279, 133} {
		add(fmt.Sprintf("whitespace-%04x", point), "/*"+string(point)+"TODO later*/", nil)
		add(fmt.Sprintf("quote-%04x", point), "/*TODO"+string(point)+"later*/", nil)
		add(fmt.Sprintf("directive-%04x", point), "/*"+string(point)+"eslint no-warning-comments*/", map[string]any{"terms": []string{"eslint"}, "location": "anywhere"})
	}
	for i, value := range []string{"ſTODO", "TODOſ", "KTODO", "TODOK", "ſ", "K"} {
		add(fmt.Sprintf("iu-boundary-%d", i), "// "+value, map[string]any{"terms": []string{"todo", "s", "k"}, "location": "anywhere"})
	}
	return rows
}

func TestWarningCommentsJavaScriptFixtures(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	rows := warningJavaScriptRows(t)
	path := manifest(t, rows)
	want := compare(t, goOracle(t), buildPort(t, directory, true), directory, path)
	parts := strings.Split(string(want), "case ")
	for index, row := range rows {
		if index+1 >= len(parts) {
			t.Fatal("oracle case missing")
		}
		t.Logf("%s: PASS, %d findings", filepath.Base(strings.Split(row, "\t")[0]), strings.Count(parts[index+1], "  no-warning-comments  "))
	}
	t.Logf("%d JavaScript fixture rows agree byte for byte", len(rows))
	// Case 108 is retained in generated(), with its original row number and source.
	case108 := manifest(t, generated(t)[108:109])
	compare(t, goOracle(t), buildPort(t, directory, true), directory, case108)
	t.Logf("case 108 new Go count: %s", execute(t, "", goOracle(t), "--manifest", case108, "--count").output)
}

func TestWarningCommentsWhitespaceReversion(t *testing.T) {
	t.Parallel()
	path := manifest(t, warningJavaScriptRows(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	directory := mutant(t, "!space(character)", `![' ', '\t', '\n', '\r', '\f'].includes(character)`, "comments.ts")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)}, {"emitted JavaScript", emittedNode(t, directory, path, false)},
		{"native", execute(t, "", binary, "--manifest", path)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("whitespace reversion survived on %s", side.name)
		}
		t.Logf("Go-five whitespace reversion caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

// A complete reversion to the base's matcher must compile and run, then disagree.
func TestWarningCommentsCompleteReversion(t *testing.T) {
	t.Parallel()
	current, err := os.ReadFile("comments.ts")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := os.ReadFile("rules/no-warning-comments/evidence/base-comments.txt")
	if err != nil {
		t.Fatal(err)
	}
	directory := mutant(t, string(current), string(previous), "comments.ts")
	rows := append(warningJavaScriptRows(t), generated(t)[108])
	path := manifest(t, rows)
	want := execute(t, "", goOracle(t), "--manifest", path).output
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)}, {"emitted JavaScript", emittedNode(t, directory, path, false)},
		{"native", execute(t, "", binary, "--manifest", path)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("complete reversion survived on %s", side.name)
		}
		t.Logf("complete base reversion caught on %s: %s", side.name, difference(side.run.output, want))
	}
}
