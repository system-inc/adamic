package native_test

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Library validation and runtime's entry-point status must share one exitCode word.
func hostProviderExitState(test *testing.T, mutant bool) {
	test.Helper()
	const truthSource = `process.exitCode=17; console.log(process.exitCode); process.exitCode=4294967297; console.log(process.exitCode); process.exitCode=9007199254740991; console.log(process.exitCode); process.exitCode=-9007199254740991; console.log(process.exitCode); process.exitCode=undefined; console.log(process.exitCode===undefined);`
	truth := hostRun(test, nil, "node", "-e", truthSource)
	if truth.ExitCode != 0 || string(truth.Stdout) != "17\n1\n-1\n1\ntrue\n" || len(truth.Stderr) != 0 {
		test.Fatalf("unexpected Node exitCode contract: %+v", truth)
	}
	const source = `#include "adamic.h"
#include "host_runtime.h"
#include <stdio.h>
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 adamic_host_set_exit_code(true, 17);
 adamic_maybe_number code = adamic_process_exit_code();
 printf("%.0f\n", code.number);
 adamic_process_set_exit_code((adamic_maybe_number){.present=true, .number=4294967297.0});
 printf("%d\n", adamic_host_exit_status());
 adamic_process_set_exit_code((adamic_maybe_number){.present=true, .number=9007199254740991.0});
 printf("%d\n", adamic_host_exit_status());
 adamic_process_set_exit_code((adamic_maybe_number){.present=true, .number=-9007199254740991.0});
 printf("%d\n", adamic_host_exit_status());
 adamic_process_set_exit_code((adamic_maybe_number){0});
 puts(adamic_process_exit_code().present ? "false" : "true");
 return adamic_host_exit_status();
}
`
	directory := ""
	if mutant {
		directory = hostProviderExitMutantDirectory(test)
	}
	binary := filepath.Join(test.TempDir(), "exit-state")
	hostBuild(test, source, binary, directory, native.Options{Sanitize: true})
	got := hostRun(test, nil, binary)
	if got.ExitCode != truth.ExitCode || string(got.Stderr) != string(truth.Stderr) {
		test.Fatalf("exit-state check failed outside Node stdout: %+v", got)
	}
	if mutant {
		if string(got.Stdout) == string(truth.Stdout) {
			test.Fatal("wrong shared status survived Node comparison")
		}
		test.Log("wrong shared status compiled and ran cleanly; Node stdout caught it")
	} else if string(got.Stdout) != string(truth.Stdout) {
		test.Fatalf("library and runtime disagree with Node: got %q want %q", got.Stdout, truth.Stdout)
	}
}

func TestHostProviderExitState(test *testing.T) { test.Parallel(); hostProviderExitState(test, false) }
func TestHostProviderExitStateMutant(test *testing.T) {
	test.Parallel()
	hostProviderExitState(test, true)
}

func hostProviderExitMutantDirectory(test *testing.T) string {
	test.Helper()
	return hostMutant(
		test,
		"process.c",
		"adamic_host_set_exit_code(code.present, value);",
		"adamic_host_set_exit_code(code.present, value + 1);",
	)
}

func TestProduct_HostProviderExitStateMutant(test *testing.T) {
	test.Parallel()
	if _, err := native.RuntimeLibrary(
		hostProviderExitMutantDirectory(test),
		native.Options{Sanitize: true},
	); err != nil {
		test.Fatal(err)
	}
}
