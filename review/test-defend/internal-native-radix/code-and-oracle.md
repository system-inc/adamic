CODE UNDER TEST and ORACLE, before mutation:
- Radix row: runtime/radix.c adamic_number_to_radix, with live Node output, RangeError classification, native exit70 and diagnostic prefix.
- Benchmark: runtime record/map/string/heap behavior at sizes1000,10000,100000,1000000, unsanitized, five rounds. Live Node work checksum; hand-written timing field validation. There is no performance threshold.
- Regexp Search, Lint and Random family: runtime regexp VM and internal/regexp compilation/native bytecode, with live Node capture spans, groups and lastIndex. Node is the oracle, never mutated.
- BytecodeTest262: same compiler/runtime with recorded outside-authority test262 observations; no live test262 runner.
- StepLimit: runtime VM step budget, hand-written exit70 and diagnostic requirement. New boundary test checks exact instruction budgets independently.
- PatternUnits: regexp UTF16 pattern compiler and native VM, hand-written exact [0,1] spans for two different lone surrogates; does not use JSON-decoded pattern text as its authority.
- ReleasePaths: heap release/draining on NULL, immortal, shared strings and 100000 linked objects; live allocation counts and fixed output plus sanitizers.
- StringEquality: runtime equal for same header, separate equal headers, unequal strings and undefined; live Node strict equality.
No oracle, harness or test source was changed. Construction coverage is measured with Go coverpkg; C coverage remains unavailable. Semantic input differences are documented per mutation in mutants.json.
