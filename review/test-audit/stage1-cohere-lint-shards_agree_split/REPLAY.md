# u112 replay

Starting origin/main: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.

Source /workspace/adamic-tools/env.sh. Run npm ci in stage3/api. Obtain the TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8 under /tmp/u112/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8, as recorded in corpus.json.

Run time-baseline.py from the repository root. It writes bounded clean JSON logs and stops on an assertion-red baseline. Four full-corpus ShardsAgree rows could not finish preparation within 90 seconds in this session. Replaying those rows on a prepared machine may give additional evidence.

Run run-checks.py first on restored production sources. Its scratch overlays are under /tmp/u112/weak. W1 deliberately disables only the harness byte comparison for witness checks. S1 through S5 break suite construction. The corrected S3 changes the constructed filename extension, leaving the extension assertion intact. The discarded assertion-edit observation is retained separately and supports no verdict.

Create /tmp/u112/source.ts containing /*😀*/debugger; and /tmp/u112/manifest.txt containing /tmp/u112/source.ts followed by a tab and no-debugger, with a final newline. Then run run-audit.py. It records combined family timings, instruments stable port and coordinator sources, compares clean switched controls, runs the bounded matrix, records direct changed-output witnesses, restores sources in finally, and compiles every standalone diff. Native products use the normal content-addressed build cache; the selector is a runtime input read from /tmp/u112/selector. No compiler source changes. Each standalone port source differs in its content key and is rebuilt through TestProduct_WitnessScriptKindNative, which lowers a copied port with the unchanged no-debugger registry and calls native.Build with sanitizers and split compilation.

The matrix runs the suggestion family and script-kind family, individually per mutant. The latter includes its corpus coverage union. A family kill is one row regardless of how many leaves fail. The Go coordinator is not called by these two families. Its full-corpus caller did not finish its clean baseline, so coordinator survivor observations are bounded and do not establish unguarded behavior in ShardsAgree.

M1 through M4 are production mutants. P1 and P2 are empty-entry probes and never establish uniqueness or subsumption. W1 and S1 through S5 are allowed harness checks, not production kills. Raw logs are the observations; matrix.json and rows.json summarize them. summarize.py regenerates those summaries after all runs complete.

Apply each diffs/M*.diff independently to the starting commit for central replay. Apply P*.diff only as probes. Apply W1/S*.diff only for their named witness or construction checks. Do not apply all diffs together. The committed repository has restored production and test sources; scratch switched sources are retained as .txt evidence.
