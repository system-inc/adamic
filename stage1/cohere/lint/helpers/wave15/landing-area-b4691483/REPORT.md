Rebased completed helpers onto lint area b46914832d70e00847d82d5d221ab7bb24040c53, containing main c7991b90.
Fresh Go, Node, emitted JavaScript and sanitized native comparisons pass for all three helpers.
198584 CSS string, 8836 declaration and 1197583 flags cases match; upstream consumer Go tests pass.
Escape, important-range and duplicate-flag mutants are caught only by output comparison on all three Adamic paths.
No new claim; full gate and the 17 newly required checks remain unverified.

Ran each owned validate.py with /workspace/adamic-tools/env.sh sourced. Helper sources unchanged. Previous reports name six Tailwind dependency consumers for each CSS helper and four flag consumers; these are helper proofs, not full consumer ports. Accepted the expanded registered shared driver unchanged. Broad rule parity still explicitly refuses case-936/Octal.ts.
