package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// This validates the owned normalization boundary without enabling lowering.
func TestCheckedViewTupleArrayNormalizerProbe(t *testing.T) {
	const source = `#include "adamic.h"
#include "view_tuples.h"
#include "view_arrays.h"
#include "view_unions_mixed.h"
#include <stdio.h>
static void check(unsigned char storage, adamic_value input) {
 adamic_array *array = adamic_array_new(1, storage == 10);
 adamic_array_view_storage(array, storage);
 adamic_array_push(array, input);
 adamic_value result;
 if (!adamic_tuple_array_union_at(array, 0, false, true, "probe", "selected", &result)) { adamic_release(array); return; }
 adamic_release(array);
 adamic_view_union_value value = adamic_view_union_heap(result.reference);
 if (value.kind == adamic_view_union_number) printf("%.0f\n", value.payload.number);
 else if (value.kind == adamic_view_union_boolean) printf("%s\n", value.payload.boolean ? "true" : "false");
 else if (value.kind == adamic_view_union_string) adamic_write_line(adamic_stdout, value.payload.reference);
 adamic_release(result.reference);
}
int main(void) {
 setvbuf(stdout, NULL, _IONBF, 0);
 check(1, (adamic_value){.number = 7});
 check(2, (adamic_value){.boolean = true});
 check(7, (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){true,9})});
 check(10, (adamic_value){.reference = adamic_box_number(11)});
 static adamic_string a = ADAMIC_STRING("made"), b = ADAMIC_STRING("text");
 check(10, (adamic_value){.reference = adamic_string_concat(2, (adamic_string *const[]){&a,&b})});
 adamic_array *empty = adamic_array_new(0, false);
 adamic_array_view_storage(empty, 1);
 adamic_value absent;
 if (adamic_tuple_array_union_at(empty, 0, false, true, "probe", "empty", &absent)) return 2;
 if (adamic_tuple_array_union_at(empty, -1, true, true, "probe", "relative", &absent)) return 3;
 adamic_release(empty);
 adamic_array *holes = adamic_array_holes(3, false);
 adamic_array_view_storage(holes, 1);
 if (adamic_tuple_array_union_at(holes, 1, false, true, "probe", "hole", &absent)) return 4;
 adamic_release(holes);
 check(7, (adamic_value){.number = adamic_maybe_number_pack((adamic_maybe_number){false,0})});
 check(10, (adamic_value){.reference = NULL});
 return adamic_process_status();
}`
	node := execute(t, "node", "--eval", "for (const v of [7,true,9,11,['made','text'].join('')]) console.log(v)")
	if difference := disagreement(run{stdout: []byte("7\ntrue\n9\n11\nmadetext\n")}, node); difference != "" {
		t.Fatal("Node: " + difference)
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "normalizer")
		if err := native.Build(source+"\n", binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		actual := execute(t, binary)
		if difference := disagreement(node, actual); difference != "" {
			t.Fatalf("%s; stdout %q stderr %q exit %d", difference, actual.stdout, actual.stderr, actual.exitCode)
		}
		if sanitize {
			report, err := leakcheck.Check(leakcheck.Program{C: source + "\n", Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"), Execute: func(env []string, name string, args ...string) leakcheck.Run {
				actual := executeWith(t, env, name, args...)
				return leakcheck.Run{Stdout: actual.stdout, Stderr: actual.stderr, ExitCode: actual.exitCode}
			}})
			if err != nil || report != "" {
				t.Fatalf("leaks: %v %s", err, report)
			}
		}
	}
}
