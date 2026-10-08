# ecmascript/literal package claim

Branch: lint-helpers/ecmascript-literal. Base: origin/area/stage1-lint.
Triage d7ab0bc4 rank 22; no retained port, four missing helpers: CookedToRaw,
cookedBytesProducedBy, escapeWidthInStringLiteral, producesNoCookedBytes.
All 1,516 origin refs checked for ports and all helper branch claims inspected.
Earlier eligible packages are reserved, excluded tonight, or require runtime regex.
Yielded rules/next to earlier remote dcaf0a30 (local claim never pushed).
No literal package reservation or implementation found; regexpattern's dependency
mention is not a literal reservation. This package scans literals, without regex compilation.

Consumer: no-regex-spaces. Forecast: zero alone; one additional with preceding
complete packages (131 cumulative). Its other dependency ecmascript/regexpattern
is reserved: use its completed shared port if available, otherwise stop dependent
rule work explicitly. Counts are conditional, not findings parity claims.

Claim before code; fetch again and yield to earlier competing claim. Build fresh,
one helper per file; actual Go consumer captures compared across Node, emitted
JavaScript and sanitized native; one caught Node/native mutant per helper.
Prove the consuming rule if its shared prerequisites are available. Run helpers
and lint with every input set before the finished-unit implementation push.
