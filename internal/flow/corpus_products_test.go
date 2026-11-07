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
		directory, builtAheadError := buildcache.BuiltAhead(
			inputs,
			fmt.Errorf("flow corpus lowering not built ahead; run TestProduct_FlowCorpusRemainder tests first"),
		)
		if builtAheadError != nil {
			test.Fatal(builtAheadError)
		}
		return directory
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		absolute, absolutePathError := filepath.Abs(path)
		if absolutePathError != nil {
			return absolutePathError
		}
		source, loadDisplayedError := load.LoadDisplayed([]string{absolute}, buildcache.Relative)
		if loadDisplayedError != nil {
			return loadDisplayedError
		}
		program, lowerError := Lower(context.Background(), source)
		if lowerError != nil {
			return lowerError
		}
		ir.RegisterGob()
		file, createError := os.Create(filepath.Join(directory, "program.gob"))
		if createError != nil {
			return createError
		}
		encodeError := gob.NewEncoder(file).Encode(program)
		closeError := file.Close()
		if encodeError != nil {
			return encodeError
		}
		return closeError
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
	file, openError := os.Open(filepath.Join(directory, "program.gob"))
	if openError != nil {
		test.Fatal(openError)
	}
	defer file.Close()
	var program ir.Program
	if decodeError := gob.NewDecoder(file).Decode(&program); decodeError != nil {
		test.Fatal(decodeError)
	}
	// Declaration positions are metadata; restore their checkout spelling.
	for index := range program.Functions {
		position := &program.Functions[index].Position
		position.File = buildcache.Absolute(position.File)
	}
	absolute, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	value, _ := loweredPrograms.LoadOrStore(absolute, &loweredProgram{})
	entry := value.(*loweredProgram)
	entry.once.Do(func() { entry.program = &program })
}
