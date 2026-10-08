package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCheckedViewIntersectionOriginalHeritageUnion(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	for _, name := range []string{"good", "wrong"} {
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane7/original/class-augments-good.a"))
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(input), "console.log(`${heritageName(augmentsTag)} ${emitHeritage(augmentsTag)}`);", "console.log(`${emitHeritage(augmentsTag)}`);", 1)
			source = strings.Replace(source, "heritage.end > heritage.pos", "heritage.end > 0", 1)
			if name == "wrong" {
				source = strings.Replace(source, "pos: 13", "pos: 'wrong'", 1)
			}
			source = strings.Replace(source, "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
			path := filepath.Join(t.TempDir(), "heritage-union-"+name+".a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			requireIntersectionOriginalComplete(t, program, manifest, []string{"ExpressionWithTypeArguments", "Identifier", "Node"})
			found := false
			for i := range program.Functions {
				function := &program.Functions[i]
				if function.Name != "emitHeritage" {
					continue
				}
				for index, statement := range function.Body {
					declare, ok := statement.(ir.Declare)
					if !ok {
						continue
					}
					property, ok := declare.Value.(ir.Property)
					if !ok || property.View != "tag.class" {
						continue
					}
					found = true
					if property.ViewContract == 0 {
						t.Fatal("union field read has no interned contract")
					}
					selected := program.ViewContracts[property.ViewContract-1]
					fields := []string{}
					for _, field := range selected.Fields {
						fields = append(fields, field.Name)
					}
					slices.Sort(fields)
					for _, pair := range manifest.Pairs {
						if pair.ID == 92175 && (!selected.IntersectionBounded || selected.Unsupported != "" || !slices.Equal(fields, pair.PresentFields)) {
							t.Fatal("synthetic union read lost original complete bounded obligations")
						}
					}
					if os.Getenv("ADAMIC_INTERSECTION_HERITAGE_UNION_MUTANT") != "" {
						property.ViewContract = 0
						declare.Value = property
						function.Body[index] = declare
					}
				}
			}
			if !found {
				t.Fatal("original union field read missing")
			}
			actual, binary := nativelyUncached(t, program)
			if name == "good" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if name == "good" {
					if difference := disagreement(run{stdout: []byte("true\n")}, got); difference != "" {
						t.Error(difference)
					}
				} else if difference := disagreement(run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: tag.class.pos is not a number; expected number, found string\n")}, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
