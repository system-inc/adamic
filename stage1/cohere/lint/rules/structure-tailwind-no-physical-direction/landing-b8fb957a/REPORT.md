Rebased all twelve owned rules onto main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 and rebuilt the standalone comparisons.
All supported upstream fixtures and 4176 compiler/stage1 file-rule pairs match Go on Node, emitted JavaScript and sanitized native.
Fresh commands: six-rule validate.py, no-multi-str/validate.py and validate-standalone.py with --mutants; verify-regex-reset.py also passes.
Twelve rule mutants, interpolation mutant and RegExp reset mutant are caught only by comparison on all three Adamic paths.
Ten documented parser-gap inputs still refuse; the shared harness remains the named parking blocker, and no new claim was taken.

All validation logs are beside this report. Each main validation used --scratch /tmp/wave15-b8-{six,literals,three} --typescript /tmp/lint-wave1-15-typescript --mutants with the setup environment sourced. Six-rule fixtures 83165 bytes and corpus 79646442 bytes; literals fixtures 26425 and corpus 39664290; final three fixtures 51192 and corpus 39671632. Corpus has 348 files per suite, including 77 TypeScript compiler sources.

Current-main shared TestOwnedWitnesses cannot compile because profile_test.go ranges over portFiles without calling it. Published harness ab70f38d4 remains unchanged and its Go serializer panics on multiple automatic fixes, reproduced previously for prefer-as-const in harness-ab70f38d4. These unowned shared files were not edited. Repeated foundation merge conflicts were resolved using exact published efeb3f66 file versions, as in the prior landing. The integration harness has not landed on main.

rule.json kinds use registry-validated ast.Kind names. No new rule, declaration format or shared driver change is included. Prior numeric exports are not evidence of shared dispatch behavior. Fixed regex uses the published JS literal with lastIndex reset; no matcher was introduced. Full repository gate, full shared-driver parity and throughput were not rerun.
