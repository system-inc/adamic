package markdownblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

type auditInput struct {
	Corpus bool   `json:"-"`
	Name   string `json:"name"`
	Text   string `json:"text"`
}
type auditOutput struct {
	Name      string   `json:"name"`
	Auto      string   `json:"auto"`
	Off       string   `json:"off"`
	AutoError string   `json:"autoError"`
	OffError  string   `json:"offError"`
	Embeds    []string `json:"embeds"`
}

// census reports whether the corpus covers every Markdown file in the repository and its submodules. It is
// opt-in: that set moves with every commit and every cohere bump, and a gate must not turn red on a README this
// unit never owned. Without it the corpus is this unit's own files plus the generated documents.
func census() bool {
	return os.Getenv("ADAMIC_MARKDOWNBLOCKS_CENSUS") != ""
}

func auditCorpus(t *testing.T, root string) ([]auditInput, int) {
	t.Helper()
	var inputs []auditInput
	patterns := []string{"*.md", "*.markdown", "*.mdown", "*.mkd"}
	roots := []string{"stage1/cohere/markdownblocks"}
	if census() {
		roots = []string{"."}
	}
	paths := corpusfiles.Repository(t, root, roots, patterns)
	if census() {
		paths = append(paths, corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{"CHANGELOG.md", "CONTRIBUTING.md", "README.md", "THIRD_PARTY_NOTICES.md", "TypeScript-shim", "editors", "internal", "schema", "swift"}, patterns)...)
		paths = append(paths, corpusfiles.Upstream(t, filepath.Join(root, "cohere/TypeScript"), corpusfiles.TypeScriptGoCommit, []string{".github", "CODE_OF_CONDUCT.md", "CONTRIBUTING.md", "README.md", "SECURITY.md", "SUPPORT.md", "packages", "tsc"}, patterns)...)
	}
	sort.Strings(paths)
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, auditInput{Name: filepath.ToSlash(name), Text: string(content)})
	}
	files := len(inputs)
	for _, marker := range []string{"*", "-", "+", "1.", "1)", "10.", "999999999."} {
		for depth := 1; depth <= 5; depth++ {
			var lines []string
			for level := 0; level < depth; level++ {
				lines = append(lines, strings.Repeat("  ", level)+marker+" item *text* `x`")
			}
			inputs = append(inputs, auditInput{Name: "generated/list/" + marker + "/" + strings.Repeat("n", depth), Text: strings.Join(lines, "\n") + "\n"})
		}
	}
	for _, text := range []string{"# x #", "## x", "x\n=", "x\n---", "a\nb\n==="} {
		inputs = append(inputs, auditInput{Name: "generated/heading/" + text, Text: text + "\n"})
	}
	for _, prefix := range []string{"", "> ", "- ", "  "} {
		for _, block := range []string{"```\nx\n```", "~~~\nx\n~~~", "    x\n    y", "<div>\nx\n</div>", "<!-- x -->", "***", "---\na: b\n---", "|a|b|\n|-|-|\n|x|y|"} {
			inputs = append(inputs, auditInput{Name: "generated/block/" + prefix + block, Text: prefix + strings.ReplaceAll(block, "\n", "\n"+prefix) + "\n"})
		}
	}
	for _, code := range []string{`if (true) { console.log("x"); }`, `for (let i=0;i<2;i++) console.log(i);`, `const x = <div className="x">x</div>;`} {
		inputs = append(inputs, auditInput{Name: "generated/embedding/" + code, Text: "```tsx\n" + code + "\n```\n"})
	}
	for _, name := range []string{"embedded_jsonc.md", "embedded_flow.md"} {
		content, err := os.ReadFile(filepath.Join("gaps", name))
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, auditInput{Name: "generated/" + name, Text: string(content)})
	}
	return inputs, files
}

func auditResults(t *testing.T, name string, result run) []auditOutput {
	t.Helper()
	clean(t, name, result)
	var answers []auditOutput
	for _, line := range bytes.Split(bytes.TrimSuffix(result.stdout, []byte("\n")), []byte("\n")) {
		var answer auditOutput
		if err := json.Unmarshal(line, &answer); err != nil {
			t.Fatal(err)
		}
		if answer.AutoError != "" || answer.OffError != "" {
			t.Fatalf("%s formatting error at %s: %s %s", name, answer.Name, answer.AutoError, answer.OffError)
		}
		answers = append(answers, answer)
	}
	return answers
}

// Not parallel: wholeDocumentPreflightShared is initialized for the parallel preflight shards. prepares the shared oracle binaries and corpus before parallel leaves run.
func TestWholeDocumentOraclePreflight(t *testing.T) {
	started := time.Now()
	fixture := wholeDocumentPreflight(t)
	for shard, cases := range fixture.shards {
		for _, input := range cases {
			digest := sha256.Sum256([]byte(input.Name))
			assigned := int(uint32(digest[0])<<24|uint32(digest[1])<<16|uint32(digest[2])<<8|uint32(digest[3])) % testWholeDocumentOraclePreflightShards
			if assigned != shard {
				t.Fatal(fmt.Sprintf("incorrect shard for %s", input.Name))
			}
		}
	}
	t.Logf("TestWholeDocumentOraclePreflight (setup) %.6fs", time.Since(started).Seconds())
}

func firstDifference(a, b string) int {
	index := 0
	for index < len(a) && index < len(b) && a[index] == b[index] {
		index++
	}
	return index
}
