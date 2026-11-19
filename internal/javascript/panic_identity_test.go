package javascript

import (
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

var temporaryName = regexp.MustCompile(`temporary_[0-9]+`)

// panicIdentityShapes is a function holding the three kinds of try: a catch alone, a finally alone, and
// both, each returning its own number from each part.
func panicIdentityShapes() *ir.Program {
	constant := func(value float64) []ir.Statement {
		return []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: value}}}
	}
	return &ir.Program{Source: "shapes.a", Functions: []ir.Function{{Name: "shapes", Returns: ir.Number, Body: []ir.Statement{
		ir.Try{Body: constant(1), Catch: constant(2), CatchLocal: -1, HasCatch: true},
		ir.Try{Body: constant(3), Finally: constant(4), CatchLocal: -1, HasFinally: true},
		ir.Try{Body: constant(5), Catch: constant(6), Finally: constant(7), CatchLocal: -1, HasCatch: true, HasFinally: true},
	}}}}
}

// panicIdentityClose is a function holding an iterable view's dispatch (ir.Block's Closes): its first
// statement declares the local the block closes, and the rest returns.
func panicIdentityClose() *ir.Program {
	return &ir.Program{Source: "close.a", Locals: []ir.Local{{Name: "held", Type: ir.Object, Function: 0}}, Functions: []ir.Function{{Name: "close", Returns: ir.Number, Body: []ir.Statement{
		ir.Block{Body: []ir.Statement{ir.Declare{Local: 0, Value: ir.NumberConstant{Value: 8}}, ir.Return{Value: ir.NumberConstant{Value: 9}}}, Closes: 1},
	}}}}
}

// Every try asks the runtime's isPanic of what it caught, so a Worker's thrown AdamicPanic is never
// the program's (docs/rulings.md): a catch throws it on, and a finally body is skipped while one
// unwinds, by a flag of that try's own. The import has no fallback, so a runtime without isPanic
// fails to load. This pins the emitted shape for every program, Node's and a Worker's alike.
func TestEveryTryTestsPanicIdentity(t *testing.T) {
	t.Parallel()
	code := JavaScript(panicIdentityShapes())
	if !strings.Contains(code, "import { isPanic as adamicIsPanic } from 'adamic';\n") || strings.Contains(code, "const adamicIsPanic") {
		t.Fatalf("want the isPanic import from 'adamic', with no fallback:\n%s", code)
	}
	worker := JavaScriptWith(panicIdentityShapes(), Options{RuntimeImport: "./adamic.mjs"})
	if !strings.Contains(worker, "import { isPanic as adamicIsPanic } from \"./adamic.mjs\";\n") {
		t.Fatalf("want the isPanic import from the runtime import:\n%s", worker)
	}
	const header = "\nfunction function_0_shapes() {\n"
	start := strings.Index(code, header)
	if start < 0 {
		t.Fatalf("no function_0_shapes:\n%s", code)
	}
	end := strings.Index(code[start+len(header):], "\n}\n")
	if end < 0 {
		t.Fatalf("function_0_shapes never closes:\n%s", code)
	}
	body := code[start+len(header) : start+len(header)+end+1]
	stack := "\t\tif (temporary_N instanceof RangeError && temporary_N.message === " + quote("Maximum call stack size exceeded") + ") panic(" + quote(stackOverflowMessage) + ");\n"
	want := "\ttry {\n" +
		"\t\treturn (1);\n" +
		"\t} catch (temporary_N) {\n" +
		"\t\tif (adamicProcessExiting()) throw temporary_N;\n" +
		"\t\tif (adamicIsPanic(temporary_N)) throw temporary_N;\n" +
		stack +
		"\t\treturn (2);\n" +
		"\t}\n" +
		"\tlet temporary_N = false;\n" +
		"\ttry {\n" +
		"\t\treturn (3);\n" +
		"\t} catch (temporary_N) {\n" +
		"\t\tif (adamicIsPanic(temporary_N)) temporary_N = true;\n" +
		"\t\tthrow temporary_N;\n" +
		"\t} finally {\n" +
		"\t\tif (!adamicProcessExiting() && !temporary_N) {\n" +
		"\t\t\treturn (4);\n" +
		"\t\t}\n" +
		"\t}\n" +
		"\tlet temporary_N = false;\n" +
		"\ttry {\n" +
		"\t\ttry {\n" +
		"\t\t\treturn (5);\n" +
		"\t\t} catch (temporary_N) {\n" +
		"\t\t\tif (adamicProcessExiting()) throw temporary_N;\n" +
		"\t\t\tif (adamicIsPanic(temporary_N)) throw temporary_N;\n" +
		"\t" + stack +
		"\t\t\treturn (6);\n" +
		"\t\t}\n" +
		"\t} catch (temporary_N) {\n" +
		"\t\tif (adamicIsPanic(temporary_N)) temporary_N = true;\n" +
		"\t\tthrow temporary_N;\n" +
		"\t} finally {\n" +
		"\t\tif (!adamicProcessExiting() && !temporary_N) {\n" +
		"\t\t\treturn (7);\n" +
		"\t\t}\n" +
		"\t}\n"
	if got := temporaryName.ReplaceAllString(body, "temporary_N"); got != want {
		t.Fatalf("emitted try shapes, temporaries as temporary_N:\ngot:\n%s\nwant:\n%s", got, want)
	}
	// Each finally's guard reads the flag its own try declared, and its catch around the try sets it.
	declared := regexp.MustCompile(`let (temporary_[0-9]+) = false;`).FindAllStringSubmatch(body, -1)
	set := regexp.MustCompile(`\) (temporary_[0-9]+) = true;`).FindAllStringSubmatch(body, -1)
	guarded := regexp.MustCompile(`&& !(temporary_[0-9]+)\) \{`).FindAllStringSubmatch(body, -1)
	if len(declared) != 2 || len(set) != 2 || len(guarded) != 2 {
		t.Fatalf("want two flags, each declared, set and guarded once:\n%s", body)
	}
	for index := range declared {
		if declared[index][1] != set[index][1] || declared[index][1] != guarded[index][1] {
			t.Fatalf("finally %d reads a flag its try did not declare or set:\n%s", index, body)
		}
	}
	// An iterable view's close calls a generator's return(), which runs its finally, so it is skipped
	// while a panic unwinds, by a flag of the block's own that a catch around the rest sets.
	closing := JavaScript(panicIdentityClose())
	if !strings.Contains(closing, "import { isPanic as adamicIsPanic } from 'adamic';\n") {
		t.Fatalf("want an iterable view's close to import isPanic:\n%s", closing)
	}
	const closeHeader = "\nfunction function_0_close() {\n"
	closeStart := strings.Index(closing, closeHeader)
	if closeStart < 0 {
		t.Fatalf("no function_0_close:\n%s", closing)
	}
	closeEnd := strings.Index(closing[closeStart+len(closeHeader):], "\n}\n")
	if closeEnd < 0 {
		t.Fatalf("function_0_close never closes:\n%s", closing)
	}
	closeBody := closing[closeStart+len(closeHeader) : closeStart+len(closeHeader)+closeEnd+1]
	wantClose := "\t{\n" +
		"\t\tlet local_0_held = (8);\n" +
		"\t\tlet temporary_N = false;\n" +
		"\t\ttry {\n" +
		"\t\t\treturn (9);\n" +
		"\t\t} catch (temporary_N) {\n" +
		"\t\t\tif (adamicIsPanic(temporary_N)) temporary_N = true;\n" +
		"\t\t\tthrow temporary_N;\n" +
		"\t\t} finally {\n" +
		"\t\t\tif (!temporary_N) adamicIteratorClose(local_0_held);\n" +
		"\t\t}\n" +
		"\t}\n"
	if got := temporaryName.ReplaceAllString(closeBody, "temporary_N"); got != wantClose {
		t.Fatalf("emitted iterable view close, temporaries as temporary_N:\ngot:\n%s\nwant:\n%s", got, wantClose)
	}
	closeFlag := regexp.MustCompile(`let (temporary_[0-9]+) = false;[\s\S]*\) (temporary_[0-9]+) = true;[\s\S]*if \(!(temporary_[0-9]+)\) adamicIteratorClose`).FindStringSubmatch(closeBody)
	if closeFlag == nil || closeFlag[1] != closeFlag[2] || closeFlag[1] != closeFlag[3] {
		t.Fatalf("the close reads a flag its block did not declare or set:\n%s", closeBody)
	}
	// A program with no try asks nothing of the runtime's isPanic.
	plain := &ir.Program{Source: "plain.a", Functions: []ir.Function{{Name: "plain", Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 1}}}}}}
	if strings.Contains(JavaScript(plain), "adamicIsPanic") {
		t.Fatal("a program with no try imports isPanic")
	}
}
