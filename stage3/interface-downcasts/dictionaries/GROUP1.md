Built dictionary read components and recursive descriptor preparation on integration ba59427c; source admission is still blocked.
Commit: this checkpoint on codex/views-dictionaries, for the integrator through the user; no other lane was merged.
Validation: focused lower/oracle, record/lazy regressions, package vet, ranking reproduction and whitespace checks pass.
Mutants: 6 requested component mutations plus 5 guard mutations lose their pinned refusals; all build and execute valid code.
Remaining: 29 candidate pairs / 228 candidate reads; exact production reachability is unmeasured and no source pair is completed.

Revised whole-family working date: October 19, 2026, 23:00 UTC, conditional on
source record production and shared read-dispatch handoffs being resolved by
October 9. The integrated branch has lazy admission, but still no compiler
record representation or dictionary indexed-read dispatch. The remaining work
is larger than attaching a contract selector. This estimate is not an
unconditional promise to complete dependent arrays/unions or encode null without
the shared representation owner. The prior October 16 estimate is superseded.

The integration merge was a clean fast-forward from 753cca96 to ba59427c, which
already contains the earlier dictionary checkpoint. Integration/REPORT.md was
read; its conflict choices were preserved. Direct lane merges remain forbidden.
The original CLAUDE.md and unit rule reserve existing shared files to their
owners. This unit has changed only its new files and its plan additions.

## Demand and limits

The cited lazy/REPORT.md labels its read table static candidates, not runtime
allocation-flow reachability. Its ADAPTED-CENSUS.md explicitly records exact
remaining runtime reachability as unmeasured because checker-rejected entries
have no executable IR/shared allocation graph. There is no exact production
read table in that report to measure successful per-pair completion against.
This discrepancy was raised to the user while implementation continued.

rank-lazy-demand.py now consumes the lazy owner's compressed candidate pair
inventory, preserving witnesses and input SHA-256. Its dictionary-contract set
is 18 pairs / 174 reads; its dictionary-read set is 14 pairs / 66 reads. The
deduplicated union is 29 pairs / 228 candidate reads, not 32 / 240. Exact counts
are null, not zero. Leading pair ParsedCommandLine.options has 111 reads.
The old 409 / 2,256 inventory remains historical and does not set this queue.
Both candidate-pair-progress.json and candidate-summary.json reproduce byte for
byte with --check. All candidate rows remain pending after this push.

## Implemented components

The native helper reads the existing record.c Map wrapper. It adds no second
record table, readiness bitmap or heap representation. Values in its supported
producer representation are existing counted union boxes or ordinary heap
references. Each selected read classifies the actual box header before accessing
the payload, checks a logical-kind mask, and returns a borrowed normalized value.
Missing keys yield undefined only when its kind is permitted. Existing own-key
prototype refusal is preserved. Unknown/unboxed source storage refuses before
any record slot is followed. Casting to a dictionary cannot certify that storage.
The source adapter must establish record identity and boxed storage independently.

The JavaScript helper reads an own data descriptor, rejects an accessor without
running it, checks the selected value's runtime kind, and preserves record.c's
prototype-name policy. Container checks reject primitives, arrays, Map and null.
The first group covers broad primitive unions, absence, object-kind membership
and nested object reads. Finite literals, complete array/callable/union contracts
and native null encoding remain dependent on shared adapters. A kind mask alone
never authorizes those contracts.

Reference results preserve a nonzero child contract. The nested oracle consumes
that id and invokes the existing readiness-aware adamic_object_view at the later
entry.name read. Its JavaScript adapter retains the same later field check.
This proves the component adapter seam, not source aliases/generics/callbacks.
No object field is recursively inspected at dictionary admission or key lookup.
Readiness of record entries follows actual definitions in the existing table;
no staged/uninitialized record entry encoding has been introduced.

The lower helper reserves an id before recursively building the index element
and named fields in the shared registry. Failed child construction rolls back
new ids. Numeric/template/symbol index contracts refuse in this adapter rather
than disappear. The root deliberately stays ViewUnknown with Unsupported set to
'dictionary source dispatch'. It cannot become a supported-slot certificate while
shared production and reads are unwired. Tests cover optional named strict fields,
recursive dictionary/object graphs, deferred certification and failure rollback.

## Node and refusal evidence

Five source .a controls under components/ run on source Node. Their output is
compared with the component native/JavaScript adapters for valid values. Negative
runs pin exit 70 and the exact owned diagnostic. Normal native runs use ASan,
UBSan and leak checking through native.Build(Sanitize:true); mutants use release C.

| Control | Source Node | Component backends |
| --- | --- | --- |
| options-good | 42 then undefined, exit 0 | identical |
| options-wrong | [object Object] then undefined, exit 0 | exit 70, view.options['value'], expected string \| number \| boolean \| undefined, found object |
| nested-good | name, exit 0 | identical |
| nested-wrong | 42, exit 0 | exit 70 at later entry.name, expected string, found number |
| array-wrong | undefined, exit 0 | exit 70, view.symbols['item'], expected Entry \| undefined, found array |

The component programs are not compiler output from those .a sources. The
integrated compiler was built separately and run on the original source probes:
both c and js still exit 1 with the index-signature syntax refusal. Their exact
frontier observations are in probe-observations.json. Whole-tsc diagnostics were
not used to excuse this per-fixture source frontier.

## Mutants and tests

Three requested mutations were run independently in C and JavaScript:

- Skip membership check: options-wrong runs on with [object Object]/undefined,
  exit 0. The pinned diagnostic catches both mutations.
- Accept array as object: array-wrong runs on with undefined, exit 0. The pinned
  element-read diagnostic catches both mutations.
- Drop child contract: nested-wrong runs on with 42, exit 0. The pinned later
  entry.name failure catches both mutations. This is an adapter-seam mutation.

Additional independent JavaScript mutations remove the container, accessor,
nonzero child-contract and required-missing-key guards. Each normal run stops
with its exact diagnostic; each mutant prints passed and exits 0. Removing the
native source-storage guard reads a deliberately uncertified but allocated
record and prints 42/undefined instead of its unsupported-representation stop.
No mutant is counted for a compilation warning, compiler refusal or sanitizer.
The first array-shape mutation was initially masked by the separate later field
check in its JS harness; the isolated harness now prints the source array result
without that second guard. The final rerun loses the intended refusal in both
components. That initial failed harness run is preserved separately.

All output is captured in logs. Commands and completed output:

```text
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run '^TestViewDictionary|^TestCheckedViewDictionary' -count=1 -v -timeout 10m
ok internal/lower 0.318s
ok internal/oracle 0.897s
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestViewDictionary|TestCheckedViewDictionary|TestRecordsAgainstNode|TestLazyView|TestCheckedViewLazy' -count=1 -timeout 10m
ok internal/lower 0.125s
ok internal/native 8.324s
ok internal/javascript 0.004s [no tests to run]
ok internal/oracle 4.978s
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
exit 0, no output
python3 stage3/interface-downcasts/dictionaries/rank-lazy-demand.py --check
29 candidate pairs / 228 candidate reads remain; exact runtime reachability unmeasured
git diff --check
exit 0, no output
```

The first regression attempt failed because stage3/api's pinned @types/node was
absent. npm ci --prefix stage3/api fixed that prerequisite; the rerun passes.
The complete JavaScript package also passes in 0.520s (logs/group1-javascript-package.log). No full lower/native/
repository gate is claimed, and no counts.md changes are needed for component
programs outside the production fixture registry.

Setup rerun: GOPROXY=https://proxy.golang.org|direct, source the printed
/workspace/adamic-tools/env.sh, nproc=5, quota 4 CPUs. Go ready 0.021s,
Node ready 0.024s, Markdown ready 0.072s, clang ready 0.153s, submodules
ready 3.141s, Go build ready 83.447s, cache warm 83.575s, done 83.605s.
The actual timing log is retained if a displayed timing differs.

## Source wiring handoff

The plan names the exact entry points and owner-held files. End-to-end source
support still needs these together:

1. Index-signature syntax must become a supported/deferred obligation instead of
   unconditional refusal in lower/refusals.go. Do not remove the refusal before
   dictionary producers and all reachable read paths have sound alternatives.
2. Wire viewDictionaryContractHook into lazy viewContract classification. Supply
   a proper dictionary contract kind/representation after complete source support;
   currently returning Unsupported is intentional and must stay fail-closed.
3. Add record literal/lookup/write/delete operations through shared IR and lower
   expression dispatch, implemented by lane-owned helpers using record.c. Certify
   producer storage independently and preserve observable identity/mutating aliases.
   Never reinterpret an ordinary scalar table as boxes or copy a record on a read.
4. Wire indexed and named dictionary reads to these helpers with evaluated operands,
   the complete selected element contract and exact source expression. Carry child
   contracts through helper arguments, returns, captures, stored fields and generic
   instances using the existing shared flow. Unknown retains the guard/refusal.
5. Include DictionaryRuntime in JavaScript assembly. Native runtime files already
   participate in the existing runtime embed/build. Exact literals, null and array/
   callable element consumers must retain their owner checks or named refusals.

This group's tip is ready for the integrator through the user. It does not claim
that lazy admission alone implements the missing record compiler representation.
