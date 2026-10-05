package fuzz

import "fmt"

// Long inputs reach what short ones can't. TimSort gallops only when it merges runs, and runs exist
// only from 64 elements on; the string index (runtime/string_index.c) is built only for a string of
// 64 bytes or more, and checked only by reads at scattered positions. Everything else the generator
// makes is kept short, so these two live in globals of their own, long and never cut: an array of
// patterned runs, and a non-ASCII string with surrogate pairs in it. Reviewer R2's round four found
// a planted galloping bug and a planted index bug both invisible to the generator without them.

// longArray and longText are the long globals' names.
const (
	longArray = "long"
	longText  = "longText"
)

// declareLong declares the long array and the long string: built at runtime, from runs ascending
// and descending, duplicates and noise, and from pieces of one, two and three bytes and four (a
// surrogate pair).
func (g *generator) declareLong(add func(*Statement)) {
	pattern := g.pick(
		"(index % 23) + Math.floor(index / 23) * 3",
		"Math.floor(index / 17) % 2 === 0 ? index % 17 : 40 - (index % 17)",
		"(index * 37) % 101",
		"index % 7 === 0 ? (index * 13) % 50 : index",
		"Math.floor(index / 9) % 3",
	)
	add(statement(fmt.Sprintf("const %s: number[] = Array.from({ length: %d }, (_, index) => %s);", longArray, 64+g.random.IntN(240), pattern)))
	g.declare(longArray, NumberArray, false)
	pieces := g.pick("['a', 'é', '🌍', '世', 'bc']", "['🌍', 'x', '\\uD83C', 'ß']", "['世界', 'a', '🌍🌍', '']")
	add(statement(fmt.Sprintf("const %s: string = Array.from({ length: %d }, (_, index) => %s[(index * 7) %% 5] ?? 'z').join('');", longText, 30+g.random.IntN(120), pieces)))
	g.declare(longText, String, false)
}

// longStatement sorts the long array, by a comparator consistent or not, or reads the long string at
// scattered positions, and prints what came of it.
func (g *generator) longStatement() *Statement {
	if g.chance(1, 2) {
		comparator := g.pick(
			"(left, right) => left - right",
			"(left, right) => right - left",
			"(left, right) => (left % 3 === 0 ? 0 / 0 : left - right)",
			"(left, right) => ((left * 7 + right * 13) % 5) - 2",
			"(left, right) => capture(0) * 0 + Math.floor(left / 4) - Math.floor(right / 4)",
		)
		return statement("@b", &Block{Statements: []*Statement{
			statement(longArray + ".sort(" + comparator + ");"),
			statement("console.log(`${" + longArray + ".slice(0, 6).join(',')} ${" + longArray + ".reduce((sum, value) => (sum * 3 + value) % 1000003, 0)}`);"),
		}})
	}
	stride := g.pick("7919", "31", "97", "1")
	read := g.pick(
		"`${"+longText+".charCodeAt(position)}`",
		"`${"+longText+".codePointAt(position) ?? -1}`",
		"("+longText+".at(position) ?? '')",
		longText+".slice(position, position + 3)",
		"`${"+longText+".indexOf("+longText+".slice(position, position + 2))}`",
	)
	return statement("console.log(Array.from({ length: 24 }, (_, step) => @b).join(','));", &Block{Statements: []*Statement{
		statement("const position = (step * "+stride+" + @e) % "+longText+".length;", compose(Number, "Math.abs(Math.trunc(@e))", g.expression(Number, 1))),
		statement("return " + read + ";"),
	}})
}
