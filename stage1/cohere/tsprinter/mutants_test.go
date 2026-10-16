package tsprinter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func mutatedPort(t *testing.T, change oracles.Mutation) string {
	t.Helper()
	directory := t.TempDir()
	parser, err := filepath.Abs("../../typescript/parser")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	applied := false
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		if file == change.File {
			if strings.Count(source, change.From) != 1 {
				t.Fatalf("mutant %q must change one place", change.Name)
			}
			source = strings.Replace(source, change.From, change.To, 1)
			applied = true
		}
		source = strings.ReplaceAll(source, "../../typescript/parser/", parser+"/")
		if err = os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !applied {
		t.Fatal("mutant not applied")
	}
	return filepath.Join(directory, change.Entry)
}
