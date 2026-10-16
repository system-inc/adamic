package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

func shippedLintInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	profdata, err := native.Stage1ProfileDataTool()
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Inputs{Name: "stage1-shipped-lint-" + name, Files: append(compilerinputs.Lowering(), "stage1/cohere/lint", "stage1/typescript", "stage1/profiles/training.json"), Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool(profdata, "--version")}}
}
func shippedLintSource(t *testing.T) string {
	t.Helper()
	directory := buildcache.Product(t, shippedLintInputs(t, "source"), func(directory string) error {
		if _, err := registry.Generate("."); err != nil {
			return err
		}
		program, err := load.Load([]string{"main.ts"})
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
		t.Fatal(err)
	}
	return string(data)
}
func shippedLintBinary(t *testing.T) string {
	t.Helper()
	directory := filepath.Join("profiles", runtime.GOOS+"-"+runtime.GOARCH)
	text, err := native.ReadStage1Profile(filepath.Join(directory, "profile.txt"))
	if os.IsNotExist(err) && os.Getenv("ADAMIC_REQUIRE_STAGE1_PROFILES") != "1" {
		t.Skip("profile is a subsequent target-specific piece")
	}
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record native.Stage1ProfileManifest
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	training, err := os.ReadFile("../../profiles/training.json")
	if err != nil {
		t.Fatal(err)
	}
	source := shippedLintSource(t)
	if err := native.CheckStage1ProfileManifest(source, text, fmt.Sprintf("%x", sha256.Sum256(training)), record); err != nil {
		t.Fatal(err)
	}
	inputs := shippedLintInputs(t, "binary")
	inputs.Flags = []string{source}
	product := buildcache.Product(t, inputs, func(output string) error {
		return native.Build(source, filepath.Join(output, "binary"), native.Options{Release: true, Split: true, Profile: filepath.Join(directory, "profile.txt")})
	})
	return filepath.Join(product, "binary")
}
func TestProduct_LintShippedSource(t *testing.T)  { t.Parallel(); shippedLintSource(t) }
func TestProduct_LintShippedProfile(t *testing.T) { t.Parallel(); t.Log(shippedLintBinary(t)) }
func TestLintShippedProfileAgreesWithNode(t *testing.T) {
	t.Parallel()
	binary := shippedLintBinary(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	row := compilationCase(t)
	// compilationCase is repository-relative; the driver reads it from the root.
	fields := strings.SplitN(row, "\t", 2)
	path, err := filepath.Abs(filepath.Join("../../..", fields[0]))
	if err != nil {
		t.Fatal(err)
	}
	manifest := manifest(t, []string{path + "\t" + fields[1]})
	got := execute(t, "", binary, "--manifest", manifest)
	want := execute(t, "", "node", "--disable-warning=ExperimentalWarning", "../../../oracle/node.mjs", "main.ts", "--manifest", manifest)
	oracle := execute(t, "", shippedLintNoVarOracle(t), "--manifest", manifest)
	if !bytes.Equal(got.output, oracle.output) {
		t.Fatal("shipping lint differs from Go")
	}
	if !bytes.Equal(got.output, want.output) {
		t.Fatal("shipping lint differs from Node", directory)
	}
}

func lintProfilePartBytes(part native.Stage1ProfileUnit, content []byte) error {
	if fmt.Sprintf("%x", sha256.Sum256(content)) != part.SHA256 {
		return fmt.Errorf("profile data part changed: %s", part.Name)
	}
	return nil
}
func TestLintProfilePartByteControl(t *testing.T) {
	t.Parallel()
	content := []byte("profile data")
	part := native.Stage1ProfileUnit{SHA256: fmt.Sprintf("%x", sha256.Sum256(content))}
	if err := lintProfilePartBytes(part, content); err != nil {
		t.Fatal(err)
	}
	content[0] ^= 1
	if err := lintProfilePartBytes(part, content); err == nil {
		t.Fatal("profile part byte mutant survived")
	}
}
func TestLintProfileDataParts(t *testing.T) {
	t.Parallel()
	directory := filepath.Join("profiles", runtime.GOOS+"-"+runtime.GOARCH)
	data, err := os.ReadFile(filepath.Join(directory, "parts.json"))
	if os.IsNotExist(err) {
		t.Skip("profile data pieces have not landed")
	}
	if err != nil {
		t.Fatal(err)
	}
	var parts []native.Stage1ProfileUnit
	if err := json.Unmarshal(data, &parts); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(directory, "profile-*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != len(files) || len(parts) == 0 {
		t.Fatal("profile part inventory differs")
	}
	tool, err := native.Stage1ProfileDataTool()
	if err != nil {
		t.Fatal(err)
	}
	for i, part := range parts {
		if part.Name != fmt.Sprintf("profile-%02d.txt", i) {
			t.Fatal("invalid part inventory", part.Name)
		}
		content, err := os.ReadFile(filepath.Join(directory, part.Name))
		if err != nil {
			t.Fatal(err)
		}
		if err := lintProfilePartBytes(part, content); err != nil {
			t.Fatal(err)
		}
		command := toolchain.Command(tool, "merge", filepath.Join(directory, part.Name), "-o", filepath.Join(t.TempDir(), "profile.profdata"))
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("invalid profile records: %v %s", err, output)
		}
	}
}

// This product references only the upstream no-var rule. The shared all-rules Go
// oracle exceeds one cold leaf's budget; its other owners keep their own tests.
func shippedLintNoVarOracle(t *testing.T) string {
	t.Helper()
	inputs, err := buildcache.GoBuildInInputs("cohere", "oracle", "./internal/lint/rules/core", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "stage1-profile-no-var-oracle"
	inputs.Files = append(inputs.Files, "stage1/cohere/lint/testdata/oracle.go", "stage1/cohere/lint/rules/no-var/oracle.go", "stage1/cohere/lint/shipped_profile_test.go")
	directory := buildcache.Product(t, inputs, func(directory string) error {
		root, err := filepath.Abs("../../../cohere")
		if err != nil {
			return err
		}
		main, err := filepath.Abs("testdata/oracle.go")
		if err != nil {
			return err
		}
		rule, err := filepath.Abs("rules/no-var/oracle.go")
		if err != nil {
			return err
		}
		registry := filepath.Join(directory, "registry.go.txt")
		text := "package main\nimport \"github.com/system-inc/cohere/internal/lint/rule\"\ntype registeredRule struct{subject rule.Rule;options func([]string)any}\nfunc registeredRules()[]registeredRule{return []registeredRule{{oracleNoVar(),oracleNoVarOptions}}}\n"
		if err := os.WriteFile(registry, []byte(text), 0644); err != nil {
			return err
		}
		defer os.Remove(registry)
		virtual := filepath.Join(root, "adamic_lint_profile")
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(virtual, "main.go"): main, filepath.Join(virtual, "rule.go"): rule, filepath.Join(virtual, "registry.go"): registry}})
		if err != nil {
			return err
		}
		overlay := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(overlay, data, 0644); err != nil {
			return err
		}
		defer os.Remove(overlay)
		args := append(buildcache.ReproducibleGoFlags(), "-tags=lintoracle", "-overlay="+overlay, "-o", filepath.Join(directory, "oracle"), "./adamic_lint_profile")
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		command := lintCommand(ctx, root, "go", append([]string{"build"}, args...)...)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("no-var Go oracle: %v %s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}
func TestProduct_LintShippedNoVarOracle(t *testing.T) { t.Parallel(); shippedLintNoVarOracle(t) }
func lintShippedCompilerCase(t *testing.T, name, digest string) {
	t.Helper()
	root := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if root == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to pinned TypeScript 050880ce")
	}
	path := filepath.Join(root, "src/compiler", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		t.Fatal("compiler content pin differs", name)
	}
	binary := shippedLintBinary(t)
	manifest := manifest(t, []string{path})
	want := execute(t, "", "node", "--disable-warning=ExperimentalWarning", "../../../oracle/node.mjs", "main.ts", "--manifest", manifest)
	got := execute(t, "", binary, "--manifest", manifest)
	if !bytes.Equal(got.output, want.output) {
		t.Fatalf("shipping compiler lint differs from Node: native=%d Node=%d", len(got.output), len(want.output))
	}
}
func TestLintShippedCompiler000(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "_namespaces/ts.moduleSpecifiers.ts", "d858679e4d64f2879cb625cecf199bb93f44e414e3c34fd55540fcccfef9c07e")
}
func TestLintShippedCompiler001(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "_namespaces/ts.performance.ts", "e3d39c7ea64f27347b698a1951a76a014cd0c1c51565d6ed9182d7d4f064ddda")
}
func TestLintShippedCompiler002(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "_namespaces/ts.ts", "74b899b9d4a9c709b9daa248cb9107227a705f0189eee82271ee84d3bbfc482b")
}
func TestLintShippedCompiler003(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "binder.ts", "b2a0d4d09bfdd389297e346776d12297e4add7e0b62ba2015862e6b50dcb2aa3")
}
func TestLintShippedCompiler004(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "builder.ts", "6e5b07ce2c67226096cad19391791073b2e37164d4e702bf77e7b2bc28ae9814")
}
func TestLintShippedCompiler005(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "builderPublic.ts", "a6a304c41be6a7b33febcae8dd98955f90313c7e2d56af02f476e6b9722b82d6")
}
func TestLintShippedCompiler006(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "builderState.ts", "51ef79dcaebb2e9b09c5270d07ca8f4abf3eb04e4849eb9014bf2d55f2c570a5")
}
func TestLintShippedCompiler007(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "builderStatePublic.ts", "93138f0f93fd9dff071c3def9feebc5f88fbe65bcb028d4ca8871d3155b9b522")
}
func TestLintShippedCompiler008(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "checker.ts", "ffe288edd0eae68f65e4b81b5bbfd4fe5fbed62b55246dc140813078555050fb")
}
func TestLintShippedCompiler009(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "commandLineParser.ts", "e48e42605eef2b1880947b658187cf455b3c91aa4bcf32007ab7296b088009ab")
}
func TestLintShippedCompiler010(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "core.ts", "c35b2cfa5216d4979aff3f963071c979d0b57d2242649df3dc6cb7c9f4a10b1b")
}
func TestLintShippedCompiler011(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "corePublic.ts", "899b3d3421836d6345090bce67d16d6c73c944762bfb0b5483a1e2ae192c9b09")
}
func TestLintShippedCompiler012(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "debug.ts", "38dfd2cdfd58fb8a03a6c815444aa8d6876c69f7e37d96702ece4e3da836bc8b")
}
func TestLintShippedCompiler013(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "emitter.ts", "910f74394f8f4cbef48fe25387e5da6224742754b9a7699ef77312af6f453d63")
}
func TestLintShippedCompiler014(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "executeCommandLine.ts", "1bb233b3cac80fcf8572358b07e59988e59d6fb40445c16728a4a7c7ece43cbf")
}
func TestLintShippedCompiler015(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "expressionToTypeNode.ts", "43c6b56dfabaf84c3591fc363763aecbd5b4adfa420879b283aeebca2d405462")
}
func TestLintShippedCompiler016(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/baseNodeFactory.ts", "86826be85abc6406cec8915574950ab349ba829167a4d3f28332c38dd8b90e0a")
}
func TestLintShippedCompiler017(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/emitHelpers.ts", "aad8b194e1221e2cf2495cdc824d0a99aef86343da11023f39a1666d5c62bdf8")
}
func TestLintShippedCompiler018(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/emitNode.ts", "d4e8f25e7c30e6c6f5fbb72d56812d140f8d15384be4a7bd4da226a47559583a")
}
func TestLintShippedCompiler019(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/nodeChildren.ts", "bded8155f089e40a86d99d53e135b8f3a4cbc49f70ff2bdd0b80f0683a2e9606")
}
func TestLintShippedCompiler020(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/nodeConverters.ts", "3d5979059f70d4c381bbd5998181441dd6aaa7e15f6564a8f97d446ad85b041d")
}
func TestLintShippedCompiler021(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/nodeFactory.ts", "45b87af49f05d3b75f17537c50ffaecc610fe1496318a36cf51fcf25a0c1a48c")
}
func TestLintShippedCompiler022(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/nodeTests.ts", "47cc22cce04a5fa143e8182bd0f6e5aea2b1b34d58539701a0785fb0b60cf14f")
}
func TestLintShippedCompiler023(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/parenthesizerRules.ts", "94aa00739f9d2b7f41ca6c5d0aaa5495a1d4d4b7c8af1919361a64a7abf90764")
}
func TestLintShippedCompiler024(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/utilities.ts", "7b2e75c131e73ee2b98d097fa80786b6a03965d199bd2fddff4e7c953f093a31")
}
func TestLintShippedCompiler025(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "factory/utilitiesPublic.ts", "a91c2dcc86ec6711aeae4fb0ad512209abe7a220ef2a0e7d2f04c3adcb7a0ff1")
}
func TestLintShippedCompiler026(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "moduleNameResolver.ts", "b2c971b6674b3312e43cd198ee250dba86a55fea8b06b28388737e4db7b32c7f")
}
func TestLintShippedCompiler027(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "moduleSpecifiers.ts", "9a7623ef273cfe670f11132590583b7628dec59c2d85e36e897ca5deed8f168a")
}
func TestLintShippedCompiler028(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "parser.ts", "7d63eca71e53e31f7c26b3804eefe61a8809aa6989ce0a7ce1377919e81b963f")
}
func TestLintShippedCompiler029(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "path.ts", "4b777f1a03b9f39b4864106e60a43828b5b749a5fae00f68d6faf8f02e0c91bc")
}
func TestLintShippedCompiler030(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "performance.ts", "b12e3cdedb651fe1d94fceae1bb236c2745530ba7fe0c9dc996d61ba22b6c473")
}
func TestLintShippedCompiler031(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "performanceCore.ts", "7f6e15fcf84111901e95303399ba69c9baacc3466a3ad5b5b21633c3d382f7b4")
}
func TestLintShippedCompiler032(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "program.ts", "ffb29a3565664b6da4ddeb62aeb4afeb093fb8935b0fb1793e75aa4387305301")
}
func TestLintShippedCompiler033(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "programDiagnostics.ts", "1def95bb98fb69151be4295c58ccc4599cdb38b4cb710b807f19b7e8433fa8a5")
}
func TestLintShippedCompiler034(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "resolutionCache.ts", "7de45db4fd5e7c66131198ed2eb8b40f9a908056a8a3963614c1314bfca14da6")
}
func TestLintShippedCompiler035(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "scanner.ts", "02f9e8f02e234b0e917e753a7b74f975c9b9b95f7f01ee4bc151081ed43dea49")
}
func TestLintShippedCompiler036(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "semver.ts", "1b5cbc707e9a7e1dd413b3dd69e2e5cb2c816a6675ce6ae6f3a3cc2d6fc0eabe")
}
func TestLintShippedCompiler037(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "sourcemap.ts", "036ab3c39165eb06a3fb145e1467f4f0dfebdb1f76b5faccc34a79db756473e4")
}
func TestLintShippedCompiler038(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "symbolWalker.ts", "6aec8fecf7d57abd557bdbd4a9744ba2a1f3d8fcc9e9b84721158bd4f284300a")
}
func TestLintShippedCompiler039(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "sys.ts", "9fe83b83c72ee53545fd63384f67c0beb31e864d1cc315ba126014d801e2212a")
}
func TestLintShippedCompiler040(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "tracing.ts", "c47b9a7aa59c84205a27249628c671cba4f4f7e60efda89ddaf48d4aedee6f7a")
}
func TestLintShippedCompiler041(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformer.ts", "d13096beeb8d41a78cc48add41d5028b456e23f73c9973fe175c35941257a066")
}
func TestLintShippedCompiler042(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/classFields.ts", "d3a0359fad8430943ca68282cc32b12bc61a6631fa93901f6c480f411c4542f4")
}
func TestLintShippedCompiler043(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/classThis.ts", "41c65d5bb7518697a301cce563f6f800ed3ba8856372e86998d259a3eb439a6e")
}
func TestLintShippedCompiler044(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/declarations/diagnostics.ts", "dbc7849a0535ef63307656557c1037ef15006f7d6246197c095350854b71e721")
}
func TestLintShippedCompiler045(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/declarations.ts", "ad431e7fd5bfc1e0c60b4925ff6e4a6d1c56e038733a60d07794477063b1268e")
}
func TestLintShippedCompiler046(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/destructuring.ts", "d38490b41ddbc56be090ddebcec8de5cd11ec88104cc4b859a19cdf132e7c640")
}
func TestLintShippedCompiler047(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2015.ts", "79d70178e4923796212e8900090695875e51c9df60a6c604fbe9e43cd0ffbdf4")
}
func TestLintShippedCompiler048(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2016.ts", "3e3ba9728ee63a9c68e8aee6356e84b611ec19c4acbfc7b4b806b49903c2967e")
}
func TestLintShippedCompiler049(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2017.ts", "b18e65782e3ff096fb09a4609db90dddcb0cbdd8a2f1e72caeea19426dbd92b5")
}
func TestLintShippedCompiler050(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2018.ts", "846876b89a5be66536d351c7c89769acc7b8ed45577411f361107963caa09c2d")
}
func TestLintShippedCompiler051(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2019.ts", "99d755a8cd5bdd07d23590330990828ae7c8a33dae9d55926f5b0ab07db84878")
}
func TestLintShippedCompiler052(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2020.ts", "763fb25fb5c40acf6d2482631c79709c9eddd299303292750a57216e392ae1dd")
}
func TestLintShippedCompiler053(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/es2021.ts", "29798cddd385947caa004effdf997df6bc97b9a9b1d9919413dcee3a6a8b5205")
}
func TestLintShippedCompiler054(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/esDecorators.ts", "bcef2e876468210c48941dad975a492f49eab7abeb4f75ca603e76ffd5e06c7c")
}
func TestLintShippedCompiler055(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/esnext.ts", "7c566610dda1adbd37eec796f52b00a0ab2c08f42f678028ecf3b9f3516f60d1")
}
func TestLintShippedCompiler056(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/generators.ts", "93e6814b36032fd49fbc2f9ff7fa60695430431f3c11a3f03207827fe44c07f5")
}
func TestLintShippedCompiler057(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/jsx.ts", "2d9f268e5764a96ef52c2fd4d8b889bf248914917c6152358666a3effb041e6b")
}
func TestLintShippedCompiler058(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/legacyDecorators.ts", "8321e776d83fe3892df9331baa9c1b837736ba5e4130d379a75ef9bfe34d86e8")
}
func TestLintShippedCompiler059(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/module/esnextAnd2015.ts", "6c163c8dc720431f02b335d8d5c25efe849039d5805bfa722899f8a2cc2e6945")
}
func TestLintShippedCompiler060(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/module/impliedNodeFormatDependent.ts", "9170b66a9feaa1b340e0723e0a0b10c085385c42ff25d39e9a24dd44e2d84a45")
}
func TestLintShippedCompiler061(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/module/module.ts", "69802404b3e40beec0a1707813bf556841b81728a31af29ac554e95bc87f9c36")
}
func TestLintShippedCompiler062(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/module/system.ts", "89f55cac43736092d29266dbe77b6e3a94c8ff5684bb7d41306a1e544b2caef7")
}
func TestLintShippedCompiler063(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/namedEvaluation.ts", "7529078853a659502ab39832a4eec7e0ee6b94df675f63960254c56cc31060e7")
}
func TestLintShippedCompiler064(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/taggedTemplate.ts", "800dc3bebb7e7879e68b5f3ed684397e41dd76b73a1130764fefe6007f627ea9")
}
func TestLintShippedCompiler065(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/ts.ts", "f7a8c9a88c3db3ef48c01b4764f9518c6cf40d37d4d4f9e8192107981973caf2")
}
func TestLintShippedCompiler066(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/typeSerializer.ts", "8219081416c67c9ea7f1f8da6125775652eb7d065e8897fd430f50d83dbac5d6")
}
func TestLintShippedCompiler067(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "transformers/utilities.ts", "cbddbf83ed9af717b94398ab8b2617f0098406d26712ae816130f40d3ce9c2e5")
}
func TestLintShippedCompiler068(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "tsbuild.ts", "fb3789e399c0c7e1548d3949eab83a47da545cc88fe0ba2eee356926be607b78")
}
func TestLintShippedCompiler069(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "tsbuildPublic.ts", "3337fb2a5fcb5757f2800b03849452f809227ebd74fcffd2b597e17e13dcb2a5")
}
func TestLintShippedCompiler070(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "types.ts", "a7324b3c17fe5eae0e84bd537a9822bdf8351854a81231c82a4fa59e7bfb7a11")
}
func TestLintShippedCompiler071(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "utilities.ts", "9ad3da097b0b7dca95b6866c9f07216bf562e1706fe2fe05553f966709a0ea81")
}
func TestLintShippedCompiler072(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "utilitiesPublic.ts", "4b5b5726bc08d7d4f08645ee66b7322c6d2019fa239e7158e2f642ae2bd964e2")
}
func TestLintShippedCompiler073(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "visitorPublic.ts", "51ed57171507f1e3cb0a268fca6eb42eaa3882ccb184b8e8cb384a08aa38045e")
}
func TestLintShippedCompiler074(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "watch.ts", "c9d074948345a04d85e87d3c4b3e1a2c8328422c20e0b36b3ee9ae1c7a7ad24e")
}
func TestLintShippedCompiler075(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "watchPublic.ts", "74fdacba504a4962aa0b591c385b6f7061742d37122842e55886a77ffa8d2bef")
}
func TestLintShippedCompiler076(t *testing.T) {
	t.Parallel()
	lintShippedCompilerCase(t, "watchUtilities.ts", "0a4919e1976ea0683ea06137ec02c97fe5ad44ca8d0bc709f69eb0230b7a5487")
}
