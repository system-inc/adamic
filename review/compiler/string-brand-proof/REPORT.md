Built a structural phantom-string ABI proof beyond __String, documented in docs/0.1.md, toward Outcome 11 and native tsc.
Implementation commit ca29d315104a2428787b11671b691cae4fad7fff; base 55347dce1537c0bd05d35eab0dc5d663c4823803; delivery branch compiler/string-brand-proof.
Validation: six C/JavaScript Node oracles with sanitizer, release and leak checks; 1,354-file admission delta +6/-0; both local __String reasons 1 -> 0; counts and call-target guard green.
Mutants caught: real-field-phantom, drop-undefined, read-brand, literal-refinement, unchecked-missing-cast; all exit 1 for their intended assertions.
Not covered: the historical full closure 28/16 census was not rerun; const-enum lowering, arbitrary host brands and the existing runtime view boundary were not changed.

## Proof and scope

BASELINE.md records what main does with each shape. types-pin.txt records the exact upstream alias at 050880ce59e30b356b686bd3144efe24f875ebc8. ACCEPTANCE.md and the new docs/0.1.md section describe the structural expansion, including optional phantom fields. The existing phantom proof excludes actual fields, primitive-property collisions, indices and callable layouts. Runtime brand reads, indexed reads and destructuring name the member in the refusal. Real fields carried by runtime expressions or implemented signatures are refused; unused type declarations and overload signatures retain their existing handling.

The string ABI and its normal ownership rules are reused. Undefined is the missing string pointer. The void intersection gets no object layout and conservatively permits that missing representation; it is not an unchecked route for producing a value. Standalone void brands remain outside this proof. Casts preserve assignable underlying string types and insert Defined when removing possible undefined. Unknown casts and literal refinement remain refused. No native emitter or runtime source was changed.

## Admission and census observations

Both CLIs evaluated all 1,354 .a/.ts oracle testdata programs except .d.ts declarations. The main CLI uses source overlays of the unchanged main lowering and cast files. Final CLI hashes are in admission.json. The six newly admitted programs are exactly the six fixtures below; each passed its source-Node comparison. There are no lost admissions. The final comparison reused baseline observations and reran every delivery compilation.

| Reason | Main local files | Delivery local files | Historical full mode |
| --- | ---: | ---: | ---: |
| value of type __String or undefined | 1 | 0 | 28 |
| function returning __String or undefined | 1 | 0 | 16 |

The local reasons use the exact checker spelling __String | undefined. The historical result is from 807d65d9, saved separately in historical-counts.json. The local scan is first-error mode and includes the reductions, not a rerun of the adapted tsc closure. Therefore it does not establish that all historical 44 sites now close. The ordinary enum reductions retain actual InternalSymbolName string values; they avoid the independent const-enum refusal. Symbol is the reduced tsc object interface, not the JavaScript symbol primitive.

## Successful commands

Commands ran from the repository root after sourcing /workspace/adamic-tools/env.sh. Test output went directly to files. The final named runs were:

- timeout 300 go test ./internal/lower -run '^(TestStringBrand|TestClockGenericReturnsT01DeclinesBrandedPrimitives)' -v -count=1 -timeout 90s: PASS, 7.584s (lower-delivery.log).
- ADAMIC_GATE_UNCACHED=1 timeout 240 go test ./internal/oracle -run '^TestStringBrand' -v -count=1 -timeout 90s: PASS, 10.712s (fixtures-delivery.log).
- timeout 180 go test ./internal/lower -run '^(TestMixedUnionContractPhantomBrandUsesPrimitiveBase|TestMixedUnionContractPhantomVoidIsUndefined|TestPhantomPrimitiveNames)$' -v -count=1 -timeout 90s: the two existing view tests passed, 0.399s; no TestPhantomPrimitiveNames exists in this package (existing-brand-views.log).
- timeout 780 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 12m -args -update-counts: PASS, 272.988s. counts.md changed only by the six new rows, all balanced with zero live allocations (counts-delivery.log).
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -v -count=1 -timeout 90s: PASS, 11.652s (call-target-readers-delivery.log).
- timeout 120 go vet ./internal/lower ./internal/oracle: exit 0, no output (vet.log).
- timeout 780 python3 review/compiler/string-brand-proof/run-mutants.py: all five mutants caught; each nested Go test retains -timeout 90s and a 300s build/run outer bound (mutants-green.log, mutants.json).
- timeout 600 python3 review/compiler/string-brand-proof/admission.py: 1,354 programs, six admissions, no losses, both reasons 1 -> 0 (admission-delivery.log, admission.json).

## Leaf seconds

| Test | Seconds |
| --- | ---: |
| TestStringBrandCastChecksMissing | 0.82 |
| TestStringBrandCastKeepsLiterals | 0.83 |
| TestClockGenericReturnsT01DeclinesBrandedPrimitives | 1.02 |
| TestStringBrandRepresentation | 5.46 |
| TestStringBrandRepresentationRejectsLayout | 7.40 |
| TestStringBrandDestructureRefused | 1.24 |
| TestStringBrandIndexRefused | 1.25 |
| TestStringBrandReadRefused | 1.31 |
| TestStringBrandUnknownCastRefused | 0.51 |
| TestStringBrandValue | 4.86 |
| TestStringBrandPlain | 4.89 |
| TestStringBrandEscape | 6.61 |
| TestStringBrandRealRefused | 0.53 |
| TestStringBrandOptional | 6.89 |
| TestStringBrandMap | 5.23 |
| TestStringBrandUnescape | 4.40 |

Each new or touched leaf is below 60 seconds. The existing table-wide counts root is not a new leaf.

## Mutation evidence

| Mutation | Catcher | Observed failure |
| --- | --- | --- |
| Real field treated as phantom | TestStringBrandRealRefused | expected named refusal for actual; got nil |
| Drop nullable-string undefined case | TestStringBrandOptional | Node exit 0, native exit 70; exit codes differ |
| Permit reading phantom member | TestStringBrandReadRefused | named __brand refusal assertion fails |
| Permit literal refinement | TestStringBrandCastKeepsLiterals | unproven string literal refinement admitted |
| Remove Defined from brand cast | TestStringBrandCastChecksMissing | missing checked undefined removal |

Overlay sources are .go.txt, never compilable evidence Go files. Production sources were never replaced during mutation runs. The undefined mutant initially failed the fixture as intended but the harness expected stdout rather than exit-code disagreement; the final harness checks the actual Node exit mismatch and all five completed.

## Setup and bounded retries

GOPROXY was set to https://proxy.golang.org|direct before cloud/setup.sh. Two bounded 240s setup attempts exited 124 during cold builds. The final bounded retry succeeded and printed: Go ready 0.057s; Node ready 0.068s; submodules ready 0.163s; markdown ready 0.182s; clang ready 0.617s; shared cache ready 1.561s; Go build ready 525.464s; test binaries deferred 525.767s; build cache warm 525.770s; done 525.846s. nproc is 5; cpu.max is 400000 100000 (4 CPUs). Tools are Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup printed /workspace/adamic-tools/env.sh, which was sourced. setup-final.log preserves its timing lines.

The initial push window was missed during cold compilation and validation; this was reported while work continued. Bounded preliminary build/test attempts and the first guard/counts attempts timed out. The guard dependency exports were then warmed with go list -export -deps -test -json under a 360s bound; the final 90s guard test passed. The first counts refresh also exposed missing pinned Node declarations and an overbroad refusal of an unused overload type. npm ci --prefix stage3/api installed the pinned declarations, and the refusal was narrowed to runtime uses. The final counts command passed. The optional enum conditional was reduced to array lookup to isolate the nullable result from a separate existing enum rule.

void-view-probe.a is an observation of the existing checked view boundary: an asserted undefined __String payload is rejected natively, while Node erases the assertion. It is not one of the source-Node agreement fixtures and was not used to claim the void arm is inhabited. No runtime view boundary change is included.

No whole package test run or full gate was run. Integration lane results are recorded after the implementation commit. No PR is opened.

## Integration lane

After implementation commit ca29d315104a2428787b11671b691cae4fad7fff, the required command ran under timeout 180:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Exit 0: `lane checks 15.9 s: gofmt and tools on 8 Go files, t.Parallel on 2 test packages; a-check 1 .a files; vet 2 packages`. The remote integration refs needed an explicit refspec fetch because this checkout initially tracked only main. lane-checks.log preserves the output. The final report/log commit changes evidence only.
