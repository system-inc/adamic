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
	parser, absolutePathError := filepath.Abs("../../typescript/parser")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	files, globError := filepath.Glob("*.ts")
	if globError != nil {
		test.Fatal(globError)
	}
	applied := false
	for _, file := range files {
		data, readError := os.ReadFile(file)
		if readError != nil {
			test.Fatal(readError)
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
		if writeError := os.WriteFile(filepath.Join(directory, file), []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if !applied {
		test.Fatal("mutant not applied")
	}
	return filepath.Join(directory, change.Entry)
}
