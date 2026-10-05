# Stream R2, round four: stream A's fuzzer, run for an hour

On `cloud/integrate-7` at 7774d94 (not yet on main, which is 4ddd17f), alone on this 4-core machine. `adamic-fuzz -parallel 4`, in batches of 100 seeds, from seed 200000, for 3,313 seconds (`review/r2/fuzz/fuzz-hour.sh`; every batch's summary is in `fuzz/batches.txt`). Nothing on any branch was changed. The one change I made to the fuzzer, in my own copy, is below, and it's why the run tested anything.

## Finding 0: on integrate-7, and on main, the fuzzer as merged tests nothing

Every program the generator makes declares

```ts
interface Link {
	value: number;
	label: string;
	next: Link | undefined;
}
```

for the `optional-chains` feature, which is on by default. Main's cycle finder (8bc3e68, already on main) refuses that mutable `next` as cycle-capable. So every program is refused before it runs:

- a calibration run of 40 seeds (199900 to 199939) on 7774d94 came back all `invalid: refused ... Link.next ... (adamic/cycle-capable)`;
- 20 of 20 programs from seeds 300000 to 300019, put through `adamic js`, were refused for it.

`-without optional-chains` gets past it, but it leaves out the linked list, `?.` and the narrowing paths, which is where round three's findings were. The generator never assigns `next` after a node is made (`chainWrite` builds a new node, or takes `chain?.next`), so `readonly next` means exactly the same thing and is acyclic. My run used that one-line change to `linkDeclaration` in internal/fuzz/widen.go (`fuzz/readonly-next.patch`). With it, no program was refused as cycle-capable. The fix belongs in A's generator. Until then, the fuzzer on main reports every seed as invalid, and a run that only counts findings would read as green.

## The counts

| | Programs |
|---|---|
| Run (seeds 200000 to 207399) | **7,400** |
| Agreed: Node, native under ASan and UBSan, and the backend, byte for byte, and no leaks | 6,158 |
| Checked: an inserted check fired, native held to the backend | 1,180 |
| Not yet: "a function returning never" (an arrow in a dead branch, as A's lower_test.go names it) | 25 |
| Invalid: the checker refused what the generator wrote | 30 |
| Unfit: Node printed more than a megabyte (6), or never finished (1) | 7 |
| **Findings** | **0** |

No disagreement came up, so there was nothing to shrink or to judge against Node's side.

**Can this run find anything?** I ran the same fuzzer, same patch, against mutant compilers:

- `%` emitted as C's `remainder()` instead of `fmod()` (emit.go): **21 findings in 100 programs** (seeds 300000 to 300099; 14 stdout, 7 exit code). The shrinker made each a short program: `fuzz/mutant-remainder-shrunk-300010.a`, 30 lines, whose `3 % number3` sends a branch the other way. Log: `fuzz/mutant-remainder.txt`.
- TimSort's `minimum_gallop` 8 instead of 7, and string-index checkpoints without their low-half bit, together: **0 findings in 400 programs** (seeds 310000 to 310399). Both are caught at once by the hand-written sweeps of rounds one and three (`i7/timsort_sweep.a` from length 64, `probes/string_index_stale.a` from its first line). Log: `fuzz/mutant-gallop-and-checkpoints.txt`.

So the zero is evidence for what the generator reaches, and none for what it doesn't. Its arrays stay short, and its strings are cut to 40 units with `.slice(0, 40)`. Galloping (a merge of runs from 64 elements on) and the string index (from 64 bytes on, read at scattered positions) are out of its reach.

## The 30 invalid programs: the generator writing what tsc refuses

Read in seeds 200000 to 200999, where three of the 30 fall (`fuzz/200129.a`, `fuzz/200334.a`, and the checker's words beside them):

- `200129`: `if (chain !== undefined) { chain.value = tally5(...); }` inside a branch, which tsc calls `TS18048: 'chain' is possibly 'undefined'`. tsc doesn't keep the narrowing there. I didn't work out which statement before it makes it drop it.
- `200334` and `200915`: `'İ' !== `<${5}`` and the like, two literals of types tsc proves can't overlap: `TS2367`.

0.4% of programs are wasted, and nothing is wrong in the compiler. The generator could compare a literal against a variable, and narrow through a `const` copy.

## What the generator never produced

From the source of all 7,400 programs (`fuzz/coverage.py`, a regular expression per construct; `fuzz/coverage.txt` has every count). Never, in 7,400:

- **Narrowing and unions:** `typeof`, `instanceof`, a `string | number` union, a checked cast (`as`), `switch`.
- **Control:** `do...while`, nested function declarations, an explicit `panic`.
- **Types:** generics, tuples, `Set`.
- **Library:** `normalize`, `String.fromCharCode` and `fromCodePoint`, spreading a Map (`[...table]`), `console.error`, `readTextFile`, `writeTextFile`, `programArguments`.
- **Operators:** the bitwise operators. A first count said 7,400, but that regular expression matched the type union `Link | undefined`; the corrected one finds none, and the generator has none to write.

Some of these aren't on integrate-7 at all (bitwise, fromCharCode, tuples: P's branch), but most are in 0.1 and on main. The ones with findings by hand in rounds one to three are absent: `number | undefined` fields, `typeof` on a narrowed value, tuples. Every one of `fuzz.Features`' named features appeared (string-index and array-index counted together, since one expression can be either), the least often `array-write` (2,716) and `array-spread` (3,118); of the methods outside the list, `split` (1,258) and `trim` (1,779) were the rarest. Four constructs are in every program by construction (`class`, `while`, the Map iteration that prints `table`, and `array-methods`), and `slice` in 7,396.

## What I didn't cover

- The 1,180 checked programs were held to the backend, not to Node, as the fuzzer is designed. I didn't sample them to see whether each check fired where Node would have done the same, which is the one place a wrong check could hide.
- The program Node never finished (one unfit) wasn't found again: the batch summaries don't name unfit seeds. Running with `-v` would.
- Only 7774d94. The allocator branch, which changes every allocation, would be the next thing worth an hour of this.
