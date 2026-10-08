# Regex protocol implementation

The unit claim was committed/pushed first (`c5045bd`). The requested separate
literal-freshness fixture and semantic hoisting mutant were subsequently pushed
to codex/regex-matcher as `50a1dc8217f6d715fc27038e6ce12ab429416a30`, directly above
df959ee, then merged into this branch (`1e5258b`). No protocol implementation
commit preceded the requested matcher fixture.

Implemented: intrinsic Symbol.match/replace/search/split/matchAll on a proven
RegExp and their RegExp.prototype method .call forms; default String dispatch;
zero/one-parameter string replacement callbacks with collect-before-callback
semantics; exact builtin exception identities for seven constructors, including
runtime SyntaxError from invalid constant RegExp constructors and TypeError
from String global validation; named/absent d indices, pair identity and UTF-16
positions; default RegExp.toString/string concatenation. Catch/finally exception
edges and ownership are included. Diagnostic-name mutation cannot change the
constructor identity.

Runtime patterns/flags remain explicitly refused: the native runtime has no
ECMAScript pattern compiler. The Go compiler is used for every accepted static
pattern; no alternate regexp grammar or silent constant freezing is used.
Observable V8 SyntaxError diagnostic text is refused, including caught values
that escape to helpers. The source/flags conversion uses only the proven default
RegExp representation; custom overrides are refused.

Not implemented: custom Symbol hooks, overridden exec, species/getters, generic
receivers, first-class prototype method values, callback capture/index/input/
group parameter layouts, and general array reflection/methods on indices pairs.
These remain refusals. Four stock-TypeScript-valid String custom-hook forms and
constructor/exec/freeze overrides have explicit refusal controls. Stock TypeScript
rejects the structural matchAll hook and the string replacement .call overload;
those exclusions are listed separately. All 223 baseline diagnostic refusals
are independently rejected by stock TypeScript 6.0.3 and listed by path/code.

The runner's classifier, verdicts, and constructor-assertion guard are unchanged.
`HANDOFF.md` explains why intrinsic protocol fixtures do not turn currently
classified Symbol skips into passes, and why guard-refused rows have not run
untouched test262 on Node. Protected native/lower/oracle implementation files
and the cohere pattern corpus are not edited.

Validation on the final implementation:

- 127,369 test262 matcher executions plus 10,000 fixed-seed differential pairs,
  each in three Go and three C configurations: zero disagreements, zero
  unavailable properties. Go tests 17.303s; native tests 486.393s under load.
- Complete affected package gate: load 5.403s, lower 55.433s, flow 238.159s,
  fresh 91.016s, runner 238.187s; ir/javascript have no separate test files.
- Remaining native package tests: 266.642s.
- Existing/new regex and exception oracle fixtures, original Node, JavaScript
  backend, sanitized/release C, leak checks and complete recorded-counts gate:
  green in 96.072s.
- Freshness fixture after merge plus complete counts and targeted counts controls:
  green in 49.186s.
- Thirteen protocol feature mutants all caught, plus the separate retained-cache
  literal-hoisting mutant. Actual stdout witnesses/exception-edge/refusal checks
  are archived in logs; scripts restore every mutated source.

The full 1,879-file normal-runner survey is still running at this implementation
commit. Its final outcome table will be recorded in a separate report commit;
no partial progress is presented as a final total.

Counts: one existing exception fixture gains one balanced retain/release;
non-global String validation now allocates its TypeError; all new completing
fixtures allocate/free equally. The complete counts gate passes after recording
these changes. A wrongly filtered counts run and a table-order mismatch were
corrected by running the complete gate; no unrelated rows were rewritten.

Setup timings: go/clang/node/submodules 0s each; warm build cache 119s; done119s.
`nproc=5`, CPU quota 4, 17.6GB memory. Node24.19.0, Go1.27.1, clang20.1.8.
