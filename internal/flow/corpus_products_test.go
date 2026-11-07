package flow

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func flowCorpusRemainderProduct(test *testing.T, path string, build bool) string {
	test.Helper()
	files := append(compilerinputs.Lowering(), "dedication")
	inputs := buildcache.Inputs{
		Name:      "flow corpus lowering " + filepath.ToSlash(path),
		Files:     files,
		Toolchain: []string{runtime.Version()},
	}
	if !build {
		directory, err := buildcache.BuiltAhead(
			inputs,
			fmt.Errorf("flow corpus lowering not built ahead; run TestProduct_FlowCorpusRemainder tests first"),
		)
		if err != nil {
			test.Fatal(err)
		}
		return directory
	}
	return buildcache.Product(test, inputs, func(directory string) error {
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
		ir.RegisterGob()
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

func runFlowCorpusRemainderProductsShard(test *testing.T, shard int) {
	beginFlowUnit(test)
	for _, path := range flowCorpusRemainderShardPaths(test, shard) {
		test.Run(filepath.ToSlash(path), func(test *testing.T) { flowCorpusRemainderProduct(test, path, true) })
	}
}

func installFlowCorpusRemainderProduct(test *testing.T, path string) {
	test.Helper()
	directory := flowCorpusRemainderProduct(test, path, false)
	ir.RegisterGob()
	file, err := os.Open(filepath.Join(directory, "program.gob"))
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	var program ir.Program
	if err := gob.NewDecoder(file).Decode(&program); err != nil {
		test.Fatal(err)
	}
	// Declaration positions are metadata; restore their checkout spelling.
	for index := range program.Functions {
		position := &program.Functions[index].Position
		position.File = buildcache.Absolute(position.File)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		test.Fatal(err)
	}
	value, _ := loweredPrograms.LoadOrStore(absolute, &loweredProgram{})
	entry := value.(*loweredProgram)
	entry.once.Do(func() { entry.program = &program })
}
