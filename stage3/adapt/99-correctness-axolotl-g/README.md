# 99g: preserve cleared init options as JSON null

JavaScript-changing. A command-line option explicitly cleared with `null` is stored as an own property with value undefined. generateTSConfig now writes the JSON literal null for that option, preserving the active or commented form selected by emitOption. An absent option and every other value retain their previous output. This is a fork bug fix, separate from 99b's soundness adaptation; no required-value assertion is added.

lane: run as part of Correctness's wave

Base: main 470e16f8315f7a83c5c7b7f6f2a77d8328db0a25. Source: stock TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Branch: typescript/axolotl-99g. The branch is independent of 99b and 99a. The main base contains neither, so whole-program counts use this main tip rather than the earlier 99b measurement tree. The original 99b decline is left untouched.

## Repair and source contract

The only source changes are inside generateTSConfig's nested emitOption in src/compiler/commandLineParser.ts:

```diff
-        const value = (options[setting] ?? defaultValue) as PresetValue;
+        const value = options[setting];
+        const formattedValue = hasProperty(options, setting) && value === undefined ? "null" : formatValueOrArray(setting, (value ?? defaultValue) as PresetValue);
         if (comment) {
-            result.push(`${tab}${tab}// "${setting}": ${formatValueOrArray(setting, value)},`);
+            result.push(`${tab}${tab}// "${setting}": ${formattedValue},`);
         }
         else {
-            result.push(`${tab}${tab}"${setting}": ${formatValueOrArray(setting, value)},`);
+            result.push(`${tab}${tab}"${setting}": ${formattedValue},`);
         }
```

The option value is still read once. If the input has an own undefined property, formatting uses the literal null directly, including for fixed settings with defaults and enum maps. Otherwise formatting gets exactly the previous nullish selection and the previous PresetValue cast. The original comment-selection code and remaining-options removal are untouched. There is still exactly one formatting evaluation for each emitted option.

Stock TypeScript 6.0.3 parses and locates the one generateTSConfig and one emitOption, checks their nesting, and pins each value declaration, formatter declaration, output template and formatting call by exact text and count. emitOption must have six original statements or seven fully adapted statements. It refuses partial adaptation, altered conditions, missing own-property tests, extra statements, template drift or call-count drift. It plans all three edits, reparses and validates the complete after shape, checks the unchanged input, then writes once. It preserves the input's newline bytes. A fully adapted file causes no write.

## Go behavior and the explicit difference

The repository pins system-inc/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f through cohere 06af2e399c6c1fb93fd570b774027942747400b4. Its Go CLI is built from cohere/TypeScript/tsc/cmd/tsc.

`--init --outFile null` exits 0 and writes `"outFile": null,`. The fixed setting `--init --strict null` also exits 0 and writes `"strict": null,`. The Go parser records nil, emitOption distinguishes present nil from an absent map key, and JSON marshaling preserves null. 99g follows this behavior.

There is one named exception: the pinned Go CLI panics for `--init --target null`, exits 2, and writes no config. Its enum lookup tries to match nil and reports `panic: No matching value of <nil>`. Following the user's ruling, 99g instead writes `"target": null,` and reads that file cleanly. No attempt is made to copy the Go panic.

The earlier upstream investigation at Microsoft/TypeScript main [6161772c1370bdd4a82f808e6823bd755eb1870d](https://github.com/microsoft/TypeScript/commit/6161772c1370bdd4a82f808e6823bd755eb1870d), dated 2026-10-30, found active Go init.go identical to the pinned Go source and a retained TypeScript fixture whose generateTSConfig, including emitOption, is unchanged from 6.0.3. The fork issue and fix are recorded in [the upstream ledger](../../upstream/LEDGER.md).

## Runnable proofs

[witness.cjs](witness.cjs) runs the actual built before and after compiler libraries through executeCommandLine, using isolated directories with an index.ts input. It records the generated tsconfig.json, init stdout/stderr, and a second invocation's --showConfig readback. It also runs the pinned Go executable for the same commands and retains every line difference. It covers ordinary --init, --init --strict false, and null-cleared outFile, strict, target and rootDir. Additional API controls cover absent and inherited properties, own null, empty string, zero, false and empty arrays.

The mutant restores generateTSConfig's original compiled function in the real after library, reverting this repair's lines. The identical after checker runs as a separate Node process and must fail specifically because outFile did not retain null. Its stdout, stderr and exit status are saved.

[verify.cjs](verify.cjs) checks the exact source diff, equality of the baseline snapshot and 99g's input snapshot, unchanged public API bytes, JavaScript-changing transpilation, the no-write second run, and six refusal fixtures. [prove.py](prove.py) builds both upstream artifacts and runs all local checks before the sequential unfiltered oracle suites. It applies 91's strict count, sanctioned API failure, memory-event, exact baseline-byte and public-API gates. No test, baseline, sanction, expected count or oracle implementation is changed.

All raw logs, configs, readbacks and oracle reports are under /workspace/scratch/99g. Commands:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99g/before > /workspace/scratch/99g/apply-before.log 2>&1
# The before build was started before this adaptation directory was created.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99g/after > /workspace/scratch/99g/apply-after.log 2>&1
go build -o /workspace/scratch/99g/adamic ./cmd/adamic > /workspace/scratch/99g/build-adamic.log 2>&1
go build -o /workspace/scratch/pinned-tsgo ./cohere/TypeScript/tsc/cmd/tsc > /workspace/scratch/build-pinned-tsgo.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/adapt/99-correctness-axolotl-g/prove.py /workspace/scratch/99g > /workspace/scratch/99g/prove.log 2>&1
```

The oracle commands, sequentially, are:

```sh
bash stage3/oracle/run.sh /workspace/scratch/99g/before /workspace/scratch/99g/oracle-before --workers 4 --limit-seconds 1800
bash stage3/oracle/run.sh /workspace/scratch/99g/after /workspace/scratch/99g/oracle-after --workers 4 --limit-seconds 1800
```

Before each suite, the wrapper evicts clean cached workspace file pages using read-only file descriptors and POSIX_FADV_DONTNEED. This changes no source bytes, tests, Node flags, worker counts or cgroup settings. It keeps pages from older completed scratch trees from consuming the test memory budget. The cache and memory measurements are retained.

The first control suite completed with the expected 106366/1/0 counts, but the wrapper stopped when memory.events max increased from 0 to 30215; oom and all kill counters stayed zero. Its full result remains at /workspace/scratch/99g/oracle-before-pressure. Releasing clean cached pages reduced accounted memory from 8819761152 to 3657322496 bytes. The unchanged control and after oracles were then retried once, sequentially, with the same stopping gates. Completed local proofs and built artifacts were reused with `prove.py /workspace/scratch/99g --oracles-only`; the retry log is prove-retry.log. No expectation or memory-event gate was relaxed.

The wrapper gives each oracle a 2400-second hard limit. It stops on any new checker diagnostic, unexpected suite result, new baseline path, changed baseline diff bytes, API change or memory event. A single inserted source line is mapped back explicitly when comparing complete diagnostic strings; all columns, codes, messages, multiplicities and elaboration chains remain part of the comparison. No checker options or overlays change.

## Results

Both full apply.sh runs passed end to end. The 99g row is one file, four lines added and three removed. The exact changes are the value/formatter declarations and two output templates shown above.

| Command | Before config value | After config value | Before/after readback exit | Go comparison |
| --- | --- | --- | --- | --- |
| --init | unchanged defaults | unchanged defaults | 0 / 0 | entire file byte-identical |
| --init --strict false | false | false | 0 / 0 | entire file byte-identical |
| --init --outFile null | undefined token | null | 1 / 0 | entire after file byte-identical |
| --init --strict null | true default | null | 0 / 0 | entire after file byte-identical |
| --init --rootDir null | "./src" default | null | 0 / 0 | entire after file byte-identical |
| --init --target null | "esnext" default | null | 0 / 0 | Go panics and writes no file |

The full [invalid before file](proofs/outFile-before.tsconfig.json.txt), [valid after file](proofs/outFile-after.tsconfig.json.txt) and [Go file](proofs/outFile-go.tsconfig.json.txt) are retained. The invalid line is `"outFile": undefined,`; the after and Go line is `"outFile": null,`. The actual before readback reports:

```text
tsconfig.json(44,16): error TS5024: Compiler option 'outFile' requires a value of type string.
```

The after readback exits 0 with no diagnostic. [witnesses.json](proofs/witnesses.json) records every comparison, all readback exits and diagnostics, API controls, and the mutant's exit 1 with `outFile did not retain null`.

Whole-program [measurement](proofs/measurement.json): 79 sorted roots, **112 before and 112 after**, zero additions and zero removals. The single inserted line is mapped back for comparison; full diagnostic strings otherwise match exactly. D036 remains the same TS2322 at after line 2958:13, previously 2957:13. No `!` is added, and 99b's decline validation is not edited.

[Verification](proofs/verification.json) passes all six refusal fixtures, confirms no second-run write and byte-identical public API declarations, and records commandLineParser.ts transpilation at 161985 bytes before and 162076 after. This is intentionally JavaScript-changing.

The sequential [oracle comparison](proofs/oracle-comparison.json) passes:

| Oracle | Passing | Failing | Pending | Wrapper seconds |
| --- | ---: | ---: | ---: | ---: |
| Control | 106366 | 1 | 0 | 221.785 |
| 99g | 106366 | 1 | 0 | 335.800 |

Both install/build phases exit 0. Both test phases exit 1 solely for the sanctioned `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change` failure. The existing normalizer accepts exactly all 223 sanctioned declarations, with no missing, extra or changed declaration. No expectation or sanction was edited.

**Zero baseline moves.** Both complete baseline diffs are byte-identical at 51123 bytes, SHA-256 `6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`. The only baseline path is the inherited api/typescript.d.ts exception. Public typescript.d.ts bytes match between the actual control and 99g artifacts. All memory-event counters stayed unchanged during both retry suites; oom and kill counters remained zero throughout. The original pressure-affected control result is retained and disclosed above.

Only this adaptation directory and stage3/upstream/LEDGER.md are committed. No lane manifest, baseline, test input, checker option or production compiler implementation is edited.
