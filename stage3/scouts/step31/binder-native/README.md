Built: a Node-held native binder observer candidate, both split-mode attempts, fifteen ordered stop witnesses, and replayable throwing discovery evidence.
Commits: main 45487a809f89885a3fc651cd590e7dabf31362dc; compiler candidate 0c99fc545e0f49d8407467321355fd45e2934c21; scout 2de9fc1b; delivery SHA is in the handoff.
Results: split 0 and 1 fail before C emission, first commandLineParser.ts:1576:96 TS2345; source Node matches stock on 301 projects, 3,475,070 bytes.
Mutants: fifteen witness source changes, fifteen throwing-placeholder erasures, three independent process/output changes, one observer flag change and five ledger/manifest changes are caught.
Limits: no native binder binary, native acceptance difference, allocation/leak measurement, or full gate; five walk stops depend on placeholders rather than unchanged source.

# Native result and compiler handoff

Area-next is **not** on the fetched main. The compiler measurement therefore uses a detached, never-pushed scratch of `origin/cloud/land-area-next-auto-4885cec5`, SHA **0c99fc545e0f49d8407467321355fd45e2934c21**. No compiler code or compiler options were changed. The pinned cohere **7945d102a6c18dd36adf9114a758ce646e8b2359** and its TypeScript module **d92d9bfee114c80be2c375d72edae966176e3a4f** match the candidate; their existing checkout was reused via a scratch symlink.

The main.a a-check header records its unstaged repository-relative import error (TS2307); the native attempts stage it at the gathered tree root and record their actual dependency diagnostics separately. The native candidate [main.a](main.a) implements the scout's `adamic-binder-v1` observer with named compiler imports and Adamic pathname/file input. It parses, binds with converted options, preserves node/symbol identity, includes attached JSDoc, emits ordered locals/exports/members, and serializes both diagnostic families. It does not call the checker. The pinned scout contained a Node `.cjs` observer, not a native `.a` driver; this candidate is new. Node validation does **not** prove its casts, JSON request boundary, graph ownership, or object projections admissible in Adamic. Those obligations remain visible to the compiler; no proof or source gate was bypassed.

Following the scanner README, both builds ran from the actual compiler candidate:

```
ADAMIC_NATIVE_SPLIT=0 /workspace/cache/step31-binder-area-adamic build /workspace/cache/step31-binder-driver-slice-v2/main.a -o /workspace/cache/step31-binder-split0
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 /workspace/cache/step31-binder-area-adamic build /workspace/cache/step31-binder-driver-slice-v2/main.a -o /workspace/cache/step31-binder-split1
```

Both exit **1**, with identical diagnostics:

- `commandLineParser.ts:1576:96 TS2345`: `string | undefined` passed to `string`, through `arrayFrom(opt.type.keys())` and the deprecated-key filter.
- `debug.ts:113:19 TS2339`: missing `ErrorConstructor.captureStackTrace`.
- `debug.ts:114:19 TS2339`: the second use of the same missing member. One throwing replacement for `Debug.fail` removes both diagnostics; they are not misreported as two independent repairs.

These locations are in the audited gathered source. No C, clang build, native output or native timing exists. The two split settings therefore exercise the same frontend admission boundary; they do not establish split compilation correctness.

The bounded walk exposes **module initialization ordering before binder body lowering**. Orders 3-9 are enum-initialization obligations, including Map construction, a host call and timestamp initialization. Order 13 is the original performance-hook call blocked by namespace ordering. Do not infer that these are unsupported Map operations or retired binder casts. The minimal programs deliberately reproduce the relevant initialization order.

Owner assignments in [STOPS.md](STOPS.md) and [stops.json](stops.json): **9 compiler**, **1 library**, **5 adaptation**. The library owner is the missing Error declaration; implementing its native host behavior is separate runtime work. The adaptation owner denotes obligations introduced by the throwing discovery replacement, not recommended production source changes.

# Fifteen-stop walk and its limits

[STOPS.md](STOPS.md) gives every file, one-based line/column, exact message, owner, scanner-list comparison and minimal `.a` witness. The walk starts from a copy of the original audited slice. Each build runs normal checking/lowering, records the actual first stop, then replaces the smallest enclosing body or initializer with a throw. Original inferred result/initializer types are retained where needed. No declaration is given a successful fake implementation.

Every replacement is present in the **uncommitted cache copy**, never the delivered compiler or an acceptance slot. The evidence contains each exact replacement, the diagnosed source **before** that replacement (`stop-NN.source.gz`), and an exact per-step text diff (`patch-NN.diff.gz`). The combined [discovery.patch.gz](evidence/discovery.patch.gz) is evidence only. Source snapshots preserve measured bytes; text diffs normalize CRLF through Python's text reader. Coordinates refer to the progressive copy, not stock TypeScript line numbers.

Orders **10, 11 and 15** concern calls introduced by throwing initializer IIFEs while runtime namespaces remain pending. Their direct-throw replacements retain typed declarations. Order **12**, an unassigned enumMemberCache read, follows one such replacement. Order **14**, optional nativePerformanceTime in a callback, follows changing the performance helper body. These five are explicitly marked **adaptation**, with their dependency qualification. They must not be added to an unchanged-source blocker census. The walk stops after fifteen recorded sites; the fifteenth replacement is retained but no sixteenth stop is claimed.

The initial walk attempts were discarded: `throw new Error` inserted constructor calls into the enum-initialization path, and inferred return types changed when bodies were replaced. The final fresh walk uses primitive throws, preserves inferred types, and rejects unchanged repeated stops. Only its fifteen snapshots/diffs constitute the reported sequence.

For scanner overlap, the task id `#whkxbc7` is not present in the repository and cannot identify a different external ledger. The concrete comparison here is main's latest available fifteen-stop artifact, `stage3/drivers/scanner/evidence/combined-records-library/stops.json`, introduced by **8822db17657db9b75e0e81243b24b125f29b0332** and copied verbatim to [scanner-stops.json](evidence/scanner-stops.json). **None** of these fifteen binder messages exactly matches that list. Missing captureStackTrace declarations are nevertheless a historical scanner control/blocker described in scanner README/BLOCKERS. That related observation is not an exact fifteen-list membership claim. No retirement from an unidentified external task ledger is claimed.

# Node acceptance control and comparator sensitivity

The request uses the pinned scout's unchanged `prepare-corpus.py`: **301 acceptance projects**, no upstream-case expansion. [request.json.gz](evidence/request.json.gz) retains the exact request. The independent stock TypeScript binder and the new candidate through [node.mjs](node.mjs) produce identical bytes, empty stderr and exit 0. The pinned `compare.py` passes all three checks.

Output: **3,475,070 bytes**, SHA256 **3f83fbf1a9c1587e2923484f7a22911f21c3ac47382ba80dbca597af04d846aa**, matching the original scout's acceptance-only hash. [stock-golden.jsonl.gz](evidence/stock-golden.jsonl.gz), [manifest.json](manifest.json) and [node-control-report.json](evidence/node-control-report.json) retain the proof. The source runner uses stock transpileModule and barrel-first initialization, as scanner/node.mjs does; neither Adamic backend output nor a Node forwarder is called native.

Changing the candidate's `flags: node.flags` to `flags: node.flags + 1` still finishes on Node with exit 0 and empty stderr. The same comparator rejects stdout at byte **405**, with equal total byte lengths. Its mutated dump, report and exit/stderr are retained. Three separate output/process mutants independently corrupt stdout by one byte, add one stderr byte, or change exit 0 to 1; each is caught only by its named comparator check. There is **no native first difference** to report because compilation failed.

# Fixtures, counts and checks

Each of the fifteen `.a` witnesses finishes on Node and reproduces the corresponding compiler message/code on this exact area compiler. [fixtures.json](fixtures.json) records every golden observation, compiler outcome and mutant. Each source mutant changes only the tested value/behavior: key/error message, cache size, timestamp result, initialized-versus-uninitialized cache, or callback result. All fifteen fixed Node comparisons reject their mutants. Headers record measured first checker errors; NotYet witnesses use explanatory comments because NotYet is not Refused.

[audit.py](audit.py) independently parses the raw first-stop logs, checks coordinates/messages and owner qualifications, and executes **each exact replacement** on source Node. Every replacement exits 70 with its named discovery marker. Replacing its throw with `void` exits 0 and is caught. It also rejects missing-stop, wrong-line, wrong-message, false-owner and false-native-green ledger/manifest mutants. [placeholder-mutants.json](placeholder-mutants.json) records all fifteen throw/erasure outcomes. No native guard-erasure or ownership proof is claimed from these source checks.

[counts.md](counts.md) records these local witnesses and acceptance observations. No fixture was registered in internal/oracle; no native allocation counts are invented. Test output is retained under [evidence](evidence); no whole package or full gate was run.

# Source and toolchain provenance

TypeScript source is **6.0.3**, commit **050880ce59e30b356b686bd3144efe24f875ebc8**, outside the repository. The input is the existing cache produced by `stage3/apply.sh` on exact main **45487a809f89885a3fc651cd590e7dabf31362dc**, verified through its source pin and the recorded hashes. No new adaptation pipeline or scanner-specific temporary slice profile is applied: this closure includes createSourceFile, so scanner's createScanner-only profile does not apply.

The slice gathers bindSourceFile, createSourceFile, forEachChild, convertCompilerOptionsFromJson, flattenDiagnosticMessageText and version. It retains **79 ordered evaluation modules**, **28 code files**, **2,225 code declarations**, and **2,316 byte-identical copied spans**; the original slice audit passes. Copied declaration bytes are **2,701,646**, including facades. [slice.json.gz](evidence/slice.json.gz) records every span; [input-hashes.json](input-hashes.json) records staged source bytes. This is deliberately broader than the scout's binder-only declaration closure: the native observer also needs parsing, options and diagnostic serialization.

Setup ran with the required GOPROXY workaround, then `/workspace/adamic-tools/env.sh` was sourced. Timing lines: Node ready **0.074s**, Go **0.086s**, markdown **0.220s**, submodules **0.263s**, clang **0.925s**, Go build **26.022s**, deferred test binaries **26.279s**, warm cache **26.281s**, done **26.383s**. nproc **5**, cgroup quota **4**, memory **17.6 GB**; Node **24.19.0**, Go **1.27.1**, clang **20.1.8**. Setup passed.

The first scratch-build invocation accidentally ran from main and is excluded. The reported compiler was rebuilt from the candidate's checkout with its pinned module linked and `-buildvcs=false`; its SHA256 is in manifest.json. The first full-driver gather named the wrong owner for flattenDiagnosticMessageText; the successful gather uses program.ts. Observer-owned typing errors were corrected before the reported split attempts. These corrections did not weaken compiler flags or original compiler source.

# Commands and replay

All build/test commands redirect output to logs. The cache paths below are concrete measurement paths; repository files changed only under this unit's directory.

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step31-binder-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# In detached /workspace/cache/step31-binder-area at 0c99fc54:
go build -buildvcs=false -o /workspace/cache/step31-binder-area-adamic ./cmd/adamic > /tmp/step31-binder-area-build.log 2>&1
# From main, using the pinned prepared tree:
SLICE_TYPESCRIPT=/workspace/cache/tsc-census/npm/node_modules/typescript/lib/typescript.js bash stage3/slice/run.sh /workspace/cache/checker-stops-adapted /workspace/cache/step31-binder-driver-slice-v2 src/compiler/binder.ts:bindSourceFile src/compiler/parser.ts:createSourceFile src/compiler/parser.ts:forEachChild src/compiler/commandLineParser.ts:convertCompilerOptionsFromJson src/compiler/program.ts:flattenDiagnosticMessageText src/compiler/corePublic.ts:version --no-adapt > /tmp/step31-binder-driver-slice.log 2>&1
node stage3/slice/verify.cjs /workspace/cache/step31-binder-driver-slice-v2 > /tmp/step31-binder-driver-slice-verify.log 2>&1
# Stage main.a at the slice root; the two split attempts are given above.
export STEP31_TYPESCRIPT=/workspace/cache/tsc-census/npm/node_modules/typescript/lib/typescript.js
export BINDER_RUNTIME=/workspace/adamic/oracle/adamic.mjs
WALK_ROOT=/workspace/cache/step31-binder-walk-final python3 stage3/scouts/step31/binder-native/walk.py > /tmp/step31-binder-walk-final.log 2>&1
python3 stage3/scouts/step31/binder-native/check-fixtures.py > /tmp/step31-binder-fixtures-final.log 2>&1
python3 stage3/scouts/step31/binder-native/check-comparator.py > /tmp/step31-binder-comparator-mutants.log 2>&1
python3 stage3/scouts/step31/binder-native/package.py > /tmp/step31-binder-package.log 2>&1
python3 stage3/scouts/step31/binder-native/audit.py > /tmp/step31-binder-audit.log 2>&1
```

The acceptance comparison uses the pinned scout's prepare-corpus.py, binder-dump.cjs and compare.py, exported into the never-pushed scratch. `node.mjs DRIVER REQUEST` takes the same request pathname as the future native executable. The raw requests/golden are committed for replay without rebuilding the scout corpus. `walk.py` can start with a new WALK_ROOT and retains per-step source/diff evidence; existing recorded roots resume only after completing the final recorded placeholder. Never run a discovery tree in a native acceptance slot.

Not covered: native admission past module-initialization guards, binder-body lowering/ownership, native host facilities, malformed-request native decoding, sanitized/leak-clean execution, the 6,262 upstream scout cases, or the repository gate. The next native attempt should repair or prove the original module initialization boundaries and library declarations, then rerun the untouched candidate; the fifteen throwing replacements are not a proposed adaptation.
