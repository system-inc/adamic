Built: numeric symbol-description listener; atomic transfer/meet kernel; require-await syntax kernel; three manifests.
Commits: React parking bd93c67d3; fourth-batch claim fcfd9568c; implementation/evidence follows.
Commands and outputs: symbol listener 17 controls/6 findings and both corpora byte-identical; 16 atomic and 8 await kernel controls agree with production Go.
Mutants: symbol argument-count inversion, atomic refresh deletion and await-using mask mutation caught by byte comparison; new handle-retention mutant caught by required panic.
Not covered: native shared-driver integration, atomic full rule and require-await full rule; these three claims remain incomplete.

# Fourth batch partial implementation

The prior six ports are green on origin/main f8013f0b and pushed. The three React
claims are explicitly parked: globals awaits JSX integration; immutability and
no-deriving-state-in-effects await React HIR/SSA/capture passes plus JSX. Ahra
identified #dnv6f2c and area/stage1-lint as the dependency owners. Those branches
were not modified or pushed by this unit.

After fetching 529 origin refs, the volume ranking had 197 checker-dependent
entries, 154 named in origin claims and 25 already ported on main/bridge. React
and structure/react-hook-no-any-type entries were skipped while React is parked.
The next three non-React candidates were require-atomic-updates, require-await
and symbol-description, each with zero compiler/repository volume. An immediate
pre-claim refresh found none reserved; the claim was pushed before code.

## Symbol listener

symbol_description/rule.a takes a supplied HandedNode with numeric SyntaxKind
and UTF-8 ranges. It neither reads a string kind nor retrieves the current node
from a parser. rule.json listens to CallExpression (214). The raw bridge question
call-symbol-shape, in its own Go and .a files, supplies argument count, the
parenthesis-unwrapped callee's numeric kind/text, symbol presence and each
resolved declaration's declaration-file flag. The Adamic handler chooses the
finding and exact production message. No fix or suggestion is emitted.

The isolated test driver receives numeric syntax captures from the independent
Go oracle. It is deliberately a driver replacement for these tests: it is not
proof that the shared native parser supplies the right nodes or spans. Production
Go rules remain unchanged and import no bridge code. The listener matches all
17 controls (6 findings, 2812 bytes), 77 compiler roots (0 findings, 5318 bytes)
and 287 frozen repository roots (0 findings, 18485 bytes). All three comparisons
also match under ASan/UBSan and leak checks. Tests include shadowing, parentheses,
optional calls, new/member calls, explicit undefined, spread and Unicode.

The argument-count mutant compiles, exits 0 with empty stderr and disagrees only
at byte comparison. Released call-symbol-shape inspection panics with exit 70 and
`adamic: panic: invalid or released checker handle`. A registry-retention overlay
mutant exits 0 with empty stderr, proving that check can fail. Checker package
tests PASS 0.123s. No shared registration file or harness was changed.

Three interleaved process timings per corpus, with captured syntax generation
excluded: compiler Go median 0.292923093s, native 0.821113525s (2.803x); repository
Go 0.126508479s, native 0.168059858s (1.328x). Both processes load a Go checker;
Go traverses its AST, while native consumes the capture. These observations are
not a full native-parser benchmark or a claimed speed improvement.

## Atomic kernel

require_atomic_updates/state.a ports the production fresh/outdated state,
clone, union meet and per-event transfer. Both sets travel across joins; a read
refreshes its symbol, a suspension stales fresh symbols and a write tests the
outdated set. Numeric event kinds and checker symbol IDs carry identity.

Sixteen transfer/join scenarios agree byte for byte with the unchanged production
lattice through a test-only Go overlay. Original and refresh-deletion mutant
compile and exit cleanly under sanitizers; Go bytes catch the mutant. Full rule
replay still needs numeric syntax/event delivery, deferred suspension/store
placement, binding resolution, escape filtering, fixed-point solving and final
finding rendering. The kernel is not presented as a complete rule.

## Await syntax kernel

require_await/syntax.a ports the empty-body and await search predicates using
numeric node kinds, declaration flags and a supplied indexed node graph. It
recognizes await, for-await and the composite await-using mask, and skips nested
functions/classes. Eight source controls parsed independently by typescript-go
agree with the unchanged production predicates. Two additional native-only
controls exercise an empty block and concise await. The mask mutant compiles
and exits cleanly under sanitizers, then fails the eight-case Go comparison.

Full require-await replay still needs shared numeric function/parent delivery,
checker type/signature facts for contextual and declared-generic promise demand,
heritage demand, exact head/name rendering, async-token repair and suggestion
serialization. No full-rule oracle, fixes, suggestions or timing is claimed for
this rule. The manifest lists the seven production function kinds.

## Current blocker and scope

The shared parser on this branch still has ParseNode.kind:string. There is no
shared numeric node-delivery API here. New listeners must consume a supplied
node, so the new code declares numeric kinds and avoids adding old string-based
per-rule dispatch. Numeric delivery and the shared Diagnostic integration are
pending with their owners. Full native integration cannot be validated against
the shared driver until that API arrives. The additional atomic/await work
listed above is unfinished implementation, not a claim that their entire
semantics are blocked by the harness alone.

No new batch may be claimed until these three are completed or explicitly
parked. Evidence, timings, independent captures, exact streams and mutants are
in validation. verify_symbol.py documents the runnable listener checks; the
kernel .a entries and Go overlay helpers document the smaller comparisons.
All new Adamic source files use .a. The preceding six-port oracle gate was not
repeated for this additive partial batch; only the new checker package and the
new listener/kernel checks were run.
