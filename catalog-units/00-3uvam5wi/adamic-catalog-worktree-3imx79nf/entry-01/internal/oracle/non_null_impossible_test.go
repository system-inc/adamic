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

func TestImpossibleNonNullFixturesAreRefused(t *testing.T) {
	t.Parallel()
	paths := []string{}
	for _, extension := range []string{"a", "ts"} {
		for _, operand := range []string{"undefined", "null"} {
			paths = append(paths, "non_null_refuse_"+operand+"."+extension)
		}
	}
	for _, form := range []string{"argument", "return", "expression", "comparison", "conditional", "logical", "spread", "element", "default"} {
		paths = append(paths, "non_null_refuse_"+form+".a")
	}
	for _, name := range paths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if filepath.Ext(path) == ".ts" {
				if err != nil {
					t.Fatal(err)
				}
				if program.NonNullChecks.Proven != 0 || program.NonNullChecks.Checked != 1 {
					t.Fatalf("want proven 0 checked 1, got %#v", program.NonNullChecks)
				}
				expression := "undefined!"
				if strings.Contains(name, "null.ts") {
					expression = "null!"
				}
				if difference := disagreement(run{stdout: []byte(strings.TrimSuffix(expression, "!") + "\n")}, onNode(t, path)); difference != "" {
					t.Fatal("source Node: " + difference)
				}
				want := run{stderr: []byte("adamic: panic: non-null assertion failed at " + path + ":1:13: " + expression + " is null or undefined\n"), exitCode: 70}
				native, _ := nativelyUncached(t, program)
				for _, got := range []run{native, onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s: %#v", difference, got)
					}
				}
				return
			}
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want impossible assertion Refused, got %v", err)
			}
			if refused.What != "the non-null assertion !" {
				t.Fatalf("wrong refusal: %v", err)
			}
			if refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
				t.Fatalf("wrong fix: %v", err)
			}
			if !strings.HasSuffix(refused.Error(), ": Adamic 0.1 refuses "+refused.What+"; write ?? panic('why it can't be missing'), or narrow and handle the missing case") {
				t.Fatalf("wrong diagnostic: %v", err)
			}
		})
	}
}

func TestPossibleNonNullAdamicAssertionsAreRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"read", "typeerror"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_possible_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "the non-null assertion !" || refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
				t.Fatalf("wrong .a refusal: %v", err)
			}
		})
	}
}
