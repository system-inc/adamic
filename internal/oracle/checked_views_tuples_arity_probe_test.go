package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func TestCheckedViewTupleArityRuntimeProbe(t *testing.T) {
	const source = `#include "view_tuples.h"
#include <stdio.h>
#include <stdlib.h>
int main(int argc,char **argv) {
 if(argc!=4) return 2;
 size_t count=(size_t)atoi(argv[1]);
 static const char *const names[]={"0","1","2","3"};
 static const bool references[]={false,false,false,false};
 adamic_shape shape={count,names,references,NULL};
 adamic_object *value=adamic_object_new(&shape);
 value->tuple=atoi(argv[2])!=0;
 adamic_view_tuple_range(value,1,2,atoi(argv[3])!=0,"Tuple","selected");
 printf("%zu\n",value->shape->count);
 adamic_release(value);
 return adamic_process_status();
}
`
	code := source
	if os.Getenv("ADAMIC_TUPLE_ARITY_PROBE_MUTANT") == "range" {
		code = strings.Replace(code, `adamic_view_tuple_range(value,1,2,atoi(argv[3])!=0,"Tuple","selected");`, `(void)argv[3];`, 1)
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "arity")
		if err := native.Build(code, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		for _, sample := range []struct{ count, tuple, rest, output, found string }{
			{"1", "1", "0", "1\n", ""}, {"2", "1", "0", "2\n", ""},
			{"0", "1", "0", "", "array"}, {"3", "1", "0", "", "array"}, {"1", "0", "0", "", "object"},
			{"1", "1", "1", "1\n", ""}, {"2", "1", "1", "2\n", ""}, {"4", "1", "1", "4\n", ""},
		} {
			args := []string{sample.count, sample.tuple, sample.rest}
			want := run{stdout: []byte(sample.output)}
			if sample.found != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: selected is not a Tuple; expected Tuple, found " + sample.found + "\n")}
			}
			if sample.found == "" {
				truth := execute(t, "node", "--eval", "console.log(Array(Number(process.argv[1])).length)", sample.count)
				if diff := disagreement(want, truth); diff != "" {
					t.Fatal("Node: " + diff)
				}
			}
			if diff := disagreement(want, executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, args...)); diff != "" {
				t.Fatalf("sanitize=%v count=%s: %s", sanitize, sample.count, diff)
			}
			if sanitize && sample.found == "" {
				report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"), Arguments: func() []string { return args }, Execute: func(env []string, name string, arguments ...string) leakcheck.Run {
					return leakRun(executeWith(t, env, name, arguments...))
				}})
				if err != nil || report != "" {
					t.Fatalf("leaks: %v %s", err, report)
				}
			}
		}
	}
}
