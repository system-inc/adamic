package tsprinter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func mutatedPort(test *testing.T, change oracles.Mutation) string {
	test.Helper()
	directory := test.TempDir()
	parser, err := filepath.Abs("../../typescript/parser")
	if err != nil {
		test.Fatal(err)
	}
	files, err := filepath.Glob("*.ts")
	if err != nil {
		test.Fatal(err)
	}
	applied := false
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			test.Fatal(err)
		}
		source := string(data)
		if file == change.File {
			if strings.Count(source, change.From) != 1 {
				test.Fatalf("mutant %q must change one place", change.Name)
			}
			source = strings.Replace(source, change.From, change.To, 1)
			applied = true
		}
		source = strings.ReplaceAll(source, "../../typescript/parser/", parser+"/")
		if err = os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); err != nil {
			test.Fatal(err)
		}
	}
	if !applied {
		test.Fatal("mutant not applied")
	}
	return filepath.Join(directory, change.Entry)
}
