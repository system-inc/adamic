package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// definedHidden is every key the program's record defines hide, from the defineHidden calls in its IR.
func definedHidden(program *ir.Program) []string {
	var keys []string
	visit := func(node any) bool {
		if call, ok := node.(ir.RecordCall); ok && call.Method == "defineHidden" {
			if key, ok := call.Arguments[1].(ir.StringConstant); ok {
				keys = append(keys, program.Strings[key.Index])
			}
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return keys
}

// The admitted define, as tsc's commandLineParser writes it and in any order of its three properties, lowers
// to a defineHidden whose false throws, and turns on the program's write checks.
func TestRecordDefineHiddenLowers(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const r:Record<string,number>={a:1}; Object.defineProperty(r, "h", {enumerable:false, writable:false, value:5}); console.log(String(r['h']));`,
		`const r:Record<string,number>={}; Object.defineProperty(r, "h", {value:5, "writable":false, enumerable:(false)});`,
		`interface Source {fileName:string; lines:number[]} interface Options {[option:string]: number|string|Source|undefined; configFile?: Source} const options:Options={target:1}; const file:Source={fileName:'tsconfig.json', lines:[]}; Object.defineProperty(options, "configFile", {enumerable:false, writable:false, value:file});`,
		`function hide(r:Record<string,number[]>, list:number[]):void { Object.defineProperty(r, "main", {enumerable:false, writable:false, value:list}); } hide({}, [1]);`,
	} {
		test.Run(source, func(test *testing.T) {
			program, err := lowerSource(test, source)
			if err != nil {
				test.Fatal(err)
			}
			keys := definedHidden(program)
			if len(keys) != 1 {
				test.Fatalf("want one defineHidden, got %v", keys)
			}
			if !program.HiddenRecordKeys[keys[0]] {
				test.Fatalf("hidden keys %v, want %q among them", program.HiddenRecordKeys, keys[0])
			}
			thrown := false
			walk(program.Main, func(node any) bool {
				if check, ok := node.(ir.If); ok && len(check.Then) == 1 {
					if throw, ok := check.Then[0].(ir.Throw); ok {
						if made, ok := throw.Value.(ir.MakeError); ok {
							message, named := made.Message.(ir.StringConstant)
							name, typed := made.Name.(ir.StringConstant)
							thrown = thrown || named &&
								typed &&
								strings.HasPrefix(program.Strings[message.Index], "Cannot redefine property: ") &&
								program.Strings[name.Index] == "TypeError"
						}
					}
				}
				return true
			})
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					if throw, ok := node.(ir.Throw); ok {
						if made, ok := throw.Value.(ir.MakeError); ok {
							if name, ok := made.Name.(ir.StringConstant); ok &&
								program.Strings[name.Index] == "TypeError" {
								thrown = true
							}
						}
					}
					return true
				})
			}
			if !thrown {
				test.Fatal("a second define must throw TypeError: Cannot redefine property")
			}
		})
	}
}

// Only a write or delete that may meet a hidden key can throw: one whose key is known only as it runs,
// or is a literal some define hides. A literal key no define names, and a fixed object's field, never
// can, and without a define nothing can. Callers test for the throw after exactly those functions.
func TestRecordDefineHiddenScopesWrites(test *testing.T) {
	test.Parallel()
	const writers = `function put(r:Record<string,number>, k:string):void { r[k]=1; }
function drop(r:Record<string,number>, k:string):void { delete r[k]; }
function named(r:Record<string,number>):void { r['h']=1; }
function coalesce(r:Record<string,number>):void { r['h'] ??= 1; }
function other(r:Record<string,number>):void { r['x']=1; delete r.y; r['z'] ??= 2; }
interface Point {x:number}
function move(p:Point):void { p.x=2; }
const r:Record<string,number>={};
put(r,'x'); drop(r,'x'); named(r); coalesce(r); other(r); move({x:1});
`
	for _, probe := range []struct {
		source   string
		throwing map[string]bool
	}{
		{writers, map[string]bool{}},
		{
			writers + `Object.defineProperty(r, "h", {enumerable:false, writable:false, value:5});`,
			map[string]bool{"put": true, "drop": true, "named": true, "coalesce": true},
		},
	} {
		test.Run(probe.source, func(test *testing.T) {
			program, err := lowerSource(test, probe.source)
			if err != nil {
				test.Fatal(err)
			}
			if (len(program.HiddenRecordKeys) > 0) != (len(probe.throwing) > 0) {
				test.Fatalf("hidden keys %v", program.HiddenRecordKeys)
			}
			found := map[string]bool{}
			for _, function := range program.Functions {
				switch function.Name {
				case "put", "drop", "named", "coalesce", "other", "move":
					found[function.Name] = true
					if function.MayThrow != probe.throwing[function.Name] {
						test.Errorf(
							"%s: MayThrow %t, want %t",
							function.Name,
							function.MayThrow,
							probe.throwing[function.Name],
						)
					}
				}
			}
			if len(found) != 6 {
				test.Fatalf("found %v of the six writers", found)
			}
		})
	}
}

// Every other shape stays not yet, or refused as Object.defineProperty is on a fixed object, and nothing of
// it is lowered: other flags, a getter, an extra or missing property, a shorthand or spread descriptor, a
// key that isn't a string literal, a value the record's slot doesn't take, a result used as a value, and a
// fixed object seen as a dictionary. A getter, a spread and a fixed object's view stop however the refusal
// walk or lowering reaches them first, so those name no reason.
// Object.getOwnPropertyNames, which would list a hidden key, stays not yet.
func TestRecordDefineHiddenRefusals(test *testing.T) {
	test.Parallel()
	const record = `const r:Record<string,number>={a:1}; `
	for _, probe := range []struct {
		source, reason string
		notYet         bool
	}{
		{
			record + `Object.defineProperty(r, "h", {enumerable:true, writable:false, value:5});`,
			"property descriptors",
			false,
		},
		{
			record + `Object.defineProperty(r, "h", {enumerable:false, writable:true, value:5});`,
			"property descriptors",
			false,
		},
		{
			record + `const no = false; Object.defineProperty(r, "h", {enumerable:no, writable:false, value:5});`,
			"property descriptors",
			false,
		},
		{record + `Object.defineProperty(r, "h", {enumerable:false, writable:false, get() { return 5; }});`, "", false},
		{
			record + `Object.defineProperty(r, "h", {enumerable:false, writable:false, configurable:false, value:5});`,
			"property descriptors",
			false,
		},
		{record + `Object.defineProperty(r, "h", {enumerable:false, value:5});`, "property descriptors", false},
		{
			record + `const value = 5; Object.defineProperty(r, "h", {enumerable:false, writable:false, value});`,
			"property descriptors",
			false,
		},
		{
			record +
				`const flags = {enumerable:false, writable:false}; Object.defineProperty(r, "h", {...flags, value:5});`,
			"",
			false,
		},
		{
			record + `const k = "h"; Object.defineProperty(r, k, {enumerable:false, writable:false, value:5});`,
			"property descriptors",
			false,
		},
		{
			record + "Object.defineProperty(r, `h`, {enumerable:false, writable:false, value:5});",
			"property descriptors",
			false,
		},
		{
			record +
				`const same = Object.defineProperty(r, "h", {enumerable:false, writable:false, value:5}); console.log(Object.keys(same).join(','));`,
			"whose result is used",
			true,
		},
		{
			record + `Object.defineProperty(r, "h", {enumerable:false, writable:false, value:'text'});`,
			"index signature doesn't take",
			true,
		},
		{
			`interface Options {[option:string]: string|number|undefined; configFile?: string} const o:Options={}; Object.defineProperty(o, "configFile", {enumerable:false, writable:false, value:5});`,
			"named property configFile",
			true,
		},
		{
			record + `console.log(Object.getOwnPropertyNames(r).join(','));`,
			"Object.getOwnPropertyNames on a record",
			true,
		},
		// The rest of the descriptor and prototype family is refused on a record as on any object, by
		// library_object.go's own refusal; getOwnPropertyNames above is reflection not yet built.
		{
			record + `Object.defineProperties(r, {h: {enumerable:false, writable:false, value:5}});`,
			"property descriptors",
			false,
		},
		{record + `console.log(String(Object.getOwnPropertyDescriptor(r, "a")));`, "property descriptors", false},
		{record + `console.log(String(Object.getOwnPropertyDescriptors(r)));`, "property descriptors", false},
		{record + `console.log(String(Object.getPrototypeOf(r)));`, "prototypes", false},
		{record + `Object.setPrototypeOf(r, null);`, "prototypes", false},
		// Object.assign never reaches a hidden entry through a runtime copy loop: onto a record it's not yet
		// (records.go admits keys, values, entries and hasOwn only), and from one it's refused, since every
		// source must be a literal's proven shape (library_object.go). Language's question on #3xq2xzc.
		{
			record +
				`Object.defineProperty(r, "h", {enumerable:false, writable:false, value:5}); Object.assign(r, {h: 6});`,
			"Object.assign on a record",
			true,
		},
		{
			record +
				`Object.defineProperty(r, "h", {enumerable:false, writable:false, value:5}); const target = {h: 0}; Object.assign(target, r);`,
			"",
			false,
		},
		{
			`const fixed = {a: 1}; Object.defineProperty(fixed, "h", {enumerable:false, writable:false, value:5});`,
			"Object.defineProperty",
			false,
		},
		{
			`function hide(r:{[k:string]:unknown}):void { Object.defineProperty(r, "h", {enumerable:false, writable:false, value:5}); } const fixed = {a: 1}; hide(fixed);`,
			"",
			false,
		},
	} {
		test.Run(probe.source, func(test *testing.T) {
			program, err := lowerSource(test, probe.source)
			var notYet *NotYet
			var refused *Refused
			stopped := errors.As(err, &notYet) || errors.As(err, &refused)
			if probe.reason == "" {
				if !stopped {
					test.Fatalf("got %v, want a stop", err)
				}
			} else if err == nil ||
				!strings.Contains(err.Error(), probe.reason) ||
				(probe.notYet && !errors.As(err, &notYet)) ||
				(!probe.notYet && !errors.As(err, &refused)) {
				test.Fatalf("got %v, want %q (NotYet=%t)", err, probe.reason, probe.notYet)
			}
			if program != nil {
				test.Fatalf("a refused program came back lowered")
			}
		})
	}
}
