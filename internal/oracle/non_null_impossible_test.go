package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, operand := range []string{"undefined", "null"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_refuse_" + operand + ".ts", true, true})
	}
}

func TestImpossibleNonNullFixturesAreRefused(test *testing.T) {
	test.Parallel()
	paths := []string{}
	for _, extension := range []string{"a", "ts"} {
		for _, operand := range []string{"undefined", "null"} {
			paths = append(paths, "non_null_refuse_"+operand+"."+extension)
		}
	}
	for _, form := range []string{
		"argument",
		"return",
		"expression",
		"comparison",
		"conditional",
		"logical",
		"spread",
		"element",
	} {
		paths = append(paths, "non_null_refuse_"+form+".a")
	}
	for _, name := range paths {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", func() string {
				if filepath.Ext(name) == ".a" {
					return "non_null_refused/" + name
				}
				return name
			}()))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if filepath.Ext(path) == ".ts" {
				if err != nil {
					test.Fatal(err)
				}
				if program.NonNullChecks.Proven != 0 || program.NonNullChecks.Checked != 1 {
					test.Fatalf("want proven 0 checked 1, got %#v", program.NonNullChecks)
				}
				expression := "undefined!"
				if strings.Contains(name, "null.ts") {
					expression = "null!"
				}
				if difference := disagreement(
					run{stdout: []byte(strings.TrimSuffix(expression, "!") + "\n")},
					onSliceNode(test, path),
				); difference != "" {
					test.Fatal("source Node: " + difference)
				}
				want := run{stderr: []byte("adamic: panic: non-null assertion failed at " +
					path +
					":1:13: " +
					expression +
					" is null or undefined\n"), exitCode: 70}
				native, _ := nativelyUncached(test, program)
				for _, got := range []run{native, onSliceJavaScript(test, program)} {
					if difference := disagreement(want, got); difference != "" {
						test.Fatalf("%s: %#v", difference, got)
					}
				}
				return
			}
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				test.Fatalf("want impossible assertion Refused, got %v", err)
			}
			if refused.What != "the non-null assertion !" {
				test.Fatalf("wrong refusal: %v", err)
			}
			if refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
				test.Fatalf("wrong fix: %v", err)
			}
			if !strings.HasSuffix(refused.Error(), ": Adamic 0.1 refuses "+
				refused.What+
				"; write ?? panic('why it can't be missing'), or narrow and handle the missing case") {
				test.Fatalf("wrong diagnostic: %v", err)
			}
		})
	}
}

func TestPossibleNonNullAdamicAssertionsAreRefused(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"read", "typeerror"} {
		test.Run(name, func(test *testing.T) {
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/non_null_refused/non_null_possible_"+name+".a"),
			)
			if err != nil {
				test.Fatal(err)
			}
			_, err = lowered(test, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) ||
				refused.What != "the non-null assertion !" ||
				refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
				test.Fatalf("wrong .a refusal: %v", err)
			}
		})
	}
}

// An unused literal default is a slot initializer, not an assertion expression.
func TestLiteralDefaultPlaceholderAgreesWithNode(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_refuse_default.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := onSliceNode(test, path)
	native, binary := nativelyUncached(test, program)
	for _, got := range []run{native, onSliceJavaScript(test, program)} {
		if difference := disagreement(source, got); difference != "" {
			test.Fatal(difference)
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
