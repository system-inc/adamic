Retained current main 71d7e491 and integrated lint area e667e3e1 without authored shared changes.
All six delivered helper validators rerun PASS against actual Go, source Node, emitted JavaScript and ASan/UBSan native.
Counts: CSS string 198584; CSS declaration 8836; regexp flags 1197583; SpecifierNode 82; SourceVisitors 65; AccessedName 33694.
All six compiling semantic mutants are caught only by output comparison on all three Adamic paths; original consumer test families PASS.
No new helper claimed while broad rule parity retains the explicit octal parser-recovery refusal; full gate, 17 external checks and throughput remain unverified.

Commands: source /workspace/adamic-tools/env.sh, then python3 stage1/cohere/lint/helpers/wave15/<helper>/validate.py for css_string, css_declaration, regexp_flags, import_specifier, import_source_visitors and accessed_property. Complete logs are beside this report; refreshed per-helper evidence remains in each helper directory. Setup with GOPROXY=https://proxy.golang.org|direct passed, nproc=5; timing lines preserved in setup.log.
