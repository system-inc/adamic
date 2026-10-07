Rebased the seven takeover rules onto current main 71d7e491 and retained current area bb2ece56 ancestry; owned rule sources unchanged.
Revalidated implementation c10930bcf5417fd2fbdf88bdc9f78c90c4277dd4; previous pushed 0e26763f62da23a566e7ca0a6e8ea921b3e9ed39; evidence commit follows on codex/lint-wave1-06-land only.
Fresh TestRulesAgree and TestOwnedWitnesses PASS 101.703s; no source fixes or new claims.
Seven existing clean-executing semantic mutant proofs remain applicable to unchanged compiler, stage1 and rule sources; mutants were not rerun this turn.
Full gate, setup, vet, full mutant registry and required compiler corpus were not repeated for this branch; no helper work claimed.

Fetched all origin heads. Main 71d7e491b3c9724f7a0e2ee754592149e7f9790b integrates stage3 only relative to previous main 4e0bfda5: git diff --name-only over cmd, internal, stage1 and cloud/setup.sh is empty. DEDUP_LEDGER.md is unchanged. Rebased onto current main and merged area bb2ece564842c4b2f909b9f75c27e74c2efa4f29 to preserve both ancestries; the area merge changes no source tree. Seven owned directories are byte-identical to prior pushed 0e26763f6. No shared resolution or compiler edit was authored.

Fresh command sourced /workspace/adamic-tools/env.sh:
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout 15m
PASS 101.703s. TestRulesAgree passed 65.99s with 13,188,475 identical bytes; TestOwnedWitnesses passed 35.69s with 97,561 identical bytes. Findings and fixes match independent Go on source Node, emitted JavaScript and ASan/UBSan native. Positive witnesses fire. Output went directly to tests.log. Inherited method-signature-style recovery exclusions remain explicit; excluded recovery is not parity coverage.

The unchanged executable tree retains the seven passing mutant definitions from wave1-06-land-main-report.md: wrongStrategy, wrongCssTags, wrong-global-type-name, omit-empty-export-fix, wrong-enclosing-scope, configured-restriction-ignored, report-left-reference. No fresh mutant, setup, vet, performance or full-gate result is claimed. Previous setup evidence is 450.035s, nproc 5. All seven canonical rules remain retained; losing copies and baseline duplicates remain absent. Current helper-job prerequisites on wave1-07 still fail because consistent-return and constructor-super are unfinished. No new rule or helper is claimed. Push only the own landing branch with an exact lease against 0e26763f62da23a566e7ca0a6e8ea921b3e9ed39; never main or area.
