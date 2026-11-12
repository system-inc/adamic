package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's two calls of an optional host method through a non-null assertion lower, each as o.m() does,
// keeping its receiver (#p2n17c4 head A): program.ts:471:60's `system.realpath!(path)` inside the arrow
// createCompilerHostWorker hands on, and watch.ts:747's `host.trace!(s)` in one arm of a conditional.
// They stopped as a method read as a value only because the guard didn't see a callee through the `!`;
// #0bxc4jh's assertedCallee (refusals.go) taught it, and these rows hold tsc's own shapes to that rule.
// The call itself is 1e2e35f7's ir.MethodCallee road, whose words when the member is absent the oracle
// holds (internal/oracle/asserted_method_test.go). Mutant: assertedCallee dropped from refusals.go's
// method-value guard, and both rows stop as a method read as a value.
func TestAssertedOptionalHostMethodCallsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"program.ts:471's realpath", "interface System {\n    realpath?(path: string): string;\n}\nfunction realpathOf(system: System): (path: string) => string {\n    return (path: string) => system.realpath!(path);\n}\nconsole.log(realpathOf({ realpath: (path: string): string => `/real${path}` })('/a'));\n"},
		{"watch.ts:747's trace", "interface WatchHost {\n    trace?(s: string): void;\n}\nconst noop = (_s: string): void => {};\nfunction writeLogOf(host: WatchHost, verbose: boolean): (s: string) => void {\n    return verbose ? (s => host.trace!(s)) : noop;\n}\nwriteLogOf({ trace: (s: string): void => {\n    console.log(s);\n} }, true)('traced');\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, probe.source, ".ts"); err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// tsc's three real detaches stay refused as a method read as a value, as Adamic's ruling on #p2n17c4
// excludes them: a host method taken to either side of || and called or stored loses its object, since
// Node calls the result with this undefined. program.ts:459 and 478 call
// `(compilerHost.directoryExists || system.directoryExists)(path)` and the same of createDirectory, and
// program.ts:1602 stores `host.hasInvalidatedResolutions || returnFalse`. assertedCallee sees only the
// operand of a called `!`, so none of these is a callee. Mutant: the method-value guard skipping an
// operand of a binary expression, and each row lowers or stops for another reason.
func TestHostMethodDetachesThroughOrStayRefused(t *testing.T) {
	t.Parallel()
	const hosts = "interface System {\n    directoryExists(path: string): boolean;\n    createDirectory(path: string): void;\n}\ninterface CompilerHost {\n    directoryExists?(path: string): boolean;\n    createDirectory?(path: string): void;\n}\nconst system: System = { directoryExists: (path: string): boolean => path.length > 1, createDirectory: (path: string): void => {\n    console.log(path);\n} };\n"
	for _, probe := range []struct{ name, source string }{
		{"program.ts:459's directoryExists", hosts + "function exists(compilerHost: CompilerHost, path: string): boolean {\n    return (compilerHost.directoryExists || system.directoryExists)(path);\n}\nconsole.log(String(exists({}, '/a')));\n"},
		{"program.ts:478's createDirectory", hosts + "function creator(compilerHost: CompilerHost): (path: string) => void {\n    return path => (compilerHost.createDirectory || system.createDirectory)(path);\n}\ncreator({})('/a');\n"},
		{"program.ts:1602's hasInvalidatedResolutions", "interface ProgramHost {\n    hasInvalidatedResolutions?(fileName: string): boolean;\n}\nfunction returnFalse(): false {\n    return false;\n}\nfunction invalidated(host: ProgramHost, fileName: string): boolean {\n    const hasInvalidatedResolutions = host.hasInvalidatedResolutions || returnFalse;\n    return hasInvalidatedResolutions(fileName);\n}\nconsole.log(String(invalidated({}, 'a.ts')));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			var refused *Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want it stopped as a method read as a value", err)
			}
			if !strings.Contains(err.Error(), "a method read as a value") {
				t.Errorf("got %v, want it stopped as a method read as a value", err)
			}
		})
	}
}
