# 99k: sameMap's input extends its output; builder checks reused chains

Task #b5w3ycg. Pinned upstream TypeScript 6.0.3 is
`050880ce59e30b356b686bd3144efe24f875ebc8`. Initial Adamic main is
`6761c7374aa7169c73602b6a74bc281b879a640d`.

The four sameMap overloads and implementation now use `<T extends U, U = T>`.
The slice initializes `const result: U[] = array.slice(0, i)` directly, and the
last statement is `return array`. `item as unknown !== mapped` stays unchanged;
it is an upcast. Only builder.ts's checks change JavaScript behavior, and only
when a chain invariant is broken. `javascriptChanging` is true. core.ts's
transpiled JavaScript is byte-identical with comments kept and removed.

## Builder edits

Only convertOrRepopulateDiagnosticMessageChain and
convertOrRepopulateDiagnosticMessageChainArray are edited.

The scalar function keeps both repopulation branches and recursive calls.
After converting next, its identity arm asserts all three required properties
and `value.next === next`, then returns the original chain without a cast.
The local type predicate is justified by that equality: next has already been
converted to full chains. In the changed-next arm, the three properties are
asserted before returning `{ ...chain, next }`. This removes both casts to
DiagnosticMessageChain. The existing three `chain.next as T[]` casts are
unchanged. There is no coercion, default value, or unchecked narrowing.

The array function retains its original generic signature as an overload. Its
implementation accepts the union and callback on that union; callbacks still
receive only original input chains. Undefined returns unchanged. A recursive
property check narrows full arrays, which use
`sameMap<DiagnosticMessageChain>` with T and U both DiagnosticMessageChain.
Arrays containing a reusable info-only chain use map to convert every element.
Such an element, or its ancestor, must change identity when repopulated, so
sameMap would also allocate a result for those inputs. Conversion order, output
values, and unchanged array/chain identity are preserved. The additional array
traversal checks properties before conversion; it does not call repopulateInfo.
The two external callers pass callbacks reading `chain.repopulateInfo?.()`
(builder.ts:528) or `chain.info` (builder.ts:610), respectively. The compiler
repopulation factories return metadata; the checks do not invoke those callbacks.

## Property inventory and unconditional assertion

At the fully adapted initial head, DiagnosticMessageChain is at
`src/compiler/types.ts:7259-7268`. Its required distinguishing property set is
exactly **messageText (7260), category (7261), code (7262)**. At the pinned
pristine source the corresponding lines are 7249, 7250, 7251. next,
repopulateInfo, and canonicalHead are optional. The nested next type is checked
recursively for array narrowing, and by equality to converted next for the
scalar identity return.

`builder.ts:115-118` defines ReusableRepopulateInfoChain with info and optional
next only. `builder.ts:121-123` defines SerializedDiagnosticMessageChain using
Omit of next and repopulateInfo, retaining all three required properties;
`builder.ts:126` is their union. Repopulation info is the object at
`types.ts:7241-7245` or true at 7248; its union is at 7251. No other reusable
chain variant is omitted.

Main does not initially include 99-cli-debug-assertion-level. Its specified
branch was fetched at `b1eb58de534346dba084923acbc617a0abb75913` and applied to a
separate scratch source tree. Debug.assert at that adapted tree's
`debug.ts:192-200` is unconditional: false calls fail, which throws. It does not
call shouldAssertFunction or consult the assertion level. The six specialized
node assertions are gated; this adaptation uses none of them.

## Count before building

Before creating this adaptation, only a scratch copy of the fully adapted tree
was instrumented. The original return expression stayed intact. The exact
identity-return path ran **6 times, with 0 non-full chains**. The wider path
with no repopulation info ran **7 times, with 0 non-full chains**. Missing
messageText, category, or code was counted for every call. Per-process counters
were flushed on exit. The full oracle used all runners, eight workers, and no
filter: **106,366 passing, one sanctioned API failure, zero pending**. The full
pass/fail identity multiset equals main's, including multiplicities. The first
concurrent main reference run lost a worker to the 16 GB cgroup's OOM killer;
the serial retry passed with the expected identities. No test expectation,
timeout, runner, or worker count was changed.

See count-results.json, count-instrument.py, and LEDGER.md. The count clears
Adamic's condition 3. No non-full observation was ignored.

## Adapter contract and ordering

Stock TypeScript 6.0.3 parses both complete files before and after planning.
owners.json pins exact before/after source text and SHA-256 for three named
function owners, including the complete sameMap overload set and the array
helper's overload/implementation pair. Owners must be top-level with the exact
reviewed declaration counts. No whole-file hash is an adaptation precondition.
A changed, moved, missing, duplicate, or half-adapted owner, mixed file/tree, or
parse error is refused before any write. Every original file is reread before
writing. Unrelated functions are accepted and preserved. A second run makes
zero edits and zero writes, preserving bytes and mtimes. CRLF is preserved.

Python's real directory sort places this directory after 99j and before
99-wasm-crypto-import. The user explicitly corrected the earlier requirement
about 99-cli: it may sort earlier because it edits only debug.ts.

## Reproduction

Run from the repository root with NODE_PATH pointing to stage3/api/node_modules.
Witness output arguments must be outside this directory, because all its files
participate in apply's product key.

```
node stage3/adapt/99-correctness-axolotl-k/contract.cjs BEFORE_TREE > CONTRACT.json 2> CONTRACT.log
node stage3/adapt/99-correctness-axolotl-k/witnesses.cjs BEFORE_TREE AFTER_TREE WITNESSES.json > WITNESSES.log 2>&1
node stage3/adapt/99-correctness-axolotl-k/witnesses.cjs --mutant BEFORE_TREE AFTER_TREE MUTANT.json > MUTANT.log 2>&1
```

The mutant restores the unchecked identity return and must exit 1. See PROOF.md
and CENSUS.md for the final verification results and new first stops.

## Verification

All measured checks pass. [PROOF.md](PROOF.md) records the 11 matching valid Node
rows, red mutant, no-write contract, 79-root Adamic checks, byte-identical public
declarations, complete passing lane and Go fixtures. [CENSUS.md](CENSUS.md)
records all 73 movers: every one now stops at core.ts:381:26 in unchanged flatten.
