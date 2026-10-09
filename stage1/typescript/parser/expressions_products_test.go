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

func expressionSource(t *testing.T, index int) string {
	t.Helper()
	variant := &expressionVariants[index]
	variant.once.Do(func() {
		if index == 0 {
			variant.source = compilerExpressionsProductDirectory(t)
			return
		}
		files := []string{"stage1/typescript/parser/expressions_products_test.go"}
		for _, name := range portFiles {
			files = append(files, "stage1/typescript/parser/"+name)
		}
		scanner, err := filepath.Abs("../scanner/scanner.ts")
		if err != nil {
			t.Fatal(err)
		}
		variant.source = buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprintf("expression-source-%03d", index), Files: files, Flags: []string{scanner, "precedence-14-to-12", "optional-false", "parenthesized-to-arrow"}}, func(dir string) error {
			var source string
			switch index {
			case 1:
				source = copyPort(t, "grammar.ts", "return 14;", "return 12;")
			case 2:
				source = copyPort(t, "parser.ts", "left = this.make('PropertyAccessExpression', pos, children);\n                this.node(left).optional = optional;", "left = this.make('PropertyAccessExpression', pos, children);\n                this.node(left).optional = false;")
			case 3:
				source = copyPort(t, "parser.ts", "return this.make('ParenthesizedExpression', pos, [expression]);", "return this.make('ArrowFunction', pos, [expression]);")
			default:
				t.Fatalf("unknown expression variant %d", index)
			}
			for _, name := range portFiles {
				data, err := os.ReadFile(filepath.Join(source, name))
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
					return err
				}
			}
			return nil
		})
	})
	return variant.source
}

var expressionNative [4]struct {
	once   sync.Once
	binary string
}

func expressionBinary(t *testing.T, index int) string {
	t.Helper()
	product := &expressionNative[index]
	product.once.Do(func() { product.binary = wholeMutantBuildPortProduct(t, expressionSource(t, index), true) })
	return product.binary
}
func expressionSetup(t *testing.T) *expressionProductSet {
	t.Helper()
	expressionProducts.once.Do(func() {
		expressionProducts.oracle = wholeMutantBuildOracleProduct(t)
		for i := range expressionProducts.source {
			expressionProducts.source[i] = expressionSource(t, i)
			expressionProducts.binary[i] = expressionBinary(t, i)
		}
	})
	return &expressionProducts
}

// Only comparison work receives a deadline; product recipes do not.
func expressionExecute(t *testing.T, ctx context.Context, directory, name string, args ...string) execution {
	t.Helper()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	duration := time.Since(started)
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}
func expressionNode(t *testing.T, ctx context.Context, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return expressionExecute(t, ctx, "", "node", args...)
}

func TestProduct_ExpressionsSource001(t *testing.T) { t.Parallel(); expressionSource(t, 1) }
func TestProduct_ExpressionsSource002(t *testing.T) { t.Parallel(); expressionSource(t, 2) }
func TestProduct_ExpressionsSource003(t *testing.T) { t.Parallel(); expressionSource(t, 3) }
func TestProduct_ExpressionsOracle(t *testing.T)    { t.Parallel(); wholeMutantBuildOracleProduct(t) }
func TestProduct_ExpressionsLower000(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, expressionSource(t, 0))
}
func TestProduct_ExpressionsNative000(t *testing.T) { t.Parallel(); expressionBinary(t, 0) }
func TestProduct_ExpressionsLower001(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, expressionSource(t, 1))
}
func TestProduct_ExpressionsNative001(t *testing.T) { t.Parallel(); expressionBinary(t, 1) }
func TestProduct_ExpressionsLower002(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, expressionSource(t, 2))
}
func TestProduct_ExpressionsNative002(t *testing.T) { t.Parallel(); expressionBinary(t, 2) }
func TestProduct_ExpressionsLower003(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, expressionSource(t, 3))
}
func TestProduct_ExpressionsNative003(t *testing.T) { t.Parallel(); expressionBinary(t, 3) }

func TestExpressionsAgree(t *testing.T) {
	t.Parallel()
	setupStart := time.Now()
	products := expressionSetup(t)
	t.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
	started := time.Now()
	defer func() { t.Logf("own work wall %.3fs", time.Since(started).Seconds()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	cases := []string{"x => x + 1;", "(x) => x;", "(x);", "(x = 10);", "(x, y);", "(x = 10, y = 2) => x + y;", "() => ({x: 1});", "(x: number, y?: string, ...z: unknown[]) => [x, y, ...z];", "<T extends object>(x: T): T => x;", "async x => await f(x);", "async (x) => await f(x);", "x => { const y = x + 1; return y; };", "(function f(x: number) { return x * 2; });", "(async function() { return await f(); });", "(function*() { yield x; yield* xs; });", "({f(x) {return x;}, get x() {return 1;}, set x(v) {f(v);}, async g() {return await f();}});", "a as T;", "<T[]>a;", "value satisfies { readonly x: number; y?: string };", "value as A | B & C;", "value as readonly [number, string];", "value as const;", "f<T>(x);", "new Foo<T>();", "tag<T>`a${x}b`;", "f?.<T>(x);", "x as typeof ns.value;", "((x): number => x)(1);", "f(x, ...xs,);", "a.b[c](d).e;", "a?.b.c?.[x]?.(y);", "a?.b!.c;", "(a?.b).c;", "new Foo(a).bar();", "new new Foo();", "new Foo; new Foo();", "new.target; import.meta; import('module');", "[x,,y,...xs,];", "[\n1,\n2\n];", "({x, y: 2, [key + 1]: value, ...rest});", "({x = 3,});", "`a${x + y}b${f(z)}c`;", "tag`x${a?.b}y`;", "`a\\r\\n${x}\\u{1f600}`;", "x;", "  x + y * z;", "a - b - c;", "a ** b ** c;", "(a + b) * c;", "a ? b : c ? d : e;", "a = b = c;", "x++, --y;", "typeof x === \"number\";", "a >> b >= c;", "0xff + 1_000 / 2e3;", "'héllo' + \"😀\";", "/a[b]+/gi;", "/* trivia */ (x) + y;"}
	dir := t.TempDir()
	var manifest strings.Builder
	for i, s := range cases {
		path := filepath.Join(dir, fmt.Sprintf("case-%d.ts", i))
		if err := os.WriteFile(path, []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
		manifest.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	oracle := products.oracle
	want := expressionExecute(t, ctx, "", oracle, "--manifest", path)
	absolute, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	nativeRun := expressionExecute(t, ctx, "", products.binary[0], "--manifest", path)
	nodeRun := expressionNode(t, ctx, absolute, path, false)
	for _, side := range []struct {
		name string
		data []byte
	}{{"native", nativeRun.output}, {"Node", nodeRun.output}} {
		if diff := difference(side.data, want.output); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	t.Logf("%d inputs, %d identical canonical bytes", len(cases), len(want.output))
	mutant := products.source[1]
	for _, side := range []struct {
		name string
		data []byte
	}{{"native", expressionExecute(t, ctx, "", products.binary[1], "--manifest", path).output}, {"Node", expressionNode(t, ctx, mutant, path, false).output}} {
		if difference(side.data, want.output) == "" {
			t.Fatalf("%s precedence mutant survived", side.name)
		}
		t.Logf("%s precedence mutant caught: %s", side.name, difference(side.data, want.output))
	}
	optionalMutant := products.source[2]
	for _, side := range []struct {
		name string
		data []byte
	}{{"native", expressionExecute(t, ctx, "", products.binary[2], "--manifest", path).output}, {"Node", expressionNode(t, ctx, optionalMutant, path, false).output}} {
		if difference(side.data, want.output) == "" {
			t.Fatalf("%s optional mutant survived", side.name)
		}
		t.Logf("%s optional mutant caught: %s", side.name, difference(side.data, want.output))
	}

	arrowMutant := products.source[3]
	for _, side := range []struct {
		name string
		data []byte
	}{{"native", expressionExecute(t, ctx, "", products.binary[3], "--manifest", path).output}, {"Node", expressionNode(t, ctx, arrowMutant, path, false).output}} {
		if difference(side.data, want.output) == "" {
			t.Fatalf("%s parenthesized arrow mutant survived", side.name)
		}
		t.Logf("%s parenthesized expression misclassified as arrow caught: %s", side.name, difference(side.data, want.output))
	}

}
