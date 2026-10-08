Built: first share b checkpoint, 20 original-member callable pairs / 85 candidate reads, 90 isolated .a fixtures and scoped oracle/counts/mutant tests.
Commits: based only on codex/views-callables 926a1d39d1a0d6b4cf49bbccf521a4cc02d10f56; first checkpoint 473904f4, followed by the verifier/log correction accompanying this report.
Commands: original fixture verification PASS; restored scoped oracle PASS 56.989s; scoped count refresh PASS 10.582s; oracle vet and diff check PASS; global counts refresh has 60 inherited failed fixture rows.
Mutants: all 20 per-pair expected-arity mutations fail the pinned exit/message assertions; independent release-native and JavaScript controls execute the same malformed producers successfully.
Uncovered: full original tsc call graphs and exact reaching-view counts; 864 initially uncertified share pairs / 2772 candidate reads remain, including the skipped families and four observed code dependencies below.

The ranked ledger is unknown-callable-pairs-ranked.json. Only previously
uncertified ranks with r modulo 3 equal to 1 are credited. The certified ranks
are 331, 334, 349, 352, 355, 358, 361, 364, 367, 376, 379, 394, 397, 400,
412, 481, 622, 634, 748 and 778. certified-pairs.json contains each original
receiver type ID, member and read count; it is an additive share ledger, not a
rewrite of the other workers' progress. Baseline 156 pairs / 2699 reads plus
this share gives 176 pairs / 2784 reads before other shares are integrated.

Each fixture retains the complete original member declaration and original
read expression. Adjacent carriers and producer bodies are reduced, as in the
lane's existing aggregate and later-ranked fixtures. These are member-contract
certificates and candidate inventory credits; they do not certify complete
original payload interfaces, production reachability or whole-source compilation.
The original TypeScript source pin is 050880ce59e30b356b686bd3144efe24f875ebc8,
prepared independently of cohere. Source/read spans, source hashes and original
inheritance paths are in original-witnesses.json and families.json. No cohere
code was copied. No compiler/runtime source was edited or other branch merged.

Good executions match original-source Node in release native, ASan/UBSan native
and emitted JavaScript; finishing native executions also pass leak checks.
Negatives cover non-callables, incorrect arity, incompatible result
representations where a result exists, and incompatible parameter
representations where parameters exist. Diagnostics pin the complete expected
signature from the original ranked ledger, expression, found category and exit
70. No partial stdout is admitted for these negatives.

Each rank has an independent arity mutant. It changes only that read's lowered
expected parameter descriptor to accept the malformed producer's actual arity;
the original fixture and immutable producer metadata remain unchanged. The
original negative oracle then fails on valid native execution with exit 0.
The separate mutation controls independently show release native and JavaScript
matching source Node's successful wrong execution, proving the same pin would
catch both. The 20 concrete ranks above identify every mutant. No clang error,
sanitizer fault or unavailable runtime package counts as a mutant kill.
run-mutants.py reproduces these controls and failures without editing production
files; all mutations are confined to one test's in-memory IR.

Observed code dependencies, not certification credit:

| Rank | Pair / candidate reads | Original read | Observed stop | Required owner work |
| --- | --- | --- | --- | --- |
| 1 | typeof Debug.assert / 439 | Debug.assert | existing original-signature probe refuses adamic/no-type-predicate at its assertion signature | admit and prove the original assertion-predicate callable contract without trusting an opaque assertion |
| 4 | NodeFactory.createAssignment / 143 | factory.createAssignment | read refuses unsupported callable contract | certify overloaded callable signatures and preserve their argument/result obligations |
| 10 | string[].push / 94 | keys.push | read refuses unsupported callable contract | certify the rest-parameter intrinsic callable contract and actual array receiver |
| 37 | string[].join / 42 | result.join | lowering succeeds; sanitized native build fails passing adamic_array * to adamic_object *; JavaScript exits 70, found function with unknown signature | array-to-view receiver conversion plus immutable intrinsic signature metadata |

The three new blocker declarations/read spans are independently verified in
blocked-original-witnesses.json and verify.cjs. Rank 1 reuses the lane's existing
original assertion-signature probe. No native execution or leak pass is claimed
for the rank 37 build refusal. The failed direct Node invocation of CLI-generated
JavaScript lacked the adamic loader; it is not evidence. The oracle's configured
JavaScript runner supplies the actual pinned signature refusal.
Set-intrinsic, optional-host and bind/condition families are excluded from credit.
Other untested ranks remain untested, rather than being declared blocked from
signature appearance alone. This is an initial fixed-signature checkpoint;
the higher-ranked generic, predicate, overload and intrinsic frontier is not
claimed complete.

All test output is captured directly in logs. Commands after sourcing
/workspace/adamic-tools/env.sh:

```sh
node stage3/interface-downcasts/lane5/share-b/prepare.cjs /tmp/lane5-b-original 61,331,334,349,352,355,358,361,364,367,376,379,394,397,400,412,427,481,622,634,748,778
node stage3/interface-downcasts/lane5/share-b/generate.cjs
node stage3/interface-downcasts/lane5/share-b/prepare.cjs /tmp/lane5-b-original 4,10,37 blocked-original-witnesses.json
node stage3/interface-downcasts/lane5/share-b/verify.cjs /tmp/lane5-b-original
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareBArityMutants$' -count=1 -v -timeout 5m
ADAMIC_GATE_UNCACHED=1 ADAMIC_CALLABLE_SHARE_B_MUTANT=arity go test ./internal/oracle -run '^TestCheckedViewCallableShareBFamilies$/^rank-/^wrong-arity$' -count=1 -v -timeout 5m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewCallableShareBCounts$' -count=1 -timeout 5m -args -update-counts
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareB' -count=1 -v -timeout 10m
go vet ./internal/oracle
git diff --check
```

The required global counts command failed in 41.064s in 60 unchanged fixtures;
its complete log is retained. The scoped updater measures and appends 90 rows,
then the final restored test checks them. A row-key comparison proves zero
existing rows changed. The full package tests and full gate were not run.

Setup passed with GOPROXY=https://proxy.golang.org|direct: Go ready 0.288s,
Node ready 0.297s, clang ready 0.886s, markdown ready 2.163s, submodules ready
490.545s, go build ready 784.336s, cache warm 784.462s, done 784.490s.
nproc=5; cpu.max=400000 100000. No setup failure; the initial default fetch
updated only main, so the named lane ref was fetched explicitly. An initial
unsourced gofmt command was unavailable and was corrected after sourcing the
printed environment. Initial blocker tests used relative Node paths and were
corrected to absolute paths before the passing recorded run.

Evidence correction: 473904f4 contained a syntax error in the newly extended
blocker section of verify.cjs. The earlier 20-family verifier and final runtime
oracle had passed; the syntax error was found during post-push evidence review
and is corrected here. The restored verifier now checks all 90 member fixtures
and the three new original blocker declarations. A mutation narrowing the
rank-334 original parameter declaration from string to the literal "value3"
fails the original-declaration pin; the fixture is restored. The reproducible
run-mutants.py command also passes, catching all 20 arity mutations. No runtime
or compiler source changes are involved. Raw logs are explicitly tracked here
despite the repository's general logs ignore rule.
