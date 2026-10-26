package flow

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

var flowCorpusGobOnce sync.Once

func flowCorpusRemainderProduct(t *testing.T, path string, build bool) string {
	t.Helper()
	files := append(compilerinputs.Lowering(), "dedication")
	inputs := buildcache.Inputs{Name: "flow corpus lowering " + filepath.ToSlash(path), Files: files, Toolchain: []string{runtime.Version()}}
	if !build {
		directory, err := buildcache.Get(inputs, func(string) error {
			return fmt.Errorf("flow corpus lowering not built ahead; run TestProduct_FlowCorpusRemainder tests first")
		})
		if err != nil {
			t.Fatal(err)
		}
		return directory
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		source, err := load.LoadDisplayed([]string{absolute}, buildcache.Relative)
		if err != nil {
			return err
		}
		program, err := Lower(context.Background(), source)
		if err != nil {
			return err
		}
		registerFlowCorpusIR()
		file, err := os.Create(filepath.Join(directory, "program.gob"))
		if err != nil {
			return err
		}
		err = gob.NewEncoder(file).Encode(program)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
}

func runFlowCorpusRemainderProductsShard(t *testing.T, shard int) {
	beginFlowUnit(t)
	for _, path := range flowCorpusRemainderShardPaths(t, shard) {
		t.Run(filepath.ToSlash(path), func(t *testing.T) { flowCorpusRemainderProduct(t, path, true) })
	}
}

func installFlowCorpusRemainderProduct(t *testing.T, path string) {
	t.Helper()
	directory := flowCorpusRemainderProduct(t, path, false)
	registerFlowCorpusIR()
	file, err := os.Open(filepath.Join(directory, "program.gob"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var program ir.Program
	if err := gob.NewDecoder(file).Decode(&program); err != nil {
		t.Fatal(err)
	}
	// Declaration positions are metadata; restore their checkout spelling.
	for index := range program.Functions {
		position := &program.Functions[index].Position
		position.File = buildcache.Absolute(position.File)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	value, _ := loweredPrograms.LoadOrStore(absolute, &loweredProgram{})
	entry := value.(*loweredProgram)
	entry.once.Do(func() { entry.program = &program })
}
