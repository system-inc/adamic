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

func TestLegacyClassOctalNode(test *testing.T) {
	test.Parallel()
	var cases []executionCase
	for index := 0; index < 512; index++ {
		pattern := fmt.Sprintf(`[\%o]`, index)
		for _, input := range []string{string(rune(index)), fmt.Sprint(index), "\x00\x01\n\x1f 0S\xff"} {
			cases = append(cases, executionCase{Pattern: pattern, Input: utf16.Encode([]rune(input))})
		}
	}
	cases = append(cases, executionCase{
		Pattern: `\2()(\12)(foo)\1\0[\0\1\01\123\08\8](\3\03)\5\005\9\009`,
		Input:   []uint16{10, 102, 111, 111, 0, 83, 102, 111, 111, 3, 5, 5, 57, 0, 57},
		Source:  "cohere pattern 620",
	})
	compareExecutionCases(test, cases, true)
}

func TestLargeQuantifierBoundsNode(test *testing.T) {
	test.Parallel()
	var cases []oracleCase
	for _, lo := range []string{
		"7",
		"8",
		"2147483646",
		"2147483647",
		"2147483648",
		"4294967296",
		"9007199254740992",
		"9007199254740993",
		"9223372036854775807",
		"9223372036854775808",
		"18446744073709551615",
		"18446744073709551616",
	} {
		for _, hi := range []string{
			"7",
			"8",
			"2147483646",
			"2147483647",
			"2147483648",
			"4294967296",
			"9007199254740992",
			"9007199254740993",
			"9223372036854775807",
			"9223372036854775808",
			"18446744073709551615",
			"18446744073709551616",
		} {
			for _, flags := range []string{"", "u", "v"} {
				cases = append(cases, oracleCase{fmt.Sprintf("a{%s,%s}", lo, hi), flags, "large bound sweep"})
			}
		}
	}
	// The shared Node syntax oracle is independently exercised below.
	compareSyntaxCases(test, cases)
}

func compareSyntaxCases(test *testing.T, cases []oracleCase) {
	test.Helper()
	input, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		"node",
		"-e",
		`process.stdout.write(JSON.stringify(JSON.parse(require('fs').readFileSync(0,'utf8')).map(c=>{try{new RegExp(c.Pattern,c.Flags);return true}catch(e){if(!(e instanceof SyntaxError))throw e;return false}})))`,
	)
	command.Stdin = bytes.NewReader(input)
	out, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, out)
	}
	var valid []bool
	if decodeError := json.Unmarshal(out, &valid); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(valid) != len(cases) {
		test.Fatal("Node answer count")
	}
	for index, oracleCase := range cases {
		_, parseError := Parse(oracleCase.Pattern, oracleCase.Flags)
		var divergence *V8DivergenceError
		if errors.As(parseError, &divergence) {
			if !valid[index] {
				test.Errorf("Node unexpectedly rejects divergence %s: %v", oracleCase.Pattern, parseError)
			}
			continue
		}
		if (parseError == nil) != valid[index] {
			test.Errorf(
				"DISAGREEMENT %s/%s parser=%v Node=%v",
				oracleCase.Pattern,
				oracleCase.Flags,
				parseError,
				valid[index],
			)
		}
	}
	test.Logf("syntax totals: %d cases", len(cases))
}
