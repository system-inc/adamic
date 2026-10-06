# Parameter properties and namespaces coverage

Base: adc45ca417482834336c08510be0db5b91974680.
Branch: coverage/parameter-properties-namespaces.

Seven accepted entry programs, ten new oracle .a files including dependencies.
Zero observed output disagreements on the unmodified compiler. Three additional
candidate shapes stop at lowering; they are preserved here, outside testdata.

## Inspection

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the requested
origin/main...origin/codex/parameter-properties-namespaces diff and commit log.
The initial origin/main ref was stale at f580438 and included integration
history. Refreshed origin/main to 50045bd and reread the requested diff/log:
exactly aea3215 (parameter properties) and adc45ca (namespaces). Inspected their changed
compiler paths, configuration, documentation, fixtures, and lower/oracle tests.
Read cloud/grok-params-namespaces-coverage before writing candidates. Its four
entry programs are comparison coverage, not copied or cherry-picked here.

## Case census

Existing means the feature branch's registered oracle programs. Grok means a
program on the other coverage branch. New filenames below have prefix
params_namespaces_ and suffix .a; shared means shared/main.a.

| Case in the feature code | Existing program | Grok coverage | Added coverage or limit |
| --- | --- | --- | --- |
| Parameter-property modifiers: public, private, protected, readonly, combinations | parameter_properties | levels | order, callbacks |
| Ordinary parameters mixed with property parameters | parameter_properties | levels | shared, values |
| Fields laid out with property slots before written fields | parameter_properties | levels lacks written fields | order |
| Base field initializer, defaults, stores, constructor body in order | parameter_properties | levels | order reads an initialized written field in a default |
| Derived defaults, lexical writes before super, field initializers, stores, body | parameter_properties | levels | order at three levels |
| Defaults omitted, supplied, explicitly undefined, dependent on prior parameter | parameter_properties | levels | order, dotted |
| Lexical parameter and this.property diverge after implicit copy | none | none after copy | order |
| Mutable field write and compound update after construction | parameter_properties | levels | values, overrides |
| Same-name inherited parameter slot reset and replacement | none | levels numbers | order numbers, overrides object references |
| Readonly override narrowed with unchanged native representation | none | none | overrides string literal |
| Default derived constructor forwards inherited property parameters | none | none | overrides Forward |
| Number, string, boolean property values | number/string | kinds/levels | values required boolean |
| Optional number, present/missing | parameter_properties | levels | existing and Grok |
| Optional string, object, closure present/missing | none | kinds optional generic string/object | values |
| Class-valued property, then a method call on it | none | none | values Label/read |
| Arrays, Map, Set as property values; aliasing and mutation | none | kinds arrays | values |
| Object/interface property values and owned dynamic strings | parameter_properties_ownership | kinds | values, overrides |
| Function property, captured closure, overridden method and base dispatch | parameter_properties and ownership | none new | callbacks, shared |
| Generic classes and generic inheritance at multiple concrete types | parameter_properties | kinds | existing and Grok |
| Weak property links read with a live strong owner | parameter_properties_ownership | kinds | existing and Grok |
| Optional boolean / boxed heterogeneous union field | no accepted fixture | none | refused_values, slotless representation |
| Readonly writable-view, mutable narrowing, readonly-to-mutable override guards | no successful program can violate these | refused.txt partial | lower parameter_properties tests |
| Early property/default/initializer reads and this escape guards | no successful program can violate these | refused.txt early default | lower parameter_properties tests; refused_overrides |
| Dynamic this callback guard | no successful program | none | TestParameterPropertyCallbackReceiver |
| Bound-local missing / unsupported type diagnostics | no valid successful source | none | internal invariants or unsupported representations |
| Namespace single top-level declaration and scoped names | namespaces | nested | callbacks, shared |
| Nested namespaces in ModuleBlock | namespaces | nested deep | existing and Grok |
| Dotted nested declaration with ModuleDeclaration body | none | none | dotted Outer.Inner.Deep |
| Empty statements, empty namespace, type-only recursion, no runtime pending | Types within namespaces | none | types, call before type-only declarations |
| Interface and type alias name coexistence | namespaces interface | nested interface | dotted Factory type alias |
| Exported primitive constants, private const and private mutable let | namespaces | nested | callbacks, values |
| Exported object/array/closure constants and mutable contents | none | none | values, callbacks, shared |
| Direct qualified calls, unqualified internal calls | namespaces | nested | callbacks, shared |
| Parenthesized qualified calls | none | none | dotted, named function with default and dotted nested function |
| Generic qualified call with explicit and inferred type arguments | namespaces explicit number/string | none new | dotted inferred number/string/object/array |
| Namespace function default parameter, missing/undefined/present arguments | none | none | dotted |
| Namespace rest function parameter | none | none | refused_dotted, ordinary signature limit |
| Qualified function as value, detached call, identity comparison, array callback | namespaces | nested/pair | callbacks, shared |
| Closure-valued export invocation, returned closure reads private state | none | none | callbacks |
| Structural typeof namespace object uses ordinary receiver dispatch | namespaces copy | none | already covered |
| Named imports, import alias, nested imported types/functions | namespaces_modules | pair | shared has a diamond of three callers |
| Private state and exported references shared across several importing files | none | pair one caller | shared |
| Initialization dependency order; deferred function bodies may call | namespaces_modules | pair | shared, callbacks |
| Calls/new before pending runtime namespace, early reads, initializer calls | cannot pass preflight | refused.txt partial | TestNamespaceLimitsStayLoud |
| Ambient/external, local scope, reopening, runtime merge | outside admitted subset | refused.txt reopening | namespace guards; some rejected by checker first |
| var, exported let, missing initializer, destructured namespace binding | outside admitted subset | refused.txt var | namespace guards |
| Executable namespace bodies and classes/enums inside namespace | outside admitted subset | refused.txt body | namespace guards |
| this in namespace function / explicit this parameter | outside admitted subset | refused.txt receiver | namespace receiver and limit tests |
| Namespace container escape/reflection/computed read, export replacement | outside admitted subset | none | TestNamespaceLimitsStayLoud |
| Optional namespace reads and computed member lowering | outside admitted subset | none | namespaceExpression rejects; container checks may reject first |
| Nil body/symbol, missing binding, no declaration | malformed/synthetic AST or internal failure | none | cannot reach from valid checked .a source |

## Refused candidates

These are compiler limits, not native stdout mismatches. No native binary was
produced, so native stdout is absent rather than a different completed result.
Each source is retained under its own .a filename here.

### refused_values.a

Node exit 0, stdout:

```text
base-field label3 map3 1 1,2,1
child-field true word3 note3 closure3
base-field empty missing 0  child-field false absent absent absent
base-field version2 changed 2 1,2,1,2 child-field true absent note3 closure3 version2
```

Native build exit 1, no stdout or binary. Diagnostic:

```text
adamic: /workspace/adamic/notes/parameter-properties-namespaces/refused_values.a:18:91: stage 0 can't lower a field of type boolean | undefined yet
exit status 1
```

Responsible guard: internal/lower/class_inheritance.go:173 (slotless boolean | undefined; expression.go:833 also treats boxed Union as slotless).

### refused_dotted.a

Node exit 0, stdout:

```text
8 text object 2,3 3 5 5 6
deep1 deep2 true
```

Native build exit 1, no stdout or binary. Diagnostic:

```text
adamic: /workspace/adamic/notes/parameter-properties-namespaces/refused_dotted.a:5:25: stage 0 can't lower a parameter that isn't a plain name yet
exit status 1
```

Responsible guard: internal/lower/functions.go:102 (rest parameter is not a supported plain incoming parameter).

### refused_overrides.a

Node exit 0, stdout:

```text
small child-note 9 1 2
small new-note 10 1 3
```

Native build exit 1, no stdout or binary. Diagnostic:

```text
adamic: /workspace/adamic/notes/parameter-properties-namespaces/refused_overrides.a:11:24: Adamic 0.1 refuses this escaping a base constructor before derived fields are initialized; use this only to read or write initialized base fields; call methods and publish the object after construction
exit status 1
```

Responsible guard: internal/lower/class.go:524 (method call can escape this before derived fields initialize).

## Mutation proof

Temporarily changed one line in internal/lower/parameter_properties.go:31:

```go
if !parameterProperty(parameter) || parameter.Name().Text() == "value" {
```

This omits the value property store. The order program failed with native and
JavaScript backend stdout differences, both exit 0, empty stderr, and clean
sanitizers. First line: Node `root 111 11 22`, native/backend `root 111 0 22`.
Restored the file from adc45ca. The restored compiler passes all seven fixtures.
The full failed oracle output is in mutant.log.

## Toolchain

Successful setup timing lines: go ready (0s), clang ready (1s), node ready (1s),
submodules ready (1s), build cache warm (66s), done in 66s on 5 processors.
`nproc`: 5. CPU quota: 400000/100000, four CPUs. Environment file:
/workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The initial setup warm build overlapped checkout and failed on stale package
file lists. Reran bash cloud/setup.sh on the final branch; it succeeded.

## Commands

All test runs redirect output to a log and the logs were read. No test output
was piped into a truncating process. Every Go command below sources
/workspace/adamic-tools/env.sh in that shell.

```sh
git fetch origin codex/parameter-properties-namespaces:refs/remotes/origin/codex/parameter-properties-namespaces cloud/grok-params-namespaces-coverage:refs/remotes/origin/cloud/grok-params-namespaces-coverage
git fetch origin main:refs/remotes/origin/main
git log --oneline origin/main..origin/codex/parameter-properties-namespaces
git diff origin/main...origin/codex/parameter-properties-namespaces
git show --format=fuller --stat aea3215 adc45ca
git diff adc45ca...origin/cloud/grok-params-namespaces-coverage -- internal/oracle/testdata review
git switch -c coverage/parameter-properties-namespaces origin/codex/parameter-properties-namespaces
bash cloud/setup.sh
bash cloud/setup.sh > /tmp/params-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
gofmt -w internal/oracle/parameter_properties_test.go
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/params_namespaces' -count=1 -timeout 10m -v > /tmp/params-oracle.log 2>&1
# Initial candidates: three PASS, three lowering refusals.
# Repeated after supported variants, then after seventh fixture and restoration:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/params_namespaces' -count=1 -timeout 10m -v > /tmp/params-oracle-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/params_namespaces' -count=1 -timeout 10m -v > /tmp/params-oracle-restored.log 2>&1
# With the one-line mutation:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/params_namespaces_order.a$' -count=1 -timeout 10m -v > /tmp/params-mutant.log 2>&1
git restore --source=adc45ca -- internal/lower/parameter_properties.go
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/params-counts.log 2>&1
# Repeated for the seventh fixture:
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/params-counts-final.log 2>&1
gofmt -l cmd internal > /tmp/params-format.log
go vet ./... > /tmp/params-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/params-gate.log 2>&1
```

Standalone builds and independent source runs used this exact loop, first for
six original candidates, then six supported variants plus refused notes. The
seventh fixture used the same commands individually:

```sh
mkdir -p /tmp/params-builds
for file in internal/oracle/testdata/params_namespaces*.a internal/oracle/testdata/params_namespaces_shared/main.a notes/parameter-properties-namespaces/refused_*.a; do
    name=$(basename "$file" .a)
    node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" > "/tmp/params-builds/$name.node" 2> "/tmp/params-builds/$name.node.err"
    go run ./cmd/adamic build "$file" -o "/tmp/params-builds/$name" > "/tmp/params-builds/$name.build.log" 2>&1
    result=$?
    if [ "$result" = 0 ]; then
        "/tmp/params-builds/$name" > "/tmp/params-builds/$name.native" 2> "/tmp/params-builds/$name.native.err"
        cmp "/tmp/params-builds/$name.node" "/tmp/params-builds/$name.native"
    fi
    echo "$file build=$result"
    cat "/tmp/params-builds/$name.build.log"
done
```

All accepted standalone builds exit 0 and cmp finds identical stdout. The oracle
also verifies stderr, exit status, release builds, ASan/UBSan, and LeakSanitizer.

Final counts check (without update):

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/params-counts-check.log 2>&1
gofmt -l cmd internal > /tmp/params-format-final.log
git diff --check
```

Validation observed so far: final seven-program uncached oracle 6.203s, counts
regeneration 42.173s; both pass. The full gate's native suite passed in 272.723s
and its oracle suite in 205.369s. Final parenthesized-call addition was also
built and compared individually using the standalone commands above, followed
by the same seven-program oracle and counts regeneration commands.

The complete uncached repository gate exited 0: every package passed. Unicode
properties took 691.063s, cohere/json 425.384s, parser 193.163s, scanner 52.987s.
Its complete package summary is gate.log. Formatting, vet and diff checks were
clean. Compiler files match adc45ca; no mutation remains. Seven new counts rows
are present and all previous rows are unchanged.

Commits and pushes:

```sh
git commit -m "Cover parameter property ordering and shared namespace values"
git push -u origin coverage/parameter-properties-namespaces
git commit -m "Record the full coverage gate and refreshed main comparison"
git push origin coverage/parameter-properties-namespaces
git ls-remote --heads origin coverage/parameter-properties-namespaces
```

The first push succeeded without retry. The second records the final full-gate
result and corrects the description of the initially stale origin/main ref.
