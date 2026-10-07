Built: native validator cores for the three reserved React rules, native post-dominance and an isolated raw type-name question; source-to-HIR integration remains unimplemented.
Commits: reservation 3e7267e2 and prior evidence fbbf0e59 were pushed before this code; implementation/evidence commit recorded in git history.
Commands/output: core agreement PASS 409.393s, 91 valid controls with 48 findings; prepared-HIR compiler77/repository287 match normally and under sanitizers; checker PASS 0.144s; Node PASS 16.614s; vet/gofmt clean.
Mutants: render unconditional guard byte 448, effect setter guard byte 1549, static creator propagation byte 56, adjacency-row alias byte 3115; all exit 0 with empty stderr and fail only complete-byte comparison; retained-handle and suffix mutants also caught.
Not covered: native source lowering/SSA, source compilation gates and memo shadowing, production harness integration, all upstream multifile contexts, full gate, and deterministic byte parity for Go's two-creator phi messages.

Scope and native implementation

No additional rules were claimed. This continues the three retained reservations:
react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
react-hooks/static-components. They have native validator code now, but are not
claimed as completed end-to-end native source ports.

Every new Adamic implementation is .a in wave21_react/. Each rule has its own
file. The HIR model carries function-local value identities, capture/context
edges, instructions, patterns, phi operands, source spans and nested functions.
The native post-dominator implementation runs Cooper-Harvey-Kennedy on the
reversed graph with return terminals feeding a synthetic exit. Its control
frontier and unconditional-block queries reproduce the Go shelf.

Set-state-in-render propagates setter functions through loads/stores, translates
captures into each nested function's value namespace, discards ordinary nested
findings, emits memo-callback findings, ignores optional calls and tests
post-dominance for unconditional calls. Set-state-in-effect preserves the first-
setter policy, synchronous callback reachability, useEffectEvent propagation,
ref-value and control exemptions, destructuring/default/rest taint, capture span
selection and recursion through nested graphs. Static-components retains the
single forward walk, phi creator propagation, function/call/new/method origins,
creation-source text, tag spans and Go's loop back-edge limitation.

Finding ranges, messages, repairs and suggestions use the existing canonical
Diagnostic serializer. These three Go rules produce no fixes or suggestions;
their zero counts are compared as part of the complete output bytes. No shared
registration generator, test harness, native parser or protected compiler file
was edited.

The new bridge question

react_type_names.go and wave21_react/type_names.a form the isolated pair.
The only shared source change is one dispatch arm (two formatted Go lines) in
facts.go. The question returns protocol version/mode, type-present, alias name
and type-symbol name. It does not decide whether a value is a setter, ref or
hook. Native predicates compare the names. Internal Go symbol spellings can
contain invalid UTF-8; the new question encodes their Unicode scalar projection,
matching JSON's replacement spelling, rather than putting invalid bytes in a
UTF-8 C buffer. Dispatch and RefObject are tested in their distinct alias and
symbol fields, and useEffect in the type-symbol field. Names are copied out of
the checker and cached by native syntax-node identity.

Validation boundary

The private --native-graphs fixture producer calls actual Go cohere Lower,
Construct, the memoization-erasure/inlining variant and compilation-unit selection.
It exports prepared HIR, source ranges and raw checker names into generated .a
fixture drivers. Native code then computes the rule findings. The expected bytes
come from a separate execution of unchanged Go production registry rules.
The expected-finding path imports no bridge implementation and computes no
native rule output.

This is prepared-HIR validation, not a native source frontend. The source unit
gates and memo annotations are fixture-provider dependencies. Its memo annotation
recognizes direct useMemo/React.useMemo syntax, and does not implement production
memo-shadow scope resolution. A real native source adapter must implement
Lower/Construct, capture mapping, standalone nested compilation units, the two
graph variants, source gates and the full memo scope predicate. That adapter is
still absent. Go lowering has not been placed in the production bridge as a
substitute for native execution. Static source inputs also still meet the shared
native JSX parser refusals documented in WAVE_21_REACT_REPORT.md.

Controls and independent failure detection

The extractor found 100 source strings from all three production test files plus
independent controls. Go's parser admitted 91. The source strings and valid flags
are saved in controls.json; all valid inputs are retained. React declarations use
Dispatch as an alias and RefObject as an interface, with both relative and bare
react module specifiers. This does not reproduce every upstream multifile context.

Twelve private batches compare 48 findings and 37047 total canonical bytes,
including file lines and per-batch count lines. Every batch matches normally and
under ASan, UBSan and LeakSanitizer with empty sanitizer stderr. The 77 compiler
roots match 5087 bytes and the 287 repository roots match 18485 bytes, normally
and under sanitizers, using Go-prepared inputs. These corpus checks likewise do
not establish native source lowering parity.

The first batch attempt compiled a 922779-byte generated module for several
minutes and was stopped explicitly after 271.145s. It was replaced by private
batches without dropping any valid fixture. A subsequent control caught a real
post-dominance bug: different predecessor rows reused one mutable empty array,
losing the loop-after-throw finding. Allocating a fresh row fixed the byte
mismatch. The old failure and oversized-build logs are preserved compressed.

Every native core mutant builds and exits 0 with empty stderr:

| Mutant | Independent comparison catches |
| --- | --- |
| Reverse render's unconditional membership guard | byte 448 |
| Reverse effect's typed-setter guard | byte 1549 |
| Delete call/new creator propagation | byte 56 |
| Reintroduce shared predecessor-row aliasing | byte 3115 |

The normal and sanitized raw-question probes return Dispatch and the observed
anonymous symbol spelling. After release both require exit 70 and exactly
`adamic: panic: invalid or released checker handle`. A retained-registry mutant
exits 0, demonstrating the lifetime assertion fails. Disabling the new question's
suffix guard compiles but fails the direct checker test with `malformed suffix
accepted`. These are separate from the core rule mutants.

Observed Go oracle ambiguity

An additional valid input assigns createA() and createB() on opposite incoming
branches and uses the joined value as a JSX component. Twenty-four executions
of unchanged production Go with identical source/config produced two distinct
canonical findings, with frequencies 21 and 3. The creation-site message alternates
between the two callees. Phi.Operands is a Go map, and the rule selects the first
dynamic operand it encounters. The native core preserves its supplied operand
order; it cannot promise the same random choice as a separately executed Go
process. Neither output was normalized and this input is not presented as a
passing byte agreement case. The exact source, all 24 outputs and summary are
preserved in validation-wave-21-react-core/.

Timing, commands and limits

Across the 91 controls, prepared-input native process wall time was 0.015188s;
independent full Go wall time was 0.259542s; Go fixture preparation was 0.309006s.
The native timing excludes source loading/lowering, while Go includes them.
These are different input pipelines and are not an end-to-end native speed
comparison. A meaningful native source-versus-Go time remains uncovered.

Test output was redirected directly to files:

    source /workspace/adamic-tools/env.sh
    ADAMIC_WAVE21_REACT_CORE_ARTIFACTS=/workspace/wave21-react-core-final ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave21ReactCores$' -count=1 -timeout=20m -v

PASS 409.393s. The complete checker package passes in 0.144s. The filtered
external Node oracle covers maps/text, sorting, string indexing, lone surrogates,
functions, closures, method closures, generics and number parsing across its
backends, plus TestTheOracleCatchesOneByte; PASS 16.614s. Touched-package vet and
gofmt output are empty. The full repository gate was not run.

Original setup timing remains tools 0s, submodules 0s, cache 81s, total 81s;
nproc is 5. No fresh setup was needed for this continuation. Complete canonical
outputs, prepared .a drivers, manifests, fixture text, mutants, refusals and hashes
are preserved in validation-wave-21-react-core/. Reservations remain retained,
not released; source integration and the exact Go ambiguity are outstanding.
