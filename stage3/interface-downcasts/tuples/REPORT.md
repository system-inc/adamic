Built: original EmitSignature field views using lane 1's single tuple certificate path and lane 2 array read hooks.
Commits: territory b3bc7138; lane 1 merge 13486c7c (includes 65d8a138).
Checks: original pair oracle PASS 11.507s; scoped IR/lower/native/JavaScript PASS 0.011s/4.593s/39.285s/0.875s.
Mutants: skip root, accept record, drop helper position each fail the pinned refusal oracle; none is killed by clang.
Remaining: 9 candidate pairs / 14 candidate reads after this push; optional/rest Map forms are additional fixture obligations.

The combined candidate queue is 10 pairs / 15 reads. Lane 2 supplies 6/9;
lane 1's nullish table supplies four optional-position receiver pairs / six
reads. No rest receiver is identifiable in that table. The two optional/rest
Map gap fixtures are tracked separately from production counts. Exact reaching
view coverage is not measured. This push covers pair 97934, outSignature, one
candidate read, against unreduced declarations from upstream commit
050880ce59e30b356b686bd3144efe24f875ebc8.

Preparation reuses lane 4b's original declaration emitter, verifies all nine
lane 2 read sites, and hashes all 78 generated declaration files. Tests require
the original IncrementalBundleEmitBuildInfo field set; no reduced interface is
substituted. All twelve fixtures run their unchanged source on Node, generated
JavaScript, native release and native ASan/UBSan. Successful native runs also
pass the shared leak checker. Undefined is admitted; the required field being
missing refuses. Wrong primitive, tuple length, record identity and nested
position diagnostics are pinned literally in the oracle.

Lane 1's constructor is generalized only to accept the existing recursive child
builders. Map producers still require complete descriptors and reject callable
tuple descendants. The array adapter retains lazy descendants. Empty tuples
have an explicit layout marker; Map schemas distinguish them from ordinary
records. Native producers carry tuple identity independently of casts, and
ordinary object copies do not acquire it. These hooks are listed in the plan.

Reproduction (source /workspace/adamic-tools/env.sh first):

```sh
node stage3/interface-downcasts/tuples/prepare.cjs /tmp/untagged-typescript /tmp/views-tuples-original-declarations > /tmp/views-tuples-prepare.log 2>&1
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalOutSignature$' -count=1 -v > /tmp/views-tuples-pair1-final.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'View|Map|Contract|Tuple' -count=1 -timeout 10m > /tmp/views-tuples-packages-handoff.log 2>&1
```

With the same original-declaration environment, each mutant is a separate
`go test ./internal/oracle -count=1 -v` run. `ADAMIC_TUPLE_ORIGINAL_MUTANT=skip`
on `^TestCheckedViewTupleOriginalOutSignature/emit-wrong-kind-noread$` prints
boolean and exits 0 instead of the named exit-70 refusal. `shape` on
`.../emit-record-noread$` admits an ordinary record and exits 0. `nested` on
`.../emit-helper-wrong$` returns the wrong position value and exits 0. Logs:
/tmp/views-tuples-mutant-skip-noread.log,
/tmp/views-tuples-mutant-shape-handoff.log,
/tmp/views-tuples-mutant-nested-handoff.log. IR mutations do not alter production
sources. The earlier skip probe was masked by secondary narrowing and was
replaced by the consumer that does not narrow.

Setup: GOPROXY=https://proxy.golang.org|direct; submodules 0.077s; markdown-width
skipped (ready 0.090s); clang ready 0.182s; Go build 37.033s; deferred test
binaries 37.214s; cache ready 37.216s; done 37.247s. nproc=5, cgroup quota=4.
Go 1.27.1, clang 20.1.8, Node 24.19.0. No complete repository gate is claimed.

Target for the optional/rest adapters is October 8. A completion date for the
whole original-declaration family is contingent on certifying original branded
numeric IDs: the highest-ranked tuple uses IncrementalBuildInfoFileId, whose
original required brand is any. The unchanged original read currently reports
NotYet for that numeric intersection. No number alias or void-brand replacement
will be counted as certification of that original declaration.
