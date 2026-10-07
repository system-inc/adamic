# Landing blocked after main advanced

The original REPORT.md and pre-merge evidence are retained as observations on the old main base, not as current-main green claims.

The final refresh discovered main advance from 71d7e491 to `48c05d091f0a43c31cbe051b1d6578d99eeedf19`. The new main includes compiler changes and changes cohere to `7945d102a6c18dd36adf9114a758ce646e8b2359`. It was integrated without history rewriting by own-branch merge `c3a12c1ef683e0537512c1b3e845abf4473c8e31`; lint area `2fbfe42155387f67f297a41c89d318e3b6cde310` remains an ancestor. No main or area branch was pushed.

Setup was rerun with the required GOPROXY fallback and checked out both new submodules. It succeeded in 169.554s, nproc 5. Timing lines: Go ready 0.027s, markdown dependencies ready 0.086s, clang ready 0.205s, submodules ready 7.170s, Go build ready 169.181s, test binaries deferred 169.511s, cache warm 169.512s. Complete output is in landing/setup.log.

Fresh checks, all direct log output:

- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch33 -count=1 -v -timeout=20m`: FAIL, 0.013s, exit 1. TestComponent, TestType and TestNetwork each fail with `Go pin drift`. The exact original oracle pin is 715ba94f3608a6500086b1076ce5cb7e51b836db; current cohere is 7945d102a6c18dd36adf9114a758ce646e8b2359. This is a provenance guard failure before fixture comparison or mutant execution, not evidence of helper semantic failure or success on new cohere. The pinned helpers require an explicit oracle migration and revalidation; changing only the expected hash would not establish that. No guard was bypassed or relaxed. The same old pin occurs in the other historical owned batches; they were not rerun here.
- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch34 -count=1 -v`: PASS, 1.157s, exit 0. Actual new Go Compile and source Node print true for pattern a, flag u, subject a. Current Adamic lowering still returns typed NotYet for a nonconstant pattern. The first reproduced shared compiler gap therefore remains present on current main.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v`: PASS, 0.931s, exit 0, seven probe misses and zero hits.
- `go vet ./...`: exit 0, empty log.

The branch includes current main but is explicitly not landing-ready. All 21 latest-trio semantic mutants were caught before this main advance; none is credited on the new main or cohere pin. The new Compile helper is undelivered and has no credited semantic mutant. Zero new helper blockers or rules are removed. No additional helper is claimed. Stop at the reproduced shared dynamic-RegExp gap and report the additional outstanding oracle migration. The full gate and its 17 required external checks were not run, skipped or modified.

Pre-merge evidence logs were initially ignored by the repository's log pattern. They are explicitly force-added with this landing report so all reported observations are reviewable. No prior pushed evidence was removed or replaced.
