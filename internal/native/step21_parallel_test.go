package native

import "testing"

func step21ParallelUndefined(t *testing.T, options Options) {
	t.Helper()
	binary := parallelHarness(t, "map.c", options)
	want := runWithInput(t, "", "node", "--input-type=module", "-e", "try { Array.from({length:2048}, (_,i)=>i).map(i=>{if(i===257)throw undefined;if(i===900)throw new Error(String(i));return {};}); } catch(value) {console.log('exception '+String(value));}")
	for _, threads := range []string{"1", "4"} {
		got, _ := parallelRun(t, binary, threads, "undefined_exception")
		if got != want {
			t.Fatalf("threads %s: native %q Node %q", threads, got, want)
		}
	}
}
func TestParallelUndefinedExceptionSanitized(t *testing.T) {
	t.Parallel()
	step21ParallelUndefined(t, Options{Sanitize: true})
}
func TestParallelUndefinedExceptionCounted(t *testing.T) {
	t.Parallel()
	step21ParallelUndefined(t, Options{Count: true})
}
func TestParallelUndefinedExceptionRaces(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	step21ParallelUndefined(t, Options{ThreadSanitize: true})
}
