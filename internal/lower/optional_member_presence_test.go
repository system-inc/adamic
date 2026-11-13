package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// optionalPresenceForms are the two ways a program declares the optional member it tests for: a method
// signature and a field holding a function, which optional_member_presence.go reads the same way.
var optionalPresenceForms = []struct{ name, member string }{
	{"method", "m?(): number;"},
	{"field", "m?: () => number;"},
}

// optionalPresenceDeclared is interface H, a required n beside member, the form's m.
func optionalPresenceDeclared(member string) string {
	return "interface H {\n\treadonly n: number;\n\t" + member + "\n}\n"
}

// optionalPresenceProbe tests h.m by ?:, in a function the program calls.
const optionalPresenceProbe = "function probe(h: H): string {\n\treturn h.m ? 'has' : 'none';\n}\n"

// optionalPresenceCalls hands probe an arrow field and a literal without m.
const optionalPresenceCalls = "console.log(probe({ n: 1, m: (): number => 1 }));\nconsole.log(probe({ n: 2 }));\n"

// optionalPresenceSlotReads counts a program's reads of name off its slot, an ir.Property with Absent
// of a Closure, which a presence test keeps and runs.
func optionalPresenceSlotReads(program *ir.Program, name string) int {
	count := 0
	visit := func(node any) bool {
		if read, isProperty := node.(ir.Property); isProperty && read.Absent && read.Of == ir.Closure && read.Name == name {
			count++
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return count
}

// A presence test of an optional member lowers at each consumer that only tests it, in both forms, as a
// read of its slot that runs: typeof, !, a condition, ?: and the left of &&, on an arrow field, a literal's
// method shorthand (an own slot, Compiler's ruling 1 on #p2n17c4 shape 1) and a literal without m, through
// ?., and with the call the test guards. A slot written after the literal is built is read as written;
// the literal gives m, since a write to a field some literal omits is absentOptionalWrite's not yet, and
// the form is the field's, since a write to an optional method stays refused (below). Its m's type holds
// undefined, as tsc's own host types write it, since under exactOptionalPropertyTypes no other optional
// field takes undefined, in a literal or written in. The last two rows
// were refused on main as a method read as a value: method_presence_test.go's typeof row and
// optional_method_call_test.go's presence test, whose places now hold an alias, which stays refused. The
// oracle holds what these run to (internal/oracle/optional_member_presence_test.go). Mutants are in
// optional_member_presence.mutants.md.
func TestOptionalMemberPresenceLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		// only is the one form a row runs in, "" for both; an alone row's source declares its own
		// interface.
		name, only, source, read string
		alone                    bool
	}{
		{"typeof", "", "function probe(h: H): string {\n\treturn typeof h.m;\n}\n" + optionalPresenceCalls, "m", false},
		{"!", "", "function probe(h: H): string {\n\treturn String(!h.m);\n}\n" + optionalPresenceCalls, "m", false},
		{"a condition", "", "function probe(h: H): string {\n\tif (h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';\n}\n" + optionalPresenceCalls, "m", false},
		{"the test of ?:", "", optionalPresenceProbe + optionalPresenceCalls, "m", false},
		{"the left of &&", "", "function probe(h: H): string {\n\tconst said = h.m && 'has';\n\treturn said ?? 'none';\n}\n" + optionalPresenceCalls, "m", false},
		{"a literal's method shorthand", "", optionalPresenceProbe + "console.log(probe({ n: 1, m() { return 1; } }));\nconsole.log(probe({ n: 2 }));\n", "m", false},
		{"the call the test guards keeps its receiver", "", "function probe(h: H): string {\n\tif (h.m) {\n\t\treturn String(h.m());\n\t}\n\treturn 'none';\n}\nconsole.log(probe({ n: 3, m() { return this.n; } }));\nconsole.log(probe({ n: 4 }));\n", "m", false},
		{"through ?.", "", "function probe(h: H | undefined): string {\n\treturn h?.m ? 'has' : 'none';\n}\nconsole.log(probe(undefined));\n" + optionalPresenceCalls, "m", false},
		{"a slot written after the literal is built", "field", optionalPresenceDeclared("m?: (() => number) | undefined;") + optionalPresenceProbe + "const h: H = { n: 1, m: undefined };\nconsole.log(probe(h));\nh.m = (): number => 2;\nconsole.log(probe(h));\nh.m = undefined;\nconsole.log(probe(h));\n", "m", true},
		{"typeof an optional method of a const literal", "method", "interface Maybe {\n\tm?(): number;\n}\nconst maybe: Maybe = { m() { return 1; } };\nconsole.log(typeof maybe.m);\n", "m", true},
		{"a presence test of an optional method on {}", "method", "interface Sized {\n\tsize?(): number;\n}\nfunction size(sized: Sized): string {\n\tif (sized.size) {\n\t\treturn 'sized';\n\t}\n\treturn 'unsized';\n}\nconsole.log(size({}));\n", "size", true},
	} {
		for _, form := range optionalPresenceForms {
			if probe.only != "" && probe.only != form.name {
				continue
			}
			source := optionalPresenceDeclared(form.member) + probe.source
			if probe.alone {
				source = probe.source
			}
			for _, extension := range []string{".a", ".ts"} {
				t.Run(probe.name+" "+form.name+extension, func(t *testing.T) {
					t.Parallel()
					program, err := lowerSourceAs(t, source, extension)
					if err != nil {
						t.Fatalf("want it to lower, got %v", err)
					}
					if optionalPresenceSlotReads(program, probe.read) == 0 {
						t.Errorf("no read of %s off its slot, want the test read as it runs, never folded", probe.read)
					}
				})
			}
		}
	}
}

// Where the receiver may hold anything but a plain object literal, the two backends can answer
// differently (native reads own slots, Node walks the prototype and runs getters), so the presence test
// stops, in its own words: a getter on a class or a literal (the method form at eepPropertyPresence's
// stop, the field form because finishAccessors made the read a dispatcher's call), a class's method on
// its prototype, a spread's copy, a closure's parameter, an exported function's and a Map's value. Its
// detaches stay refused: an alias, and a write to an optional method, which is why the written slot above
// is the field form's alone. A non-presence read of a field holding a literal method keeps
// erasedMethodField's stop, so the scan's bypass is the presence test's alone. A literal can't set its
// prototype (Lower's must): a __proto__ key, Object.setPrototypeOf and Object.create are each refused by
// another rule, held here. The __proto__ literal gives its own m, since one without it is
// adamic/no-optional-widening's refusal first, which says nothing of the key. No record row: a record literal reaching an H receiver needs a cast, so
// checkOptionalMemberPresence's record stop is pinned by no program yet (mutants M8).
func TestOptionalMemberPresenceStops(t *testing.T) {
	t.Parallel()
	const notLeft = "a presence test of optional member m that lowering didn't leave as one field read"
	const unknown = "a presence test of optional member m whose receiver may hold a value the program didn't build as a plain object literal"
	const getter = "a method sharing its name with a getter"
	const detached = "a method read as a value (m would lose its object, and this with it)"
	for _, probe := range []struct {
		name, extension, source string
		// method and field are what each form stops saying, "" where the row doesn't run in that form.
		method, field string
		refused       bool
		// readonly declares the field form's m readonly: a class's method or getter, or a held literal's
		// m of () => number, seen as a writable field that takes undefined is adamic/invariant-mutable's
		// refusal, which runs before lowering and would stop the row before the check it pins could.
		readonly bool
	}{
		{"a class's getter", ".a", "class G {\n\treadonly n = 1;\n\tget m(): () => number {\n\t\treturn (): number => 1;\n\t}\n}\n" + optionalPresenceProbe + "console.log(probe(new G()));\nconsole.log(probe({ n: 2 }));\n", getter, "", false, true},
		{"a literal's getter", ".a", optionalPresenceProbe + "console.log(probe({ n: 1, get m(): () => number { return (): number => 1; } }));\nconsole.log(probe({ n: 2 }));\n", getter, "", false, false},
		{"a class instance", ".a", "class C {\n\treadonly n = 1;\n\tm(): number {\n\t\treturn this.n;\n\t}\n}\n" + optionalPresenceProbe + "console.log(probe(new C()));\nconsole.log(probe({ n: 2 }));\n", "a presence test of optional member m whose receiver may be a class instance, whose method lives on its prototype, not an own slot", "a presence test of optional member m whose receiver may be a class instance, whose method lives on its prototype, not an own slot", false, true},
		{"a spread", ".a", optionalPresenceProbe + "const base: H = { n: 1, m: (): number => 1 };\nconsole.log(probe({ ...base }));\n", "a presence test of optional member m whose receiver may be an object a spread made", "a presence test of optional member m whose receiver may be an object a spread made", false, false},
		{"an arrow's parameter", ".a", "const probe = (h: H): string => (h.m ? 'has' : 'none');\n" + optionalPresenceCalls, unknown, unknown, false, false},
		{"an exported entry function's parameter", ".ts", "export " + optionalPresenceProbe + optionalPresenceCalls, unknown, unknown, false, false},
		{"a Map's value", ".ts", optionalPresenceProbe + "const held = new Map<string, H>();\nheld.set('a', { n: 1, m: (): number => 1 });\nconsole.log(probe(held.get('a')!));\n", unknown, unknown, false, false},
		{"an alias", ".a", "const h: H = { n: 1, m() { return 1; } };\nconst held = h.m;\nconsole.log(typeof held);\n", detached, "", true, false},
		{"a write to an optional method", ".a", optionalPresenceProbe + "const h: H = { n: 1, m: (): number => 1 };\nh.m = (): number => 2;\nconsole.log(probe(h));\n", detached, "", true, false},
		{"a field's literal method read as a value", ".a", "const h: H = { n: 1, m() { return 1; } };\nconst held = h.m;\nconsole.log(String(held ? held() : 0));\n", "", "a literal method through a view that erases its receiver", false, false},
		{"a __proto__ key", ".a", optionalPresenceProbe + "const proto: H = { n: 0, m: (): number => 1 };\nconst made = { n: 1, m: (): number => 2, __proto__: proto };\nconsole.log(probe(made));\n", "__proto__ in an object literal", "__proto__ in an object literal", true, true},
		{"Object.setPrototypeOf", ".a", optionalPresenceProbe + "const proto: H = { n: 0, m: (): number => 1 };\nconst made: H = { n: 1 };\nObject.setPrototypeOf(made, proto);\nconsole.log(probe(made));\n", "Object.setPrototypeOf", "Object.setPrototypeOf", true, false},
		{"Object.create", ".a", optionalPresenceProbe + "const proto: H = { n: 0, m: (): number => 1 };\nconst made: H = Object.create(proto);\nconsole.log(probe(made));\n", "Object.create", "Object.create", true, false},
	} {
		for _, form := range optionalPresenceForms {
			want := probe.method
			if form.name == "field" {
				want = probe.field
			}
			if want == "" {
				continue
			}
			member := form.member
			if probe.readonly && form.name == "field" {
				member = "readonly " + member
			}
			source := optionalPresenceDeclared(member) + probe.source
			t.Run(probe.name+" "+form.name+probe.extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, source, probe.extension)
				var notYet *NotYet
				var refused *Refused
				if probe.refused && !errors.As(err, &refused) || !probe.refused && !errors.As(err, &notYet) || !strings.Contains(err.Error(), want) {
					kind := "a not yet"
					if probe.refused {
						kind = "a refusal"
					}
					t.Errorf("got %v, want %s saying %q", err, kind, want)
				}
			})
		}
	}
}

// Lower's must on shape 1: Object.prototype holds members a plain literal inherits, so an optional member
// of one of their names diverges on a literal without its own (Node finds the prototype's, native reads
// no slot). toString is one the library's Object interface declares, which the check asks by symbol;
// __lookupGetter__ is one Node has that the library doesn't, which the check takes by its two leading
// underscores. H holds a required n beside the name, as a host's interface does: an H of optional members
// alone, one of them toString, takes a string too, so representation holds it as a union, and its read is
// dynamicProperty's not yet for a prototype name before this check is reached. Mutants M9 and M10.
func TestOptionalMemberPresenceOfAnObjectPrototypeNameStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"toString", "__lookupGetter__"} {
		for _, member := range []string{name + "?(): string;", name + "?: () => string;"} {
			source := optionalPresenceDeclared(member) + "function probe(h: H): string {\n\treturn h." + name + " ? 'has' : 'none';\n}\nconsole.log(probe({ n: 1 }));\n"
			want := "a presence test of optional member " + name + ", a name Object.prototype holds too"
			t.Run(member, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSource(t, source)
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), want) {
					t.Errorf("got %v, want a not yet saying %q", err, want)
				}
			})
		}
	}
}

// A .ts cast's checked view keeps its own stop for an optional callable read (unemittableCheckedViewRead's
// or the view's contract's, both before this check), never this check's words: the view's real object
// may hold anything in m, which no reach of literals speaks to.
func TestOptionalMemberPresenceThroughACastKeepsItsOwnStop(t *testing.T) {
	t.Parallel()
	for _, form := range optionalPresenceForms {
		source := "interface H {\n\treadonly kind: 'h';\n\t" + form.member + "\n}\nfunction probe(raw: { readonly kind: string }): string {\n\tconst viewed = raw as H;\n\treturn viewed.m ? 'has' : 'none';\n}\nconsole.log(probe({ kind: ['', 'h'].join('') }));\n"
		t.Run(form.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, ".ts")
			var notYet *NotYet
			var refused *Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want the checked view's own stop", err)
			}
			if strings.Contains(err.Error(), "a presence test of optional member") {
				t.Errorf("got %v, want the checked view's own stop, not the presence check's", err)
			}
			t.Logf("stopped: %v", err)
		})
	}
}

// Every road a presence test can be lowered by meets the check, since the refusal pass records each read
// it lets through (methodValueStop), not only those object.go's property reaches (Compiler's must and
// Lower's should on shape 1). Over a class instance, whose method lives on its prototype, Node finds m
// and native's own-slot read wouldn't, so each road must stop, whatever stops it first: an optional chain
// under typeof and under a condition, a parenthesized receiver, a !, and a member whose type is a union of
// two function types. Each reaches object.go's property today and stops at the check's class arm, which
// is what the rows pin: M14 (the record moved back into erasedLiteralMethod alone) leaves every row green
// on Home (J-6e61jk0), so the record at the exemption is a defense for a road that may come, as
// #4gpbn7w's fix was, and a row that starts stopping elsewhere says so here.
func TestOptionalMemberPresenceRoadsStopOverAClassInstance(t *testing.T) {
	t.Parallel()
	const class = "class C implements H {\n\treadonly n = 1;\n\tm(): number {\n\t\treturn this.n;\n\t}\n}\n"
	for _, probe := range []struct{ name, source string }{
		{"typeof through ?.", optionalPresenceDeclared("m?(): number;") + class + "function probe(h: H | undefined): string {\n\treturn typeof h?.m;\n}\nconsole.log(probe(new C()));\n"},
		{"a condition through ?.", optionalPresenceDeclared("m?(): number;") + class + "function probe(h: H | undefined): string {\n\tif (h?.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';\n}\nconsole.log(probe(new C()));\n"},
		{"a parenthesized receiver", optionalPresenceDeclared("m?(): number;") + class + "function probe(h: H): string {\n\treturn (h).m ? 'has' : 'none';\n}\nconsole.log(probe(new C()));\n"},
		{"a !", optionalPresenceDeclared("m?(): number;") + class + "function probe(h: H): string {\n\treturn !h.m ? 'none' : 'has';\n}\nconsole.log(probe(new C()));\n"},
		{"a union of two function types", optionalPresenceDeclared("readonly m?: (() => number) | ((x: number) => number);") + "class D {\n\treadonly n = 1;\n\tm(): number {\n\t\treturn this.n;\n\t}\n}\nfunction probe(h: H): string {\n\treturn h.m ? 'has' : 'none';\n}\nconsole.log(probe(new D()));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				var refused *Refused
				const classArm = "a presence test of optional member m whose receiver may be a class instance, whose method lives on its prototype, not an own slot"
				if !errors.As(err, &notYet) && !errors.As(err, &refused) || !strings.Contains(err.Error(), classArm) {
					t.Fatalf("got %v, want it stopped saying %q: Node finds a class's prototype method where native reads own slots", err, classArm)
				}
			})
		}
	}
}

// A field-form presence test finishAccessors made a dispatcher's call (#4gpbn7w's road) is judged by what
// the dispatcher does with each object that may reach it, by the reach and owners it was dispatched on: a
// class instance whose getter of the name the dispatcher calls reads the same on both backends, as a plain
// literal does through the fallback, so each of these lowers, while a spread's copy, an arrow's parameter
// and a class whose method of the name lives on its prototype still stop (Compiler's and Lower's ruling on
// shape 1's red against internal/oracle's accessor_dispatch_absent). The method form never dispatches: a
// getter of its name stops it first (eepPropertyPresence). Mutant M15 drops the note finishAccessors
// takes, and every lowering row stops as a read lowering didn't leave as one field read.
func TestOptionalMemberPresenceThroughAGetterDispatcher(t *testing.T) {
	t.Parallel()
	const getterClass = "class G {\n\treadonly n = 1;\n\tget m(): () => number {\n\t\treturn (): number => 7;\n\t}\n}\n"
	const methodClass = "class C {\n\treadonly n = 2;\n\tm(): number {\n\t\treturn this.n;\n\t}\n}\n"
	declared := optionalPresenceDeclared("readonly m?: () => number;")
	for _, probe := range []struct{ name, source, want string }{
		{"a class's getter and a literal", declared + getterClass + optionalPresenceProbe + "console.log(probe(new G()));\nconsole.log(probe({ n: 2 }));\nconsole.log(probe({ n: 3, m: (): number => 1 }));\n", ""},
		{"a literal's getter and a literal", declared + optionalPresenceProbe + "console.log(probe({ n: 1, get m(): () => number { return (): number => 1; } }));\nconsole.log(probe({ n: 2 }));\n", ""},
		{"a class's getter and a spread", declared + getterClass + optionalPresenceProbe + "const base: H = { n: 1 };\nconsole.log(probe(new G()));\nconsole.log(probe({ ...base }));\n", "a presence test of optional member m whose receiver may be an object a spread made"},
		{"a class's getter through an arrow's parameter", declared + getterClass + "const probe = (h: H): string => (h.m ? 'has' : 'none');\nconsole.log(probe(new G()));\nconsole.log(probe({ n: 2 }));\n", "a presence test of optional member m whose receiver may hold a value the program didn't build as a plain object literal"},
		{"a class's getter and a class's method", declared + getterClass + methodClass + optionalPresenceProbe + "console.log(probe(new G()));\nconsole.log(probe(new C()));\n", "a presence test of optional member m whose receiver may be a class instance, whose method lives on its prototype, not an own slot"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				if probe.want == "" {
					if err != nil {
						t.Fatalf("want it to lower through the dispatcher, got %v", err)
					}
					return
				}
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want) {
					t.Fatalf("got %v, want a not yet saying %q", err, probe.want)
				}
			})
		}
	}
}
