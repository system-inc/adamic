package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// A probe for #7mjgxkc's stop, never landed: roads that copy a class instance's own properties into a value typed
// with its getter, other than one spread read directly. Each row prints whether it lowers, the stop's words if it
// doesn't, and native's output beside Node's.
func TestSpreadCopyProbe(t *testing.T) {
	t.Parallel()
	head := "interface Source { readonly label: string; readonly weight: number }\nclass Sized { readonly label = \"t\"; get weight(): number { return 20; } }\nconst honest: Source = { label: \"h\", weight: 2 };\nconst built: Source = new Sized();\nconsole.log(typeof honest.weight);\n"
	cases := []struct{ name, source string }{
		{"copy of a copy", head + "const inner: Source = { ...built };\nconst value: Source = { ...inner };\nconsole.log(typeof value.weight + \" \" + String(value.weight));\n"},
		{"copy of a copy, inline", head + "const value: Source = { ...{ ...built } };\nconsole.log(typeof value.weight + \" \" + String(value.weight));\n"},
		{"Object.assign onto a literal", head + "const value: Source = Object.assign({ label: \"a\", weight: 1 }, built);\nconsole.log(typeof value.weight + \" \" + String(value.weight));\n"},
		{"copy handed through a function", head + "function pass(s: Source): Source { return s; }\nconst value = pass({ ...built });\nconsole.log(typeof value.weight + \" \" + String(value.weight));\n"},
		{"copy stored in a field", head + "const box: { held: Source } = { held: honest };\nbox.held = { ...built };\nconsole.log(typeof box.held.weight + \" \" + String(box.held.weight));\n"},
		{"copy in an array", head + "const list: Source[] = [honest];\nlist.push({ ...built });\nfor (const each of list) console.log(typeof each.weight + \" \" + String(each.weight));\n"},
		{"copy plus own field after, then copied", head + "const inner = { ...built, extra: 1 };\nconst value: Source = { ...inner };\nconsole.log(typeof value.weight + \" \" + String(value.weight));\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			t.Logf("PROBE %s: node stdout %q exit %d", c.name, truth.stdout, truth.exitCode)
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("PROBE %s: STOP %v", c.name, err)
				return
			}
			got, _ := natively(t, program)
			t.Logf("PROBE %s: native stdout %q exit %d stderr %.300q; disagreement %q", c.name, got.stdout, got.exitCode, got.stderr, disagreement(truth, got))
		})
	}
}
