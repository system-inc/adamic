Built: native case-equivalence groups and per-rune lookup for the existing no-invalid-regexp dependency.
Commits: follows pushed 8725848fa on codex/typeaware-wave-09; current main f8013f0b is already an ancestor; no new claims.
Commands/output: invalid_regexp/verify_equivalence.py PASS against production Go, normal native, sanitizers, source Node and emitted JavaScript.
Mutant: discard groups with fewer than three members instead of two; successful native build/run is caught only by the output comparison.
Not covered: class widening integration, rewrite or exact regexp engine, shared numeric handed-node execution, full-rule parity or another corpus sweep.

The component builds groups from the existing Unicode 17 primitive uppercase
and simple-fold mappings. It does not embed production group verdicts.
Each seed is canonicalized by the previously verified native canonicalizer.
Groups with fewer than two members are discarded. Lookups return an empty
array for isolated runes; production Go returns nil, with the same no-members
meaning. Callers must not mutate returned groups. Production Go group order
comes from a map; the oracle sorts it by the first member for comparison,
and native exposes that same stable order. Group membership order is numeric
on both sides. The oracle calls unchanged production CaseEquivalents and
CaseEquivalenceGroups, not a second reimplementation.

For both flag modes, the probe enumerates complete groups, then queries every
integer from -1 through 1114112, emitting each nonempty result. This covers
2,228,228 membership queries, including all Unicode scalars, surrogates and
out-of-range boundaries. Absence of a line certifies the empty lookup too.
The same bytes pass ASan, UBSan and LeakSanitizer, source Node and emitted JS.
The pair-dropping mutant builds and exits 0 with empty stderr before differing
from Go. Source hashes, results, complete stdout/stderr and generated probes
are retained in validation-equivalence.

Reproduce, with test output written to a log:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_equivalence.py > /workspace/wave-09-equivalence-test.log 2>&1
```

Whole-process table construction and exhaustive lookups took native 0.113926
seconds versus Go 0.064169 seconds; sanitized native took 0.765218 seconds.
These are single component runs, not a full lint performance measurement.
Setup remains the existing successful 88-second run; nproc 5.
No node kinds are read, no nodes are fetched and no checker ABI is changed.
The existing six numeric rule.json declarations remain intact. Shared files
were not edited. The shared string-only parser and missing handed-node
contract still block migration of legacy rule execution. This dependency is
not evidence that the complete no-invalid-regexp rule is ported: the exact
pattern engine remains unfinished. No additional rules were claimed.
