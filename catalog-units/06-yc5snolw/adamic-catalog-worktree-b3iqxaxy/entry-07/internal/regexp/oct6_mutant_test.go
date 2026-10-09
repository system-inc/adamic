package regexp

import (
	"reflect"
	"testing"
)

// Mutate compiled programs, never the implementation. Each mutant must still
// execute successfully and differ from Node on one of the new inputs.
func TestMatcherOct6Mutants(t *testing.T) {
	t.Parallel()
	mutants := []struct {
		name      string
		caseIndex int
		mutate    func(*Program)
	}{
		{"keep stale nested captures", 0, func(p *Program) {
			for i := range p.code {
				p.code[i].clear = nil
			}
		}},
		{"reverse alternation priority in lookbehind", 7, func(p *Program) {
			for _, i := range p.code {
				if i.look != nil {
					for j := range i.look.code {
						if i.look.code[j].op == opSplit {
							i.look.code[j].x, i.look.code[j].y = i.look.code[j].y, i.look.code[j].x
						}
					}
				}
			}
		}},
		{"ignore repeated backreferences", 15, func(p *Program) {
			for i := range p.code {
				if p.code[i].op == opReference {
					p.code[i].references = nil
				}
			}
		}},
		{"case sensitive non-ASCII backreference", 21, func(p *Program) {
			for i := range p.code {
				if p.code[i].op == opReference {
					p.code[i].flags.IgnoreCase = false
				}
			}
		}},
	}
	cases := oct6Cases()
	expected := nodeResults(t, cases)
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			c := cases[mutant.caseIndex]
			p, err := Compile(c.Pattern, c.Flags)
			if err != nil {
				t.Fatal(err)
			}
			control, err := matcherResult(p.New(), c.Input)
			if err != nil || !reflect.DeepEqual(control, expected[mutant.caseIndex]) {
				t.Fatalf("control: %v %+v", err, control)
			}
			mutant.mutate(p)
			r := p.New()
			r.StepLimit = 100000
			got, err := matcherResult(r, c.Input)
			if err != nil {
				t.Fatalf("mutant must execute, not hit a limit: %v", err)
			}
			if reflect.DeepEqual(got, expected[mutant.caseIndex]) {
				t.Fatal("mutant survived")
			}
			t.Logf("caught: mutant=%+v Node=%+v", got, expected[mutant.caseIndex])
		})
	}
	loops := oct6Loops()
	want := oct6NodeLoops(t, loops)
	for _, mutant := range []struct {
		name   string
		index  int
		mutate func(*Program)
	}{
		{"global ignores lastIndex", 1, func(p *Program) { p.flags.Global = false }},
		{"sticky searches across a gap", 6, func(p *Program) { p.flags.Sticky = false; p.flags.Global = true }},
		{"empty pattern consumes a unit", 3, func(p *Program) {
			p.code = append([]instruction{{op: opSet, direction: 1, set: characterSet{ranges: []RuneRange{{0, 0xffff}}}}}, p.code...)
		}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			loop := loops[mutant.index]
			p, err := Compile(loop.Pattern, loop.Flags)
			if err != nil {
				t.Fatal(err)
			}
			mutant.mutate(p)
			r := p.New()
			r.LastIndex = loop.LastIndex
			r.StepLimit = 100000
			caught := false
			for step := 0; step < loop.Steps; step++ {
				got, err := matcherResult(r, loop.Input)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want[mutant.index][step]) {
					caught = true
					t.Logf("caught at step %d: mutant=%+v Node=%+v", step, got, want[mutant.index][step])
					break
				}
			}
			if !caught {
				t.Fatal("mutant survived")
			}
		})
	}
}
