package parser

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The four variants use the existing content-addressed parser recipes. Source
// relocation is normalized by wholeMutantPortInputs, so build and test units
// resolve the same keys even though their scratch paths differ.
type expressionProductSet struct {
	once   sync.Once
	oracle string
	source [4]string
	binary [4]string
}

var expressionProducts expressionProductSet
var expressionVariants [4]struct {
	once   sync.Once
	source string
}

func expressionSource(test *testing.T, index int) string {
	test.Helper()
	variant := &expressionVariants[index]
	variant.once.Do(func() {
		if index == 0 {
			variant.source = compilerExpressionsProductDirectory(test)
			return
		}
		files := []string{"stage1/typescript/parser/expressions_products_test.go"}
		for _, name := range portFiles {
			files = append(files, "stage1/typescript/parser/"+name)
		}
		scanner, absolutePathError := filepath.Abs("../scanner/scanner.ts")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		variant.source = buildcache.Product(test, buildcache.Inputs{
			Name:  fmt.Sprintf("expression-source-%03d", index),
			Files: files,
			Flags: []string{scanner, "precedence-14-to-12", "optional-false", "parenthesized-to-arrow"},
		}, func(dir string) error {
			var source string
			switch index {
			case 1:
				source = copyPort(test, "grammar.ts", "return 14;", "return 12;")
			case 2:
				source = copyPort(
					test,
					"parser.ts",
					"left = this.make('PropertyAccessExpression', pos, children);\n                this.node(left).optional = optional;",
					"left = this.make('PropertyAccessExpression', pos, children);\n                this.node(left).optional = false;",
				)
			case 3:
				source = copyPort(
					test,
					"parser.ts",
					"return this.make('ParenthesizedExpression', pos, [expression]);",
					"return this.make('ArrowFunction', pos, [expression]);",
				)
			default:
				test.Fatalf("unknown expression variant %d", index)
			}
			for _, name := range portFiles {
				data, readError := os.ReadFile(filepath.Join(source, name))
				if readError != nil {
					return readError
				}
				if writeError := os.WriteFile(filepath.Join(dir, name), data, 0644); writeError != nil {
					return writeError
				}
			}
			// copyPort's scanner import names this checkout; the product names it <repository> (#tqrqx60).
			return buildcache.RelativeFiles(dir)
		})
	})
	// Node and the loader read the port, so callers get the copy whose imports name this checkout.
	directory, resolvedError := buildcache.Resolved(variant.source)
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	return directory
}

var expressionNative [4]struct {
	once   sync.Once
	binary string
}

func expressionBinary(test *testing.T, index int) string {
	test.Helper()
	product := &expressionNative[index]
	product.once.Do(func() { product.binary = wholeMutantBuildPortProduct(test, expressionSource(test, index), true) })
	return product.binary
}
func expressionSetup(test *testing.T) *expressionProductSet {
	test.Helper()
	expressionProducts.once.Do(func() {
		expressionProducts.oracle = wholeMutantBuildOracleProduct(test)
		for i := range expressionProducts.source {
			expressionProducts.source[i] = expressionSource(test, i)
			expressionProducts.binary[i] = expressionBinary(test, i)
		}
	})
	return &expressionProducts
}

// Only comparison work receives a deadline; product recipes do not.
func expressionExecute(
	test *testing.T,
	callerContext context.Context,
	directory, name string,
	args ...string,
) execution {
	test.Helper()
	command := exec.CommandContext(callerContext, name, args...)
	command.Dir = directory
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	runError := command.Run()
	duration := time.Since(started)
	if runError != nil || stderr.Len() != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return execution{data, duration}
}
func expressionNode(test *testing.T, callerContext context.Context, directory, manifest string, count bool) execution {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	args := []string{
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(directory, "main.ts"),
		"--manifest",
		manifest,
	}
	if count {
		args = append(args, "--count")
	}
	return expressionExecute(test, callerContext, "", "node", args...)
}

func TestProduct_ExpressionsSource001(test *testing.T) { test.Parallel(); expressionSource(test, 1) }
func TestProduct_ExpressionsSource002(test *testing.T) { test.Parallel(); expressionSource(test, 2) }
func TestProduct_ExpressionsSource003(test *testing.T) { test.Parallel(); expressionSource(test, 3) }
func TestProduct_ParserFindsExpressionsOracle(test *testing.T) {
	test.Parallel()
	wholeMutantBuildOracleProduct(test)
}
func TestProduct_ExpressionsLower000(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, expressionSource(test, 0))
}
func TestProduct_ExpressionsNative000(test *testing.T) { test.Parallel(); expressionBinary(test, 0) }
func TestProduct_ExpressionsLower001(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, expressionSource(test, 1))
}
func TestProduct_ExpressionsNative001(test *testing.T) { test.Parallel(); expressionBinary(test, 1) }
func TestProduct_ExpressionsLower002(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, expressionSource(test, 2))
}
func TestProduct_ExpressionsNative002(test *testing.T) { test.Parallel(); expressionBinary(test, 2) }
func TestProduct_ExpressionsLower003(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, expressionSource(test, 3))
}
func TestProduct_ExpressionsNative003(test *testing.T) { test.Parallel(); expressionBinary(test, 3) }

func TestExpressionsAgree(test *testing.T) {
	test.Parallel()
	setupStart := time.Now()
	products := expressionSetup(test)
	test.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
	started := time.Now()
	defer func() { test.Logf("own work wall %.3fs", time.Since(started).Seconds()) }()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	cases := []string{
		"x => x + 1;",
		"(x) => x;",
		"(x);",
		"(x = 10);",
		"(x, y);",
		"(x = 10, y = 2) => x + y;",
		"() => ({x: 1});",
		"(x: number, y?: string, ...z: unknown[]) => [x, y, ...z];",
		"<T extends object>(x: T): T => x;",
		"async x => await f(x);",
		"async (x) => await f(x);",
		"x => { const y = x + 1; return y; };",
		"(function f(x: number) { return x * 2; });",
		"(async function() { return await f(); });",
		"(function*() { yield x; yield* xs; });",
		"({f(x) {return x;}, get x() {return 1;}, set x(v) {f(v);}, async g() {return await f();}});",
		"a as T;",
		"<T[]>a;",
		"value satisfies { readonly x: number; y?: string };",
		"value as A | B & C;",
		"value as readonly [number, string];",
		"value as const;",
		"f<T>(x);",
		"new Foo<T>();",
		"tag<T>`a${x}b`;",
		"f?.<T>(x);",
		"x as typeof ns.value;",
		"((x): number => x)(1);",
		"f(x, ...xs,);",
		"a.b[c](d).e;",
		"a?.b.c?.[x]?.(y);",
		"a?.b!.c;",
		"(a?.b).c;",
		"new Foo(a).bar();",
		"new new Foo();",
		"new Foo; new Foo();",
		"new.target; import.meta; import('module');",
		"[x,,y,...xs,];",
		"[\n1,\n2\n];",
		"({x, y: 2, [key + 1]: value, ...rest});",
		"({x = 3,});",
		"`a${x + y}b${f(z)}c`;",
		"tag`x${a?.b}y`;",
		"`a\\r\\n${x}\\u{1f600}`;",
		"x;",
		"  x + y * z;",
		"a - b - c;",
		"a ** b ** c;",
		"(a + b) * c;",
		"a ? b : c ? d : e;",
		"a = b = c;",
		"x++, --y;",
		"typeof x === \"number\";",
		"a >> b >= c;",
		"0xff + 1_000 / 2e3;",
		"'héllo' + \"😀\";",
		"/a[b]+/gi;",
		"/* trivia */ (x) + y;",
	}
	dir := test.TempDir()
	var manifest strings.Builder
	for i, s := range cases {
		path := filepath.Join(dir, fmt.Sprintf("case-%d.ts", i))
		if writeError := os.WriteFile(path, []byte(s), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if writeError := os.WriteFile(path, []byte(manifest.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	oracle := products.oracle
	want := expressionExecute(test, timeoutContext, "", oracle, "--manifest", path)
	absolute, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	nativeRun := expressionExecute(test, timeoutContext, "", products.binary[0], "--manifest", path)
	nodeRun := expressionNode(test, timeoutContext, absolute, path, false)
	for _, side := range []struct {
		name string
		data []byte
	}{{"native", nativeRun.output}, {"Node", nodeRun.output}} {
		if diff := difference(side.data, want.output); diff != "" {
			test.Fatalf("%s: %s", side.name, diff)
		}
	}
	test.Logf("%d inputs, %d identical canonical bytes", len(cases), len(want.output))
	mutant := products.source[1]
	for _, side := range []struct {
		name string
		data []byte
	}{
		{"native", expressionExecute(test, timeoutContext, "", products.binary[1], "--manifest", path).output},
		{"Node", expressionNode(test, timeoutContext, mutant, path, false).output},
	} {
		if difference(side.data, want.output) == "" {
			test.Fatalf("%s precedence mutant survived", side.name)
		}
		test.Logf("%s precedence mutant caught: %s", side.name, difference(side.data, want.output))
	}
	optionalMutant := products.source[2]
	for _, side := range []struct {
		name string
		data []byte
	}{
		{"native", expressionExecute(test, timeoutContext, "", products.binary[2], "--manifest", path).output},
		{"Node", expressionNode(test, timeoutContext, optionalMutant, path, false).output},
	} {
		if difference(side.data, want.output) == "" {
			test.Fatalf("%s optional mutant survived", side.name)
		}
		test.Logf("%s optional mutant caught: %s", side.name, difference(side.data, want.output))
	}

	arrowMutant := products.source[3]
	for _, side := range []struct {
		name string
		data []byte
	}{
		{"native", expressionExecute(test, timeoutContext, "", products.binary[3], "--manifest", path).output},
		{"Node", expressionNode(test, timeoutContext, arrowMutant, path, false).output},
	} {
		if difference(side.data, want.output) == "" {
			test.Fatalf("%s parenthesized arrow mutant survived", side.name)
		}
		test.Logf(
			"%s parenthesized expression misclassified as arrow caught: %s",
			side.name,
			difference(side.data, want.output),
		)
	}

}
