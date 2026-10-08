# Open checked casts

This is a classification and fixture baseline for #kbcj5nr, awaiting the broader
checked-cast compiler feature #b5w3ycg. It changes no compiler or adaptation.
The supported sub-union cast is held against source on Node and both backends.
The other nine pairs retain their actual compile-time refusal and their required
runtime contracts. A refusal is not evidence of a runtime check.

## Population and counts

Input ledger: 855bcfaacb8919a0d61ec34eda785fce9fde8ba2.
Input source: Microsoft TypeScript v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8.
Worker base: origin/main 45487a809f89885a3fc651cd590e7dabf31362dc.

**4,228 is the ledger's stock open-obligation total, not its cast total.** It
contains 3,957 open as-casts, 112 explicit-any tokens and 159 any-typed
declarations. Every one of the 3,957 stock casts is classified exactly once in
classification.json. outside-stock-casts.json accounts for the other 271 sites
with locations and reasons.

The ledger also reports 31 adapted-only open casts and two added explicit-any
tokens. Those are retained with their adapted locations and target types in
outside-stock-casts.json; they have no stock AST anchor and are not included in
this stock classification. Consequently the ledger's current totals are 3,988
open casts and 4,261 open obligations. We do not relabel any obligations as casts
or silently mix stock and adapted source coordinates.

| Class | Stock casts | Two fixture sites, stock file:line |
| --- | ---: | --- |
| Tagged downcast | 2,051 | binder.ts:359, binder.ts:669 |
| Untagged downcast | 1,292 | builder.ts:610, checker.ts:6852 |
| Literal or enum target | 26 | moduleNameResolver.ts:2118, parser.ts:4566 |
| Generic or type-parameter cast | 240 | builder.ts:545, parser.ts:3119 |
| Other | 348 | builder.ts:1466, transformers/generators.ts:3057 |
| Total | 3,957 | |

All paths above are under src/compiler/. The JSON supplies complete paths,
columns, source expressions, target types, reasons and discriminant fields for
every cast, together with the ledger's observed Adamic proof state. These are
classification candidates for the feature, not claims that the target's complete
payload is proved by a tag.

classify.cjs uses typescript@6.0.3's compiler API. It checks every stock source
file's ledger SHA-256, resolves complete start/end AST spans, and requires zero
semantic diagnostics before classifying. It never matches source with regex.
The enum classifier distinguishes a whole enum from a proper member subdomain,
even when the API gives both the same enum symbol. Whole flags enums and
unchanged field domains do not identify a subtype. A Boolean target is treated
as a primitive, rather than a literal refinement.

The disjoint precedence is:

1. Generic: source or target contains an unresolved parameter, indexed access,
   conditional or substitution type, including callable parameters and results.
2. Literal/enum: the target is a finite literal or enum domain.
3. Tagged: the non-null target's object members share a required finite literal
   or proper enum-member field whose domain refines the source. The field must
   be present in each target member; a whole enum flags domain is excluded.
4. Generic: the target is a concrete parameterized alias, array, tuple or
   reference. This bucket explicitly includes concrete instantiations.
5. Untagged: non-null target members are non-callable objects or object
   intersections with fields, without a discriminant refinement.
6. Other: remaining primitive, branded, callable, empty or mixed targets.

These distinctions concern the resolved target form. In particular, the literal/
enum bucket includes bitmask initialization such as checker.ts:11190's
`0 as TypeFlags`. Classification does not assert that comparing with declared
enumerator values is a valid runtime check for such a flags domain. Brands and
callable signatures in Other likewise require their own implementation decisions.

The classifier was run twice on the same pinned tree. Outputs were byte-identical.
Coverage verification independently reads the original ledger and checks exact
identity coverage, totals, and the ten fixture sites' expected categories.

## Fixtures and observations

Each numbered fixture keeps the real cast expression and reduces its surrounding
support types and driver. These are reduced witnesses, not copied upstream
function implementations. Tracing, parser state, diagnostics recovery and other
unexercised work is omitted. Enum domains are reduced; 02 substitutes string
literal tags for the upstream SyntaxKind enum tags while retaining the exact
`name as Identifier | StringLiteral` cast. The manifestation of the check, rather
than TypeScript's entire AST implementation, is what these witnesses isolate.

fixtures.json gives a source site, category, Node stdout, required runtime stdout
and exit status for all ten positive programs and ten failing twins. Both twins
run on Node with exit 0 because source casts erase. Their intentionally different
Adamic contract is an owned `adamic: panic: cast failed:` message and exit 70.
For tagged casts the check precedes the cast result. For checked views it precedes
the invalid read. 04 and 07 first read a valid value, then mutate through another
alias and read again, proving that a cast-time-only check would be insufficient.

| Pair | Real expression | Observed main behavior |
| --- | --- | --- |
| 01_enum_declaration | node as EnumDeclaration | Both refused: broad Node source has no admitted check plan |
| 02_name_subunion | name as Identifier \| StringLiteral | Both build; good case agrees with Node, failing case panics before cast result |
| 03_chain_info | chain as ReusableRepopulateInfoChain | Both refused: untagged payload view |
| 04_literal_payload | type as StringLiteralType | Both refused: untagged payload view, including an alias change between reads |
| 05_extension | tryExtractTSExtension(candidate) as Extension | Both refused: standalone enum target |
| 06_parser_keyword | currentToken as SyntaxKind.WithKeyword \| SyntaxKind.AssertKeyword | Both refused: standalone enum-member union |
| 07_generic_next | chain.next as T[] | Both refused: generic array view, including an alias change between reads |
| 08_generic_node | consumeNode(node) as T | Both refused: generic structural view |
| 09_primitive_string | value as string | Both refused: primitive refinement |
| 10_generator_label | args[0] as Label | Both refused: primitive refinement of an indexed value |

observe.cjs executes independently stock-transpiled source on Node for all 20
files. It builds supported programs with ASan and UBSan, executes the JavaScript
backend with the repository's oracle/adamic.mjs runtime, and checks both against
the appropriate source or panic contract. Successful native programs also run
with LeakSanitizer. Leak reporting is disabled for the deliberate panic, whose
contract stops without cleanup. Complete commands, diagnostics, output and
statuses are in observations.json; detailed generated artifacts remain in scratch.

Observed: **20 Node goldens pass, two runtime contracts pass, 18 runtime
contracts are blocked before emission**. verify.cjs passes the measured baseline;
`verify.cjs --require-runtime` exits 1 and names all 18 blockers. That failing
command is deliberate evidence that the broader feature is not implemented here.
A feature implementation should remove the refusal headers as these sites become
admitted, refresh observations, and make the strict runtime command pass.

The first-line a-check headers are truthful about current compilation. The two
supported programs say checked. The 18 blocked programs expect the exact
adamic/no-unchecked-cast refusal. The unchanged Gate.aCheck from fast-gate
914ea6d7d3ef0aeca2ac68baa76eb4604556479b passes all 20. There are no checker-error
or NotYet substitutions hiding a runtime failure.

## Mutants and limits

| Experiment | Number | Check that caught it |
| --- | ---: | --- |
| Change each positive program's final output marker | 10 | Independent Node stdout golden; each mutant still exits 0 |
| Remove the supported failing twin's real emitted JavaScript cast check | 1 | Panic exit and stdout contract; mutant continues with exit 0 |
| Remove the same check's panic from real emitted C | 1 | Panic exit and stdout contract; valid sanitized C runs exit 0, leak-clean and with no sanitizer report |
| Remove each expected-refusal header | 18 | Unchanged Gate.aCheck expected-error predicate |
| Replace each header reason with an unrelated rule | 18 | Unchanged Gate.aCheck expected-error predicate |
| Drop one classified cast | 1 | Independent exact ledger coverage |
| Reclassify the first tagged fixture as Other, updating both counts consistently | 1 | Independently pinned fixture classification |

Additionally, all ten failing source twins provide erased-cast negative controls
for their runtime contracts. Each continues on Node and is rejected by the
required panic/70 comparison. These are **not ten compiler implementation
mutants**: nine pairs cannot yet emit code. Their missing-check implementation
mutants, full runtime success and sanitizer behavior remain blocked by #b5w3ycg.
The one supported failing twin has actual check-removal mutants in both backends.

Counts are recorded locally in counts.md. These stage3 fixtures are not new
entries in internal/oracle's compiled fixture registry, so its table is unchanged.
No whole package suite, full gate, native tsc execution or checked-cast compiler
implementation is claimed. The 31 adapted-only casts have not been classified
against an adapted source tree by this unit.

## Reproduce

Run from the repository root. Fetch the immutable ledger and fast-gate refs if
those objects are not present. The stock checkout must contain src and scripts
from the pinned upstream commit, with upstream diagnostics generated using the
relative-path command shown below. Its node_modules must resolve
@types/node@25.3.3 and @types/source-map-support@0.5.10. The classifier's API is
stock typescript@6.0.3, installed by stage3/api's lockfile.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step09-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH="$PWD/stage3/api/node_modules"
# In the pinned stock tree:
node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json
# Back in Adamic:
node stage3/fixtures/checked-casts/classify.cjs /tmp/step09-stock /tmp/classification.json > /tmp/classify.log 2>&1
cmp /tmp/classification.json stage3/fixtures/checked-casts/classification.json
go build -o /tmp/step09-adamic ./cmd/adamic > /tmp/step09-build.log 2>&1
node stage3/fixtures/checked-casts/observe.cjs /tmp/step09-adamic /tmp/step09-observations > /tmp/step09-observe.log 2>&1
node stage3/fixtures/checked-casts/verify.cjs > /tmp/step09-verify.log 2>&1
node stage3/fixtures/checked-casts/verify.cjs --require-runtime > /tmp/step09-runtime-contract.log 2>&1
# Expected exit 1 until the 18 compiler blockers are resolved.
git show 914ea6d7d3ef0aeca2ac68baa76eb4604556479b:cloud/fast-gate/run.py > /tmp/step09-gate.py
python3 stage3/fixtures/checked-casts/a-check.py /tmp/step09-gate.py /tmp/step09-a-check > /tmp/step09-a-check.log 2>&1
node stage3/fixtures/checked-casts/native-mutant.cjs /tmp/step09-observations /path/to/sanitized/runtime/cache > /tmp/step09-native-mutant.log 2>&1
```

The native-mutant runtime argument is the directory containing adamic.h and the
sanitized runtime.a produced by the preceding build. The recorded invocation is
in native-mutant.json. On this box the directory was
/home/agent/.cache/adamic/runtime/ba946f7ff2626c5ad6da2b5b463951a80f79bbf59bce52e65eb6f0bc3aa1ff20.

Setup printed Go 1.27.1, clang 20.1.8 and Node 24.19.0. Its timing lines were:
Go ready 1s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm
51s, done in 51s on 5 processors. nproc was 5; cgroup CPU quota was 4. Setup ran
before switching the submodule to the new main pin, so the first compiler build
reported the missing cohere/rule_runner package. `git submodule update --init
--recursive` aligned cohere to main's 7945d102 pin; the subsequent compiler build
passed. No module or compiler file was edited to work around that mismatch.

Source expressions and reduced support declarations derive from Microsoft
TypeScript, copyright Microsoft Corporation, licensed under Apache-2.0. Complete
upstream source remains an external scratch input, not vendored into Adamic.

## Latest-main lane follow-up

Merged origin/main 73352e874ddbda5a78c95c5c670860d43275e4d0. The default `stage3/lane/run.sh` and independent `stage3/lane/check.py` both pass: 106,366 passing, exactly one sanctioned Public APIs failure, zero pending; lane wall time 687.275 seconds. The sole baseline diff is api/typescript.d.ts. All ten emitted JavaScript artifacts and the public API bytes match the main reference. The instrumentation audit finds zero hooks. See [lane-verification.json](lane-verification.json) and evidence/latest-*.log.

The historical red lane did not reproduce; no source cause is claimed and no compiler behavior was changed. Refreshed focused observations confirm 20 Node goldens, two supported runtime contracts, 18 unchanged refusals and 21 controls. Ledger mutants and a real native check-omission mutant are caught. The compiler was built from production sources identical to this merged main. No whole package/full gate confirmation was run.
