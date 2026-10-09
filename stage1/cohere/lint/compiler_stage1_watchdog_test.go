package lint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Lowering and C emission are a shared build product, including for native shards.
func TestProduct_CompilerAgreementLowered(t *testing.T) {
	t.Parallel()
	compilerAgreementMeasureProduct(t, compilerAgreementLowered)
}

func compilerAgreementShard(t *testing.T, shard int) {
	t.Helper()
	if os.Getenv("ADAMIC_COMPILER_AGREEMENT_PROBE") == "1" {
		c := compilerAgreementCase{key: compilerAgreementCheckerKey, rule: "no-debugger"}
		target := compilerAgreementOwner(c)*compilerAgreementSides + 1
		want, got := []byte("case 0\nfixed\tunchanged\n"), []byte("case 0\nfixed\tunchanged\n")
		if shard == target {
			got = []byte("case 0\nfixed\tdisagreement\n")
		}
		compilerAgreementCompare(t, got, want)
		return
	}
	cases := compilerAgreementCases(t)
	owner, side := shard/compilerAgreementSides, shard%compilerAgreementSides
	var rows []string
	for _, c := range cases {
		if compilerAgreementOwner(c) == owner {
			rows = append(rows, c.path+"\t"+c.rule)
		}
	}
	if len(rows) == 0 {
		t.Log("empty stable bucket")
		return
	}
	path := manifest(t, rows)
	setupStarted := time.Now()
	_ = compilerAgreementGoOracle(t)
	var module, binary string
	if side == 1 {
		module = filepath.Join(compilerAgreementLowered(t), "lint.mjs")
	}
	if side == 3 || side == 4 {
		binary = compilerAgreementNative(t, side == 4)
	}
	t.Logf("preparation: %.3fs", time.Since(setupStarted).Seconds())
	// Own work is logged, not asserted: every child runs under execute's testguard budget and ceiling,
	// so a hang is bounded there and by Loom's 90 s unit kill. A panic here would end the whole package.
	workStarted := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(workStarted).Seconds()) }()
	want := compilerAgreementOracle(t, owner, path)
	if side == 0 {
		t.Logf("Go: %s, %d cases", want.duration, len(rows))
		return
	}
	var got execution
	switch side {
	case 1:
		runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		got = execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, "--manifest", path)
	case 2:
		runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		got = execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(packageDirectory, "main.ts"), "--manifest", path)
	case 3, 4:
		got = execute(t, "", binary, "--manifest", path)
	}
	compilerAgreementCompare(t, got.output, want.output)
	t.Logf("side %d: %s, oracle %s, %d cases", side, got.duration, want.duration, len(rows))
}

func TestCompilerAndStage1Agree_6484(t *testing.T) { t.Parallel(); compilerAgreementShard(t, 6484) }
