Built: react/static-property-placement and react/style-prop-object in separate native .a modules with named listener manifests.
Commits: claim 14ba089fb pushed before code; tested native code 5e8c9523e4969be8796e645523a63cf98d4aa546; based on main c01907a70.
Checks: 73 valid controls, 45 default findings; all six options matrices, both corpora, sanitizers, checker, production Go, vet and filtered uncached Node pass.
Mutants: placement byte 5981, style literal byte 321 and raw-kind byte 49; each compiles, exits 0 with empty stderr, caught only by Go bytes; stale registry violates exit 70.
Not covered: four parked HIR/capture rules, shared-driver/native-JSX-parser integration, arbitrary JSX applications, suppression/edit application or the full root gate. No third unclaimed ranked rule exists.

## Selection and ownership

All existing wave-18 work was landing-ready on main c01907a70 at a37d18a51.
Fetching every origin branch yielded 576 refs and 33 distinct claim Markdown
blobs. The same 197-rule VOLUME_REPORT ranking, combined compiler/repository
volume then lexical name, has only this final pair unported on the two baselines
and unclaimed on every origin branch. The audit and baseline port associations
are preserved in selection.json. Neither has a named source port on any origin
branch. No HIR, SSA or capture analysis is required. There was no third rule to
claim. After reserving this pair no unclaimed entry remains in that ranking.

The claim was committed and pushed before implementation. Only
codex/typeaware-wave-18 is a push target. Each rule has its own index.a and
rule.json; manifests use typescript-go ast.Kind names exactly. Placement listens
to PropertyDeclaration, GetAccessor and BinaryExpression. Style listens to
JsxAttribute and CallExpression. Distinct supplied-node methods receive the
already selected node; neither rule compares string kind names to decide primary
node relevance or refetches its primary node. The owned standalone driver indexes
callbacks by the pinned bridge enum values. Those values are an internal facts
encoding, not registry manifest values. Six earlier owned manifests were also
corrected from numeric entries to named entries under Ahra's updated contract.

The new question component-property-syntax-facts has separate Go and Adamic
files and one registration line in facts.go. It exports numeric syntax identity,
source byte spans, decoded text units, structural links, symbol declarations,
type-annotation presence, static modifiers and heritage tokens. It emits no lint
judgments. Shorthand value symbols are exported rather than the property's own
symbol. The raw parser and checker still run in typescript-go through the C
bridge; this does not port the native JSX parser. No shared registration generator,
test harness or protected compiler implementation was edited.

## Rule behavior and independent observations

Placement reproduces the six-property order, typed props/context aliases,
non-static getter silence, parenthesized React bases, class-expression receivers,
merged declaration search and class-scope assignment exemption. Its native API
supports a default position and per-property overrides. The source driver exposes
field, getter, assignment and displayName-override configurations for comparison.

Style reproduces literal versus null/template/unary/object classification,
parenthesis unwrapping, first-declaration initialization, shorthand value symbols,
exact allow-list skipping, member/namespaced tags and distinct JSX/value/identifier
report anchors. The createElement binding/import/initializer gates run natively.
Neither production Go rule uses a regex, so no regex translation is needed.

The independent Go overlay imports the unchanged production registry and loads
its own checker program; it imports no bridge implementation. Controls combine
extracted Go fixture source strings and independent class/Unicode/CRLF, bigint,
regex-literal, shorthand, declaration-order and shadowed-callee cases. All 73
candidate sources parse, so no parse exclusion was needed. Default findings are
9 placement and 36 style. All complete records have zero fixes and suggestions,
verified as fields and compared byte for byte.

| Population/options | Findings |
| --- | ---: |
| Controls/default | 45 |
| Static getter | 46 |
| Property assignment | 43 |
| Style allow list | 25 |
| Getter and allow list | 26 |
| Per-property override | 46 |
| Compiler, frozen 77 sources | 0 |
| Repository, frozen 287 sources | 0 |

Every row matches normally and under ASan/UBSan/LeakSanitizer, with empty native
and sanitizer stderr. Default controls match 8999 bytes; compiler 5318 bytes and
repository 18485 bytes. Headers use the recorded absolute paths. Full stdout,
stderr, controls, manifests, source hashes and command journals are archived.

## Mutants and build probes

- Placement reverses expected-position equality: compile succeeds, exit 0,
  empty stderr; only independent Go bytes catch difference 5981.
- Style shifts the literal-kind membership by one: compile succeeds, exit 0,
  empty stderr; only independent Go bytes catch difference 321.
- The bridge question shifts exported raw kinds by one: compile succeeds,
  exit 0, empty stderr; independent Go bytes catch difference 49.
- A query after release exits 70 with invalid or released checker handle.
  Removing the registry deletion makes it exit 0 with empty stderr; the required
  released-handle contract catches that mutant.
- The filtered uncached Node suite includes TestTheOracleCatchesOneByte.

The first copied helper incorrectly used this.facts inside ComponentFacts and
was refused by type checking. Its receiver was corrected before successful
native builds. A first proposed style mutation broke the narrowing guard and
was refused by type checking; it is not counted. The replacement literal-kind
mutation meets the byte-only criterion. No compiler or shared harness change
was made to work around either development probe.

## Commands and timings

    bash cloud/setup.sh > /tmp/wave18-component-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/typeaware/wave_18_component_props/validate.py --scratch /workspace/wave18-component > /tmp/wave18-component-final.log 2>&1
    python3 stage1/cohere/typeaware/wave_18_component_props/validate_bridge.py --scratch /workspace/wave18-component > /tmp/wave18-component-bridge.log 2>&1
    go test ./bridge/tsgo/checker -count=1 -timeout=10m
    go -C cohere test ./internal/lint/rules/react -run '^(TestStaticPropertyPlacement|TestStylePropObject)' -count=1 -timeout=10m -v
    go vet ./...
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw)\.a$' -count=1 -timeout=15m -v

The validator bootstraps fresh binaries and archives when native-asan is absent.
The corrected build created both normal and sanitized artifacts; the final run
uses those unchanged source artifacts with the compiling replacement mutant.
The corpus variables are ADAMIC_WAVE18_COMPILER_CONFIG pointing to
/workspace/wave-18-typescript/src/compiler/tsconfig.json,
ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest,
ADAMIC_WAVE18_REPOSITORY_CONFIG=/workspace/adamic/tsconfig.json and
ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest.
Checker PASS 0.125s, production Go PASS 0.099s, vet exit 0, filtered uncached
Node PASS 1.203s, with native 19 misses and Node 13 misses, zero cache hits.
Setup prints Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 41s, done 41s;
nproc 5, CPU quota 4. All test output goes directly to log files.

Three alternating isolated whole-process measurements after checks completed:

| Corpus | Native median seconds | Go median seconds |
| --- | ---: | ---: |
| Compiler | 4.055362 | 0.308105 |
| Repository | 0.557130 | 0.129653 |

Native remains slower. The earlier concurrent samples were 3.920241/0.287815
and 1.049837/0.347145; they are retained but not used as the isolated result.
The bridge batches raw syntax/declarations per file; its broader fact graph adds
serialization and native decoding work. The measurements establish total time,
not an isolated attribution of the slowdown.

## Remaining scope

The four parked claims remain blocked on native HIR/SSA/gates or callback/callee
return and capture/escape analysis, owned by #dnv6f2c. JSX integration is landing
on area/stage1-lint. Shared harness ab70f38d4 retains the seven-argument report
and wire contract; this unit does not edit its adapters. The full root gate,
expanded stage3 population, suppression engine, edit application and arbitrary
JSX projects were not run. The fixed corpora, source controls and six option
matrices are the parity evidence. Both newly reserved rules are complete for
that scope, and there is no further unclaimed ranked rule to take.
