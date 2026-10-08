Built: 10 primary pairs / 15 candidate reads certified through the single checked tuple path; optional/rest forms and index/forEach ownership covered.
Commits: code frontier fc21dde4f9445f6250d5311fb2d679254b1e4a80; mutant runner cleanup eeab7e67200b98aaafe669ba933fa9cdfd451de8; all patches pushed separately to codex/views-tuples.
Checks: final tuple suite PASS 122.144s; lane 2/4 regressions PASS 28.186s; 84 owned count rows refreshed.
Mutants: 31 logical kills documented below and in REPORT.md; final ten-mode rerun fails only at real semantic, leak or sanitizer oracles.
Not covered: overlapping lane 4b direct array-to-tuple cast remains blocked; exact reaching-view coverage and full gate are not claimed.

## Certified census

Started at f753567dd8837e90914e7c229df88d8536a3bf8b, building on the existing
fixed tuple certificates. Counts describe candidate production reads, not measured
whole-tsc execution coverage. The extra lane 4b obligations overlap this census.

| Original type/field | Pair | Reads | Certification commit |
| --- | --- | ---: | --- |
| 97934 /  | outSignature | 1 | f753567dd8837e90914e7c229df88d8536a3bf8b |
| 70161 /  | trackedSymbol | 1 | 6b58eb3c02d57ea341985643ae9528b5597115ac |
| 97926 /  | referencedMap | 1 | 0a6edef2d89f5241f7dbeb392b607fec9fc4acff |
| 97913 /  | root index | 1 | f1f6355adf5cfb4a48e766b9d575c0c7386dfe5f |
| 97898 /  | signature positions | 4 | 3c143841114922f672fad316938fa7ac383b9bd5 |
| 97931 /  | emitSignature forEach | 1 | 19418766cc313c0dc2d4fcb4fe26a7da248609e1 |
| 95604 / 1 | watch event | 2 | 21fd3587d6a32fef3f195a31be84605c02f58953 |
| 95604 / 0 | watch filename | 1 | cd2500f87ea9d532b029c5cac7348a6692342a93 |
| 68230 / 1 | module specifiers | 2 | f1f41adaa37b3173a9b085fee0999bc91fc609b3 |
| 68230 / 0 | module kind | 1 | 3c32b869f9a53a357e0de9074948f7ee194d52c9 |

Optional slots are checked present, absent and present-but-undefined against Node.
Rest forms cover zero, one and many tails, plus incompatible Map contracts.
Index ownership covers tuple, scalar, out-of-range and named missing-required
reads. forEach transfer covers tuple, scalar, an externally caught throw and a
callback retaining the element elsewhere. Finishing native witnesses use release
and sanitized builds with leak checks. Dispatch depends on the tuple plan carried
on IR; unrelated array and union consumers retain their existing paths.

The optional module worker certificate uses the complete original private return
declaration, extracted from pinned TypeScript 6.0.3 source by the owned script;
it does not substitute narrowed aliases. The source and extracted declaration
hashes are pinned by the oracle. The original 78 emitted declaration hashes stay
unchanged. No cohere implementation was copied. Minimal shared hooks are listed
in docs/checked-views-plan.md. There is one tuple constructor and one tuple read
path, including optional/rest and disjoint tuple alternatives.

## Verification

Commands ran after sourcing /workspace/adamic-tools/env.sh. Test output was
redirected to logs, never piped. Original declarations were supplied using
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCheckedViewTuple.*|TestCheckedViewMapStorageGaps)$' -count=1 -v -timeout 8m > /tmp/views-tuples-final-unit.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestCheckedViewArrays|TestCheckedViewNativeArrays|TestCheckedViewRankedArrayContracts|TestCheckedViewRankedArrayUnionContracts|TestCheckedViewRankedParserArrayContracts|TestCheckedViewObjectPrimitiveSource|TestCheckedViewMixedSelection|TestCheckedViewObjectUnions|TestCheckedViewPrimitiveArrayPairGap|TestCheckedViewMutableArrayUnionRefusal)$' -count=1 -v -timeout 6m > /tmp/views-tuples-final-lane-regressions.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1 -args -update-counts > /tmp/views-tuples-lane4b-out-signature-counts.log 2>&1
```

Results: PASS 122.144s (84 native cache misses, 190 uncached Node runs),
PASS 28.186s (221 uncached Node runs), counts PASS 54.120s (82 root tuple
fixtures plus two existing legacy Map rows). These suites precede the final
mutant-runner-only cleanup. Its mutation-disabled baseline passed 0.009s;
all affected mutations were rerun after removing fallback failures.

The final mutation rerun manifest below preserves each command's environment,
test function, actual failure status and log. Each command used go test
./internal/oracle -run '^TEST$' -count=1 with the listed mutation environment.

```json
[
  {
    "env": "ADAMIC_TUPLE_TRANSFER_MUTANT=skip",
    "test": "TestCheckedViewTupleForEachTransferMutants",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-1.log"
  },
  {
    "env": "ADAMIC_TUPLE_TRANSFER_MUTANT=double",
    "test": "TestCheckedViewTupleForEachTransferMutants",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-2.log"
  },
  {
    "env": "ADAMIC_TUPLE_OPTIONAL_CONTRACT_MUTANT=present",
    "test": "TestCheckedViewTupleOptionalAbsentMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-3.log"
  },
  {
    "env": "ADAMIC_TUPLE_REST_CONTRACT_MUTANT=present",
    "test": "TestCheckedViewTupleRestAbsentMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-4.log"
  },
  {
    "env": "ADAMIC_TUPLE_REST_CONTRACT_MUTANT=schema",
    "test": "TestCheckedViewTupleRestSchemaMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-5.log"
  },
  {
    "env": "ADAMIC_TUPLE_SIGNATURE_FOREACH_MUTANT=skip",
    "test": "TestCheckedViewTupleOriginalSignatureForEachMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-6.log"
  },
  {
    "env": "ADAMIC_TUPLE_WATCH_MUTANT=arity",
    "test": "TestCheckedViewTupleOriginalWatchArityMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-7.log"
  },
  {
    "env": "ADAMIC_TUPLE_WATCH_MUTANT=narrow",
    "test": "TestCheckedViewTupleWatchNarrowingMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-8.log"
  },
  {
    "env": "ADAMIC_TUPLE_MODULE_MUTANT=arity",
    "test": "TestCheckedViewTupleOriginalModuleArityMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-9.log"
  },
  {
    "env": "ADAMIC_TUPLE_LEGACY_MAP_MUTANT=certificate",
    "test": "TestCheckedViewTupleLegacyMapAdmissionMutant",
    "exit": 1,
    "oracle_caught": true,
    "log": "/tmp/views-tuples-final-mutant-10.log"
  }
]```

## Every mutant and its oracle

The chronological REPORT.md records exact historical commands and log paths.
Each group below counts distinct logical mutations; the legacy Map mode kills
two independent positive fixtures.

| Mutants | Count | Catch |
| --- | ---: | --- |
| Original outSignature: skip, shape, nested | 3 | Wrong admission violates named refusal / exit / stdout oracle |
| trackedSymbol: meaning, optional receiver guard | 2 | Wrong enum accepted; callback evaluates on absent receiver |
| referencedMap FileIdListId field | 1 | Boolean admitted instead of checked numeric field |
| Native and JavaScript tuple selector shape | 2 | Record accepted instead of named refusal |
| Index: presence; scalar treated as tuple | 2 | Missing slot silently returns undefined; valid scalar wrongly refused |
| Original signature: ID and value guards | 2 | Boolean accepted in numeric/string positions |
| Optional producer padding | 1 | Absent optional becomes present string |
| forEach: skip release; double release | 2 | 24-byte leak; sanitizer use-after-free |
| Arity range guard | 1 | Zero-slot tuple wrongly admitted |
| Optional admission: absent as present | 1 | Invented slot disagrees with Node |
| Rest: invented presence; late position guard; erased schema | 3 | Wrong value / missing named positional or Map refusal |
| Original signature forEach arity | 1 | Malformed tuple wrongly admitted |
| Watch event: guard, arity, narrowed demand | 3 | Boolean, undefined or NaN accepted instead of required refusal |
| Watch filename | 1 | Number accepted instead of string refusal |
| Module specifiers: slot guard, arity | 2 | Boolean or six-position tuple admitted |
| Module kind literal | 1 | Wrong string admitted |
| Legacy Map: erase optional and rest source certificates | 2 | Each valid Node object incorrectly refused |
| Lane 4b outSignature guard | 1 | Boolean accepted with exit zero instead of refusal |
| Total | 31 | All caught by semantic, leak or sanitizer evidence |

Early invalid mutation trials are not counted. No acceptance mutant is claimed
for a source program that cannot emit. The final runner cleanup ensures an
escaped mutation would pass its opt-in runner, exposing the missing kill.

## Remaining boundary and unrelated failures

Lane 4b outSignature is independently certified with exact string input, tuple
input and wrong-primitive control (PASS 2.919s; guard mutant FAIL 0.751s).
The direct-cast frontier from 8abb52a1518b9d8c3f0dbde993551d4f01ba10e2 is
preserved verbatim as frontiers/incremental-tuple-element-frontier.a. Node prints
string. Observed compiler refusal occurs at line 5 column 18 before the tuple
read: adamic/no-unchecked-cast. The pinned refusal test passes. This is an
additional blocked cast obligation overlapping four primary candidate reads;
it does not invalidate the checked-field positional certificates.

Inference from storage rules: admitting this cast needs an alias-preserving
bridge from a homogeneous mixed-scalar array to tuple object slots. A snapshot
could lose subsequent mutations and identity. Broadening mixed-array admission
would exceed the approved tuple-plan-only dispatch. No such conversion was
introduced. The blocked nested fixture has no runtime counts or native/leak
certificate. Spread/variadic rest, a production rest pair, full-tsc execution,
exact reaching-view coverage and a full gate are not claimed.

The user accepted refreshing owned counts only. Unrelated whole-count failures
belong to integration: ctor_set.a invalid free; census_overload_contracts.a
excess implementation arguments NotYet; census_small_boolean.a and
nbody_field_values.a template interpolation NotYet; maybe_number_slots.a native
argument representation error; require_perf_hooks.a Node.Text binding-pattern
panic. No full-count or full-gate success is claimed.

Toolchain setup used GOPROXY='https://proxy.golang.org|direct'. Cold retry timing
lines: Node 0.033s; Go 0.043s; submodules 0.088s; markdown 0.090s (skip 0.007s);
clang 0.258s; Go build 795.430s; test binaries deferred 795.533s; cache 795.535s;
done 795.618s. nproc=5, CPU quota=4. Initial TypeScript checkout was interrupted
(exit 143); a racing shallow fetch reported 'shallow file has changed since we
read it'. Sequential exact submodule fetch succeeded. Setup and subsequent
builds continued; no outstanding toolchain blocker remains.
