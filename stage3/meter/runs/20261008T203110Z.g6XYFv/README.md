# Stage 3 meter on main d72728e5

Per-ref run from meter writer `11e837ee510c174c8245de17421f655221de8e3f`:

```sh
source /workspace/adamic-tools/env.sh
export PATH=/workspace/stage3-meter-progress-tmp/go-wrapper:$PATH
export TMPDIR=/tmp/stage3-meter-d72728e5
STAGE3_METER_MAIN_REF=d72728e570fe09d91cf55b37b564dbad0fefc24d STAGE3_METER_COMPILER=per-ref STAGE3_METER_RUNS=$PWD/stage3/meter/runs bash stage3/meter/twice-daily.sh > /tmp/stage3-meter-d72728e5.log 2>&1
```

Main source and compiler: `d72728e570fe09d91cf55b37b564dbad0fefc24d`. Area source and compiler: `17b0161d97ede12914d1c171ac5a58491b64f43e`. Node 24.19.0; nproc 5. The Go workspace wrapper uses each ref's compiler module and clean cached dependencies at the same pins as its initialized submodules. Per-tree build-provenance.json and binary-buildinfo.log retain those checks.

| Main roots | 45487a80 (20261008T171021Z.iBLy18) | d72728e5 | Delta |
| --- | ---: | ---: | ---: |
| Own file | 56/79 | 56/79 | 0 |
| Whole program | 2/79 | 2/79 | 0 |
| Lowering | 0/79 | 1/79 | +1 |

The sole new lowering pass is `src/compiler/hostErrors.ts`, created by adaptation 47. Tsc's original 78 roots remain own file 55/78, whole program 1/78, lowering 0/78. There are no own-file fail-to-pass or pass-to-fail moves against that baseline. baseline-comparison.json records the exact comparison.

Area measures own file 56/79, whole program 2/79, lowering 0/79; its original 78 roots have the same counts as main. Both tsc entries fail with 320 whole-program diagnostics. Three adapted compiler inputs differ between main and area: commandLineParser.ts, scanner.ts, utilities.ts. adapted-input-comparison.json records this difference; paired changes cannot be attributed to the compiler alone.

Full latent mode retains the no-output guards. Main's compiler-file measurement recovers 10,551 panics: 10,550 are `latent state copy: unexported IR field argumentFacts`, and one is `Unhandled case in Node.Text: *ast.ComputedPropertyName`. Its entry measurement recovers 10,558 panics. These are overlay measurement failures; its latent counts do not establish full-body coverage. The ordinary root counts above are separate observations. Each stream's latent-diagnostics.json preserves every distinct panic and error; raw census JSONL and gzip streams remain ignored.

The writer's merge validation passed all 49 meter tests without skips using existing real census probe binaries, and caught equal-hashes-only, missing-adaptation, literal-entry-path, and full-entry first-error mutants. Its 10 progress fixtures passed again during this run. Logs are under validation/. Those fixtures do not validate snapshot support for main's new private IR field.

This meter invokes no native backend or native-output comparison. Its progress snapshot does not claim a native milestone. The scanner build-stop observation named in progress.json is historical (compiler dbd7a7c8), not a scanner build on d72728e5. Reproducible scratch copies and submodule metadata were relocated or removed to make disk space; committed prior evidence was preserved.

Completed with exit 0. The run and root progress snapshots are byte-identical; all four native milestones remain false. Full latent totals (NotYet, Refused, errors, panics): main compiler roots 1551, 8410, 0, 10551; area compiler roots 15284, 12186, 2, 25. Both entry streams and compiler-root streams retain their complete panic/error ledgers.

Go build-info logs have formatting-only trailing tabs trimmed for the whitespace check; their exact original bytes are preserved in binary-buildinfo.log.raw.gz. Diagnostic streams and progress snapshots were not reformatted.
