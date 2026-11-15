package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// checkBoundArguments counts only the functions the allocation flow says can be a bound value
// (boundValueReach, #04f5s0s). This hears the reach where each bind asks it. Language's b01, whose
// narrow holds only seen, which declares an index its bound type doesn't name, reaches seen's
// functions alone, known. known_reach's unary, holding only one, reaches one's functions alone, known,
// so two, a two-parameter function also held as Unary, never counts and the bind lowers.
// absent_property, function_bind.a's honest `absentProperty.f?.bind(otherReceiver)` moved out, is
// heard as it is today: unknown, since its field read projects through the literal, where the absent
// field or any closure call in the program (shape_flow.go's opaqueCalls) leaves the projection unknown,
// so it falls back to every () => number function and stops, a known over-refusal. #x909e6q's lift
// turns that row red, as its witness.
func TestBoundValueReach(t *testing.T) {
	var mutex sync.Mutex
	heard := []string{}
	observe := func(where string, functions []string, known bool) {
		mutex.Lock()
		defer mutex.Unlock()
		heard = append(heard, fmt.Sprintf("%s known %t functions %v", where, known, functions))
	}
	boundReachObserved.Store(&observe)
	defer boundReachObserved.Store(nil)
	for _, probe := range []struct {
		fixture, bind, reached string
		known                  bool
	}{
		{"../oracle/testdata/bind_extra_parameters/narrow_type.a", "narrow.bind(undefined)", "seen", true},
		{"../oracle/testdata/bind_extra_parameters/known_reach.a", "unary.bind(undefined)", "one", true},
		{"../oracle/testdata/bind_extra_parameters/absent_property.a", "absentProperty.f?.bind(otherReceiver)", "", false},
	} {
		source, err := os.ReadFile(probe.fixture)
		if err != nil {
			t.Fatal(err)
		}
		path := writeAgreementSource(t, "main.ts", string(source))
		site := ""
		for number, line := range strings.Split(string(source), "\n") {
			if column := strings.Index(line, probe.bind); column >= 0 && !strings.HasPrefix(line, "//") {
				site = fmt.Sprintf("main.ts:%d:%d", number+1, column+1)
			}
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		_, err = Lower(context.Background(), program)
		t.Logf("%s lowered: %v", filepath.Base(probe.fixture), err)
		mutex.Lock()
		found := []string{}
		for _, line := range heard {
			if strings.HasPrefix(line, filepath.Dir(path)) && strings.Contains(line, site+" ") {
				found = append(found, line)
			}
		}
		mutex.Unlock()
		for _, line := range found {
			t.Logf("%s heard %s", filepath.Base(probe.fixture), strings.TrimPrefix(line, filepath.Dir(path)+"/"))
		}
		if len(found) == 0 {
			t.Fatalf("%s: no reach heard at %s", filepath.Base(probe.fixture), site)
		}
		for _, line := range found {
			if !strings.Contains(line, fmt.Sprintf(" known %t ", probe.known)) {
				t.Fatalf("%s: want known %t, heard %s", filepath.Base(probe.fixture), probe.known, line)
			}
			_, list, _ := strings.Cut(line, " functions [")
			functions := strings.Fields(strings.TrimSuffix(list, "]"))
			if probe.known && (len(functions) == 0 || slices.ContainsFunc(functions, func(name string) bool { return !strings.HasPrefix(name, probe.reached) })) {
				t.Fatalf("%s: want only %s's functions reached, heard %s", filepath.Base(probe.fixture), probe.reached, line)
			}
		}
	}
}
