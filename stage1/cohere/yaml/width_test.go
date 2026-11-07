package yaml

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

func goWidth(test *testing.T, cases string) []byte {
	test.Helper()
	goBinary := goWidthOracle(test)
	return run(test, "", nil, goBinary, cases)
}

// goWidthOracle is formatter_comparison over testdata/width_go.go, a product built ahead (#5qykzj5).
func goWidthOracle(test *testing.T) string { return oracles.Width(test) }

func TestProduct_YamlFindsWidthGo(test *testing.T) {
	test.Parallel()
	goWidthOracle(test)
}

func TestWidthsMatchGo(test *testing.T) {
	test.Parallel()
	binary := yamlWidthBuild.binary(test, "width_main.ts")
	emitted := yamlWidthBuild.emitted(test, "width_main.ts")
	var input strings.Builder
	count := 0
	add := func(text string) {
		for index, point := range []rune(text) {
			if index > 0 {
				input.WriteByte(',')
			}
			fmt.Fprintf(&input, "%d", point)
		}
		input.WriteByte('\n')
		count++
	}
	for point := 0; point <= 0x10ffff; point++ {
		if point >= 0xd800 && point <= 0xdfff {
			continue
		}
		add(string(rune(point)))
	}
	for _, emoji := range []string{
		"😀",
		"👨‍👩‍👦",
		"👩🏽‍💻",
		"🏳️‍🌈",
		"🇺🇸",
		"©",
		"©️",
		"#️⃣",
		"*⃣",
		"1️⃣",
		"❤",
		"❤️",
		"❤️‍🔥",
		"🚶‍♀️",
		"⛓️‍💥",
		"🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f",
	} {
		for _, before := range []string{"", "中", "é", "["} {
			for _, after := range []string{"", "️", "‍😀", ", tail]"} {
				add(before + emoji + after)
			}
		}
	}
	cases := filepath.Join(test.TempDir(), "width-cases.txt")
	if writeError := os.WriteFile(cases, []byte(input.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	expected := goWidth(test, cases)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs("width_main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, side := range []struct {
		name string
		out  []byte
	}{
		{"native ASan/UBSan/LSan", run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)},
		{"Node", run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)},
		{
			"emitted JavaScript",
			run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases),
		},
	} {
		if !bytes.Equal(side.out, expected) {
			test.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	test.Logf(
		"%d width cases: every Unicode scalar and 256 emoji/context sequences; %d exact Go answer bytes",
		count,
		len(expected),
	)
}
