package regexp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"
	"unicode/utf16"
)

func TestLegacyClassOctalNode(t *testing.T) {
	t.Parallel()
	var cases []executionCase
	for n := 0; n < 512; n++ {
		pattern := fmt.Sprintf(`[\%o]`, n)
		for _, input := range []string{string(rune(n)), fmt.Sprint(n), "\x00\x01\n\x1f 0S\xff"} {
			cases = append(cases, executionCase{Pattern: pattern, Input: utf16.Encode([]rune(input))})
		}
	}
	cases = append(cases, executionCase{Pattern: `\2()(\12)(foo)\1\0[\0\1\01\123\08\8](\3\03)\5\005\9\009`, Input: []uint16{10, 102, 111, 111, 0, 83, 102, 111, 111, 3, 5, 5, 57, 0, 57}, Source: "cohere pattern 620"})
	compareExecutionCases(t, cases, true)
}

func TestLargeQuantifierBoundsNode(t *testing.T) {
	t.Parallel()
	var cases []oracleCase
	for _, lo := range []string{"7", "8", "2147483646", "2147483647", "2147483648", "4294967296", "9007199254740992", "9007199254740993", "9223372036854775807", "9223372036854775808", "18446744073709551615", "18446744073709551616"} {
		for _, hi := range []string{"7", "8", "2147483646", "2147483647", "2147483648", "4294967296", "9007199254740992", "9007199254740993", "9223372036854775807", "9223372036854775808", "18446744073709551615", "18446744073709551616"} {
			for _, flags := range []string{"", "u", "v"} {
				cases = append(cases, oracleCase{fmt.Sprintf("a{%s,%s}", lo, hi), flags, "large bound sweep"})
			}
		}
	}
	// The shared Node syntax oracle is independently exercised below.
	compareSyntaxCases(t, cases)
}

func compareSyntaxCases(t *testing.T, cases []oracleCase) {
	t.Helper()
	input, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", "-e", `process.stdout.write(JSON.stringify(JSON.parse(require('fs').readFileSync(0,'utf8')).map(c=>{try{new RegExp(c.Pattern,c.Flags);return true}catch(e){if(!(e instanceof SyntaxError))throw e;return false}})))`)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, out)
	}
	var valid []bool
	if err := json.Unmarshal(out, &valid); err != nil {
		t.Fatal(err)
	}
	if len(valid) != len(cases) {
		t.Fatal("Node answer count")
	}
	for i, c := range cases {
		_, err := Parse(c.Pattern, c.Flags)
		var divergence *V8DivergenceError
		if errors.As(err, &divergence) {
			if !valid[i] {
				t.Errorf("Node unexpectedly rejects divergence %s: %v", c.Pattern, err)
			}
			continue
		}
		if (err == nil) != valid[i] {
			t.Errorf("DISAGREEMENT %s/%s parser=%v Node=%v", c.Pattern, c.Flags, err, valid[i])
		}
	}
	t.Logf("syntax totals: %d cases", len(cases))
}
