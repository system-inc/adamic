# Explicit any owner adaptation (partial)

Current progress: 70 tokens removed, 140 left. The 36 phantom brand sites now use void and isArray takes unknown under the October 7 ruling; see [BRANDS-VOID.md](BRANDS-VOID.md). See [PROGRESS.md](PROGRESS.md)
for the combined 10+30+40 proof and reproduction. The section below records
the original three-owner proof and its historical commands.

This unit currently removes three of the 210 explicit any tokens in 28 original
TypeScript 6.0.3 compiler files. It does not complete the unit's requested census.
There are 207 tokens left, each retained in evidence/sites.json with a reason.

The adapter uses stock typescript@6.0.3's parser and edits only these declarations
in src/compiler/core.ts:

- length, line 22: inferred T in readonly T[]. The body reads only length and
  tests undefined; callers retain their array element types.
- toOffset, line 965: inferred T in readonly T[]. addRange and elementAt pass
  arrays of their own T; the body reads length and performs number arithmetic.
- hasProperty, line 1266: object. Its body calls Object.prototype.hasOwnProperty
  with the object and key. Consumers include Node, SourceFile, PackageJson,
  arrays, functions, and formatting settings. No dictionary value is read.

No unknown substitution or consumer edit is made. An exploratory MapLike<T>
parameter for hasProperty exposed index-signature errors, including callers
outside src/compiler. evidence/exploratory-generic-build.log preserves every
build diagnostic from that rejected contract. The final object parameter is
chosen from those actual uses at the owning declaration.

The remaining 207 sites are now classified by semantic pattern in
[CLASSIFICATION.md](CLASSIFICATION.md), with exact token locations and proposed
owner rules in evidence/classification.json. They have not received the complete
owner-and-consumer proof needed to promise their replacement types. In particular, assertions, dynamic constructors, brands,
JSON pipelines, timers, and heterogeneous callbacks cannot be repaired by a
uniform replacement. Their reasons identify the outstanding contract work;
they are incomplete sites, not claims that the adaptation is impossible.

## Reproduce

Use stage3-base at 8728405135d329efc12c837a7a6c293234abbe1c in a separate
checkout, and stock TypeScript 6.0.3 from its stage3/api lockfile. Main does not
contain the pipeline yet. Add this directory to that checkout's stage3/adapt
when integrating. The adapter accepts exactly one prepared upstream tree and
is discovered by apply.sh's existing numeric ordering.

Run commands from the Adamic repository after sourcing the setup environment.
Set NODE_PATH to the installed stock 6.0.3 API directory. A fresh pipeline tree
contains the generated diagnostic input and no source adaptation beyond setup.

```sh
npm run build --prefix "$tree" > before-build.log 2>&1
node stage3/adapt/40-explicit-any/verify.cjs before "$tree" "$proof" > before.log 2>&1
node stage3/adapt/40-explicit-any/adapt.cjs "$tree" > adapt.log 2>&1
npm run build --prefix "$tree" > after-build.log 2>&1
node stage3/adapt/40-explicit-any/verify.cjs after "$tree" "$proof" > after.log 2>&1
node stage3/adapt/40-explicit-any/verify.cjs mutants "$tree" "$proof" > mutants.log 2>&1
node stage3/adapt/40-explicit-any/adapt.cjs "$tree" > idempotence.log 2>&1
stage3/oracle/run.sh "$tree" "$oracle" > oracle.log 2>&1
```

Run build and oracle sequentially on a tree. Before/after proof snapshots retain
all original compiler source hashes, AST any sites, stock pre-emit diagnostics,
and every .js/.mjs/.cjs hash under built, including test output. Every declaration
output is compared to the pristine artifact with only the two emitted owner
signatures replaced. toOffset is private and does not appear in declarations.
Public typescript.d.ts must consequently remain byte-identical.

Source maps and .tsbuildinfo are metadata, not emitted JavaScript, and are not
claimed identical. Generic declarations shift source columns in maps. The proof
checks every emitted JavaScript file, not just the compiler bundle.

Stock compiler checks supplement the upstream locked build, which uses the
TypeScript version in upstream package-lock.json. Adamic's full uncached gate,
its adapted-tree census under sounder regex libraries, composition with units
10 and 20, and native compiler execution are not covered by this partial unit.
There are no fixture files: this is a source adaptation rather than a fixture
bucket. No other unit or production compiler file is edited.

Source snippets in evidence are from Microsoft TypeScript 6.0.3, copyright
Microsoft Corporation, licensed under Apache-2.0. Upstream source stays in the
external scratch checkout.

Validate the complete classification and its five failing mutants:

```sh
node stage3/adapt/40-explicit-any/classify-check.cjs "$tree" > classification-check.log 2>&1
```

The follow-up stops without new source edits: JSON/config is the largest class
(38), and its whole-class rule requires checked runtime input narrowing.

## Filesystem entry classification

Site 138 at sys.ts:1853 now uses `import("fs").Stats | import("fs").Dirent | undefined`.
Both assignment branches require the union: statSync can return undefined,
and readdirSync supplies Dirent. The existing continue narrows the optional
stat before isFile/isDirectory. This local annotation emits no API declaration.
The class rule checks the exact declaration line and VariableDeclaration
AST owner inside getAccessibleFileSystemEntries; a second pass makes no edit.

After apply and npm ci, reproduce the full upstream build comparison and mutant:

```sh
NODE_PATH=<stock-typescript-6.0.3-node_modules> node stage3/adapt/40-explicit-any/filesystem-proof.cjs <adapted-tree> <proof-directory> > filesystem-proof.log 2>&1
```

The proof compares every emitted JavaScript and declaration hash across builds
with only this annotation changed, plus sys.ts's own stock JavaScript emission.
It runs a real `let stat: string` mutant through upstream's npm build, requires
TS2322, then restores and successfully rebuilds. Source maps and build metadata
are outside the byte identity claim. See evidence/filesystem for measured results.
