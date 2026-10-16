package markdownblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

type auditInput = oracles.AuditInput
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
func census() bool { return oracles.Census() }

func auditCorpus(t *testing.T, root string) ([]auditInput, int) { return oracles.AuditCorpus(t, root) }

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
// Not parallel: prepares the shared oracle binaries and corpus before parallel leaves run.
func TestWholeDocumentOraclePreflight_Setup(t *testing.T) {
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
	t.Logf("TestWholeDocumentOraclePreflight_Setup %.6fs", time.Since(started).Seconds())
}

func firstDifference(a, b string) int {
	index := 0
	for index < len(a) && index < len(b) && a[index] == b[index] {
		index++
	}
	return index
}
