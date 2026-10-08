package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

type resolvedFormatRow struct {
	Directory string `json:"directory"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Config    string `json:"config"`
}

func TestResolvedFormatting(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	rows := []resolvedFormatRow{}
	add := func(name, source, config string) {
		directory := filepath.Join(temp, "trees", fmt.Sprint(len(rows)))
		write(t, filepath.Join(directory, "CohereSettings.json"), []byte(config))
		write(t, filepath.Join(directory, name), []byte(source))
		rows = append(rows, resolvedFormatRow{directory, name, source, config})
	}
	var pins []settingsPin
	if e = json.Unmarshal(read(t, "testdata/corpus.json"), &pins); e != nil {
		t.Fatal(e)
	}
	css := 0
	for _, pin := range pins {
		for _, file := range pin.Files {
			if !strings.HasSuffix(file.Fixture, ".css") {
				continue
			}
			source := settingsFixture(t, file.Fixture, file.SHA256)
			css++
			for _, config := range []string{`{}`, `{"format":{}}`, `{"format":{"tabWidth":3,"singleQuote":true,"printWidth":60}}`} {
				add("probe.css", source, config)
			}
		}
	}
	if css != 6 {
		t.Fatalf("quiet CSS=%d", css)
	}
	for _, source := range []string{"{a}", "query Q($x: Int) { a(x: $x) { first second third } }", "\ufeff{a}", "query Q { a\r\n b(c: 1) }\r\n", "# comment\n{a}"} {
		for _, config := range []string{`{}`, `{"format":{}}`, `{"format":{"tabWidth":3,"printWidth":20}}`, `{"format":{"useTabs":true}}`, `{"format":{"bracketSpacing":false}}`} {
			add("probe.graphql", source, config)
		}
	}
	for _, control := range []row{{"probe.json", `{"a":[1,2],"b":{}}`}, {"probe.yaml", "a:   1\nb: [ x,y ]\n"}, {"probe.ts", "1+2;"}} {
		add(control.Name, control.Source, `{"format":{}}`)
	}
	add("probe.css", "a{color:red}", `{"format":{"quoteProps":"consistent"}}`)
	data, e := json.Marshal(rows)
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(temp, "format.json")
	write(t, manifest, data)
	overlayData, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): filepath.Join(root, "stage1/cohere/command/testdata/resolved_format_oracle.go")}})
	overlay := filepath.Join(temp, "overlay.json")
	write(t, overlay, overlayData)
	oracle := filepath.Join(temp, "oracle")
	run(t, filepath.Join(root, "cohere"), "go", "build", "-overlay", overlay, "-o", oracle, "./command/formatter_comparison")
	expected := run(t, root, oracle, manifest)
	entry := filepath.Join(root, "stage1/cohere/command/settings/format_main.a")
	node := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", entry, manifest)
	if !bytes.Equal(node, expected) {
		g, w := bytes.Split(node, []byte("\n")), bytes.Split(expected, []byte("\n"))
		for i := range rows {
			if !bytes.Equal(g[i], w[i]) {
				t.Fatalf("resolved format row %d (%s, %s)\ngot %s\nwant %s", i, rows[i].Name, rows[i].Config, g[i], w[i])
			}
		}
		t.Fatal("format framing")
	}
	binary := build(t, root, entry, filepath.Join(temp, "release"), false)
	native := run(t, root, binary, manifest)
	if !bytes.Equal(native, expected) {
		t.Fatal("linked native resolved formatting differs from Go")
	}
	t.Logf("linked resolver+formatters: %d exact Go/Node/native outcomes, six pinned CSS sources, house/own/override/default/refusal cases", len(rows))
	// The unchanged JSON/YAML/TS APIs cannot silently receive house or arbitrary options.
	boundaries := []resolvedFormatRow{}
	for _, name := range []string{"probe.json", "probe.yaml", "probe.ts"} {
		directory := filepath.Join(temp, "boundary", name)
		boundaries = append(boundaries, resolvedFormatRow{directory, name, "1", `{}`})
	}
	data, e = json.Marshal(boundaries)
	if e != nil {
		t.Fatal(e)
	}
	unsupported := filepath.Join(temp, "boundaries.json")
	write(t, unsupported, data)
	got := run(t, root, binary, unsupported)
	if bytes.Count(got, []byte("NotYet:")) != 3 {
		t.Fatalf("formatter owner boundaries must refuse explicitly: %s", got)
	}
	if source := run(t, root, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", entry, unsupported); !bytes.Equal(source, got) {
		t.Fatal("formatter boundary Node/native disagreement")
	}
}
