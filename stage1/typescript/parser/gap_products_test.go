package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// gapNative is a gap program's native build as a keyed product, built ahead by its TestProduct_ below, so no gap test
// compiles C inside its own run (#tze35sy). TestClosedStrongAstParentGap built a sanitized and a counted native in its
// run, 29.1 s of stage1/typescript/parser's 72.7 s unit on Adam at load 58 (Nov 5), and the package's other gaps each
// built one. relative is the program's path in this package; directory, when set, is a directory of the program's own
// that it imports from, keyed whole. The key is the lowering compiler's files, the program's, this file, the options
// and what clang reads.
func gapNative(t *testing.T, relative, directory string, options native.Options) string {
	t.Helper()
	files := append(compilerinputs.Lowering(), "stage1/typescript/parser/"+relative, "stage1/typescript/parser/gap_products_test.go")
	if directory != "" {
		files = append(files, "stage1/typescript/parser/"+directory)
	}
	flags := append([]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	inputs := buildcache.Inputs{
		Name:      fmt.Sprintf("parser gap %s native %+v", relative, options),
		Files:     files,
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.CompilerReport("clang")},
	}
	product := buildcache.Product(t, inputs, func(out string) error {
		path, err := filepath.Abs(relative)
		if err != nil {
			return err
		}
		program, err := load.Load([]string{path})
		if err != nil {
			return fmt.Errorf("Load: %w", err)
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return fmt.Errorf("Lower: %w", err)
		}
		return native.Build(native.C(lowered), filepath.Join(out, "native"), options)
	})
	return filepath.Join(product, "native")
}

func TestProduct_GapStrongAstParentSanitized(t *testing.T) {
	t.Parallel()
	gapNative(t, "gaps/1_strong_ast_parent.ts", "", native.Options{Sanitize: true})
}

func TestProduct_GapStrongAstParentCounted(t *testing.T) {
	t.Parallel()
	gapNative(t, "gaps/1_strong_ast_parent.ts", "", native.Options{Count: true})
}

func TestProduct_GapPushSpreadSanitized(t *testing.T) {
	t.Parallel()
	gapNative(t, "gaps/2_push_spread.ts", "", native.Options{Sanitize: true})
}

func TestProduct_GapOptionalFunctionValueSanitized(t *testing.T) {
	t.Parallel()
	gapNative(t, "gaps/5_optional_function_value.ts", "", native.Options{Sanitize: true})
}

func TestProduct_GapConditionalEmptyArraySanitized(t *testing.T) {
	t.Parallel()
	gapNative(t, "gaps/6_conditional_empty_array.ts", "", native.Options{Sanitize: true})
}

func TestProduct_TypeOnlyImportCycleSanitized(t *testing.T) {
	t.Parallel()
	gapNative(t, "testdata/type_import_cycle/entry.a", "testdata/type_import_cycle", native.Options{Sanitize: true})
}
