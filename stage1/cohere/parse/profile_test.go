package parse

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Compiler inputs come from Go's dependency inventory, including embedded runtime
// bytes. Driver and parser inputs belong to the separate lowering product.
func driverInputs(test *testing.T, name string) buildcache.Inputs {
	test.Helper()
	inputs, err := buildcache.GoInputs("adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		test.Fatal(err)
	}
	inputs.Name = "stage1-parse-" + name
	inputs.Files = append(inputs.Files, "stage1/cohere/parse", "stage1/typescript/parser", "stage1/typescript/scanner")
	return inputs
}

func driverSource(test *testing.T, root string) string {
	test.Helper()
	inputs := driverInputs(test, root+"-lower")
	inputs.Flags = append(inputs.Flags, root)
	directory := buildcache.Product(test, inputs, func(directory string) error {
		program, err := load.Load([]string{root + ".a"})
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		ir, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "main.c"), []byte(native.C(ir)), 0644)
	})
	data, err := os.ReadFile(filepath.Join(directory, "main.c"))
	if err != nil {
		test.Fatal(err)
	}
	return string(data)
}

func driverBinary(test *testing.T, root string) string {
	test.Helper()
	source := driverSource(test, root)
	inputs := driverInputs(test, root+"-native")
	inputs.Flags = append(inputs.Flags, source)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	directory := buildcache.Product(test, inputs, func(directory string) error {
		return native.Build(source, filepath.Join(directory, "driver"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(directory, "driver")
}

func parseOracle(test *testing.T) string {
	test.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		test.Fatal(err)
	}
	virtual := filepath.Join(root, "cohere/TypeScript/tsc/adamic_stage1_parse_oracle.go")
	replacement := filepath.Join(root, "stage1/cohere/parse/testdata/go_parse.go.txt")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: replacement}})
	if err != nil {
		test.Fatal(err)
	}
	overlay := filepath.Join(test.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		test.Fatal(err)
	}
	return buildcache.GoBuildIn(
		test,
		"cohere/TypeScript/tsc",
		"oracle",
		"./adamic_stage1_parse_oracle.go",
		[]string{"-pgo=off", "-overlay=" + overlay},
	)
}

func execute(test *testing.T, name string, args ...string) []byte {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := toolchain.CommandContextIn(ctx, "", name, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil || stderr.Len() != 0 {
		test.Fatalf("%s: %v\n%s", name, err, &stderr)
	}
	return output
}

func TestProduct_ParseDriverLower(test *testing.T)  { test.Parallel(); driverSource(test, "parse") }
func TestProduct_ParseDriverNative(test *testing.T) { test.Parallel(); driverBinary(test, "parse") }
func TestProduct_ParseLintLower(test *testing.T)    { test.Parallel(); driverSource(test, "main") }
func TestProduct_ParseLintNative(test *testing.T)   { test.Parallel(); driverBinary(test, "main") }
func TestProduct_ParseGoOracle(test *testing.T)     { test.Parallel(); parseOracle(test) }

func checkAST(test *testing.T, source string) {
	test.Helper()
	fixture := filepath.Join(test.TempDir(), "fixture.a")
	if err := os.WriteFile(fixture, []byte(source), 0644); err != nil {
		test.Fatal(err)
	}
	manifest := filepath.Join(test.TempDir(), "manifest.txt")
	if err := os.WriteFile(manifest, []byte(fixture+"\n"), 0644); err != nil {
		test.Fatal(err)
	}
	want := execute(test, parseOracle(test), "--manifest", manifest, "--ast")
	got := execute(test, driverBinary(test, "parse"), "--manifest", manifest, "--ast")
	node := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../../oracle/node.mjs",
		"parse.a",
		"--manifest",
		manifest,
		"--ast",
	)
	if !bytes.Equal(got, want) || !bytes.Equal(node, want) {
		test.Fatalf(
			"full AST differs from Go: native/Node agree=%t native=%d Node=%d Go=%d",
			bytes.Equal(got, node),
			len(got),
			len(node),
			len(want),
		)
	}
	test.Logf("%d full AST bytes agree with Go and Node", len(want))
}
func TestParseDriverAST(test *testing.T) {
	test.Parallel()
	checkAST(test, "const x = 0.1 * 10 - 1; const y = {a: [1, 2]};\n")
}
func TestParseDriverUnicodeAST(test *testing.T) {
	test.Parallel()
	checkAST(test, "const text = '🐢é'; const template = `hello ${text}`;\n")
}

func TestParseDriverDoesNotLint(test *testing.T) {
	test.Parallel()
	fixture := filepath.Join(test.TempDir(), "fixture.a")
	if err := os.WriteFile(fixture, []byte("ReactDOM.findDOMNode(element);\n"), 0644); err != nil {
		test.Fatal(err)
	}
	manifest := filepath.Join(test.TempDir(), "manifest.txt")
	if err := os.WriteFile(manifest, []byte(fixture+"\n"), 0644); err != nil {
		test.Fatal(err)
	}
	got := execute(test, driverBinary(test, "parse"), "--manifest", manifest, "--count")
	if string(got) != "0\n" {
		test.Fatalf("parse-only driver ran lint: %q", got)
	}
	lint := execute(test, driverBinary(test, "main"), "--manifest", manifest, "--count")
	if string(lint) != "1\n" {
		test.Fatalf("lint traversal lost finding: %q", lint)
	}
}

func TestParseDriverNodeControls(test *testing.T) {
	test.Parallel()
	for _, root := range []string{"main", "parse"} {
		fixture := filepath.Join(test.TempDir(), "fixture.a")
		if err := os.WriteFile(fixture, []byte("ReactDOM.findDOMNode(element);\n"), 0644); err != nil {
			test.Fatal(err)
		}
		got := execute(test, driverBinary(test, root), fixture)
		want := execute(
			test,
			"node",
			"--disable-warning=ExperimentalWarning",
			"../../../oracle/node.mjs",
			root+".a",
			fixture,
		)
		if !bytes.Equal(got, want) {
			test.Fatalf("%s diagnostics/fixes differ from Node", root)
		}
		if root == "main" && !strings.Contains(string(got), "noFindDOMNode") {
			test.Fatal("missing lint control")
		}
	}
}

// Each compiler file owns a top-level case. The content pin prevents a different
// external corpus from being accepted under the v6.0.3 case names.
func checkCompilerAST(test *testing.T, name, digest string) {
	test.Helper()
	root := compilerCorpus(test)
	data, err := os.ReadFile(filepath.Join(root, "src/compiler", name))
	if err != nil {
		test.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		test.Fatal("compiler corpus content pin differs", name)
	}
	checkAST(test, string(data))
}
func TestParseDriverCompiler000(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"_namespaces/ts.moduleSpecifiers.ts",
		"d858679e4d64f2879cb625cecf199bb93f44e414e3c34fd55540fcccfef9c07e",
	)
}
func TestParseDriverCompiler001(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"_namespaces/ts.performance.ts",
		"e3d39c7ea64f27347b698a1951a76a014cd0c1c51565d6ed9182d7d4f064ddda",
	)
}
func TestParseDriverCompiler002(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "_namespaces/ts.ts", "74b899b9d4a9c709b9daa248cb9107227a705f0189eee82271ee84d3bbfc482b")
}
func TestParseDriverCompiler003(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "binder.ts", "b2a0d4d09bfdd389297e346776d12297e4add7e0b62ba2015862e6b50dcb2aa3")
}
func TestParseDriverCompiler004(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "builder.ts", "6e5b07ce2c67226096cad19391791073b2e37164d4e702bf77e7b2bc28ae9814")
}
func TestParseDriverCompiler005(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "builderPublic.ts", "a6a304c41be6a7b33febcae8dd98955f90313c7e2d56af02f476e6b9722b82d6")
}
func TestParseDriverCompiler006(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "builderState.ts", "51ef79dcaebb2e9b09c5270d07ca8f4abf3eb04e4849eb9014bf2d55f2c570a5")
}
func TestParseDriverCompiler007(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "builderStatePublic.ts", "93138f0f93fd9dff071c3def9feebc5f88fbe65bcb028d4ca8871d3155b9b522")
}
func TestParseDriverCompiler008(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "checker.ts", "ffe288edd0eae68f65e4b81b5bbfd4fe5fbed62b55246dc140813078555050fb")
}
func TestParseDriverCompiler009(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "commandLineParser.ts", "e48e42605eef2b1880947b658187cf455b3c91aa4bcf32007ab7296b088009ab")
}
func TestParseDriverCompiler010(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "core.ts", "c35b2cfa5216d4979aff3f963071c979d0b57d2242649df3dc6cb7c9f4a10b1b")
}
func TestParseDriverCompiler011(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "corePublic.ts", "899b3d3421836d6345090bce67d16d6c73c944762bfb0b5483a1e2ae192c9b09")
}
func TestParseDriverCompiler012(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "debug.ts", "38dfd2cdfd58fb8a03a6c815444aa8d6876c69f7e37d96702ece4e3da836bc8b")
}
func TestParseDriverCompiler013(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "emitter.ts", "910f74394f8f4cbef48fe25387e5da6224742754b9a7699ef77312af6f453d63")
}
func TestParseDriverCompiler014(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "executeCommandLine.ts", "1bb233b3cac80fcf8572358b07e59988e59d6fb40445c16728a4a7c7ece43cbf")
}
func TestParseDriverCompiler015(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"expressionToTypeNode.ts",
		"43c6b56dfabaf84c3591fc363763aecbd5b4adfa420879b283aeebca2d405462",
	)
}
func TestParseDriverCompiler016(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"factory/baseNodeFactory.ts",
		"86826be85abc6406cec8915574950ab349ba829167a4d3f28332c38dd8b90e0a",
	)
}
func TestParseDriverCompiler017(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "factory/emitHelpers.ts", "aad8b194e1221e2cf2495cdc824d0a99aef86343da11023f39a1666d5c62bdf8")
}
func TestParseDriverCompiler018(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "factory/emitNode.ts", "d4e8f25e7c30e6c6f5fbb72d56812d140f8d15384be4a7bd4da226a47559583a")
}
func TestParseDriverCompiler019(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"factory/nodeChildren.ts",
		"bded8155f089e40a86d99d53e135b8f3a4cbc49f70ff2bdd0b80f0683a2e9606",
	)
}
func TestParseDriverCompiler020(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"factory/nodeConverters.ts",
		"3d5979059f70d4c381bbd5998181441dd6aaa7e15f6564a8f97d446ad85b041d",
	)
}
func TestParseDriverCompiler021(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "factory/nodeFactory.ts", "45b87af49f05d3b75f17537c50ffaecc610fe1496318a36cf51fcf25a0c1a48c")
}
func TestParseDriverCompiler022(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "factory/nodeTests.ts", "47cc22cce04a5fa143e8182bd0f6e5aea2b1b34d58539701a0785fb0b60cf14f")
}
func TestParseDriverCompiler023(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"factory/parenthesizerRules.ts",
		"94aa00739f9d2b7f41ca6c5d0aaa5495a1d4d4b7c8af1919361a64a7abf90764",
	)
}
func TestParseDriverCompiler024(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "factory/utilities.ts", "7b2e75c131e73ee2b98d097fa80786b6a03965d199bd2fddff4e7c953f093a31")
}
func TestParseDriverCompiler025(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"factory/utilitiesPublic.ts",
		"a91c2dcc86ec6711aeae4fb0ad512209abe7a220ef2a0e7d2f04c3adcb7a0ff1",
	)
}
func TestParseDriverCompiler026(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "moduleNameResolver.ts", "b2c971b6674b3312e43cd198ee250dba86a55fea8b06b28388737e4db7b32c7f")
}
func TestParseDriverCompiler027(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "moduleSpecifiers.ts", "9a7623ef273cfe670f11132590583b7628dec59c2d85e36e897ca5deed8f168a")
}
func TestParseDriverCompiler028(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "parser.ts", "7d63eca71e53e31f7c26b3804eefe61a8809aa6989ce0a7ce1377919e81b963f")
}
func TestParseDriverCompiler029(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "path.ts", "4b777f1a03b9f39b4864106e60a43828b5b749a5fae00f68d6faf8f02e0c91bc")
}
func TestParseDriverCompiler030(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "performance.ts", "b12e3cdedb651fe1d94fceae1bb236c2745530ba7fe0c9dc996d61ba22b6c473")
}
func TestParseDriverCompiler031(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "performanceCore.ts", "7f6e15fcf84111901e95303399ba69c9baacc3466a3ad5b5b21633c3d382f7b4")
}
func TestParseDriverCompiler032(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "program.ts", "ffb29a3565664b6da4ddeb62aeb4afeb093fb8935b0fb1793e75aa4387305301")
}
func TestParseDriverCompiler033(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "programDiagnostics.ts", "1def95bb98fb69151be4295c58ccc4599cdb38b4cb710b807f19b7e8433fa8a5")
}
func TestParseDriverCompiler034(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "resolutionCache.ts", "7de45db4fd5e7c66131198ed2eb8b40f9a908056a8a3963614c1314bfca14da6")
}
func TestParseDriverCompiler035(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "scanner.ts", "02f9e8f02e234b0e917e753a7b74f975c9b9b95f7f01ee4bc151081ed43dea49")
}
func TestParseDriverCompiler036(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "semver.ts", "1b5cbc707e9a7e1dd413b3dd69e2e5cb2c816a6675ce6ae6f3a3cc2d6fc0eabe")
}
func TestParseDriverCompiler037(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "sourcemap.ts", "036ab3c39165eb06a3fb145e1467f4f0dfebdb1f76b5faccc34a79db756473e4")
}
func TestParseDriverCompiler038(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "symbolWalker.ts", "6aec8fecf7d57abd557bdbd4a9744ba2a1f3d8fcc9e9b84721158bd4f284300a")
}
func TestParseDriverCompiler039(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "sys.ts", "9fe83b83c72ee53545fd63384f67c0beb31e864d1cc315ba126014d801e2212a")
}
func TestParseDriverCompiler040(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "tracing.ts", "c47b9a7aa59c84205a27249628c671cba4f4f7e60efda89ddaf48d4aedee6f7a")
}
func TestParseDriverCompiler041(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformer.ts", "d13096beeb8d41a78cc48add41d5028b456e23f73c9973fe175c35941257a066")
}
func TestParseDriverCompiler042(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/classFields.ts",
		"d3a0359fad8430943ca68282cc32b12bc61a6631fa93901f6c480f411c4542f4",
	)
}
func TestParseDriverCompiler043(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/classThis.ts",
		"41c65d5bb7518697a301cce563f6f800ed3ba8856372e86998d259a3eb439a6e",
	)
}
func TestParseDriverCompiler044(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/declarations/diagnostics.ts",
		"dbc7849a0535ef63307656557c1037ef15006f7d6246197c095350854b71e721",
	)
}
func TestParseDriverCompiler045(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/declarations.ts",
		"ad431e7fd5bfc1e0c60b4925ff6e4a6d1c56e038733a60d07794477063b1268e",
	)
}
func TestParseDriverCompiler046(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/destructuring.ts",
		"d38490b41ddbc56be090ddebcec8de5cd11ec88104cc4b859a19cdf132e7c640",
	)
}
func TestParseDriverCompiler047(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2015.ts", "79d70178e4923796212e8900090695875e51c9df60a6c604fbe9e43cd0ffbdf4")
}
func TestParseDriverCompiler048(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2016.ts", "3e3ba9728ee63a9c68e8aee6356e84b611ec19c4acbfc7b4b806b49903c2967e")
}
func TestParseDriverCompiler049(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2017.ts", "b18e65782e3ff096fb09a4609db90dddcb0cbdd8a2f1e72caeea19426dbd92b5")
}
func TestParseDriverCompiler050(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2018.ts", "846876b89a5be66536d351c7c89769acc7b8ed45577411f361107963caa09c2d")
}
func TestParseDriverCompiler051(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2019.ts", "99d755a8cd5bdd07d23590330990828ae7c8a33dae9d55926f5b0ab07db84878")
}
func TestParseDriverCompiler052(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2020.ts", "763fb25fb5c40acf6d2482631c79709c9eddd299303292750a57216e392ae1dd")
}
func TestParseDriverCompiler053(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/es2021.ts", "29798cddd385947caa004effdf997df6bc97b9a9b1d9919413dcee3a6a8b5205")
}
func TestParseDriverCompiler054(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/esDecorators.ts",
		"bcef2e876468210c48941dad975a492f49eab7abeb4f75ca603e76ffd5e06c7c",
	)
}
func TestParseDriverCompiler055(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/esnext.ts", "7c566610dda1adbd37eec796f52b00a0ab2c08f42f678028ecf3b9f3516f60d1")
}
func TestParseDriverCompiler056(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/generators.ts",
		"93e6814b36032fd49fbc2f9ff7fa60695430431f3c11a3f03207827fe44c07f5",
	)
}
func TestParseDriverCompiler057(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/jsx.ts", "2d9f268e5764a96ef52c2fd4d8b889bf248914917c6152358666a3effb041e6b")
}
func TestParseDriverCompiler058(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/legacyDecorators.ts",
		"8321e776d83fe3892df9331baa9c1b837736ba5e4130d379a75ef9bfe34d86e8",
	)
}
func TestParseDriverCompiler059(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/module/esnextAnd2015.ts",
		"6c163c8dc720431f02b335d8d5c25efe849039d5805bfa722899f8a2cc2e6945",
	)
}
func TestParseDriverCompiler060(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/module/impliedNodeFormatDependent.ts",
		"9170b66a9feaa1b340e0723e0a0b10c085385c42ff25d39e9a24dd44e2d84a45",
	)
}
func TestParseDriverCompiler061(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/module/module.ts",
		"69802404b3e40beec0a1707813bf556841b81728a31af29ac554e95bc87f9c36",
	)
}
func TestParseDriverCompiler062(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/module/system.ts",
		"89f55cac43736092d29266dbe77b6e3a94c8ff5684bb7d41306a1e544b2caef7",
	)
}
func TestParseDriverCompiler063(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/namedEvaluation.ts",
		"7529078853a659502ab39832a4eec7e0ee6b94df675f63960254c56cc31060e7",
	)
}
func TestParseDriverCompiler064(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/taggedTemplate.ts",
		"800dc3bebb7e7879e68b5f3ed684397e41dd76b73a1130764fefe6007f627ea9",
	)
}
func TestParseDriverCompiler065(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "transformers/ts.ts", "f7a8c9a88c3db3ef48c01b4764f9518c6cf40d37d4d4f9e8192107981973caf2")
}
func TestParseDriverCompiler066(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/typeSerializer.ts",
		"8219081416c67c9ea7f1f8da6125775652eb7d065e8897fd430f50d83dbac5d6",
	)
}
func TestParseDriverCompiler067(test *testing.T) {
	test.Parallel()
	checkCompilerAST(
		test,
		"transformers/utilities.ts",
		"cbddbf83ed9af717b94398ab8b2617f0098406d26712ae816130f40d3ce9c2e5",
	)
}
func TestParseDriverCompiler068(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "tsbuild.ts", "fb3789e399c0c7e1548d3949eab83a47da545cc88fe0ba2eee356926be607b78")
}
func TestParseDriverCompiler069(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "tsbuildPublic.ts", "3337fb2a5fcb5757f2800b03849452f809227ebd74fcffd2b597e17e13dcb2a5")
}
func TestParseDriverCompiler070(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "types.ts", "a7324b3c17fe5eae0e84bd537a9822bdf8351854a81231c82a4fa59e7bfb7a11")
}
func TestParseDriverCompiler071(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "utilities.ts", "9ad3da097b0b7dca95b6866c9f07216bf562e1706fe2fe05553f966709a0ea81")
}
func TestParseDriverCompiler072(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "utilitiesPublic.ts", "4b5b5726bc08d7d4f08645ee66b7322c6d2019fa239e7158e2f642ae2bd964e2")
}
func TestParseDriverCompiler073(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "visitorPublic.ts", "51ed57171507f1e3cb0a268fca6eb42eaa3882ccb184b8e8cb384a08aa38045e")
}
func TestParseDriverCompiler074(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "watch.ts", "c9d074948345a04d85e87d3c4b3e1a2c8328422c20e0b36b3ee9ae1c7a7ad24e")
}
func TestParseDriverCompiler075(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "watchPublic.ts", "74fdacba504a4962aa0b591c385b6f7061742d37122842e55886a77ffa8d2bef")
}
func TestParseDriverCompiler076(test *testing.T) {
	test.Parallel()
	checkCompilerAST(test, "watchUtilities.ts", "0a4919e1976ea0683ea06137ec02c97fe5ad44ca8d0bc709f69eb0230b7a5487")
}

// The stage 3 pin names the corpus. Read its committed blobs as key-time work;
// the caller's checkout only supplies objects, never working-tree contents.
func compilerCorpus(test *testing.T) string {
	test.Helper()
	pinBytes, err := os.ReadFile("../../../stage3/source.json")
	if err != nil {
		test.Fatal(err)
	}
	var source struct {
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(pinBytes, &source); err != nil {
		test.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:  "stage1-parse-compiler-corpus",
		Files: []string{"stage3/source.json", "stage1/cohere/parse/profile_test.go"},
		Flags: []string{"TypeScript compiler at " + source.Commit},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		root := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if root == "" {
			return fmt.Errorf(
				"environment: ADAMIC_TYPESCRIPT_SOURCE must supply stage3/source.json's TypeScript commit; the corpus never fetches",
			)
		}
		files, err := buildcache.PinnedFiles(root, inputs, source.Commit, "parse compiler", func(name string) bool {
			return strings.HasPrefix(name, "src/compiler/") && strings.HasSuffix(name, ".ts")
		})
		if err != nil {
			return err
		}
		for _, file := range files {
			path := filepath.Join(directory, filepath.FromSlash(file.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(path, file.Data, 0644); err != nil {
				return err
			}
		}
		return nil
	})
}
func TestProduct_ParseCompilerCorpus(test *testing.T) { test.Parallel(); compilerCorpus(test) }
