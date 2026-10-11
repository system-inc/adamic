package yaml

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func goWidth(t *testing.T, cases string) []byte {
	t.Helper()
	goBinary := goWidthOracle(t)
	return run(t, "", nil, goBinary, cases)
}

// goWidthOracle is formatter_comparison over testdata/width_go.go, a product built ahead (#5qykzj5).
func goWidthOracle(t *testing.T) string {
	t.Helper()
	return formatterComparison(t, "go-width", "width_go.go")
}

func TestProduct_YAMLWidthGo(t *testing.T) {
	t.Parallel()
	goWidthOracle(t)
}

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units).
func TestWidthsMatchGo(t *testing.T) {
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
	for _, emoji := range []string{"😀", "👨‍👩‍👦", "👩🏽‍💻", "🏳️‍🌈", "🇺🇸", "©", "©️", "#️⃣", "*⃣", "1️⃣", "❤", "❤️", "❤️‍🔥", "🚶‍♀️", "⛓️‍💥", "🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f"} {
		for _, before := range []string{"", "中", "é", "["} {
			for _, after := range []string{"", "️", "‍😀", ", tail]"} {
				add(before + emoji + after)
			}
		}
	}
	cases := filepath.Join(t.TempDir(), "width-cases.txt")
	if err := os.WriteFile(cases, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	expected := goWidth(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("width_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "width")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	emitted := filepath.Join(t.TempDir(), "width.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		out  []byte
	}{
		{"native ASan/UBSan/LSan", run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)},
		{"Node", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)},
		{"emitted JavaScript", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)},
	} {
		if !bytes.Equal(side.out, expected) {
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	t.Logf("%d width cases: every Unicode scalar and 256 emoji/context sequences; %d exact Go answer bytes", count, len(expected))
}
