Step 31 scouts the real TypeScript 6.0.3 binder, checker and emitter. The delivered
implementations are Node binder, checker and emitter observers, with fixtures,
mutants and native process comparators. REPORT.md retains the initial research;
COMPONENTS.md specifies the checker/emitter protocol and changed-input workflow.
COMPONENTS-REPORT.md records their measured results. TRAIN-REPORT.md records
the named views-train rebuild, 301-project confirmation and native first stops.

Run after sourcing the environment printed by cloud/setup.sh:

```sh
export STEP31_TYPESCRIPT="$HOME/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js"
bash stage3/scouts/step31/run.sh /absolute/pinned-adapted-tree /tmp/new-step31-proof > /tmp/step31.log 2>&1
```

The adapted tree must be constructed from this unit's pinned main with apply.sh.
The stock API is the integrity-checked 6.0.3 dependency installed by apply.sh.
The command runs fixture assertions, checker/printer/emitter seam probes, three
actual binder-body mutants, three seam mutants, and three process-output mutants.
It then compares the real adapted binder with the independent stock binder on
301 acceptance projects and 6,262 selected upstream compiler/conformance cases.
Every subprocess writes stdout, stderr and exit separately. Source mutations use
new disposable copies, never the input tree. Existing result paths are refused.

A native implementation receives the same request pathname as its only argument:

```sh
bash stage3/scouts/step31/run.sh /absolute/pinned-adapted-tree /tmp/new-native-proof /absolute/native-binder > /tmp/step31-native.log 2>&1
```

The native slot is optional and never impersonated by a Node forwarder in the
reported results. Without it, report.json says native_run=false. compare.py can
also compare a smaller fixture directly:

```sh
python3 stage3/scouts/step31/compare.py --output /tmp/new-binder-comparison \
  stage3/scouts/step31/fixtures/golden.jsonl -- \
  /absolute/native-binder stage3/scouts/step31/fixtures/projects.json > /tmp/compare.log 2>&1
```

Input is UTF-8 JSON: {"projects": [{"id": string, "options": object, "files":
[{"path": relative virtual path, "text": string, "scriptKind": optional numeric
TypeScript ScriptKind}]}]}. Options use TypeScript's JSON option names/values.
Each file is freshly parsed with parents enabled and passed to bindSourceFile.
The binder does not resolve imports or create a checker. Its alias symbols are
observed as bound aliases, without asking the checker to resolve their targets.
.a input names are exposed to the parser as .ts; output retains their .a names.
The default file kinds are TS, TSX for .tsx and JS for .js. Explicit scriptKind
supports other parser modes. Diagnostics are output data, not process failure.
Malformed requests and unsupported compiler options fail the process.

The stable output contract is adamic-binder-v1, one JSON object per line with a
final LF. JSON property order is the order in binder-dump.cjs. Project order is
request order, files sort by virtual path, and options/table keys sort by UTF-16
code units using JavaScript's string comparison. JSON strings use JSON.stringify,
so embedded newlines do not create extra records. The observer never prints
absolute execution paths, memory addresses or compiler allocation IDs.

A project record contains format, TypeScript version, project id and options.
A file record contains the virtual path, AST node rows, symbol rows, parse
and bind diagnostics. AST IDs are preorder indices from forEachChild, including
node-array children and attached JSDoc. Positions are UTF-16 code-unit offsets;
negative positions, when present, retain their meaning. Node rows contain kind,
pos/end, flags, symbol/localSymbol IDs, sorted locals and optional flowFlags.
They do not claim a complete control-flow graph comparison.

Symbol IDs are assigned at first encounter through AST rows and sorted tables,
then through symbol rows in queue order. Rows contain escapedName as name,
numeric flags, ordered declaration AST IDs, valueDeclaration, parent,
exportSymbol, sorted exports/members and constEnumOnlyModule. Missing references
and tables are null; missing declarations are []. IDs preserve shared symbol
identity and can represent cycles through parent/table references. A declaration
outside the observed AST fails loudly. Symbol allocation id, links caches and
Debug methods are omitted because they are runtime/compiler-state details.
The symbol row's declaration references make the source name recoverable from
its span and the input text; this format does not attempt full type dumps.

Diagnostics retain numeric category/code/start/length, string or recursive
message-chain structure, and related-information order. Binding is file-local;
this first format does not include cross-file checker diagnostics. Exact bytes,
empty stderr and exit zero are independently checked by compare.py. There is no
whitespace, newline, message, or path normalization on actual output.

Fixtures are .a source inputs to the real binder. scopes.a drives block-scoped
variables and function declarations; members.a drives class instance/static
members and type declarations; aliases.a drives imports and export aliases.
projects.json fixes their wire inputs, golden.jsonl is the independent stock
binder's output, and test.cjs verifies source bytes and explicit semantic facts.
counts.md records Node AST/symbol counts. These inputs were not registered in
internal/oracle, and no native allocation counts are claimed.

research.py takes the completed meter directory, its adapted main tree, the
parent directory containing step31-{binder,checker,emitter}-slice, and a new
output directory. It attributes unique findings to actual diagnostic sites,
not the attempting file. It retains every family location, direct closure file
row, source hash, and skipped-body extent. Full-file census rows are explicitly
broader than reached declarations. reached_in_slice distinguishes those cases.
inspect.cjs independently records all three source files' function spans and
import bindings using stock TypeScript. Compressed evidence is ordinary gzip JSON.

```sh
python3 stage3/scouts/step31/audit-research.py /absolute/exact-adapted-main > /tmp/research-audit.log 2>&1
python3 stage3/scouts/step31/audit-research.py /absolute/exact-adapted-main --mutant attribution > /tmp/research-mutant.log 2>&1
```

The first command must exit zero; each --mutant meter/family/hidden/attribution
command must exit one at its named recount assertion. Original-source historical
rankings come from stage3/census/data/rankings.json at ef3d907 and are kept apart
from current adapted-source observations.
