package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This pins the refusal before touching pending record storage, not support for MapLike<any>.
func TestCheckedJSONRecordRemainsUnsupported(t *testing.T) {
	const harness = `#include "adamic.h"
#include "checked_json.h"
int main(void) {
 static const adamic_json_contract number={"number","number",NULL,0,NULL,NULL,NULL,0,0,false};
 static const adamic_json_contract dictionary={"object","NumericConfig",&number,0,NULL,NULL,NULL,0,0,false};
 adamic_record *record=adamic_record_new(false);
 adamic_check_json(&record->heap,&dictionary,"raw");
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "record-boundary")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	failure, ok := err.(*exec.ExitError)
	want := "adamic: panic: checked any: raw needs JSON value | undefined, found record object (awaits compiler/records-maplike)\n"
	if !ok || failure.ExitCode() != 70 || stdout.Len() != 0 || stderr.String() != want {
		t.Fatalf("pending record must stop before reading storage: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}

// External data can cycle even though source ownership proofs refuse creating a cycle.
func TestCheckedJSONCycleStopsAtDepth(t *testing.T) {
	const harness = `#include "adamic.h"
#include "checked_json.h"
int main(void) {
 static const char *const names[]={"next"};
 static const bool references[]={true};
 static const adamic_shape shape={1,names,references,NULL};
 static const adamic_json_contract contract={"object","RecursiveConfig",NULL,0,NULL,NULL,NULL,0,0,false};
 adamic_object *object=adamic_object_new(&shape);
 object->slots[0].reference=&object->heap;
 adamic_object_initialized(object)[0]=1;
 adamic_check_json(&object->heap,&contract,"raw");
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "cycle-boundary")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	failure, ok := err.(*exec.ExitError)
	want := "adamic: panic: checked any: raw" + strings.Repeat(".next", 65) + " needs JSON value | undefined, found non-JSON recursion\n"
	if !ok || failure.ExitCode() != 70 || stdout.Len() != 0 || stderr.String() != want {
		t.Fatalf("cyclic data must stop at the depth boundary: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}
