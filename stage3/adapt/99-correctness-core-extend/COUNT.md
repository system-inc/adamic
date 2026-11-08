# Complete lane count on main 698965fe6

One bounded run of the unchanged whole-lane oracle passed in 604.846 seconds.
The normal API, test-count, baseline and failure guards all passed: 106366 passes,
one sanctioned API failure, zero skips. Its complete test-identity multiset and
raw baseline diff match the independently passing uninstrumented whole-lane
control exactly. No timeout or baseline guard was relaxed.

| Owner | Calls | First argument plain | Second argument plain | Counterexamples |
| --- | ---: | ---: | ---: | ---: |
| extend | 10169 | 10169 | 10169 | 0 |
| copyProperties | 490 | 490 | 490 | 0 |

For each argument of each owner: non-plain objects, inherited enumerable strings,
user accessors, own symbol keys, null-prototype objects, arrays and sparse arrays
all have count 0. All invocations returned normally. The observer records throws
as well, preserves exact exception identity, and does not invoke getters.
The standard inherited Object.prototype __proto__ accessor is excluded from the
user-accessor category. Node witnesses agree on all 14 owner/class pairs,
including inherited properties, accessors, symbols and sparse arrays.

| Runtime caller | Calls |
| --- | ---: |
| commandLineParser.ts:3095, extend | 9902 |
| commandLineParser.ts:3101, extend | 0 |
| harness/fourslashImpl.ts:352, extend | 267 |
| watch.ts:873, copyProperties | 408 |
| tsbuildPublic.ts:326, copyProperties | 82 |

Every recorded stack resolves to a listed caller. The unexercised WatchOptions
merge call is covered by byte-identical JavaScript, as are all other inputs.
This is a type-only adaptation: equality follows from exact emitted-code identity,
not from a shadow alternative implementation. There is no JavaScript-changing
road, off-count semantic difference, new cast, caller edit or ledger amendment.

The complete 81504 non-build inputs, including all 80520 fixture files, were
verified before and after counting. The measurement tree differs only by two
wrappers that forward the exact original bodies once and preserve their returns
and exceptions. Full raw records are in lane-count.tar.gz. COUNT.json and
RUNTIME-CALLERS.json are independently reconstructed summaries. Its internal
SHA256SUMS and adjacent SHA-256 cover all retained evidence.

Earlier setup, fixture-copy and contention attempts are explicitly excluded.
Historical f027 data is in historical-f027.tar.gz; current-main contention data
is in core-proof.tar.gz. Accepted totals contain only the passing bounded run.
