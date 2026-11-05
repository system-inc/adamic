package lower

import (
	"errors"
	"strings"
	"testing"
)

// A template or a concatenation outside a try can throw past V8's longest string, so an async function
// that makes one may reject, and the promise consumption proof keeps that edge (Adamic's ruling on
// #c3bk0fp). word below rejects only that way. Two of its promises pending, awaited in one template,
// stop as NotYet named for the string-length limit, at word's template: if the first await throws,
// the second's rejection goes unobserved. One pending promise awaited first in a template still lowers:
// nothing in the statement can run code or throw before that await observes it (leadingAwait). A template between a
// pending promise and its await stops the same way, at that template. A callee that throws for
// another reason, with nothing else between, keeps the general stop.
func TestAsyncStringLengthConsumption(t *testing.T) {
	t.Parallel()
	const word = "async function word(text: string): Promise<string> {\n\treturn `${text}!`;\n}\n"
	const fail = "async function fail(text: string): Promise<string> {\n\tif (text === 'b') {\n\t\tthrow new Error(text);\n\t}\n\treturn text;\n}\n"
	const general = "a promise not awaited in the same function before an escaping completion"
	for _, probe := range []struct {
		name, source, where, want string
	}{
		{"two pending awaits in one template", word + "const first = word('a');\nconst second = word('b');\nconsole.log(`${await first}|${await second}`);\n", "main.a:2:9", stringLengthConsumption},
		{"one pending await first in a template", word + "const first = word('a');\nconsole.log(`${await first}|`);\n", "", ""},
		{"a template between a pending promise and its await", fail + "const count = 3;\nconst first = fail('a');\nconst label = `n${count}`;\nconst value = await first;\nconsole.log(label + value);\n", "main.a:9:15", stringLengthConsumption},
		{"a callee that throws otherwise", fail + "const first = fail('a');\nconst second = fail('b');\nconsole.log(`${await first}|${await second}`);\n", "main.a:8:7", general},
		// An awaited function that rejects at the limit only through a helper it calls keeps the general
		// text: telling it apart needs a whole-program may-throw fixpoint (Adamic, 07:11:58Z), so a later
		// one changes this row on purpose.
		{"an indirect cause", "function shout(text: string): string {\n\treturn `${text}!`;\n}\nasync function loud(text: string): Promise<string> {\n\treturn shout(text);\n}\nconst first = loud('a');\nconst second = loud('b');\nconsole.log(`${await first}|${await second}`);\n", "main.a:8:7", general},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.want == "" {
				if err != nil {
					t.Fatalf("must lower: %v", err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.want || !strings.HasSuffix(notYet.Where, "/"+probe.where) {
				t.Fatalf("want NotYet %q at %s, got %v", probe.want, probe.where, err)
			}
		})
	}
}

// Language's and Adamic's probes (#c3bk0fp, 07:11:58Z and 07:12:12Z): two of word's promises pending
// in one statement, and in each something can run code or throw before the second await, or before
// the first: a concatenation after the first await, an argument or an operand between the two, a
// condition evaluated first, a callee, a property read, a computed key, a tagged template's tag, a
// destructuring default, a comma expression. Each stops: named for the string-length limit where the
// proof is what stops it, since an await consumes its own promise only, never the statement's later
// awaits; and a computed key, a tagged template and a destructuring default stop earlier, at
// lowering's own NotYet for each, and are pinned there, so lowering one of them shows up here.
func TestAsyncStringLengthProbes(t *testing.T) {
	t.Parallel()
	const word = "async function word(text: string): Promise<string> {\n\treturn `${text}!`;\n}\n"
	const lowers = "lowers"
	const general = "a promise not awaited in the same function before an escaping completion"
	const setup = word + "const big = 'x';\nconst more = 'y';\nconst p = word('a');\nconst q = word('b');\n"
	for _, probe := range []struct{ name, source, want string }{
		{"a: a concatenation between the awaits", setup + "console.log(`${await p}${big}${await q}`);\n", ""},
		{"b: an argument between the awaits", setup + "function f(a: string, b: string, c: string): string {\n\treturn a + b + c;\n}\nconsole.log(f(await p, big + more, await q));\n", ""},
		{"c: a binary chain", setup + "console.log((await p) + big + (await q));\n", ""},
		{"d: a condition first", setup + "const flag = big === 'x';\nconsole.log(flag ? await p : `${big}${await q}`);\n", ""},
		{"a1: a callee first", setup + "function show(text: string): string {\n\treturn text;\n}\nconsole.log(show(`${await p}|${await q}`));\n", ""},
		{"a2: a property read first", setup + "const box = { x: 'k' };\nconsole.log(`${box.x}${await p}|${await q}`);\n", ""},
		{"a3: a computed key first", setup + "const key = 'k';\nconst record = { [key]: `${await p}|${await q}` };\nconsole.log(record.k);\n", "a computed field name"},
		{"a4: a tagged template's tag first", setup + "function tag(parts: TemplateStringsArray, ...values: string[]): string {\n\treturn values.join('|');\n}\nconsole.log(tag`${await p}|${await q}`);\n", "a value of type TemplateStringsArray"},
		{"a5: a destructuring default", setup + "const holder: { v?: string } = {};\nconst { v = `${await p}|${await q}` } = holder;\nconsole.log(v);\n", "a destructured name that isn't plain"},
		{"a6: a comma expression", setup + "console.log((big.length, `${await p}|${await q}`));\n", ""},

		// One pending promise (Language's probes, 11:34Z, on Language's later(), which rejects only at
		// the limit, and in d1 to d3 otherwise): an await consumes its promise only when nothing in the
		// statement that can run code or throw is evaluated before it, a callee among them unless it's a
		// function declaration (Adamic, 11:36Z).
		{"c1: a declared function as callee", "async function later(text: string): Promise<string> { return `${text}!`; }\nfunction wrap(s: string): string { return s; }\nasync function run(): Promise<string> { const first = later(\"a\"); return wrap(`${await first}|`); }\nconsole.log(await run());\n", lowers},
		{"c2: a property read before the await", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst o = { prefix: \"p\" };\nasync function run(): Promise<string> { const first = later(\"a\"); return `${o.prefix}${await first}|`; }\nconsole.log(await run());\n", ""},
		{"c6: a comma before the await", "async function later(text: string): Promise<string> { return `${text}!`; }\nlet counter = 0;\nasync function run(): Promise<string> { const first = later(\"a\"); return (counter++, `${await first}|`); }\nconsole.log(await run());\n", ""},
		{"c7: the await, then a concatenation", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst big = \"x\";\nasync function run(): Promise<string> { const first = later(\"a\"); return `${await first}${big}`; }\nconsole.log(await run());\n", lowers},
		{"c8: a condition before the await", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst big = \"x\"; let flag = false;\nasync function run(): Promise<string> { const first = later(\"a\"); return flag ? await first : `${big}|`; }\nconsole.log(await run());\n", ""},
		{"c9: the await, then a binary concatenation", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst big = \"x\";\nasync function run(): Promise<string> { const first = later(\"a\"); return (await first) + big; }\nconsole.log(await run());\n", lowers},
		{"c10: a read concatenated before the await", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst big = \"x\";\nasync function run(): Promise<string> { const first = later(\"a\"); return big + (await first); }\nconsole.log(await run());\n", ""},
		// c11 stops first at flow's NotYet for an await among a structural method's operands
		// (async_normalize.go), before the consumption proof runs, so lowering one shows up here.
		{"c11: a method as callee", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst o = { wrap(s: string): string { return s; } };\nasync function run(): Promise<string> { const first = later(\"a\"); return o.wrap(`${await first}|`); }\nconsole.log(await run());\n", "await among structural method operands is not yet proven"},
		{"c12: a concatenation inside the template before the await", "async function later(text: string): Promise<string> { return `${text}!`; }\nconst big = \"x\"; const more = \"y\";\nasync function run(): Promise<string> { const first = later(\"a\"); return `${big + more}${await first}`; }\nconsole.log(await run());\n", ""},
		{"c13: the await as a second argument", "async function later(text: string): Promise<string> { return `${text}!`; }\nfunction join2(a: string, b: string): string { return a + b; }\nconst big = \"x\";\nasync function run(): Promise<string> { const first = later(\"a\"); return join2(big + \"y\", await first); }\nconsole.log(await run());\n", ""},
		{"d1: a const callee in its dead zone, the promise rejecting otherwise", "async function later(text: string): Promise<string> { if (text === 'a') throw new Error('boom'); return `${text}!`; }\nasync function run(): Promise<string> { const first = later('a'); return wrap(`${await first}|`); }\ntry { console.log(await run()); } catch (error) { console.log(error instanceof Error ? error.name : 'other'); }\nconst wrap = (s: string): string => s;\nconsole.log('end');\n", general},
		{"d2: a declared callee, the promise rejecting otherwise", "async function later(text: string): Promise<string> { if (text === 'a') throw new Error('boom'); return `${text}!`; }\nfunction wrap(s: string): string { return s; }\nasync function run(): Promise<string> { const first = later('a'); return wrap(`${await first}|`); }\ntry { console.log(await run()); } catch (error) { console.log(error instanceof Error ? error.message : 'other'); }\nconsole.log('end');\n", lowers},
		{"d3: a const read in its dead zone before the await", "async function later(text: string): Promise<string> { if (text === 'a') throw new Error('boom'); return `${text}!`; }\nasync function run(): Promise<string> { const first = later('a'); return `${prefix}${await first}|`; }\ntry { console.log(await run()); } catch (error) { console.log(error instanceof Error ? error.name : 'other'); }\nconst prefix = 'p';\nconsole.log('end');\n", general},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			want := probe.want
			switch want {
			case "":
				want = stringLengthConsumption
			case lowers:
				if err != nil {
					t.Fatalf("must lower: %v", err)
				}
				return
			}
			if !errors.As(err, &notYet) || notYet.What != want {
				t.Fatalf("want NotYet %q, got %v", want, err)
			}
		})
	}
}
