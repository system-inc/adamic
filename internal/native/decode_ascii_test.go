package native

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Each gate-discoverable leaf runs one complete prefix range. The corpus and the decoder are products built ahead and
// shared by every unit of a target; each unit's own work is running its range and the comparison, which always run.
func runDecodeASCIIUnit(test *testing.T, target string, index int) {
	test.Helper()
	if target == "wasi" && os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip("set ADAMIC_TEST_WASI=1")
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		test.Fatal(err)
	}
	phase := time.Now()
	corpusDirectory := decodeASCIICorpus(test)
	binary := decodeASCIIDecoder(test, target)
	test.Logf("product fetch phase %.3fs", time.Since(phase).Seconds())
	// The unit's clock and deadline start after the fetch, as a product's build is its own unit's.
	defer checkGrainBudget(test)()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	run := func(test *testing.T, name string, args ...string) []byte {
		test.Helper()
		command := programCommand(ctx, repository, name, args...)
		output, err := command.CombinedOutput()
		if err != nil {
			test.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
		return output
	}
	phase = time.Now()
	piece := unitRanges(decodePrefixes, decodePrefixUnitSize)[index]
	corpus := filepath.Join(corpusDirectory, piece.name()+".bin")
	var output []byte
	if target == "wasi" {
		output = run(test, "node", "--disable-warning=ExperimentalWarning", "oracle/wasi.mjs", binary, corpus)
	} else {
		output = run(test, binary, corpus)
	}
	test.Logf("run phase %.3fs", time.Since(phase).Seconds())
	want := fmt.Sprintf(
		"prefixes=%d cases=%d Node and baseline identical",
		piece.last-piece.first,
		(piece.last-piece.first)*65*8*2,
	)
	if !strings.Contains(string(output), want) {
		test.Fatalf("missing completed decoder partition %s: %s", piece.name(), output)
	}
	test.Log(string(output))
}

func TestDecodeASCIIUnit00(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 0)
}

func TestDecodeASCIIUnit01(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 1)
}

func TestDecodeASCIIUnit02(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 2)
}

func TestDecodeASCIIUnit03(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 3)
}

func TestDecodeASCIIUnit04(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 4)
}

func TestDecodeASCIIUnit05(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 5)
}

func TestDecodeASCIIUnit06(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 6)
}

func TestDecodeASCIIUnit07(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 7)
}

func TestDecodeASCIIUnit08(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 8)
}

func TestDecodeASCIIUnit09(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 9)
}

func TestDecodeASCIIUnit10(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 10)
}

func TestDecodeASCIIUnit11(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 11)
}

func TestDecodeASCIIUnit12(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 12)
}

func TestDecodeASCIIUnit13(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 13)
}

func TestDecodeASCIIUnit14(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 14)
}

func TestDecodeASCIIUnit15(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 15)
}

func TestDecodeASCIIUnit16(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 16)
}

func TestDecodeASCIIUnit17(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 17)
}

func TestDecodeASCIIUnit18(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 18)
}

func TestDecodeASCIIUnit19(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 19)
}

func TestDecodeASCIIUnit20(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 20)
}

func TestDecodeASCIIUnit21(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 21)
}

func TestDecodeASCIIUnit22(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 22)
}

func TestDecodeASCIIUnit23(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 23)
}

func TestDecodeASCIIUnit24(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 24)
}

func TestDecodeASCIIUnit25(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 25)
}

func TestDecodeASCIIUnit26(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 26)
}

func TestDecodeASCIIUnit27(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 27)
}

func TestDecodeASCIIUnit28(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 28)
}

func TestDecodeASCIIUnit29(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 29)
}

func TestDecodeASCIIUnit30(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 30)
}

func TestDecodeASCIIUnit31(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 31)
}

func TestDecodeASCIIUnit32(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 32)
}

func TestDecodeASCIIUnit33(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 33)
}

func TestDecodeASCIIUnit34(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 34)
}

func TestDecodeASCIIUnit35(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 35)
}

func TestDecodeASCIIUnit36(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 36)
}

func TestDecodeASCIIUnit37(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 37)
}

func TestDecodeASCIIUnit38(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 38)
}

func TestDecodeASCIIUnit39(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 39)
}

func TestDecodeASCIIUnit40(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 40)
}

func TestDecodeASCIIUnit41(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 41)
}

func TestDecodeASCIIUnit42(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 42)
}

func TestDecodeASCIIUnit43(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 43)
}

func TestDecodeASCIIUnit44(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 44)
}

func TestDecodeASCIIWASIUnit00(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 0)
}

func TestDecodeASCIIWASIUnit01(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 1)
}

func TestDecodeASCIIWASIUnit02(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 2)
}

func TestDecodeASCIIWASIUnit03(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 3)
}

func TestDecodeASCIIWASIUnit04(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 4)
}

func TestDecodeASCIIWASIUnit05(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 5)
}

func TestDecodeASCIIWASIUnit06(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 6)
}

func TestDecodeASCIIWASIUnit07(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 7)
}

func TestDecodeASCIIWASIUnit08(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 8)
}

func TestDecodeASCIIWASIUnit09(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 9)
}

func TestDecodeASCIIWASIUnit10(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 10)
}

func TestDecodeASCIIWASIUnit11(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 11)
}

func TestDecodeASCIIWASIUnit12(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 12)
}

func TestDecodeASCIIWASIUnit13(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 13)
}

func TestDecodeASCIIWASIUnit14(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 14)
}

func TestDecodeASCIIWASIUnit15(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 15)
}

func TestDecodeASCIIWASIUnit16(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 16)
}

func TestDecodeASCIIWASIUnit17(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 17)
}

func TestDecodeASCIIWASIUnit18(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 18)
}

func TestDecodeASCIIWASIUnit19(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 19)
}

func TestDecodeASCIIWASIUnit20(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 20)
}

func TestDecodeASCIIWASIUnit21(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 21)
}

func TestDecodeASCIIWASIUnit22(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 22)
}

func TestDecodeASCIIWASIUnit23(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 23)
}

func TestDecodeASCIIWASIUnit24(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 24)
}

func TestDecodeASCIIWASIUnit25(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 25)
}

func TestDecodeASCIIWASIUnit26(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 26)
}

func TestDecodeASCIIWASIUnit27(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 27)
}

func TestDecodeASCIIWASIUnit28(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 28)
}

func TestDecodeASCIIWASIUnit29(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 29)
}

func TestDecodeASCIIWASIUnit30(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 30)
}

func TestDecodeASCIIWASIUnit31(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 31)
}

func TestDecodeASCIIWASIUnit32(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 32)
}

func TestDecodeASCIIWASIUnit33(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 33)
}

func TestDecodeASCIIWASIUnit34(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 34)
}

func TestDecodeASCIIWASIUnit35(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 35)
}

func TestDecodeASCIIWASIUnit36(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 36)
}

func TestDecodeASCIIWASIUnit37(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 37)
}

func TestDecodeASCIIWASIUnit38(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 38)
}

func TestDecodeASCIIWASIUnit39(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 39)
}

func TestDecodeASCIIWASIUnit40(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 40)
}

func TestDecodeASCIIWASIUnit41(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 41)
}

func TestDecodeASCIIWASIUnit42(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 42)
}

func TestDecodeASCIIWASIUnit43(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 43)
}

func TestDecodeASCIIWASIUnit44(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 44)
}
