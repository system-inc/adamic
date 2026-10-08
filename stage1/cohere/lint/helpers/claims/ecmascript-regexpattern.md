# ecmascript/regexpattern package claim

Branch: lint-helpers/ecmascript-regexpattern. Base: origin/area/stage1-lint at cd56db1dd7ab7eda7a5f492804dcd1849a1494aa.
Triage: d7ab0bc4, rank 19. No retained port (0/18); build fresh. Every higher-ranked eligible package is reserved; tonight exclusions and runtime regex compilation exclusions honored.
Forecast: 0 rules alone, 1 additional with all earlier packages (127 cumulative); conditional helper readiness.
Consumers: no-control-regex and no-regex-spaces. Completing regexpattern and regexsyntax removes no-control-regex's remaining helper blockers; no-regex-spaces also needs ecmascript/literal.

Checked every claims/ file on all origin/lint-helpers/* and origin/codex/lint-helpers* branches after fetch: 30 files, no regexpattern reservation.
No runtime regex compiler is used; this package scans pattern source.
The full walker depends on ecmascript/regexsyntax. Its package is reserved by origin/lint-helpers/ecmascript-regexsyntax, currently claim-only f997f145cceca508f0c2767778fe503463c31354. Build independent regexpattern helpers first; stop dependent helpers rather than duplicate that reservation unless a finished port becomes available.
One helper per file, live consumer Go captures, source Node/emitted JavaScript/sanitized native agreement and one output mutant per helper on every backend required before the finished-unit push. One consuming rule is required when its prerequisite closure is available.
