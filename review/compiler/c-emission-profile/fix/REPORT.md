Built: per-emitter dynamicProperties and fieldTypesNeeded caches computed together in one program walk.
Commits: baseline a75e71dd0fa2472680cb65cbbbd2f8f2dd803786; fix SHA is reported with the pushed branch.
Commands and outputs: 387 C files compared, diff exit 0 and 0 bytes; emission timings below; focused tests and go vet passed.
Mutants: wrong field-type key, wrong dynamic key, and stale source key each failed the intended emitted-C assertion; the field-type key also changed two oracle C files.
Not covered: newly provisioned cold-instance timings; measurements use fresh processes on the existing four-CPU instance; remaining readiness and record scans are unchanged.

## Change

The emitter owns a nullable objectMetadataNeeds value. Its first metadata query
walks the lowered program once, recording checked ObjectCall/DynamicProperty
presence, Property.View presence and CheckedFields. It derives fieldTypes as
those contracts or dynamic properties. All later queries read the two cached
booleans, including false answers. No cache is shared between emitters or keyed
by program/source identity. Lazy initialization also supports hand-built emitters.

Emission only changes borrowing facts in locals, not the expressions and contracts
this cache inspects. Each C call builds a new emitter, so an IR change between
calls is observed. No declaration ordering, shape keys or runtime code changed.
The new test emits the same hand-built program with a field contract absent,
added, and removed, requiring the first and third C strings to match exactly.
It also requires checked fields alone to avoid dynamic shape registration.

## Byte comparison

Baseline is the pushed profiling tip a75e71dd, built before these compiler edits.
Both binaries use validation-probe.go.txt copied into a temporary scratch Go
package inside the repository. It calls load.Load, lower.Lower and native.C
directly. prepare-validation.py extracts every entry in oracle_test.go's
fixtures declaration: **384 entries, all 384 emitting C**. Each input is loaded
and lowered afresh by each compiler; output names and absolute source paths
are identical. No oracle observation or build-product cache is involved.

The three additional inputs follow their actual test recipes:

- sf-yepesta: yepesta_build_test.go's JSX textnode mutant. Copy the lint port,
  rewrite imports leaving the port to the original absolute targets as copyPort
  does, apply rules/react-jsx-no-comment-textnodes/mutant.json's single mutation,
  regenerate the registry, then load the copied main.ts. Witnesses are copied too.
- g-lint compiler agreement: compiler_stage1_split_test.go's unmodified absolute
  stage1/cohere/lint/main.ts, with its generated registry.
- estree scalar edges: scalar_edges_preparation_test.go's absolute
  stage1/cohere/estree/main.ts.

No code was copied from cohere. The copy used for sf-yepesta contains Adamic's
own port; external imports still reference the existing source/submodule.

Exact executed comparison:

```sh
diff -ru --exclude='*.pb.gz' /tmp/c-emission-validation/before /tmp/c-emission-validation/after > /tmp/c-emission-validation/emitted-c.diff
```

**Exit 0; the diff is empty (0 bytes).** The only files other than C in those
directories are profiler outputs, deliberately excluded because profiles encode
measurements. Both directories contain exactly 387 C files. emitted-c.diff is
preserved here; c-sha256.json records every C file's size and SHA-256 on both sides,
plus the mutant side. The oracle input manifest is ../oracle-manifest.json and
the three port manifests are here. Full generated C remains under the named
scratch directories rather than expanding this evidence commit.

## Emission time

Each port is run by a separate new process, with GOMAXPROCS=4 and no concurrent
builds/tests during the measurement. Each call is the first native.C call in
that process; all emitter and compiler observation caches start empty. The clock
covers only native.C, after lowering. CPU profiling brackets that call only;
profile startup/shutdown and output writes are outside the wall timer.

| Program | Before native.C | After native.C | Saving | C bytes |
|---|---:|---:|---:|---:|
| sf-yepesta | 62.863882907s | 11.356153963s | 81.9% | 6,405,621 |
| g-lint compiler agreement | 66.079272263s | 10.722321147s | 83.8% | 6,405,618 |
| estree scalar edges | 20.272838215s | 3.127246540s | 84.6% | 4,057,785 |

The attached instance is the previously provisioned instance, not a new cold
machine. Go builds and filesystem state are warm. Thus these are cold emission
cache/fresh-process timings on a four-CPU quota, **not newly provisioned
cold-instance measurements**. No tool available in this turn provisions an
additional instance; that part of the requested measurement is not claimed.
The native.C step itself uses the already-lowered IR and does not perform file IO.

nproc=5, cpu.max=400000 100000, Go 1.27.1, clang 20.1.8, Node 24.19.0.
The original setup.log preserves setup timings and the tool environment is
/workspace/adamic-tools/env.sh. CPU profiles before-*.pb.gz and after-*.pb.gz
are preserved. after-pprof.txt attributes the remaining g-lint scan cost to
fieldReadinessNeeded (5.07 CPU seconds) and hasRecordStorage (4.75 seconds).
The requested two queries no longer appear in its top 20.

## Checks and mutants

Commands run from the repository root with the tool environment sourced; all
test output went directly to log files:

```sh
go test ./internal/native -run '^TestObjectMetadataCacheIsPerEmission$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(library_object_keys|narrowed_fields)\.a$' -count=1 -v
go vet ./internal/native
```

The new top-level parallel test passed in 0.00s, package command 0.007s.
The two runtime oracle leaves passed in 13.05s and 12.99s (whole command 13.202s):
source Node, backend Node, native, release native, ASan/UBSan/LeakSanitizer.
Vet exited 0 without diagnostics. No whole package test or full gate was run.
No new oracle fixture was added, so counts.md needs no refresh; byte-identical
C also preserves all existing runtime counts.

Mutations were applied through Go build overlays, never the working compiler:

| Mutant | C change and catcher |
|---|---|
| fieldTypesNeeded reads dynamic, a wrong query key | drops field-type writes; new test fails at the added-contract comparison; all-fixture C diff exits 1 on casts.a and cast_fails.a |
| dynamicProperties reads fieldTypes, the other wrong query key | adds dynamic registration for a checked-only field; new test fails at the registration assertion |
| fieldTypes persisted in a source-keyed global cache | retains true after the same IR loses its contract; new test fails because the third C differs from the first |

All three mutant test commands exited 1 with their intended assertion (not a
compiler/clang diagnostic). Patch sources are ../wrong-query-key.patch,
../wrong-dynamic-key.patch, ../stale-source-key.patch. mutant-c.diff records the
first mutant's independent 384-fixture comparison. Individual logs are preserved.
The final unmutated test was rerun and passed. This proves each assertion in the
new test can fail, including the stale-cache byte comparison requested here.

## Reproduction and delivery

Copy ../validation-probe.go.txt to .scratch-c-emission/main.go, run
../prepare-validation.py from the repository root, and build the scratch command.
The actual commands for each captured side are:

```sh
go build -o /tmp/c-emission-before ./.scratch-c-emission
GOMAXPROCS=4 /tmp/c-emission-before /tmp/c-emission-validation/oracle.json /tmp/c-emission-validation/before
GOMAXPROCS=4 /tmp/c-emission-before /tmp/c-emission-validation/port-sf-yepesta.json /tmp/c-emission-validation/before
GOMAXPROCS=4 /tmp/c-emission-before /tmp/c-emission-validation/port-g-lint.json /tmp/c-emission-validation/before
GOMAXPROCS=4 /tmp/c-emission-before /tmp/c-emission-validation/port-estree-scalar-edges.json /tmp/c-emission-validation/before
# After the compiler edit, use /tmp/c-emission-after and the after output directory.
```

For reproduction after this fix is checked out, build the baseline with a Go
overlay replacing internal/native/emit.go and emit_objects.go with their
`git show a75e71dd:<path>` contents. Keep the same probe and input paths for both
binaries. Mutant commands use `go test -overlay=<json> ...` and the same source
replacement mechanism with the named patches.

Lane-check output is preserved after committing the complete change. This lands
the measured metadata cache improvement toward C emission unit #3he8f8g. The
remaining whole-program scans and newly provisioned cold-instance measurements
are the explicit limits of this delivery.
