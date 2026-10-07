Built source-qualified stable symbols, function-local temporary/cache counters, and module-local main counters.
Commits: implementation 36fd309df907a4729e2f0bbf53a15e6dc65adfe9; landing merge 1ae93921 includes main 71d7e491; implementation merge db583013; emitter-speed 52f9bae.
Verification: native 369.479s, lower 39.845s, oracle 330.665s, split oracle 336.853s PASS; go vet ./... PASS.
Fifteen independent mutants fail their intended assertions, including the splitter's whole-program-counter mutant.
Not covered: one-module acceptance for added functions until developer-tools changes grouping and declaration dependencies; the full repository test gate.

Names use the entry file's directory as the program root. SourceIdentity retains the
relative, slash-normalized module path, declaration ancestry, concrete specialization,
and generated role. Bundled checker types and the prelude have fixed virtual identities.
Anonymous functions are numbered within their enclosing declaration. Repeated identical
named paths get a suffix counted only among that spelling and ancestry in their module.
A generated allocator, initializer or value adapter has a role distinct from a user
method with the same spelling.

Every name includes a deterministic 128-bit SHA-256 suffix over the full identity.
Escaping and truncation do not depend on encounter order. Long readable prefixes keep
both the module prefix and declaration tail. Native identifiers remain bounded, with
space for emitted suffixes below C11's 127 significant internal-identifier characters.
Local names retain their spelling after a decimal digest; those digits are not ordinals.
Compiler-generated source-less helpers use semantic IR identities, resolving function,
local, class, string and regexp references instead of whole-program allocation IDs.
Repeated helper bodies and class descriptors are shared only after equality checks;
conflicting definitions fail explicitly.

Strings and regexp bytecode use content identities. Regexp renaming scans identifier
tokens and preserves literal/comment bytes. Shapes, method adapters, comparator adapters,
JSON schemas and class descriptors have stable identities. Private accessor storage uses
the stable getter/setter name; public property bytes retain their original spelling.

Named hooks and territory

The one lower.go orchestration hook is moduleStatements. It records MainModules counts
while retaining the original statement sequence. sourceIdentity indexes declaration
ancestry before lowering order can affect numbering; functionSource and sourceTypeKey
supply specialization identity. Small registration hooks are in functions.go (signature),
locals.go (noteLocal), generic.go, class.go, class_inheritance.go, class_static.go,
class_accessors.go and expression.go (the value adapter).

The emit.go hooks are resetCounters and moduleMain. Each function/region variant resets
its temporary counter, inline-cache counter and temporary-name-keyed region facts.
moduleMain resets counters at retained module boundaries while keeping one scope,
initialization order and cleanup behavior. Main temporaries and labels have a module
namespace; file-scope cache names include their owning function/variant/module. The
additional emit.go naming hooks deduplicate identical declarations and print module
ownership comments. Field-store, switch-lowering and check-elision paths were not
restructured. No own changes were made to native.go, oracle_test.go, the splitter or grouping.
The current-main merge retained request-handler global lifetime behavior. Its newer
writeFieldSlot implementation superseded the older emitter-speed copy in emit_speed.go,
which was removed to resolve duplicate definitions. The old worker test now recognizes
main's data-write guard and includes its manually attached static layout in the IR, as
main's layout proof requires. These are named merge adaptations, not new lowering paths.
No cohere source was copied or changed.

Measured lint changes

Entry: stage1/cohere/lint/main.ts. Absolute normalized overlays target
stage1/typescript/parser/grammar.ts. The final before/after comparison uses the same
updated lint registry and splitter on main 4e0bfda5. The baseline is that exact commit,
compiled in an isolated scratch archive with cohere referenced through the existing
submodule. The final measurement uses the landed splitter directly. Neither its lexer,
splitC nor grouping is modified.

| Edit | Body/state units before | Body/state units after | Header before / after | Effective invalidations before / after |
|---|---:|---:|---|---|
| Add isolatedProbe before precedence | 53/53 | 38/53 | changed / changed | 53 / 53 |
| Add const isolatedProbe = [14].length in precedence | 40/53 | 1/53 | changed / unchanged | 53 / 1 |
| Replace return 14 with return [14].length | 40/53 | 1/53 | changed / unchanged | 53 / 1 |

units-before-current.log and units-after-current.log preserve those observations.
The earlier, smaller lint port at baseline 52f9bae had 29 units: body/state changes
29,20,20 before and 18,1,1 after; effective invalidations 29,29,29 before and 29,1,1
after. Its pinned developer-tools instrument was 14e8372816b1d3bf5bcc37ca57336e46d41b7a41.
Those historical logs remain separate; the 53-unit table is the landing result.

The add-function remainder is observed splitter behavior: inserting one function shifts
its groups and adds a prototype to the omnibus header. It is not acceptance-complete.
SPLITTER_HANDOFF.md specifies the required module grouping, generated-helper ownership,
main initialization extraction, per-unit declaration closure, single shared definitions,
and proof/ABI dependencies. The user requested that handoff for developer tools instead
of editing their code. Stable names are not permission to reuse stale dependent objects.

Verification and reproduction

Setup: bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh (the printed path;
/opt/adamic-tools/env.sh is absent). Timing lines: Go ready 0s; clang, Node and submodules
ready 1s; build cache warm and done 110s. nproc=5; cpu.max=400000 100000. Versions:
Go 1.27.1, clang 20.1.8, Node 24.19.0. setup-final.log preserves the output.

Test output was written to logs, never piped. Commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/stable-names-land-gate.log 2>&1
python3 internal/native/stable_emitter_evidence/split_overlay.py
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/stable-names-land-split.log 2>&1
go vet ./... > /tmp/stable-names-land-vet.log 2>&1
python3 internal/native/stable_emitter_evidence/reproduce.py > /tmp/stable-names-mutants-verified.log 2>&1
```

The initial split overlay supplied unchanged native.go and units.go from the pinned
developer-tools revision. Current main now includes that splitter, so the landing
command uses it directly. split_overlay.py emits an empty overlay when it is present;
reproduce.py uses the landed splitter rather than redeclaring its implementation. Results after the current-main merge: native PASS 369.479s; lower PASS 39.845s;
single-unit oracle PASS 330.665s; split oracle PASS 336.853s. go vet ./... passed, gofmt reported no files, and
Own source/format checks pass. The complete diff against origin/main includes
pre-existing diagnostic whitespace in the merged emitter-speed evidence; it was preserved. Logs are preserved as native-lower-oracle-current.log,
split-oracle-current.log, vet-current.log and format-final.log. Both modes run
all oracle fixtures, including Node output comparisons, counted builds and sanitizer
checks. The new naming tests cover escaped-module collisions, unrelated function/main
text, imported same-spelling class specializations, program-root relocation, generated
role collisions, adapter sharing and regexp preservation.

| Mutant | What catches it |
|---|---|
| Remove main's data-write guard | TestFieldStoresMatchNode detects the missing guard |
| Keep only the first 60 readable characters | Declaration tail disappears in TestStableIdentifierBoundsAndEscaping |
| Private accessor storage uses function index | TestSourceNamesDoNotMove observes changed unrelated module main |
| Remove generated role identity | TestGeneratedRolesDoNotCollide observes allocator/method collision |
| Deduplicate method adapters by IR index | TestMethodAdaptersShareStableDefinition observes duplicate definitions |
| Remove type declaration module qualification | TestImportedSpecializationsHaveSourceIdentity observes colliding specializations |
| Remove temporary reset | TestFunctionCountersDoNotMove and the splitter probe; equivalent 14 edit changes 38 units |
| Remove inline-cache reset | TestFunctionCountersDoNotMove observes moved cache |
| Remove regionValues reset | regions.a oracle executable reports ASan heap-use-after-free |
| Restore ordinal function names | TestSourceNamesDoNotMove observes unrelated text moving |
| Remove digest suffix | TestStableIdentifierBoundsAndEscaping observes escape collision |
| Remove module-main reset | TestSourceNamesDoNotMove observes unrelated module main moving |
| Replace regexp bytes globally | TestRegexpSymbolsDoNotMove observes changed literal/comment bytes |
| Include absolute program root | TestImportedSpecializationsHaveSourceIdentity observes relocation changing C |
| Remove declaration module qualification | TestSourceNamesDoNotMove observes escaped-module symbol collision |

All fifteen mutants were rerun after merging current main. The whole-program
temporary mutant makes the current splitter assert with 38 changed units.
All mutants use independent Go overlays and compile their test binaries. Assertions,
C naming validity or ASan catch the specified failure; unrelated build errors are not
accepted. reproduce.py requires the intended diagnostic for each mutant and enforces
one-unit invalidation for the equivalent temporary edit. It removes its own temporary
probe. The add-function count is reported, not asserted as satisfied. Raw diagnostics
are gzip-compressed with deterministic headers in stable_emitter_evidence.

The full repository go test ./... gate and the complete pinned TypeScript corpus were
not run. No rebuild-speed claim beyond the measured unit-change counts is made.

Landing note: 71d7e491 was merged as 1ae93921 after the green 4e0bfda5 gate.
Its changes are confined to stage3. git diff 4e0bfda5 71d7e491 --exit-code --
internal bridge go.mod go.work cohere stage1 exits 0: compiler dependencies,
oracle fixtures and measured lint inputs are identical to the green snapshot.
No repeat of the unchanged tests is claimed. The branch is pushed only to
codex/stable-emitter-names; no PR was opened.
