# Recorded validation

All test output was redirected directly to log files. This unit ran scoped gates,
not the full repository test suite. No internal compiler/runtime code changed.
The branch starts at origin/main 5d4c8012a0877094134e6c6bac367ff68f9313e8.
Cohere and its nested TypeScript submodule remained unchanged.

| Command | Observed result | Log |
|---|---|---|
| bash cloud/setup.sh | Go/clang/Node/submodules ready; cache warm 73s; total 73s; nproc 5, cgroup quota four CPUs | results/setup.log |
| go test -v -count=1 ./stage1/cohere/graphql/printer, with Prettier and benchmark enabled | PASS, 141.098s; 3,564 texts x four options on native/Node/JS backend; all three mutants caught; file driver and constructor proof pass | results/printer-tests.log |
| go test -v -count=1 ./stage1/cohere/graphql/printer -run 'TestPrinter(WhitespaceGap\|ConstructorGap)' | PASS, 9.585s, final standalone gap inputs | results/whitespace-proof.log |
| go test -v -count=1 -timeout 15m ./stage1/cohere/graphql ./stage1/cohere/json | PASS, parser 44.977s, JSON 326.718s, optional external oracles enabled | results/regressions.log |
| go test -v -count=1 -timeout 15m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(classes\|class_layouts\|strings\|case_mapping\|ascii_scan\|exceptions\|from_char\|from_code)' | PASS, 14 fixtures; native held to Node and backend, sanitizer/leak checks | results/oracle.log |
| go vet ./... | exit 0, empty output | results/vet.log |
| cohere --no-fix --no-cache, restricted to the three port .ts files | exit 0; 276 rules, three checked, 100% Adamic-ready | results/cohere-gate.log |
| go run ./cmd/adamic build stage1/cohere/graphql/printer/main.ts -o /tmp/adamic-graphql-printer | exit 0; ordinary-file output verified | results/build.log, results/driver.log |
| git diff --cached --check | exit 0 | checked before commit |

Use README.md's pinned scratch installation before reproducing the tests.
`ADAMIC_GRAPHQL_PRETTIER` points to that directory. For the regression run,
`ADAMIC_GRAPHQL_LIBRARY` and `ADAMIC_JSON_PRETTIER` pointed there too. For the
printer suite, `ADAMIC_GRAPHQL_PRINTER_BENCH=1` enabled verified throughput and
`ADAMIC_GRAPHQL_PRINTER_KEEP=/tmp/graphql-printer-corpus` retained generated data.
The final two proof tests ran after the main suite; their source adds an explicit
JSON fixture for the five whitespace inputs. Formatting-only Go import changes
and a source comment clarification followed the main suite; the latter passed
cohere's gate again. The printer package was vetted again after those changes.

The independent Go AST coverage table is results/ast-coverage.json. Exact pinned
npm dependency versions and integrities are in results/oracle-package-lock.json.
GAPS.md records the limits of the comparison: five upstream whitespace acceptance
differences, no repository .graphql files present, and syntax errors compared to
Prettier by refusal status rather than error code-frame bytes. Go errors are
always held byte for byte.
