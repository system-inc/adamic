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
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	phase := time.Now()
	corpusDirectory := decodeAsciiCorpus(test)
	binary := decodeAsciiDecoder(test, target)
	test.Logf("product fetch phase %.3fs", time.Since(phase).Seconds())
	// The unit's clock and deadline start after the fetch, as a product's build is its own unit's.
	defer checkGrainBudget(test)()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	run := func(test *testing.T, name string, args ...string) []byte {
		test.Helper()
		command := programCommand(timeoutContext, repository, name, args...)
		output, combinedOutputError := command.CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("%s %v: %v\n%s", name, args, combinedOutputError, output)
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

func TestDecodeAsciiUnit00(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 0)
}

func TestDecodeAsciiUnit01(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 1)
}

func TestDecodeAsciiUnit02(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 2)
}

func TestDecodeAsciiUnit03(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 3)
}

func TestDecodeAsciiUnit04(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 4)
}

func TestDecodeAsciiUnit05(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 5)
}

func TestDecodeAsciiUnit06(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 6)
}

func TestDecodeAsciiUnit07(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 7)
}

func TestDecodeAsciiUnit08(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 8)
}

func TestDecodeAsciiUnit09(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 9)
}

func TestDecodeAsciiUnit10(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 10)
}

func TestDecodeAsciiUnit11(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 11)
}

func TestDecodeAsciiUnit12(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 12)
}

func TestDecodeAsciiUnit13(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 13)
}

func TestDecodeAsciiUnit14(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 14)
}

func TestDecodeAsciiUnit15(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 15)
}

func TestDecodeAsciiUnit16(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 16)
}

func TestDecodeAsciiUnit17(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 17)
}

func TestDecodeAsciiUnit18(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 18)
}

func TestDecodeAsciiUnit19(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 19)
}

func TestDecodeAsciiUnit20(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 20)
}

func TestDecodeAsciiUnit21(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 21)
}

func TestDecodeAsciiUnit22(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 22)
}

func TestDecodeAsciiUnit23(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 23)
}

func TestDecodeAsciiUnit24(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 24)
}

func TestDecodeAsciiUnit25(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 25)
}

func TestDecodeAsciiUnit26(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 26)
}

func TestDecodeAsciiUnit27(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 27)
}

func TestDecodeAsciiUnit28(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 28)
}

func TestDecodeAsciiUnit29(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 29)
}

func TestDecodeAsciiUnit30(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 30)
}

func TestDecodeAsciiUnit31(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 31)
}

func TestDecodeAsciiUnit32(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 32)
}

func TestDecodeAsciiUnit33(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 33)
}

func TestDecodeAsciiUnit34(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 34)
}

func TestDecodeAsciiUnit35(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 35)
}

func TestDecodeAsciiUnit36(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 36)
}

func TestDecodeAsciiUnit37(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 37)
}

func TestDecodeAsciiUnit38(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 38)
}

func TestDecodeAsciiUnit39(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 39)
}

func TestDecodeAsciiUnit40(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 40)
}

func TestDecodeAsciiUnit41(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 41)
}

func TestDecodeAsciiUnit42(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 42)
}

func TestDecodeAsciiUnit43(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 43)
}

func TestDecodeAsciiUnit44(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "native", 44)
}

func TestDecodeAsciiwasiUnit00(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 0)
}

func TestDecodeAsciiwasiUnit01(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 1)
}

func TestDecodeAsciiwasiUnit02(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 2)
}

func TestDecodeAsciiwasiUnit03(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 3)
}

func TestDecodeAsciiwasiUnit04(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 4)
}

func TestDecodeAsciiwasiUnit05(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 5)
}

func TestDecodeAsciiwasiUnit06(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 6)
}

func TestDecodeAsciiwasiUnit07(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 7)
}

func TestDecodeAsciiwasiUnit08(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 8)
}

func TestDecodeAsciiwasiUnit09(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 9)
}

func TestDecodeAsciiwasiUnit10(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 10)
}

func TestDecodeAsciiwasiUnit11(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 11)
}

func TestDecodeAsciiwasiUnit12(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 12)
}

func TestDecodeAsciiwasiUnit13(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 13)
}

func TestDecodeAsciiwasiUnit14(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 14)
}

func TestDecodeAsciiwasiUnit15(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 15)
}

func TestDecodeAsciiwasiUnit16(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 16)
}

func TestDecodeAsciiwasiUnit17(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 17)
}

func TestDecodeAsciiwasiUnit18(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 18)
}

func TestDecodeAsciiwasiUnit19(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 19)
}

func TestDecodeAsciiwasiUnit20(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 20)
}

func TestDecodeAsciiwasiUnit21(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 21)
}

func TestDecodeAsciiwasiUnit22(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 22)
}

func TestDecodeAsciiwasiUnit23(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 23)
}

func TestDecodeAsciiwasiUnit24(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 24)
}

func TestDecodeAsciiwasiUnit25(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 25)
}

func TestDecodeAsciiwasiUnit26(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 26)
}

func TestDecodeAsciiwasiUnit27(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 27)
}

func TestDecodeAsciiwasiUnit28(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 28)
}

func TestDecodeAsciiwasiUnit29(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 29)
}

func TestDecodeAsciiwasiUnit30(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 30)
}

func TestDecodeAsciiwasiUnit31(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 31)
}

func TestDecodeAsciiwasiUnit32(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 32)
}

func TestDecodeAsciiwasiUnit33(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 33)
}

func TestDecodeAsciiwasiUnit34(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 34)
}

func TestDecodeAsciiwasiUnit35(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 35)
}

func TestDecodeAsciiwasiUnit36(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 36)
}

func TestDecodeAsciiwasiUnit37(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 37)
}

func TestDecodeAsciiwasiUnit38(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 38)
}

func TestDecodeAsciiwasiUnit39(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 39)
}

func TestDecodeAsciiwasiUnit40(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 40)
}

func TestDecodeAsciiwasiUnit41(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 41)
}

func TestDecodeAsciiwasiUnit42(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 42)
}

func TestDecodeAsciiwasiUnit43(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 43)
}

func TestDecodeAsciiwasiUnit44(test *testing.T) {
	test.Parallel()
	runDecodeASCIIUnit(test, "wasi", 44)
}
