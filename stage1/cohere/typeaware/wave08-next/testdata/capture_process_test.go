// Owned overlay around upstream tests; production calls and assertions are unchanged.
package nexus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
	"os"
	"path/filepath"
	"testing"
)

func wave08Save(t *testing.T, subject rule.Rule, files map[string]string, filename, config string) {
	payload, err := json.Marshal(struct {
		Name    string
		Rule    string
		Subject string
		Config  string
		Files   map[string]string
	}{t.Name(), subject.Name, filename, config, files})
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(payload)
	name := hex.EncodeToString(hash[:])
	if directory := os.Getenv("ADAMIC_WAVE08_CAPTURE"); directory != "" {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name+".json"), payload, 0644); err != nil {
			t.Fatal(err)
		}
	}
}
func wave08Capture(t *testing.T, subject rule.Rule, files map[string]string, filename string) rule_testing.Result {
	wave08Save(t, subject, files, filename, wave08DefaultConfig)
	return rule_testing.RunTypedFiles(t, subject, files, filename)
}
func wave08CaptureWithSetup(t *testing.T, subject rule.Rule, files map[string]string, filename string, setup func(string)) rule_testing.Result {
	config := ""
	result := rule_testing.RunTypedFilesWithSetup(t, subject, files, filename, func(directory string) {
		if setup != nil {
			setup(directory)
		}
		text, err := os.ReadFile(filepath.Join(directory, "tsconfig.json"))
		if err != nil {
			t.Fatal(err)
		}
		config = string(text)
	})
	wave08Save(t, subject, files, filename, config)
	return result
}

const wave08DefaultConfig = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"moduleDetection": "force",
		"types": []
	},
	"include": ["**/*.ts", "**/*.tsx"]
}`
