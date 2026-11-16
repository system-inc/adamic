# pidigits: not ported

The holdout keeps pidigits in its list, as refused for now.

**Source.** `pidigits/pidigits.node-2.node` from `public/download/benchmarksgame-sourcecode.zip` (blob `05648fc9a0d6`) at benchmarksgame commit `40296663ed350d5fe4a6ab5e367bab61cb77c219` (March 7, 2025). The game has no `pidigits.node` or `.node-1`, so `.node-2` is its lowest-numbered Node.js entry. The game's size is 10,000 digits.

**What it needs.** Arbitrary-precision integers. Isaac Gouy's transliteration of Alexander Fyodorov's spigot holds its eight state values (`k`, `k1`, `a`, `d`, `m`, `n`, `t`, `u`) as `bigint`, written with `0n` and `1n` literals, and grows them past any fixed width: `+=`, `*=`, `<<`, `/`, `%`, `-` and `>` on `bigint`, and `Number(t)` to turn each digit into a `number`. At 10,000 digits the numbers run to tens of thousands of decimal digits, so neither a `number` nor any fixed-width integer can hold them.

**Why it isn't ported.** Adamic has no `bigint` yet. docs/0.1.md's library row lists `bigint` as not built yet, and the compiler stops a program that uses one as NotYet. That's "not yet", not "never": the type isn't refused. Writing the port with an arbitrary-precision integer of our own, built from `number` limbs, would replace the program's algorithm with ours, which is the forcing the holdout rules out.

**When it can be ported.** Once Adamic lowers `bigint` arithmetic and `Number(bigint)`, the port needs only what the language already requires of the others: the `function`'s untyped parameters get types, `i % 10 == 0` becomes `===` (0.1 has no `==`), the `var` becomes `let`, and `+process.argv[2] || 30` becomes the game's 10,000. Its output can then be checked against Node with `cmp`, as the other ports are.
