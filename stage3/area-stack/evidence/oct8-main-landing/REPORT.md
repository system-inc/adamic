Merged origin/main d72728e570fe09d91cf55b37b564dbad0fefc24d into compiler/area-stack without conflicts, merge f91e85d93945e9ebbd7816cbcf0de41ebf7f87f8. No topic or compiler behavior changed in this landing follow-up.

Commands and outcomes:

* npm ci in stage3/api: exit 0, three locked packages installed.
* go test ./internal/lower ./internal/ir ./internal/flow -count=1: exit 0; lower 50.591s, IR 1.228s, flow 180.462s.
* go test ./stage3/fixtures -count=1 -timeout 10m -v: the initial run found seventeen stale stage0 records. Thirteen previously blocked fixtures now compile. Four remaining diagnostics are the cycle-capable refusal in objects/29_parameter_default.a:6:13, closed-origin indexed-mutation NotYet in objects/10_build_options.a:8:5, generic-function-value NotYet in cycles/04_directory_callback/main.a:9:71 and namespace-object-value NotYet in namespaces/10_tracing_escape.a:38:11.
* go test ./stage3/fixtures -count=1 -timeout 10m -v -args -update: exit 0, 29.395s. The existing updater changed compiling records only after source Node and native output agreed, including ASan/UBSan and leak checks. Four noncompiling records were updated to the exact current diagnostics separately.
* go test ./stage3/fixtures -count=1 -timeout 10m -v after refresh: exit 0, 30.235s. All 173 fixture records pass. An audit confirms only seventeen stage0 fields changed; all Node observations, source files and provenance remain unchanged.
* go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1: exit 0, 76.901s; no count refresh needed.

All output is retained alongside this report. No new compiler check or lowering was introduced, so no new implementation mutant was required. Previous topic and catalog mutant evidence remains in the stack. No full gate or new census was run. Push target is compiler/area-stack only.
