package estree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
)

// prettierParseScript is the oracle: the fork's own parse (parser "typescript", with the file path, so
// JSX and sourceType are decided the way Prettier decides them for a real file), dumped in the shape
// dumpValue prints. Offsets are UTF-16 there, and the Go side converts to match.
const prettierParseScript = `
import * as prettier from "./src/index.js";
import fs from "node:fs";
const skip = new Set(["loc","range","comments","tokens","parent","start","end","extra"]);
function dump(v) {
  if (v === null || v === undefined) return null;
  if (Array.isArray(v)) return v.map(dump);
  if (typeof v === "bigint" || v instanceof RegExp) return null;
  if (typeof v !== "object") return v;
  if (v.type === undefined) { const o = {}; for (const k of Object.keys(v).sort()) o[k] = dump(v[k]); return o; }
  const o = { type: v.type, range: [v.range[0], v.range[1]] };
  if (v.__contentEnd !== undefined) o.__contentEnd = v.__contentEnd;
  for (const k of Object.keys(v).sort()) { if (skip.has(k) || k === "type" || k.startsWith("__")) continue; o[k] = dump(v[k]); }
  if (o.type === "Literal" && (v.regex || v.bigint)) o.value = null;
  return o;
}
let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", async () => {
  for (const f of input.split("\n").filter(Boolean)) {
    const text = fs.readFileSync(f, "utf8");
    try {
      const { ast } = await prettier.__debug.parse(text, { parser: "typescript", filepath: f });
      const comments = (ast.comments ?? []).map(c => ({ type: c.type, range: c.range, value: c.value }));
      process.stdout.write(JSON.stringify({ file: f, ast: dump(ast), comments }) + "\n");
    } catch (e) { process.stdout.write(JSON.stringify({ file: f, error: String(e.message).split("\n")[0] }) + "\n"); }
  }
});
`

// TestConvertAgreesWithPrettier is the converter's acceptance test: for every .ts and .tsx file in the
// corpora, the Go tree and comment list must equal what the fork's own parse produces, property for
// property. Off unless COHERE_PRETTIER_FORK names the fork and COHERE_ESTREE_CORPORA lists repository
// roots (colon-separated; each is enumerated with git ls-files, so a nested repository is its own root).
//
// A file the fork cannot parse has no answer to match and is skipped, but counted, and the test fails
// if that count passes a tenth of the corpus: an oracle that rejects everything would otherwise read as
// perfect agreement.
func TestConvertAgreesWithPrettier(t *testing.T) {
	t.Parallel()
	corpora := os.Getenv("COHERE_ESTREE_CORPORA")
	if corpora == "" {
		t.Skip("set COHERE_ESTREE_CORPORA to compare the converter against the fork")
	}
	root := corpus.PrettierFork.Root(t)
	seen := map[string]bool{}
	var files []string
	for _, corpus := range strings.Split(corpora, ":") {
		corpus = strings.TrimSpace(corpus)
		if strings.HasPrefix(corpus, "~/") {
			home, _ := os.UserHomeDir()
			corpus = filepath.Join(home, corpus[2:])
		}
		listing, err := exec.Command("git", "-C", corpus, "ls-files", "--", "*.ts", "*.tsx").Output()
		if err != nil {
			t.Fatalf("listing %s: %v", corpus, err)
		}
		for _, name := range strings.Split(string(listing), "\n") {
			path := filepath.Join(corpus, name)
			if name == "" || seen[path] {
				continue
			}
			if _, err := os.Stat(path); err != nil {
				continue
			}
			seen[path] = true
			files = append(files, path)
		}
	}

	command := exec.Command("node", "--input-type=module", "-e", prettierParseScript)
	command.Dir = root
	command.Stdin = strings.NewReader(strings.Join(files, "\n"))
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	oracle := map[string]map[string]any{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		oracle[record["file"].(string)] = record
	}

	identical, refused, different := 0, 0, 0
	for _, path := range files {
		want, present := oracle[path]
		if !present {
			t.Fatalf("the oracle printed nothing for %s", path)
		}
		if _, failed := want["error"]; failed {
			refused++
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got, err := dumpFile(path, string(source))
		if err != nil {
			different++
			t.Errorf("%s: the fork parses it and the converter refused: %v", path, err)
			continue
		}
		for _, key := range []string{"ast", "comments"} {
			if difference := firstDifference(want[key], got[key], key); difference != "" {
				different++
				if different <= 10 {
					t.Errorf("%s: %s", path, difference)
				}
				break
			}
		}
		if difference := firstDifference(want["ast"], got["ast"], "ast") + firstDifference(want["comments"], got["comments"], "comments"); difference == "" {
			identical++
		}
	}
	t.Logf("%d files, %d identical, %d different, %d the fork could not parse", len(files), identical, different, refused)
	if refused > len(files)/10 {
		t.Errorf("the fork refused %d of %d files; the comparison is thinner than it looks", refused, len(files))
	}
}

// dumpFile converts one file and renders it the way the oracle script does, through JSON, so both sides
// compare as the same Go values.
func dumpFile(path string, text string) (map[string]any, error) {
	program, comments, _, err := ParseTypeScript(path, text, nil)
	if err != nil {
		return nil, err
	}
	offsets := utf16Offsets(text)
	dumped := make([]any, len(comments))
	for index, comment := range comments {
		dumped[index] = map[string]any{
			"type":  comment.Type(),
			"range": []int{offsets[comment.Range[0]], offsets[comment.Range[1]]},
			"value": comment.String("value"),
		}
	}
	encoded, err := json.Marshal(map[string]any{"ast": dumpValue(program, offsets), "comments": dumped})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	err = json.Unmarshal(encoded, &result)
	return result, err
}

// firstDifference names the first path where two decoded JSON values differ, or "".
func firstDifference(want any, got any, path string) string {
	switch typed := want.(type) {
	case map[string]any:
		other, isMap := got.(map[string]any)
		if !isMap {
			return fmt.Sprintf("%s: want an object, got %v", path, got)
		}
		keys := map[string]bool{}
		for key := range typed {
			keys[key] = true
		}
		for key := range other {
			keys[key] = true
		}
		sorted := make([]string, 0, len(keys))
		for key := range keys {
			sorted = append(sorted, key)
		}
		sort.Strings(sorted)
		for _, key := range sorted {
			wantValue, wantPresent := typed[key]
			gotValue, gotPresent := other[key]
			if wantPresent != gotPresent {
				return fmt.Sprintf("%s.%s: present in the fork %v, in Go %v (%s)", path, key, wantPresent, gotPresent, typed["type"])
			}
			if difference := firstDifference(wantValue, gotValue, path+"."+key); difference != "" {
				return difference
			}
		}
		return ""
	case []any:
		other, isList := got.([]any)
		if !isList || len(other) != len(typed) {
			return fmt.Sprintf("%s: want %d elements, got %v", path, len(typed), got)
		}
		for index := range typed {
			if difference := firstDifference(typed[index], other[index], fmt.Sprintf("%s[%d]", path, index)); difference != "" {
				return difference
			}
		}
		return ""
	}
	if fmt.Sprint(want) != fmt.Sprint(got) {
		return fmt.Sprintf("%s: want %v, got %v", path, want, got)
	}
	return ""
}

// utf16Offsets maps every byte offset that starts a character (and the end) to its UTF-16 index.
func utf16Offsets(text string) []int {
	offsets := make([]int, len(text)+1)
	index := 0
	for position, character := range text {
		offsets[position] = index
		size := utf8.RuneLen(character)
		for continuation := 1; continuation < size; continuation++ {
			offsets[position+continuation] = index
		}
		index += len(utf16.Encode([]rune{character}))
	}
	offsets[len(text)] = index
	return offsets
}

func dumpValue(value any, offsets []int) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case *Node:
		if typed == nil {
			return nil
		}
		result := map[string]any{"type": typed.Type(), "range": []int{offsets[typed.Range[0]], offsets[typed.Range[1]]}}
		if typed.HasContentEnd {
			result["__contentEnd"] = offsets[typed.ContentEnd]
		}
		keys := typed.Keys()
		sort.Strings(keys)
		for _, key := range keys {
			if key == "comments" || key == "tokens" || strings.HasPrefix(key, "__") {
				continue
			}
			result[key] = dumpValue(typed.Get(key), offsets)
		}
		return result
	case []*Node:
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = dumpValue(child, offsets)
		}
		return result
	case *TemplateValue:
		var cooked any
		if typed.Cooked != nil {
			cooked = *typed.Cooked
		}
		return map[string]any{"cooked": cooked, "raw": typed.Raw}
	case *Regex:
		return map[string]any{"flags": typed.Flags, "pattern": typed.Pattern}
	default:
		return typed
	}
}
