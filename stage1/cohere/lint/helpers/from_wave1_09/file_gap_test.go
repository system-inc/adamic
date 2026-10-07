package fromwave109

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestStylesheetLoaderByteInputGap(t *testing.T) {
	// Not parallel: this feasibility probe records one shared-blocker transcript.
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs("file_loader_gap")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_wave109_file_gap.go")
	export := filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_wave109_file_gap.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "oracle.go.txt"), export: filepath.Join(dir, "export.go.txt")}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(scratch, "overlay.json")
	write(t, overlayPath, overlay)
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	binary, emitted := build(t, dir)
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, sample := range []struct {
		name    string
		value   []byte
		blocked bool
	}{
		{"ASCII", []byte("A"), false},
		{"Unicode", []byte("é😀"), false},
		{"invalid80", []byte{0x80}, true},
		{"invalid81", []byte{0x81}, true},
		{"invalidFF", []byte{0xff}, true},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "theme.css")
			css := append([]byte("@theme { --color-probe: "), sample.value...)
			css = append(css, []byte("; }")...)
			write(t, path, css)
			want := run(t, "", oracle, path)
			expected := []byte{}
			for _, value := range sample.value {
				expected = append(expected, []byte(fmt.Sprintf("%d\n", value))...)
			}
			check(t, want, expected)
			for _, command := range [][]string{
				{"node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "main.a"), path},
				{"node", "--disable-warning=ExperimentalWarning", runner, emitted, path},
				{binary, path},
			} {
				got := run(t, "", command[0], command[1:]...)
				if sample.blocked {
					if bytes.Equal(got, want) {
						t.Fatal("documented byte-input gap unexpectedly closed")
					}
					t.Logf("BLOCKED %s: actual Go loadFile theme bytes [%s], decoded file-input bytes [%s]; compiled exit 0, no stderr", sample.name, strings.TrimSpace(string(want)), strings.TrimSpace(string(got)))
				} else {
					check(t, got, want)
					t.Logf("valid-text control matches Go on %s", command[0])
				}
			}
		})
	}
}
