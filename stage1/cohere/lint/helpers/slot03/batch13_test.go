package slot03

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func batch13Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch13/testdata", []string{"/collapse.InferDataType", "/collapse.matchesDataType", "/collapse.isFamilyName"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch13/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch13.go")
	overlay := filepath.Join(directory, "overlay.json")

	raw, err := os.ReadFile(filepath.Join(root, "internal/lint/rules/tailwind/collapse/data_type.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	start := strings.Index(text, "func matchesDataType(")
	end := strings.Index(text[start:], "\n/* -------------------------------------------------------------------------- */") + start
	body := text[start:end]
	anchor := "switch dataType {"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go dispatch anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording && adamicRecordTypes {adamicTrace+=\"type:\"+string(dataType)+\";\"}; "+anchor, 1)
	anchor = "return IsColor(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:0;\"}; "+anchor, 1)
	anchor = "return IsLength(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:1;\"}; "+anchor, 1)
	anchor = "return isPercentage(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:2;\"}; "+anchor, 1)
	anchor = "return isFraction(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:3;\"}; "+anchor, 1)
	anchor = "return isNumber(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:4;\"}; "+anchor, 1)
	anchor = "return IsPositiveInteger(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:5;\"}; "+anchor, 1)
	anchor = "return isURL(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:6;\"}; "+anchor, 1)
	anchor = "return isBackgroundPosition(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:7;\"}; "+anchor, 1)
	anchor = "return isBackgroundSize(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:8;\"}; "+anchor, 1)
	anchor = "return isLineWidth(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:9;\"}; "+anchor, 1)
	anchor = "return isImage(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:10;\"}; "+anchor, 1)
	anchor = "return isFamilyName(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:11;\"}; "+anchor, 1)
	anchor = "return isGenericName(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:12;\"}; "+anchor, 1)
	anchor = "return isAbsoluteSize(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:13;\"}; "+anchor, 1)
	anchor = "return isRelativeSize(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:14;\"}; "+anchor, 1)
	anchor = "return isAngle(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:15;\"}; "+anchor, 1)
	anchor = "return isVector(value)"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go predicate anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecording {adamicTrace+=\"check:16;\"}; "+anchor, 1)
	text = text[:start] + body + text[end:]
	start = strings.Index(text, "func isFamilyName(")
	end = strings.Index(text[start:], "\nvar absoluteSizes") + start
	body = text[start:end]
	anchor = "for _, part := range segment(value, ',') {"
	if strings.Count(body, anchor) != 1 {
		t.Fatal("Go family anchor changed")
	}
	body = strings.Replace(body, anchor, "if adamicRecordFamily {adamicTrace+=\"segment:,;\"}; "+anchor, 1)
	text = text[:start] + body + text[end:]
	modified := filepath.Join(directory, "data_type.go")
	if err = os.WriteFile(modified, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch13.go"): filepath.Join(here, "collapse_export.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/data_type.go"): modified}})

	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	command(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.json")
	want := command(t, "", binary, filepath.Join(here, "sources.jsonl.gz"), cases)
	return cases, want
}

// Not parallel: native sanitizer builds and large fixture observations bound memory.
func TestBatch13Helpers(t *testing.T) {
	cases, want := batch13Oracle(t)
	entry, _ := filepath.Abs("batch13/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch13JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch13Mutants(t *testing.T) {
	cases, want := batch13Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"matches_data_type.a", "case 'color': return check(value, 0);", "case 'color': return check(value, 1);", ""},
		{"matches_data_type.a", "case 'length': return check(value, 1);", "case 'length': return check(value, 2);", ""},
		{"matches_data_type.a", "case 'percentage': return check(value, 2);", "case 'percentage': return check(value, 3);", ""},
		{"matches_data_type.a", "case 'ratio': return check(value, 3);", "case 'ratio': return check(value, 4);", ""},
		{"matches_data_type.a", "case 'number': return check(value, 4);", "case 'number': return check(value, 5);", ""},
		{"matches_data_type.a", "case 'integer': return check(value, 5);", "case 'integer': return check(value, 6);", ""},
		{"matches_data_type.a", "case 'url': return check(value, 6);", "case 'url': return check(value, 7);", ""},
		{"matches_data_type.a", "case 'position': return check(value, 7);", "case 'position': return check(value, 8);", ""},
		{"matches_data_type.a", "case 'bg-size': return check(value, 8);", "case 'bg-size': return check(value, 9);", ""},
		{"matches_data_type.a", "case 'line-width': return check(value, 9);", "case 'line-width': return check(value, 10);", ""},
		{"matches_data_type.a", "case 'image': return check(value, 10);", "case 'image': return check(value, 11);", ""},
		{"matches_data_type.a", "case 'family-name': return check(value, 11);", "case 'family-name': return check(value, 12);", ""},
		{"matches_data_type.a", "case 'generic-name': return check(value, 12);", "case 'generic-name': return check(value, 13);", ""},
		{"matches_data_type.a", "case 'absolute-size': return check(value, 13);", "case 'absolute-size': return check(value, 14);", ""},
		{"matches_data_type.a", "case 'relative-size': return check(value, 14);", "case 'relative-size': return check(value, 15);", ""},
		{"matches_data_type.a", "case 'angle': return check(value, 15);", "case 'angle': return check(value, 16);", ""},
		{"matches_data_type.a", "case 'vector': return check(value, 16);", "case 'vector': return check(value, 0);", ""},
		{"matches_data_type.a", "default: return false;", "default: return check(value, 0);", ""},
		{"matches_data_type.a", "check(value, 0)", "check(value + 'x', 0)", ""},
		{"infer_data_type.a", "/^var\\(/u.test(value)", "false", ""},
		{"infer_data_type.a", "/^var\\(/u.test(value)", "/^var\\(/iu.test(value)", ""},
		{"infer_data_type.a", "for (const dataType of types)", "for (const dataType of types.slice().reverse())", ""},
		{"infer_data_type.a", "matches(value, dataType)", "matches(value + 'x', dataType)", ""},
		{"is_family_name.a", "/^[0-9]/u", "/^[1-9]/u", ""},
		{"is_family_name.a", "/^var\\(/u", "/^VAR\\(/u", ""},
		{"is_family_name.a", "count++;", "count += 0;", ""},
		{"is_family_name.a", "return count > 0;", "return count >= 0;", ""},
		{"is_family_name.a", "segment(value, ',')", "segment(value, ';')", ""},
		{"is_family_name.a", "segment(value, ',')", "segment(value + 'x', ',')", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"infer_data_type.a", "matches_data_type.a", "is_family_name.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch13", file))
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if file == mutant.file {
					if strings.Count(text, mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					text = mutant.prefix + strings.Replace(text, mutant.old, mutant.new, 1)
				}
				if file == "main.a" {
					reader, _ := filepath.Abs("../options_json.ts")
					text = strings.ReplaceAll(text, "../../options_json.ts", filepath.ToSlash(reader))
				}
				if err = os.WriteFile(filepath.Join(scratch, file), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := command(t, "", build(t, filepath.Join(scratch, "main.a")), cases)
			if bytes.Equal(got, want) {
				t.Fatal("semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q Go %q", i+1, a[i], b[i])
					return
				}
			}
			t.Fatal("no changed output line")
		})
	}
}

func batch13JavaScript(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "helper.mjs")
	if err = os.WriteFile(path, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
