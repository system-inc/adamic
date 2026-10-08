package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestWASIHostErrorIdentity(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/wasi/host_error_identity.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	t.Logf("fixture counts: %s", counted(t, "internal/oracle/testdata/wasi/host_error_identity.a", false, nil, false, false))
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	source := native.C(program)
	nativeRun, _ := natively(t, program)
	for name, actual := range map[string]run{"JavaScript": onJavaScriptBackend(t, program), "WASI": onWASI(t, source), "native": nativeRun} {
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatalf("%s: %s\nNode %+v\nactual %+v", name, difference, truth, actual)
		}
	}
	if leaked := leaks(t, program, ""); leaked != "" {
		t.Fatal(leaked)
	}
	helper := `static adamic_object *drop_host_error_tag(void) { if(adamic_thrown!=NULL){adamic_thrown->class=NULL;} return adamic_thrown; }
static adamic_string *mutant_read(const adamic_string *path,const adamic_string *flag){adamic_string *result=adamic_fs_file_read_file(path,flag);(void)drop_host_error_tag();return result;}
static adamic_array *mutant_readdir(const adamic_string *path,const adamic_object *options){adamic_array *result=adamic_node_fs_readdir(path,options);(void)drop_host_error_tag();return result;}
static double mutant_close(double descriptor){double result=adamic_fs_file_close(descriptor);(void)drop_host_error_tag();return result;}
static adamic_object *mutant_stat(const adamic_string *path,bool throwing){adamic_object *result=adamic_fs_file_stat(path,throwing);(void)drop_host_error_tag();return result;}
`
	changed := source
	for from, to := range map[string]string{"adamic_fs_file_read_file(": "mutant_read(", "adamic_fs_file_close(": "mutant_close(", "adamic_node_fs_readdir(": "mutant_readdir(", "adamic_fs_file_stat(": "mutant_stat("} {
		changed = strings.ReplaceAll(changed, from, to)
	}
	if changed == source {
		t.Fatal("tag mutant changed nothing")
	}
	changed = strings.Replace(changed, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+helper, 1)
	actual := onWASI(t, changed)
	if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
		t.Fatalf("tag mutant must fail only Node comparison: %+v", actual)
	}
	t.Logf("drop-tag mutant caught by Node stdout: %q", actual.stdout)
}
