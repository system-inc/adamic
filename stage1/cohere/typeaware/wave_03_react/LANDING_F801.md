Rebased wave-03 onto current origin/main f8013f0ba; no new claims in this landing unit.
Validated implementation tip 1c6950c0c748351e1527813be78c7bdee2b3c6e2; this evidence commit follows it.
Four owned rule oracle suites PASS in 701.284s; checker PASS in 1.233s; touched-package vet PASS.
All 12 rule mutants caught by byte comparison; alias, guards and four released-registry mutants caught.
Arbitrary configured regex, legacy numeric listener conversion and emitted-JavaScript integration remain unfinished; full repository gate not run.

Main advanced from e8ba3d5d to f8013f0baac41ddc340d76f83bddde38536a8f07.
The rebase was clean and changed no owned implementation. A fresh main fetch
after validation returned the same tip. The only branch this unit pushes is
codex/typeaware-wave-03; main and area branches are reserved for integration.

All four complete owned suites ran once with fresh landing-f801 artifact paths.
Each compared complete findings, fixes and suggestions against unchanged Go
production rules, including generated controls and the frozen 77 compiler and
287 repository roots. Normal and ASan/UBSan/leak-check runs matched; stale checker
handles produced the required panic 70. Complete logs are retained in validation/
as landing-f801-oracle.log.gz, landing-f801-checker.log.gz and landing-f801-vet.log.gz.

After sourcing /workspace/adamic-tools/env.sh, the commands were:

```sh
ADAMIC_WAVE03_ARTIFACTS=/workspace/wave-03/landing-f801-original \
ADAMIC_WAVE03_NEXT_ARTIFACTS=/workspace/wave-03/landing-f801-next \
ADAMIC_WAVE03_MORE_ARTIFACTS=/workspace/wave-03/landing-f801-more \
ADAMIC_WAVE03_REACT_ARTIFACTS=/workspace/wave-03/landing-f801-react \
ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest \
ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus \
go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants)$' -timeout 30m -count=1 -v
go test ./bridge/tsgo/checker -count=1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
```

Every rule mutant compiled, exited 0 and was caught by differing output bytes:
unsafe-call, imports, template, race, output, blocking, throw, backreference,
callback, unsupported, memo and boolean. The React mutants preserved finding
counts. The memo alias-filter mutant was caught by bytes. Regex and metadata
guard mutants escaped their required panic and were caught by that expectation.
All four released-registry mutants exited 0 and failed the required panic 70.
These are actual reruns against this main baseline, rather than earlier evidence.

Observed native/Go wall times are recorded below in the test's own timing lines.
They are individual observations, not benchmark medians or speed improvements.
Native remains slower. Compiler builds and sanitizer runs are outside the timed
commands; no additional quiet benchmark was run for this landing unit.

```text
wave_03_more_test.go:176: repository native 730.943119ms Go 230.617595ms; native tsgo: load_ns=129234552 query_ns=16552803 queries=311 first_query_ns=17806 run_ns=560380334
wave_03_more_test.go:176: compiler native 5.118086946s Go 799.551492ms; native tsgo: load_ns=377034610 query_ns=459623301 queries=538 first_query_ns=89676 run_ns=4436277666
wave_03_next_test.go:219: repository timing: native 568.132319ms Go 192.371228ms; native tsgo: load_ns=107120911 query_ns=7894833 queries=319 first_query_ns=16357 run_ns=427311508
wave_03_next_test.go:219: compiler timing: native 3.446478s Go 481.002671ms; native tsgo: load_ns=383275637 query_ns=43286279 queries=110 first_query_ns=29691 run_ns=2882377849
wave_03_react_test.go:182: repository native 1.42731367s Go 512.372574ms; native tsgo: load_ns=163833067 query_ns=473387345 queries=287 first_query_ns=1758651 run_ns=1173196117
wave_03_react_test.go:182: compiler native 13.368134771s Go 981.767567ms; native tsgo: load_ns=723480068 query_ns=5771716036 queries=77 first_query_ns=131691 run_ns=12155145865
wave_03_test.go:161: repository timing: native 664.89902ms Go 335.204959ms; native tsgo: load_ns=126666834 query_ns=185928902 queries=27298 first_query_ns=44647 run_ns=529135066
wave_03_test.go:161: compiler timing: native 4.832161201s Go 1.883239998s; native tsgo: load_ns=388990179 query_ns=1738665587 queries=185745 first_query_ns=23715 run_ns=4412527293
```

No new rule was claimed while the branch required landing validation. The
remaining boolean naming limitation and shared checker JavaScript gap remain
as documented in REPORT.md and LANDING.md. No new rule implementation or
rule.json listener declaration is claimed in this landing pass. No shared
harness, registration generator or compiler file was edited.
