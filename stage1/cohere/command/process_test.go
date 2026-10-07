package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProcessPrimitiveAgreesWithNode(t *testing.T) {
	source := absolute(t, "process_probe.ts")
	program := lowered(t, source)
	binary := filepath.Join(t.TempDir(), "probe")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	node, backend := wrapper(t, source, false), wrapper(t, source, true)
	t.Setenv("ADAMIC_PARENT_VALUE", "must disappear")
	scripts := []string{
		"printf 'hello\\n'; exit 0",
		"printf 'before failure\\n'; exit 19",
		"printf 'ordinary exit 127\\n'; exit 127",
		"kill -TERM $$",
		"printf 'child=%s parent=%s\\n' \"$ADAMIC_CHILD_VALUE\" \"${ADAMIC_PARENT_VALUE-unset}\"",
		"printf 'stdin='; cat; printf 'EOF\\n'",
		"printf '\\360\\237\\230\\200\\377\\r\\000'",
		"head -c 196608 /dev/zero | tr '\\000' x; exit 3",
	}
	for _, script := range scripts {
		// An independent direct spawnSync call observes stdout, decoding and status.
		// Go's runProject fixture separately verifies combined stdout/stderr ordering.
		reference := `const cp=require('node:child_process');const os=require('node:os');const env={...process.env};delete env.ADAMIC_PARENT_VALUE;env.ADAMIC_CHILD_VALUE='two';const r=cp.spawnSync('/bin/sh',['-c',process.argv[1]],{env,encoding:'utf8',input:''});if(r.error)throw r.error;console.log(JSON.stringify({output:r.stdout,exitCode:r.status??1,signal:r.signal?os.constants.signals[r.signal]:0}));`
		observation := execute(t, nil, "node", "-e", reference, script)
		if observation.exitCode != 0 || len(observation.stderr) > 0 {
			t.Fatal("Node control failed", observation)
		}
		var decoded struct {
			Output           string
			ExitCode, Signal int
		}
		if err := json.Unmarshal(observation.stdout, &decoded); err != nil {
			t.Fatal(err)
		}
		want := run{stdout: []byte(fmt.Sprintf("%s\t%d\t%d\t\"\"\n", strconv.Quote(decoded.Output), decoded.ExitCode, decoded.Signal))}

		for _, side := range []string{binary, node, backend} {
			got := execute(t, nil, side, script, "")
			if got.exitCode != 0 || len(got.stderr) > 0 || !bytes.Equal(got.stdout, want.stdout) {
				t.Fatalf("%s %q: exit %d stderr %s stdout %q want %q", side, script, got.exitCode, got.stderr, got.stdout, want.stdout)
			}
		}
	}
	t.Logf("%d independent Node spawnSync cases, three execution paths, ASan/UBSan/LeakSanitizer", len(scripts))
}

func TestProcessRuntimeVerdictMutant(t *testing.T) {
	// An isolated mutant run must first pass the independent Node oracle.
	TestProcessPrimitiveAgreesWithNode(t)
	source := absolute(t, "process_probe.ts")
	program := lowered(t, source)
	code := native.C(program)
	if !strings.Contains(code, "adamic_run_process(") {
		t.Fatal("no process call to mutate")
	}
	code = strings.ReplaceAll(code, "adamic_run_process(", "process_verdict_mutant(")
	code = `#include "adamic.h"
static adamic_object *process_verdict_mutant(const adamic_string *file, const adamic_array *args, const adamic_string *dir, const adamic_array *env) {
 adamic_object *result=adamic_run_process(file,args,dir,env);
 result->slots[1].number=0;
 return result;
}
` + code
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := execute(t, nil, binary, "exit 19", "")
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must execute cleanly: %+v", got)
	}
	want := onNode(t, source, "exit 19", "")
	if want.exitCode != 0 || len(want.stderr) != 0 || bytes.Equal(got.stdout, want.stdout) {
		t.Fatalf("verdict mutant survived or oracle failed: got %q want %q", got.stdout, want.stdout)
	}
	t.Logf("native child actually exits 19; clean runtime mutant reports %q instead of Node's %q; output comparison catches it", got.stdout, want.stdout)
}
