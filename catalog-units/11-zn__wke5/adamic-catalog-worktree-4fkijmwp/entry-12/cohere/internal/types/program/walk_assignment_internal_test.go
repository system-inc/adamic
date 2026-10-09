package program

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Each worker walks the files of exactly one checker, and every file is walked once: two workers sharing a
// checker would contend for its lock, and a file handed to no worker or to two would be skipped or walked
// twice (#zqsdzbq).
func TestTheWalkGivesEachWorkerTheFilesOfOneChecker(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	config := `{"compilerOptions": {"target": "ES2022", "module": "esnext", "moduleResolution": "bundler", "strict": true, "noEmit": true}}`
	if err := os.WriteFile(filepath.Join(root, "tsconfig.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	for index := range 40 {
		text := fmt.Sprintf("export const value%d = %d;\n", index, index)
		if index > 0 {
			text = fmt.Sprintf("import { value%d } from \"./f%d\";\nexport const value%d = value%d + 1;\n", index-1, index-1, index, index-1)
		}
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("f%d.ts", index)), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := Build(Options{ConfigFileName: filepath.Join(root, "tsconfig.json"), Checkers: 4})
	if err != nil {
		t.Fatal(err)
	}
	files := graph.ProjectFiles()
	workers := graph.Workers()
	if workers != 4 {
		t.Fatalf("%d workers for 4 checkers", workers)
	}

	assignments := graph.assignFilesToWorkers(context.Background(), files, workers, true)
	walked := make([]int, len(files))
	owners := map[*checker.Checker]int{}
	for worker, indices := range assignments {
		var owner *checker.Checker
		for _, index := range indices {
			walked[index]++
			fileOwner, release := graph.Program.GetTypeCheckerForFile(context.Background(), files[index])
			release()
			if owner == nil {
				owner = fileOwner
			}
			if fileOwner != owner {
				t.Errorf("worker %d holds files of two checkers", worker)
			}
		}
		if owner != nil {
			if other, taken := owners[owner]; taken {
				t.Errorf("workers %d and %d share a checker", other, worker)
			}
			owners[owner] = worker
		}
	}
	for index, count := range walked {
		if count != 1 {
			t.Errorf("%s is walked %d times", files[index].FileName(), count)
		}
	}

	strided := graph.assignFilesToWorkers(context.Background(), files, workers, false)
	total := 0
	for _, indices := range strided {
		total += len(indices)
	}
	if total != len(files) {
		t.Errorf("without checkers, %d of %d files are handed out", total, len(files))
	}
}
