# 99h evidence

Runtime branch: typescript/axolotl-99h, commit 0ee6e4d4c9012137944f3f15f2644096adff76c1. This directory is intentionally outside stage3/adapt: logs, configurations and proof scripts must not participate in apply's adaptation input key. The runtime adaptation directory holds only adapt.cjs and README.md.

## Composition and checker phase

Main is 470e16f8315f7a83c5c7b7f6f2a77d8328db0a25. The before composition adds the exact adaptation directories from 99b a299eddff730973c769df5b13ff7d9fd747435a9 and 99g debc7962f9b3ea545ac9818cbe1a93c88b020f2e. The after composition adds 99h. No 99a is included, no lane manifest is changed, and neither prerequisite directory is committed on the runtime 99h branch.

Both full apply.sh runs pass. Both whole-program checks use 79 sorted .ts/.a roots recursively under src/compiler, the same adamic binary, the unchanged strict checker options and prelude, and no overlay. They run before installing the upstream tree's npm dependencies, matching the classification workflow. There are 112 diagnostics before and 111 after. The complete normalized diagnostic multisets differ only by removal of:

```text
src/compiler/commandLineParser.ts:2958:13: error TS2322: Type 'string | undefined' is not assignable to type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

There are zero additions anywhere. All codes, columns, messages, multiplicities and elaboration chains participate; only the before/after tree root paths are normalized. No coordinate offset is needed because 99h adds no lines.

The inherited TS2307 for source-map-support is included in both fresh-tree counts. npm ci subsequently supplies this module for the upstream builds. The earlier 99g count of 112 was taken after npm ci, which explains why subtracting 99b from that earlier number would predict the wrong fresh-tree count. An initial auxiliary check on main plus 99g, without 99b and before npm ci, measured 113 to 112; it is not the requested primary composition and is not substituted for it.

## Local proofs

[verify.cjs](verify.cjs) proves all five exact source regions, equality of the before composition's 99g-after snapshot and the after composition's 99h-before snapshot, the exact three JavaScript text changes, unchanged cast/non-null assertion counts, byte-identical public API declarations, zero second-run edits and unchanged file mtime. Eight invalid fixtures fail without writes: missing 99g, one header only, templates only, array only, a wrong return type, an extra scalar call, a non-null assertion, and a callback coercion that would change custom map behavior.

JavaScript changes from 162084 bytes to 162108 bytes under stock ts.transpileModule (ES2024, ESNext, retained comments). The only JavaScript differences are the two config-template String wrappers and `String(value.map(v => formatSingleValue(v, map)).join(", "))` at the array interpolation. The two declared optional return types erase completely. Callback results are unchanged, including when a custom map filters undefined values. The preexisting 43 casts and 25 non-null expressions are unchanged; none is introduced.

[witness.cjs](witness.cjs) exercises the actual built before and after libraries. All complete generated files compare byte for byte. It covers:

- ordinary --init and --init --strict false;
- --init with outFile, strict, target, rootDir and lib cleared with null;
- the scalar paths object whose toJSON returns undefined;
- plugin array elements whose toJSON returns undefined, including two such elements;
- array holes, explicit undefined elements, and a custom map that filters undefined callback results;
- null, empty string, zero, empty arrays and inherited properties;
- identical thrown toJSON and invalid-enum errors.

Template interpolation formerly rendered undefined as "undefined", so its consumer now uses String directly. Array.join renders undefined elements as empty fields, so the callback must retain its optional result. String wraps the joined text, preserving both the native join rule and custom map methods that observe callback results. The paths witness still writes `"paths": undefined,`. A single plugin hook still writes `"plugins": [],`, and two hooks still write `"plugins": [, ],`. The explicit array-undefined case remains `"types": [, "x", ],`.

All requested ordinary and cleared scalar readbacks exit 0 before and after. The extra cleared-lib readback case reveals an existing TypeError in serializeOptionBaseObject, which tries to map the undefined lib value. This first stopped the witness's overly strict assumption that every readback exits 0. The final witness retains the case, verifies identical config bytes and the same before/after exception, and does not repair or hide it. The original failing witness logs and directory are retained under /workspace/scratch/99h/witness-initial.log and witnesses-initial. The core adapter was not changed in response to that witness assumption.

## Sequential oracle

[prove.py](prove.py) first checks the fresh trees, then runs npm ci and npm run build serially for both trees, followed by the local witnesses and verification. Both complete unchanged oracles then run serially with --workers 4 --limit-seconds 1800, all runners, no test filter, and a 2400-second hard command limit. Source, tests, baselines, expectations, API sanctions and oracle implementation are untouched.

Clean cached workspace pages are evicted through read-only descriptors before each oracle, as in the accepted 99g proof. No file bytes, cgroup settings, Node flags, tests or worker counts change. All memory counters are compared before and after. The only permitted failure is the existing public API acknowledgment failure; its normalized 223 sanctioned declarations must match exactly. Both full baseline.diff byte strings and actual public declaration bytes must match. A baseline move or any other gate failure stops the run.

The initial local proof run completed both builds and checker measurements, then stopped on the extra lib readback assumption. The final witness and scope checks were run on the same built artifacts, with no runtime source changes. `prove.py --oracles-only` then reused those completed local results for the sequential oracles. commands.json records the stopped and successful steps; no checker or runtime source was changed to obtain the final result.

## Reproduction

Use a disposable clone of the runtime branch and fresh scratch output directories. Install stage3/api's locked dependencies and supply the pinned TypeScript checkout through ADAMIC_TYPESCRIPT_SOURCE. Extract both prerequisite adaptation directories by the exact commits above. Temporarily move the 99h directory outside stage3/adapt while building the before product, restore it, then build the after product. The prerequisite extraction and temporary staging are proof preparation only; they are not commits on the runtime branch.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream
# With main + 99b + 99g, before adding 99h:
bash stage3/apply.sh /workspace/scratch/99h/before > /workspace/scratch/99h/apply-before.log 2>&1
# Restore/add the 99h adaptation directory:
bash stage3/apply.sh /workspace/scratch/99h/after > /workspace/scratch/99h/apply-after.log 2>&1
go build -o /workspace/scratch/99h/adamic ./cmd/adamic > /workspace/scratch/99h/build-adamic.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/evidence/99-correctness-axolotl-h/prove.py /workspace/scratch/99h > /workspace/scratch/99h/prove.log 2>&1
```

The exact prerequisite extraction is a git archive of each pinned commit's stage3/adapt/99-correctness-axolotl-b or stage3/adapt/99-correctness-axolotl-g directory, unpacked only into a disposable checkout. A normal full run with fresh outputs performs both fresh-tree checks before npm ci. The --oracles-only path exists solely to resume after already completed local proofs; it asserts the saved witness and verification success.

## Results

See proofs/measurement.json, proofs/verification.json, proofs/witnesses.json and proofs/oracle-comparison.json for the machine-readable results. Raw phase logs and oracle reports are retained under logs and oracle-before/oracle-after; complete witness configs and readback logs are retained under witnesses. The runtime branch has no generated proof files.

## Array consumer refinement

The first candidate converted each array callback result with String(result ?? ""). Its ordinary-array outputs were identical, but a custom map can observe the difference between undefined and an empty string. The final candidate leaves callback results unchanged and instead wraps the joined text with String. The custom-map witness detects the old candidate and passes on the final implementation.

The first candidate's complete oracle results are retained separately. Its control tree never changed and was reused for the final comparison. The final after tree was rebuilt through full apply, checked while fresh against the saved unchanged fresh-control diagnostic strings, built again, and rerun through all local witnesses, verification and its own full oracle. The --after-only proof mode reuses only that unchanged, already validated control artifact and result; it still performs every after-side build, witness, scope, checker, API, baseline and memory gate. No baseline expectation is changed.

## Final measured result

The final after oracle passes all gates against the unchanged validated control. Each has 106366 passing, one sanctioned public API failure, and zero pending. There are zero baseline moves. Complete baseline.diff bytes match at 51123 bytes, SHA-256 6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031. Both public declaration artifacts are byte-identical, and every memory-event counter stayed unchanged during the compared suites. All 223 sanctioned API declarations match exactly.

The final local result is seven CLI cases and eleven API cases with byte-identical complete generated files, plus unchanged thrown-error cases. The custom-map check is red on the superseded callback-coercion candidate and green on the final implementation. The final helper-script run also corrected a misplaced assertion in the witness factory; that ReferenceError and its logs are retained, and the fully corrected witness, scope check and final oracle all pass. No runtime source change was made for that witness-only correction.

Fresh-tree whole-program diagnostics are 112 to 111, with only D036 removed and no addition anywhere. The second adapter run makes zero edits and no write. Generated evidence stays outside the runtime adaptation input key. Raw logs and diffs are gzip-compressed without changing their uncompressed bytes; witness configurations are retained verbatim.
