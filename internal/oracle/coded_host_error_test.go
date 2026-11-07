package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A Node-coded host error nothing catches ends the program with String(error) after "adamic: panic: ",
// and String() of an error Node builds with makeNodeErrorWithCode or as a SystemError is its class's own
// toString, `${name} [${code}]: ${message}`. So readFileSync of a NUL path, process.exitCode of a fraction,
// utimesSync of an infinite time and rmSync of a directory panic with the code in brackets, and an errno
// error, whose UVException has no toString of its own, panics as `Error: ENOENT: ...`, held to Node
// natively, sanitized and released, and on the JavaScript backend. The fixtures sit in coded_host_error,
// since the flow and fresh corpora lower every program at the top of testdata. Mutant: error_text ignoring
// the coded marker (exceptions.c) turns the four coded rows red natively and leaves the ENOENT row green
// (#z28avqd).
func TestUncaughtCodedHostErrorsAgreeWithNode(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, stderr string }{
		{
			"uncaught_invalid_arg_value.a",
			"adamic: panic: TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\\x00b'\n",
		},
		{
			"uncaught_out_of_range.a",
			"adamic: panic: RangeError [ERR_OUT_OF_RANGE]: The value of \"code\" is out of range. It must be an integer. Received 1.5\n",
		},
		{
			"uncaught_invalid_arg_type.a",
			"adamic: panic: TypeError [ERR_INVALID_ARG_TYPE]: The \"time\" argument must be an instance of Date or an Time in seconds. Received type number (Infinity)\n",
		},
		{
			"uncaught_fs_eisdir.a",
			"adamic: panic: SystemError [ERR_FS_EISDIR]: Path is a directory: rm returned EISDIR (is a directory) .\n",
		},
		{
			"uncaught_enoent.a",
			"adamic: panic: Error: ENOENT: no such file or directory, open '/adamic-coded-host-error-missing/file'\n",
		},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/coded_host_error", probe.file),
			)
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, path)
			if problem := disagreement(
				run{exitCode: 70, stdout: []byte("before\n"), stderr: []byte(probe.stderr)},
				truth,
			); problem != "" {
				test.Fatalf(
					"source Node: %s: exit %d, stdout %q, stderr %q",
					problem,
					truth.exitCode,
					truth.stdout,
					truth.stderr,
				)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatalf("an uncaught host error must lower: %v", err)
			}
			native, _ := natively(test, program)
			for name, got := range map[string]run{
				"native":     native,
				"release":    released(test, program),
				"javascript": onJavaScriptBackend(test, program),
			} {
				if problem := disagreement(truth, got); problem != "" {
					test.Errorf(
						"%s: %s: exit %d, stdout %q, stderr %q",
						name,
						problem,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
		})
	}
}

// A caught Node-coded host error through String and a template over an unknown, which lower to
// ToPrimitive and so to the error's toString (#09dvwny): `name [code]: message` for the four codes a
// program reaches, and `name: message` for a descriptor's RangeError and an errno error, with each
// error's name and code, held to Node natively, sanitized and released, and on the JavaScript backend.
// Mutant: error_text ignoring the coded marker turns it red natively (#z28avqd).
func TestCaughtCodedHostErrorsAgreeWithNode(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/coded_host_error/caught.a"))
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	if problem := disagreement(run{stdout: []byte(caughtCodedHostErrors)}, truth); problem != "" {
		test.Fatalf(
			"source Node: %s: exit %d, stdout %q, stderr %q",
			problem,
			truth.exitCode,
			truth.stdout,
			truth.stderr,
		)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatalf("String and a template of a caught host error must lower: %v", err)
	}
	native, _ := natively(test, program)
	for name, got := range map[string]run{
		"native":     native,
		"release":    released(test, program),
		"javascript": onJavaScriptBackend(test, program),
	} {
		if problem := disagreement(truth, got); problem != "" {
			test.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// caughtCodedHostErrors is what caught.a prints on Node.
const caughtCodedHostErrors = `TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b'
<TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b'>
TypeError ERR_INVALID_ARG_VALUE
TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b'
<TypeError [ERR_INVALID_ARG_VALUE]: The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\x00b'>
TypeError ERR_INVALID_ARG_VALUE
RangeError [ERR_OUT_OF_RANGE]: The value of "code" is out of range. It must be an integer. Received 1.5
<RangeError [ERR_OUT_OF_RANGE]: The value of "code" is out of range. It must be an integer. Received 1.5>
RangeError ERR_OUT_OF_RANGE
SystemError [ERR_FS_EISDIR]: Path is a directory: rm returned EISDIR (is a directory) .
<SystemError [ERR_FS_EISDIR]: Path is a directory: rm returned EISDIR (is a directory) .>
SystemError ERR_FS_EISDIR
TypeError [ERR_INVALID_ARG_TYPE]: The "time" argument must be an instance of Date or an Time in seconds. Received type number (Infinity)
<TypeError [ERR_INVALID_ARG_TYPE]: The "time" argument must be an instance of Date or an Time in seconds. Received type number (Infinity)>
TypeError ERR_INVALID_ARG_TYPE
RangeError: The value of "fd" is out of range. It must be >= 0 && <= 2147483647. Received -1
<RangeError: The value of "fd" is out of range. It must be >= 0 && <= 2147483647. Received -1>
RangeError ERR_OUT_OF_RANGE
Error: ENOENT: no such file or directory, open '/adamic-coded-host-error-missing/file'
<Error: ENOENT: no such file or directory, open '/adamic-coded-host-error-missing/file'>
Error ENOENT
`

// error.toString() of a caught Node-coded error is the class's own bracketed toString, and
// Error.prototype.toString.call(error) is Error.prototype.toString itself, `TypeError: ...`. The first
// is NotYet, and the second refused, a method read as a value. This is the tripwire: whichever change
// lowers one goes red here, and must pin its output to Node, the second plain, which
// adamic_error_to_string isn't for a coded error (#z28avqd).
func TestCaughtCodedHostErrorToStringNotYet(test *testing.T) {
	test.Parallel()
	const message = "The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'a\\x00b'\n"
	for _, probe := range []struct {
		file, stdout, at, says string
		notYet                 bool
	}{
		{"caught_to_string.a", "TypeError [ERR_INVALID_ARG_VALUE]: " + message, "caught_to_string.a:8:", "", true},
		{
			"caught_prototype_to_string.a",
			"TypeError: " + message,
			"caught_prototype_to_string.a:10:15",
			"refuses a method read as a value",
			false,
		},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/coded_host_error", probe.file),
			)
			if err != nil {
				test.Fatal(err)
			}
			if truth := onNode(test, path); disagreement(run{stdout: []byte(probe.stdout)}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(test, path)
			var notYet *lower.NotYet
			if err == nil ||
				errors.As(err, &notYet) != probe.notYet ||
				!strings.Contains(err.Error(), probe.at) ||
				!strings.Contains(err.Error(), probe.says) {
				test.Fatalf(
					"want the caught error's toString stopped at %s, got %v: if it lowers now, pin its output to Node here",
					probe.at,
					err,
				)
			}
		})
	}
}
