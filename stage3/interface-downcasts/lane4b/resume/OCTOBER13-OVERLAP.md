Built: adopted seven independently held original intersection certificates, 45 candidate reads, after fresh Node/backend/leak/mutant validation.
Commits: integration-only group 9cfb94ba pushed; this certificate and count group follows it on codex/views-object-primitive-unions.
Checks: owner controls PASS 95.513s; exact source/declaration overlap audit PASS; five scoped count rows updated PASS 6.632s and verified PASS 6.594s.
Mutants: all 28 inherited pair mutants and one ancestor-pos omission caught in sanitized native, release native and JavaScript; no compile failures or crashes counted.
Uncovered: nine pairs / 23 reads remain blocked; unique certified lane inventory is 33 pairs / 158 reads; whole-tsc reachability remains unmeasured.

This is certificate reuse with exact provenance, not duplicate global completion.
The adoption audit matches every original (file, start, end, text) read site and
read count, requires certified owner status and identical declaration hashes,
and pins the complete shared certificate file hash. The seven rows reference
../../lane7/original/certification.json rather than inventing reduced targets.
All original sources remain pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.

| Pair | Reads | Independently caught owner mutants |
| --- | ---: | --- |
| 9474.expression | 11 | expression skip, wrong kind, nested symbol omission |
| 7642.parent | 10 | parent bounded skip, wrong kind, nested parameters omission |
| 9476.left | 9 | left bounded skip, wrong kind, nested symbol omission |
| 9485.left | 9 | access bounded skip, wrong kind, nested symbol omission |
| 9475.expression | 4 | expression skip, wrong kind, nested symbol omission |
| 7644.parent | 1 | absorbed ancestor pos obligation omission |
| 36241.parent | 1 | optional root bounded skip, wrong kind, nested symbol omission |

These seven certificates have 19 owner mutants, each independently caught in
three execution modes. The existing runner also reproves nine mutants for other
owner rows sharing the harness, for 28 runner mutants / 84 catches, plus the
separate absorption mutant's three successful true counterfactuals. The report
does not credit those other rows to lane 4b. Positive owner controls include
complete original field-set checks and separate leak checks. Negative controls
pin named exit-70 reads. The four reduced legacy probes add four demanded-path
omission mutants; their 12 catches are documented in OCTOBER13-INTEGRATION.md.

Five new measured labels in original/batch-counts.json cover all seven original
rows by reusing the existing owner fixtures and the owner's existing generated
union-parent control. No new standalone .a source is added, so there is no new
self-contained registry row or general counts refresh. Ordinary verification
requires the golden rows to match; all five positives balance allocations
against frees plus region reclamation. Their certificate behavior is held by
the existing original oracles, not duplicated by the count test.

Commands, tool environment sourced and output directly redirected to logs:

```
node stage3/interface-downcasts/lane7/original/prepare.cjs /tmp/lane4b-upstream /tmp/lane4b-declarations
ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewIntersectionOriginal(Bindable|Nodes|Absorption|ArrayBlockers)$' -count=1 -v -timeout 15m
ADAMIC_GATE_UNCACHED=1 python3 stage3/interface-downcasts/lane7/original/run-pair-mutants.py /tmp/lane4b-declarations /tmp/lane4b-overlap-mutants
ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_INTERSECTION_ABSORPTION_MUTANT=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewIntersectionOriginalAbsorption$/^wrong$' -count=1 -v -timeout 5m
ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveOverlapCounts$' -count=1 -v -timeout 5m -args -update-counts
ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveOverlapCounts$' -count=1 -v -timeout 5m
python3 stage3/interface-downcasts/lane4b/resume/adopt-intersection-certificates.py --update
python3 stage3/interface-downcasts/lane4b/resume/adopt-intersection-certificates.py
python3 stage3/interface-downcasts/lane4b/resume/rank-lazy-demand.py
```

The absorption mutant command must fail; its log contains exactly three exit-0,
empty-stderr true counterfactuals caught by the exact named guard. The runner
validates the same semantic criterion for its 28 commands. Individual mutant
logs are preserved under resume/logs/overlap-mutants/. No production source was
mutated and no complete package or repository gate was run.

The remaining ranked frontier mapping is in remaining-blockers.json. Every old
probe was rechecked on ffe428ab, including source Node before strict loading.
CompilerOptions and BuildOptions still refuse cast admission (two pairs / 11
reads); a tuple-element cast refuses (one / four); EmitHelper.text's callable
union field remains unsupported (one / three); branded tuple arrays refuse
storage (two / two); outSignature reaches the named tuple union refusal (one /
one); fileInfo forEach consumers refuse mixed array storage (two / two).
Receiver-bearing calls are used for all forEach probes. Owner array controls
also cover real original FileInfo objects and string members. These are compile
blockers, with no backend program, leak run, or runtime certificate to claim.
The separate template interpolation of a comment object/array union also
remains unsupported, even though its field contracts are certified.

No time cutoff caused this handoff. The canceled compiler-area merge contributes
no changes. Further certification needs the listed owner storage/admission or
callable changes to reach views-integration; no refusal was weakened locally.
