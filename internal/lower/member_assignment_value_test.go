package lower

import (
	"strings"
	"testing"
)

// An assignment to a member used as a value is the value stored, evaluated once, after the receiver and
// before the store (#4gxxhvf): revcomp's two shapes, a chain, a short circuit that skips the store, and a
// value with a side effect, a this receiver, and a setter whose getter gives back twice what was stored, so
// the result is the value and never the slot read back, each agreeing with Node on the JavaScript backend
// and natively under the sanitizers. Mutant: the result evaluates the stored value again in place of reading it back, and the
// side-effect row counts two calls.
func TestMemberAssignmentValueAgrees(t *testing.T) {
	t.Parallel()
	const linked = "class LinkedArray {\n  next: LinkedArray | undefined = undefined;\n  pos = 0;\n  constructor(readonly prev: LinkedArray | undefined) {}\n}\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"revcomp next", linked + "function grow(n: number): number {\n  let la = new LinkedArray(undefined);\n  const first = la;\n  for (let i = 0; i < n; i++) {\n    la = la.next || (la.next = new LinkedArray(la));\n  }\n  let count = 0;\n  for (let at: LinkedArray | undefined = first; at !== undefined; at = at.next) {\n    count++;\n  }\n  return count;\n}\nconsole.log(String(grow(3)));\nconsole.log(String(grow(0)));\n", "4\n1\n"},
		{"revcomp pos", linked + "const la = new LinkedArray(undefined);\nlet lnIdx = 7;\nlnIdx = la.pos = 1;\nconsole.log(String(lnIdx) + ' ' + String(la.pos));\n", "1 1\n"},
		{"chain", "class Box { f = 0; g = 0; }\nconst b = new Box();\nconst a = b.f = b.g = 5;\nconsole.log(String(a) + ' ' + String(b.f) + ' ' + String(b.g));\n", "5 5 5\n"},
		{"short circuit", "class Box { f = 1; }\nconst o = new Box();\nconst skip = Date.now() > 0;\nconst r = skip || (o.f = 9) > 0;\nconsole.log(String(r) + ' ' + String(o.f));\n", "true 1\n"},
		{"side effect once", "class Box { f = 0; }\nconst o = new Box();\nlet calls = 0;\nfunction bump(): number { calls++; return calls * 10; }\nconst r = (o.f = bump());\nconsole.log(String(r) + ' ' + String(o.f) + ' ' + String(calls));\n", "10 10 1\n"},
		{"this", "class Box {\n  f = 0;\n  set(v: number): number { return (this.f = v) + 1; }\n}\nconst o = new Box();\nconsole.log(String(o.set(4)) + ' ' + String(o.f));\n", "5 4\n"},
		{"setter", "class Box {\n  private held = 0;\n  get f(): number { return this.held * 2; }\n  set f(v: number) { this.held = v; }\n}\nconst o = new Box();\nconst r = (o.f = 3);\nconsole.log(String(r) + ' ' + String(o.f));\n", "3 6\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				program := lowersAndAgreesWithNodeAs(t, probe.source, extension)
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
				}
			})
		}
	}
}

// Node resolves the receiver before it evaluates the value, so a value that rebinds the receiver's name, by
// itself or through a closure, still stores into the object named first. Mutant: the store reads its
// receiver after the value in place of the receiver's local, and both rows store into b.
func TestMemberAssignmentValueResolvesTheReceiverFirst(t *testing.T) {
	t.Parallel()
	const boxes = "class Box { f = 0; }\nconst a = new Box();\nconst b = new Box();\nlet o = a;\n"
	for _, probe := range []struct{ name, source string }{
		{"direct", boxes + "const r = (o.f = (o = b, 2));\nconsole.log(String(r) + ' ' + String(a.f) + ' ' + String(b.f));\n"},
		{"closure", boxes + "const swap = (): number => { o = b; return 2; };\nconst r = (o.f = swap());\nconsole.log(String(r) + ' ' + String(a.f) + ' ' + String(b.f));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				program := lowersAndAgreesWithNodeAs(t, probe.source, extension)
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2 2 0\n" {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "2 2 0\n")
				}
			})
		}
	}
}

// A write the statement form stops stays stopped as a value, in the same words: a write to an optional field
// a literal left out has no slot to write (absentOptionalWrite, judged once the program is lowered).
func TestMemberAssignmentValueKeepsTheStatementFormsStop(t *testing.T) {
	t.Parallel()
	const prelude = "interface P { f?: number }\nconst p: P = {};\n"
	statement, valued := prelude+"p.f = 3;\nconsole.log(String(p.f));\n", prelude+"const r = (p.f = 3);\nconsole.log(String(r));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, statementErr := lowerSourceAs(t, statement, extension)
			_, valuedErr := lowerSourceAs(t, valued, extension)
			if statementErr == nil || valuedErr == nil {
				t.Fatalf("want both stopped, got statement %v and value %v", statementErr, valuedErr)
			}
			words := func(err error) string { return err.Error()[strings.Index(err.Error(), " ")+1:] }
			if !strings.Contains(statementErr.Error(), "optional") || words(statementErr) != words(valuedErr) {
				t.Fatalf("want the value stopped as the statement is, got statement %v and value %v", statementErr, valuedErr)
			}
		})
	}
}

// A plain Record<K, V> write is an ordinary field store, so as a value it lowers and agrees with Node,
// natively too. Only an entries record's write defines its key, and that one stays NotYet.
func TestMemberAssignmentValueRecordAgrees(t *testing.T) {
	t.Parallel()
	const source = "const rec: Record<string, number> = {};\nconst r = (rec.k = 4);\nconsole.log(String(r) + ' ' + String(rec.k));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, source, extension)
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "4 4\n" {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "4 4\n")
			}
		})
	}
}
