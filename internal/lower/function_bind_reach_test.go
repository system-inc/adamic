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
// (boundValueReach, #04f5s0s). This hears the reach where each bind asks it: Language's b01, whose
// narrow holds only seen, which declares an index its bound type doesn't name, reaches seen's
// functions alone, and function_bind.a's honest `absentProperty.f?.bind(otherReceiver)`, whose field
// is absent, reaches no function, so it lowers as on main. Neither is unknown, so neither falls back to
// every function of its type.
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
	for _, probe := range []struct{ fixture, extension, site string }{
		{"../oracle/testdata/bind_extra_parameters/narrow_type.a", ".ts", "main.ts:30:15"},
		{"../oracle/testdata/function_bind.a", ".a", "main.a:46:29"},
	} {
		source, err := os.ReadFile(probe.fixture)
		if err != nil {
			t.Fatal(err)
		}
		path := writeAgreementSource(t, "main"+probe.extension, string(source))
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		_, err = Lower(context.Background(), program)
		t.Logf("%s lowered: %v", filepath.Base(probe.fixture), err)
		mutex.Lock()
		found := []string{}
		for _, line := range heard {
			if strings.HasPrefix(line, filepath.Dir(path)) && strings.Contains(line, probe.site+" ") {
				found = append(found, line)
			}
		}
		mutex.Unlock()
		for _, line := range found {
			t.Logf("%s heard %s", filepath.Base(probe.fixture), strings.TrimPrefix(line, filepath.Dir(path)+"/"))
		}
		if len(found) == 0 {
			t.Fatalf("%s: no reach heard at %s", filepath.Base(probe.fixture), probe.site)
		}
		for _, line := range found {
			if !strings.Contains(line, " known true ") {
				t.Fatalf("%s: want a known reach, heard %s", filepath.Base(probe.fixture), line)
			}
			_, list, _ := strings.Cut(line, " functions [")
			functions := strings.Fields(strings.TrimSuffix(list, "]"))
			if probe.extension == ".a" && len(functions) != 0 {
				t.Fatalf("%s: want no function reached, heard %s", filepath.Base(probe.fixture), line)
			}
			if probe.extension == ".ts" && (len(functions) == 0 || slices.ContainsFunc(functions, func(name string) bool { return !strings.HasPrefix(name, "seen") })) {
				t.Fatalf("%s: want only seen's functions reached, heard %s", filepath.Base(probe.fixture), line)
			}
		}
	}
}
