The checker and emitter use the same project request as the binder. They create
fresh stock TypeScript 6.0.3 Programs and use getPreEmitDiagnostics and program.emit.
The compiler itself runs; diagnostics and emitted text are never synthesized.

Run the complete focused proof after cloud/setup.sh and sourcing its environment:

```sh
export STEP31_TYPESCRIPT="$HOME/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js"
bash stage3/scouts/step31/run-components.sh /absolute/pinned-adapted-tree /tmp/new-components-proof > /tmp/components-proof.log 2>&1
```

This prepares the exact same 301 acceptance and 6,262 selected upstream projects,
checks fixtures against TypeScript's standard compiler host, runs semantic and
comparison mutants, times stock Node, and compares the actual adapted source
checker and emitter against the stock bundle over all 6,563 projects. It does
not run the full upstream test suite or a repository-wide gate. Optional third
and fourth arguments are native checker and emitter executables, respectively.
Each native executable accepts one request pathname and writes protocol JSONL
to stdout, empty stderr and exit zero. Diagnostics are observations, including
errors; they do not make the observer exit nonzero. No native result is claimed
unless that executable has actually run.

For a Node baseline alone:

```sh
node stage3/scouts/step31/checker-dump.cjs /absolute/projects.json /tmp/new-checker-cache > /tmp/checker.log 2>&1
node stage3/scouts/step31/emitter-dump.cjs /absolute/projects.json /tmp/new-emitter-cache > /tmp/emitter.log 2>&1
```

A results directory contains golden.stdout, request.json, manifest.json and one
JSONL shard per project. stdout prints a small timing/hash summary, not the dump.
The native comparator reads golden.stdout and appends request.json to the command:

```sh
python3 stage3/scouts/step31/component-compare.py /tmp/new-checker-cache /tmp/new-native-checker -- /absolute/native-checker > /tmp/native-checker.log 2>&1
python3 stage3/scouts/step31/component-compare.py /tmp/new-emitter-cache /tmp/new-native-emitter -- /absolute/native-emitter > /tmp/native-emitter.log 2>&1
```

The comparator first validates request and golden hashes and the selected input
population, then delegates exact stdout, empty stderr and exit-zero checks to
compare.py. --timeout SECONDS defaults to 120. There is no diagnostic flattening,
path substitution or newline normalization on the native output. The command,
actual stdout, stderr, exit and first differing byte are retained separately.

For changed inputs, retain the previous Node cache:

```sh
node stage3/scouts/step31/checker-dump.cjs /absolute/updated-projects.json /tmp/new-checker-delta --previous /tmp/new-checker-cache --changed-only > /tmp/checker-delta.log 2>&1
python3 stage3/scouts/step31/component-compare.py /tmp/new-checker-delta /tmp/new-native-delta -- /absolute/native-checker > /tmp/native-delta.log 2>&1
```

Use emitter-dump.cjs for the identical emitter workflow. The delta's request and
golden contain only changed projects, while its manifest and shards retain all
current projects, so the delta can become the next --previous cache. Zero changed
projects produce an empty golden and a request with projects=[]. Native execution
of this request must also produce empty stdout. --ids JSON_FILE accepts an explicit
array of project IDs for a focused run; unknown IDs fail. An --ids cache contains
only that subset. Omitting --changed-only includes all projects while reusing
unchanged Node observations, which is useful for comparing a new native build.

Input identity is SHA256 of recursively key-sorted JSON, with files sorted by
path and array order otherwise retained. It includes ID, options, every source
byte and explicit scriptKind. Cache compatibility also includes protocol, exact
Node version, stock compiler bytes, all lib.*.d.ts bytes and driver bytes. A
compiler/library/driver change invalidates the baseline conservatively. Every
reused shard is hash-checked. manifest.json records per-project compile time,
output size/hash, input hash, reuse status, selected IDs and total driver time.
benchmark-components.py additionally measures complete subprocess wall time.

Changed-only selection proves changed input behavior. A native implementation
change can affect unchanged inputs: compare the full cached baseline unless an
explicit affected-input set has been justified. Node observations still need
no recomputation for those unchanged inputs. There is no native dependency
analysis in this unit.

The host is deterministic and closed over project files plus the pinned stock
standard libraries. Working directory is /project, libraries are /lib, paths
are case-sensitive and the default host newline is LF. It does not read ambient
node_modules, tsconfig files, package.json, disk imports or external @types.
Relative imports among supplied files resolve through TypeScript's own resolver.
The selected corpus already excludes virtual-host and option-variant cases;
this driver does not broaden that selection. The tiny multi-file project is
included. .a names are parsed as .ts and diagnostics restore the original .a
name; imports should use the same extensionless spelling as fixtures/components.json.
The native side must use the same standard library text and host contract.

Project ASTs are always fresh. Standard-library ASTs are reused by parser option
identity, as TypeScript permits across Programs, to avoid reparsing its libraries
6,563 times. --fresh-libs disables AST reuse. Fixture and one-changed-project
checks compare this mode with the reused mode, and source-versus-stock runs use
separate compiler implementations. This is Node memory reuse, not an Adamic
allocation or lifetime design.

The output protocols are adamic-checker-v1 and adamic-emitter-v1. They are UTF-8
JSONL, JSON.stringify property order as in component-dump.cjs, one final LF per
record. Projects retain request order. Options have recursively sorted keys;
file and output paths use UTF-16 lexical order. No timings or allocation IDs
enter the golden bytes.

Checker records, in order:

- project: record, format, typescript, id, options.
- global: record, diagnostics for diagnostics without a SourceFile.
- file: record, path, diagnostics. Every input file has a record even when clean;
  libraries with diagnostics receive @lib/<basename> records too.

Diagnostics retain file, code, start, length, category, message and related.
Absent file/start/length values are null. Offsets count UTF-16 code units. A
message is either the original string or {message, code, category, next}; next
preserves the recursive chain and compiler order. related contains diagnostics
in compiler order, including their locations. Diagnostics sort by file, start,
length, code, then serialized record as a total tie-break. getPreEmitDiagnostics
includes configuration/option, syntax, global, semantic and requested declaration
diagnostics, matching TypeScript's public tsc pre-emit API; this is broader than
checker semantic diagnostics alone. File names restore .a after serialization.
Fixed virtual paths can occur inside TypeScript's original message strings.

Emitter records, in order:

- project: the same header with adamic-emitter-v1.
- result: record, emitSkipped, sorted diagnostics returned by program.emit.
- output: record, path, text, bom, sources; sorted by emitted path.

Every writeFile callback is retained, including JS, declarations, source maps and
build info if requested. text is the exact string passed to writeFile, including
its own line endings and final-newline presence; bom separately records whether
a writer should prefix a byte-order mark. sources are sorted original input
names supplied to that callback. Outputs outside /project use @host plus their
fixed virtual absolute path. Duplicate output paths fail loudly. Emit diagnostics
and emitSkipped remain observable when options suppress output. A successful
emit does not claim that the checker found no errors.

fixtures/components.json fixes the three .a sources; checker-golden.jsonl and
emitter-golden.jsonl hold stock Node results. test-components.cjs checks nested
message chains, exact spans, clean files, import resolution, JS/declarations/maps,
file ordering, repeated observation and fresh-versus-reused libraries. It also
compares fixtures to an independently materialized standard compiler host.
component-mutants.py changes actual checker and emitter bodies in disposable
source copies, requires successful Node execution with empty stderr, and rejects
unchanged output. test-component-cache.py separately mutates diagnostic fields,
emitted bytes/population/BOM/status, source input, environment identity and cache
bytes. Its protocol producers are explicitly test doubles, not native results.
