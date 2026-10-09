package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// This area predates the compiler's built-in Error classes. Probe the actual
// host constructors and nominal runtime check against independent source Node.
func TestWASIHostErrorIdentity(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/wasi/host_error_identity.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	source := `#include "adamic.h"
#include <stdio.h>
static const adamic_class root = {NULL,0,3,NULL,1u<<30,NULL,NULL,0,false,0,NULL};
static const adamic_class type = {&root,3,3,NULL,(1u<<30)+1,NULL,NULL,0,false,0,NULL};
static const adamic_class range = {&root,3,3,NULL,(1u<<30)+2,NULL,NULL,0,false,0,NULL};
static void report(void) {
 adamic_object *error=adamic_thrown;
 if(error==NULL) {adamic_panic("expected a host error",21);}
#ifdef DROP_TAG
 error->class=NULL;
#endif
 bool is_error=adamic_instanceof(error,&root);
 puts(is_error ? "Error" : "not Error");
 puts(adamic_instanceof(error,&type) ? "TypeError" : adamic_instanceof(error,&range) ? "RangeError" : "other");
 if(is_error) {
  adamic_string *code=error->slots[2].reference,*message=error->slots[1].reference;
  printf("%.*s\n%.*s\n",(int)code->length,code->bytes,(int)message->length,message->bytes);
 }
 adamic_thrown=NULL;
 adamic_release(error);
}
int main(void) {
 adamic_start(0,NULL);
 adamic_string empty=ADAMIC_STRING(""),utf8=ADAMIC_STRING("r"),nul=ADAMIC_STRING("a\0b");
 (void)adamic_fs_file_read_file(&empty,&utf8);report();
 (void)adamic_node_fs_readdir(&empty,NULL);report();
 (void)adamic_fs_file_stat(&empty,true);report();
 (void)adamic_fs_file_read_file(&nul,&utf8);report();
 (void)adamic_node_fs_readdir(&nul,NULL);report();
 (void)adamic_fs_file_close(-1);report();
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "host-errors")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for name, actual := range map[string]run{"native": executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary), "WASI": onWASI(t, source)} {
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatalf("%s: %s\nNode stdout %q stderr %q exit %d\nactual stdout %q stderr %q exit %d", name, difference, truth.stdout, truth.stderr, truth.exitCode, actual.stdout, actual.stderr, actual.exitCode)
		}
	}
	mutant := onWASI(t, "#define DROP_TAG 1\n"+source)
	if mutant.exitCode != 0 || len(mutant.stderr) != 0 || disagreement(truth, mutant) != "stdout differs" {
		t.Fatalf("drop-tag mutant must fail only Node comparison: %+v", mutant)
	}
	t.Logf("drop-tag mutant caught by Node stdout: %q", mutant.stdout)
	if !strings.Contains(string(truth.stdout), "ERR_OUT_OF_RANGE") {
		t.Fatal("Node witness omitted RangeError")
	}
}
