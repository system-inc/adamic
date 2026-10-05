# Stream R2, round five: the allocator fuzzed in release builds, and the checked programs held to Node

Two jobs. Fuzz stream C's size-class allocator (`cloud/hx74vgw-alloc` at 3e98c99) in release builds as well as sanitized ones, since the allocator only runs unsanitized. And hold the programs the fuzzer files as "checked" to Node, to see whether each inserted check fired for a reason Node's own run shows. Nothing on any branch was changed. What I built for this is on this branch under `review/r2/fuzz2/`, and my branch has main (4ff4657) merged in.

## How it was run

- **The tree.** The allocator branch has neither the fuzzer nor A's fixes, so I merged it locally into integrate-7 (7774d94; commit 74ebf8f, never pushed). The one conflict was native.go: integrate-7's `Flags` function against the allocator's `slabs` test option. I kept `Flags` and added `-DADAMIC_SLABS` to it.
- **The fuzzer**, in my copy only (`fuzz2/fuzzer-release-and-checked.patch`), plus round four's `readonly next`:
  - `Prepare` also compiles the runtime with release flags (`native.Flags(native.Options{})`, -O2, no sanitizers, so the size classes on).
  - `Try` builds and runs each program a second time that way.
  - `judge` reports any difference between the release run and the sanitized one, before anything else.
  - The driver keeps every checked program with the panic it stopped at.
- **Proof the release comparison can fail.** A mutant giving every class above 32 bytes the slots of the class below it (heap.c) made all 40 of 40 programs a finding: "release build exit code differs from the sanitized one". It can't show in a sanitized build, which takes every value from malloc. Log: `fuzz2/mutant-size-class.txt`.
- **The run.** `fuzz2/fuzz-alloc.sh`, alone on the 4 cores, in batches of 100 from seed 500000. I stopped it after 56 batches and part of a 57th, because the shrinker was spending minutes on the one slow program below.

## The counts

| | Programs |
|---|---|
| Complete batches (seeds 500000 to 505599) | **5,600** |
| Agreed (Node, native under ASan and UBSan, the backend, and now the release build too) | 4,669 |
| Checked (an inserted check fired) | 895 |
| Not yet ("a function returning never") | 11 |
| Invalid (the checker refused the generator's program) | 23 |
| Unfit (Node printed more than a megabyte) | 2 |
| **Release build differs from the sanitized one** | **0** |
| Findings | 0 |

The partial batch (505600 on) added 11 checked programs and one finding, seed 505616, which follows.

**The allocator held.** In 5,600 programs no release build said anything its sanitized build didn't, and every sanitized build agreed with Node or stopped at a check.

## Confirmed findings

### 1. 616 of the 906 checked programs stopped at a check that shouldn't have fired (A's narrowing check)

Every checked program was run once more with its JavaScript backend output patched so each check reports its caller's line instead of panicking (`fuzz2/site.py`; all 906 sites are in `fuzz2/sites.tsv`). Every narrowing check that fired fired at one of three sites. At all three, JavaScript reads undefined and goes on correctly:

| Site | Programs |
|---|---|
| A: `chain?.next`, where `chain` was narrowed and a call put undefined back. The check panics with JavaScript's own `TypeError: Cannot read properties of undefined (reading 'next')`, where JavaScript, at `?.`, gives undefined | 91 |
| B: `let walk: Link \| undefined = chain;`, a copy into a variable typed to hold undefined | 388 |
| C: `chain = { ..., next: chain }`, a store into a field typed to hold undefined | 137 (116 seen at once, 21 more once their 220-character site was printed whole) |

None of the 616 narrowing checks fired where JavaScript would have gone on with a wrong value. The fuzzer can't see any of this: both of Adamic's backends stop at the same check and print a prefix of Node's output, which is its definition of "checked".

Minimized by hand, each on main at 4ff4657:

`fuzz2/optional_chain_narrowed.a` (site A):

```ts
interface Link {
	readonly value: number;
	readonly next: Link | undefined;
}
let chain: Link | undefined = { value: 1, next: undefined };
function drop(): void {
	chain = undefined;
}
drop();
chain = chain?.next;
console.log(`${chain === undefined}`);
```

```
== node     exit=0   true
== native   exit=70  adamic: panic: TypeError: Cannot read properties of undefined (reading 'next')
== backend  exit=70  adamic: panic: TypeError: Cannot read properties of undefined (reading 'next')
```

Before A's check (main at 2385966, this branch before today's merge), native printed `true`. `fuzz2/narrowed_copied.a` (sites B and C): Node prints `copied true` and `stored true`, and both backends panic with "undefined where the checker narrowed it away". It's round three's finding 2 seen at scale, and site A is new.

The cause, read from internal/lower/narrowed.go: `defined()` wraps a narrowed read in `ir.Defined` unless it's compared with `undefined`. It doesn't look at whether the property access it feeds is `?.` (A), or whether the read is converted to a type that holds undefined (B and C). A fix: no check when the parent is an optional property access, or when the read's contextual type includes undefined.

### 2. Round two's tuple finding is on main now: a tuple read as an array gives a silent wrong answer

C2's "tuples as values" (5f0b50d, merged in 1e8ec6b) landed on main with the hole round two found on P's branch. `p/tuple_length_silent.a` on main at 4ff4657:

```
== node     exit=0  length 2
== native   exit=0  length 94319282874752
== backend  exit=0  length 2
```

The release build printed `length 94720206621728`, exit 0. `p/tuple_through_closure.a` is a heap-buffer-overflow in `adamic_array_join` under ASan, and `p/tuple_as_array.a` reaches clang as bad C. On main the backend is now right and native isn't, so the oracle would catch it with a fixture, but none has one.

### 3. Searching a long string costs a pass over it, whatever the answer: 46 to 60 times Node (a cost, on main, not the allocator)

Seed 505616 came back "native never finished": the sanitized binary passed the fuzzer's 20-second limit. It isn't a hang. Run whole:

| Build | Time | Output |
|---|---|---|
| Node | 0.12 s | 127,961 bytes |
| merged tree, release | 8.6 s | identical |
| merged tree, sanitized | 39.6 s | identical |
| main at 4ff4657, release, no allocator | 8.6 s | identical |

gprof on the release build: 42% of the time in `units_next` (2.6 billion calls), 23% in `adamic_string_concat`, and most of the rest in `adamic_string_index_of` and `adamic_string_last_index_of`.

- `lastIndexOf` converts the whole haystack to UTF-16 before searching from the end, even for an empty search (`text.lastIndexOf(text.slice(1e21))`, which is the length) or a match at the very end.
- `indexOf` does the same only when its search holds surrogate halves. Its ordinary path scans bytes only as far as the match, then takes the unit count from the string index by binary search, so it costs the distance to the match, as V8's does. I first wrote that it counted every unit before the match; reading `adamic_string_units_before` showed it doesn't.
- The concat time is a string built by `+=` to 127,000 characters, copied whole each time.

`fuzz2/last_index_cost.a` isolates the searches: 20,000 rounds of `lastIndexOf('')`, `lastIndexOf('9é')` and `indexOf('0é')` on a 40,000-unit string. Node takes 0.096 s and native release 4.46 s, with the same answer. Nearly all of the time is `lastIndexOf`. Round one's string index already turns a byte offset into units by binary search, so `lastIndexOf` could compare bytes from the end, as `indexOf` compares them from the start (falling back to units only when the search holds surrogate halves), and answer an empty search with the length.

## The other 290 checked programs: index writes and shrinking callbacks, all held to the source on Node

These are checks 0.1 makes on purpose: a write at an index the array doesn't have (257 at NaN, 28 at an integer past the end), and `map` or `findIndex` over an array the callback shrank (5). For each, the source ran on Node through `fuzz2/probe.mjs`, with the same condition tested on Node's own state:

- each `target[index] = value;` statement rewritten to a probe that panics, as `adamicSetIndex` does, when the index isn't one the array has;
- `map`, `find` and `findIndex` replaced by loops that panic where the array shrank under them.

`fuzz2/hold.py` compared that run with the backend's (a checked program's backend run is byte for byte its native one; that's the fuzzer's own verdict). **All 290 stopped at the same check, after the same stdout, with the same message** (`fuzz2/hold.tsv`). **Mutant:** the probe's write condition made one element stricter (`index < array.length - 1`) turned 12 of the first 60 to DIFFERENT, so the comparison can fail.

## Main at 4ff4657, today's other findings

Rechecked after the merge:

- Round three's narrowing findings: all still there. The `number | undefined` field reads NaN natively (`i7/narrowed_field_silent.a`), the check panics at `typeof` and at arguments (`i7/narrowed_reads_undefined.a`), and a narrowed array element reaches clang (`i7/narrowed_element_plain.a`).
- Round two's P-branch findings, other than tuples: `function_identity.a` and `constructor_number_later.a` are still NotYet on main, since P's branch isn't merged.

## What I didn't cover

- Shrinking seed 505616 through the fuzzer: stopped, since each try takes 20 seconds or more. I read the cause by profiling instead.
- The fuzzer still never generates long strings or long sorted arrays (round four). This hour was blind there too, which is where finding 3 came from only by luck: a string grown inside a loop.
- macOS, and the allocator under memory pressure.
