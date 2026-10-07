Rebased helpers onto current main 71d7e491 and cleanly merged lint area b28757f3; helper sources unchanged.
CSS string 198584, declaration 8836 and flags 1197583 cases match Go, Node, emitted JavaScript and sanitized native.
Escape, important-range and duplicate-flag mutants compile, execute and differ only by output comparison on all three Adamic targets.
Every represented upstream consumer test passes; six Tailwind consumer dependencies per CSS helper and four flag consumer dependencies remain supplied.
No new claim in this commit; full gate, 17 required external checks and throughput were not rerun.

Ran each owned validate.py with /workspace/adamic-tools/env.sh sourced. Complete logs are beside this report. Consumer and corpus coverage boundaries are unchanged from the original helper reports. Both owned branches contain current main; the latest published lint-area changes were integrated without authoring shared file edits. The rule branch records the explicit octal parser-recovery refusal.
