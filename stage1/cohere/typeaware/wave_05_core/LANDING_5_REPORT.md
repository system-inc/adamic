Built: all nine completed wave-05 ports rebased onto current main b8fb957a; no new rules remain unclaimed.
Commits: tested rebased tip ce0d73eee7afba37915fcf202759d06ca8bbc8b5; enclosing commit records landing evidence.
Commands and outputs: nine-rule controls/corpora, sanitizers, released handles, bridge packages, filtered uncached Node oracle, metadata and vet PASS.
Mutants: nine rule verdicts and CFG corruption caught only by Go bytes; release registry and named/numeric listener checks caught their mutations.
Not covered: three parked React hook implementations, full option matrices, shared-registry integration and full repository gate.

Main advanced from c01907a7 to b8fb957aa839a9e8cb0b54279dd9864fa317bd30 with its inherited-static-field fix. The requested rebase applied all 24 commits without conflicts; main's changes are retained. The shared macOS leak-helper migration and JSX work have not reached this main, and no shared file was reverted. Only codex/typeaware-wave-05 is published, using an exact lease on its previous tip adb2e90778557b38a7f44056ee9430d5843f4480 after rebase.

A fresh all-origin inventory has 584 refs, 33 unique Markdown claim blobs, 172 claimed ranked names and 25 ranked baseline ports. All 197 ranked checker-dependent rules are ported or claimed; the unclaimed list is empty. This worker takes no further claims. The prior three React hook reservations remain parked with JSX/HIR/SSA/capture blockers already named; they are not represented as implemented.

Revalidation uses the same frozen repository (287 roots) and TypeScript compiler (77 roots) manifests, pinned Go cohere and typescript-go. Commands are those recorded in REPORT.md and wave_05_next/LANDING_3_REPORT.md, with landing5 log/artifact prefixes. Both corpus manifest environment variables and ADAMIC_TYPESCRIPT_SOURCE were set for the prior-six validators. All toolchain shells source /workspace/adamic-tools/env.sh and all subprocess test output is saved to files.

- TestWave05AgreementAndMutants runs the original three rules, all controls, both corpora, sanitizer builds, three verdict mutants and released-handle/registry-deletion checks.
- WAVE05_OUTPUT_SKIP_BENCH=1 validate_output.py covers output controls/modules, DOM/Node timer regressions, two output verdict mutants, CFG corruption, releases and both corpora normally and sanitized.
- WAVE05_NEXT_SKIP_BENCH=1 validate.py covers timer controls, its verdict mutant, release, retained constructor-gap probe and both corpora normally and sanitized.
- wave_05_core/validate.py covers 106 parse-valid controls, 76 findings including complete suggestions, all three verdict mutants, two releases and both corpora normally and sanitized.
- go test ./bridge/tsgo/... -count=1 -timeout=15m checks bridge packages; go vet ./bridge/tsgo/... runs separately.
- ADAMIC_GATE_UNCACHED=1 filtered internal/oracle test covers the one-byte oracle mutant and maps_and_text, sorting, string_index, lone_surrogates, functions and closures.
- Both owned metadata scripts check three named manifests and six earlier listener exports with their mutations; rule bodies still have no string-kind relevance comparisons.

Exact logs are retained in evidence/landing5-*.log; inventory and source base are in evidence/landing5-inventory.json. No new isolated performance benchmark is claimed: these validation processes overlap. The previously isolated measurements remain historical results in REPORT.md. Setup remains the prior successful 82-second run, nproc 5. The full repository gate, full option matrix, application of suggested edits and shared emitted-JavaScript rule harness remain outside this validation.

Observed final results: first-wave oracle PASS 156.147s; output and timer validators PASS; core validator PASS 106 controls/76 findings with full suggestions and both zero-finding corpora. Bridge PASS 98.989s and checker PASS 0.208s. Fresh Node oracle PASS 1.881s, native 0 hits/28 misses and Node 0 hits/19 misses. All metadata checks, vet and diff checks PASS. Remote main was rechecked at b8fb957a immediately before publication.
