package regexp

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

func TestV8CompatibilityRefusals(test *testing.T) {
	test.Parallel()
	for _, row := range [][3]string{
		{`[\q{a}]`, "iv", "singleton"}, {`[\q{Ss|x}]`, "iv", "singleton"},
		{`(?i:a)[b]`, "v", "scoped"}, {`(?-i:a)[b]`, "iv", "scoped"},
		{`(?i:a)\w`, "u", "scoped"}, {`(?i:a)[\w]`, "u", "scoped"},
		{`(?i:x|[^a-z])`, "", "scoped"},
		{`[\q{ab|a|}]`, "v", "empty"},
	} {
		test.Run(row[0]+"/"+row[1], func(test *testing.T) {
			program, compileError := Compile(row[0], row[1])
			if compileError != nil {
				test.Fatal(compileError)
			}
			_, nativeDeclarationsError := program.NativeDeclarations("probe")
			var divergence *V8DivergenceError
			if !errors.As(nativeDeclarationsError, &divergence) {
				test.Fatalf("expected V8 compatibility refusal, got %v", nativeDeclarationsError)
			}
			test.Log(nativeDeclarationsError)
		})
	}
}

func TestV8VisibleShapesNode(test *testing.T) {
	test.Parallel()
	var cases []executionCase
	for _, prefix := range []string{"", "(?i:^)", "(?-i:^)", "(?i:^)(?:)", "(?i:^)(?-i:)"} {
		for _, atom := range []string{
			`[b]`,
			`[a-z]`,
			`[0-9]`,
			`\w`,
			`[\w]`,
			`\W`,
			`[\W]`,
			`[\w\W]`,
			`[^\w]`,
			`[^\W]`,
			`[^b]`,
			`\p{Lowercase_Letter}`,
			`[\q{a}]`,
			`[a\q{a}]`,
			`[\q{AB}]`,
			`[\q{ab}]`,
			`[\q{Ss|x}]`,
			`[[\q{a}]&&[A]]`,
			`[[A]--[\q{a}]]`,
		} {
			for _, flags := range []string{"u", "iu", "v", "iv"} {
				pattern := prefix + atom
				if _, compileError := Compile(pattern, flags); compileError != nil {
					continue
				}
				for _, input := range []string{
					"",
					"a",
					"A",
					"b",
					"B",
					"x",
					"X",
					"AB",
					"ab",
					"sſ",
					"ſ",
					"K",
					"k",
					"1",
					"🌍",
				} {
					cases = append(
						cases,
						executionCase{Pattern: pattern, Flags: flags, Input: utf16.Encode([]rune(input))},
					)
				}
			}
		}
	}
	results := nodeResults(test, cases)
	accepted, refusedAgree, refusedDisagree := 0, 0, 0
	refusedPatterns := map[string]bool{}
	for index, executionCase := range cases {
		program, compileError := Compile(executionCase.Pattern, executionCase.Flags)
		if compileError != nil {
			test.Fatal(compileError)
		}
		exp := program.New()
		exp.StepLimit = 1000000
		got, matcherResultError := matcherResult(exp, executionCase.Input)
		if matcherResultError != nil {
			test.Fatal(matcherResultError)
		}
		agrees := reflect.DeepEqual(got, results[index])
		if program.NativeCompatibility() != nil {
			refusedPatterns[executionCase.Pattern+"/"+executionCase.Flags] = true
			if agrees {
				refusedAgree++
			} else {
				refusedDisagree++
			}
			continue
		}
		accepted++
		if !agrees {
			test.Errorf(
				"accepted divergence %s/%s input=%v got=%+v Node=%+v",
				executionCase.Pattern,
				executionCase.Flags,
				executionCase.Input,
				got,
				results[index],
			)
		}
	}
	test.Logf(
		"visible shapes: %d cases; %d accepted with zero disagreements; %d refused patterns, %d refused cases agreeing with Node and %d differing",
		len(cases),
		accepted,
		len(refusedPatterns),
		refusedAgree,
		refusedDisagree,
	)
}

func TestV8CompatibilityAcceptedNode(test *testing.T) {
	test.Parallel()
	cases := []executionCase{}
	for _, row := range [][3]string{
		{`[\q{AB}]`, "iv", "ab"}, {`[\q{ab}]`, "iv", "AB"}, {`[\q{Ss}]`, "iv", "sſ"},
		{`[a\q{a}]`, "iv", "A"}, {`[\q{12}]`, "iv", "12"},
		{`(?i:a)b`, "v", "AB"}, {`(?i:a)(?:[b])`, "v", "AB"},
		{`(?i:a)[0-9]`, "v", "A1"}, {`(?i:a)[b]`, "u", "aB"},
		{`(?i:^)[\w\W]`, "u", "K"}, {`(?-i:^)[\w\W]`, "iu", "ſ"},
		{`(?s:a).`, "", "a\n"}, {`(?-s:a).`, "s", "a\n"},
		{`(?m:)^a`, "", "\na"}, {`(?-m:)^a`, "m", "\na"},
		{`(?i:[^a-z]|x)`, "", "B"}, {`(?i:x|[^0-9])`, "", "B"},
		{`(?i:x|[^aA])`, "", "A"}, {`(?i:x|[^a-z])`, "u", "B"},
		{`(?-i:^)[\q{AB}]`, "iv", "ab"},
		{`(?i:a)\w`, "", "aK"},
		{`[[\q{ab|a|}]--[\q{}]]`, "v", "a"}, {`[\q{ab|}]`, "v", "a"},
	} {
		program, compileError := Compile(row[0], row[1])
		if compileError != nil {
			test.Fatal(compileError)
		}
		if nativeCompatibilityError := program.NativeCompatibility(); nativeCompatibilityError != nil {
			test.Fatalf("agreed shape over-refused %s/%s: %v", row[0], row[1], nativeCompatibilityError)
		}
		cases = append(cases, executionCase{Pattern: row[0], Flags: row[1], Input: utf16.Encode([]rune(row[2]))})
	}
	compareExecutionCases(test, cases, true)
}

// Pin the observed Node 24 slow-path failure without allowing a test hang.
// Direct exec ordering agrees with CompileAtom; this is not an ordering oracle.
func TestV8EmptyClassStringReplacement(test *testing.T) {
	test.Parallel()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, combinedOutputError := exec.CommandContext(
		timeoutContext,
		"node",
		"-e",
		`const r=/[\q{ab|a|}]/gv; r.lastIndex=0; console.log("🌍a🌍".replace(r,"")); console.log(JSON.stringify(/[\q{ab|a|}]/v.exec("a"))); console.log(process.version,process.versions.v8);`,
	).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("control: %v: %s", combinedOutputError, out)
	}
	test.Logf("optimized control and conforming direct exec: %s", out)
	if !strings.HasPrefix(string(out), "🌍🌍\n") {
		test.Fatalf("unexpected control %q", out)
	}
	ctxSlow, cancelSlow := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelSlow()
	out, combinedOutputError = exec.CommandContext(
		ctxSlow,
		"node",
		"-e",
		`const r=/[\q{ab|a|}]/gv; const [zero]=[0,NaN]; r.lastIndex=zero; console.log("🌍a🌍".replace(r,""));`,
	).CombinedOutput()
	if ctxSlow.Err() != context.DeadlineExceeded {
		test.Fatalf("recorded slow-path hang changed: err=%v output=%q", combinedOutputError, out)
	}
	test.Log("slow-path reproduction: timed out after 2s (process killed)")
}

func TestV8SurrogateAssertionsNode(test *testing.T) {
	test.Parallel()
	var cases []executionCase
	for _, pattern := range []string{
		`\B`,
		`\b`,
		`(?!\W)`,
		`(?=\W)`,
		`(?!.)`,
		`(?<!.)`,
		`(?:)`,
		`\B(a?)`,
		`\B.`,
		`(?!\W).`,
		`(?=\B)`,
	} {
		for _, flags := range []string{"g", "y", "gu", "yu", "giu", "yiu", "gv", "yv"} {
			for _, input := range []string{"a🌍b", "🌍", " 🌍 ", "ab", ""} {
				for index := uint64(0); index < 6; index++ {
					cases = append(cases, executionCase{
						Pattern:   pattern,
						Flags:     flags,
						Input:     utf16.Encode([]rune(input)),
						LastIndex: index,
						Source:    "V8 surrogate assertions",
					})
				}
			}
		}
	}
	compareExecutionCases(test, cases, true)
}
