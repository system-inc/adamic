The Node host embeds the unchanged published @types/node 25.3.3 package and
undici-types 7.18.2 dependency. No npm install or checkout is needed for node:*
imports. The loader serves them at bundled:///node/node_modules/@types/node,
ignores caller @types/node packages, and reports damaged bundles with a request
to reinstall Adamic. Ordinary Adamic sources keep the existing prelude.

node_types_manifest.json records SHA512 archive integrity and SHA256 for each
published file. Offline integrity tests compare every embedded byte with the
published tarballs in testdata/node_types and reject additions or omissions.

Host lowerers identify declarations with load.IsNodeLibrary and register only
implemented runtime members with lower.RegisterNodeLibraryMembers during init.
Names are node:<declaring module>.<owner, when present>.<member>, for example node:fs.statSync,
node:fs.StatsBase.size and node:fs.StatsBase.isFile. Import aliases resolve to their declarations.
Unregistered runtime uses fail up front with NotYet naming the member, including
function values, properties and constructors. Type-only references are allowed.
Registration does not implement a member; its unit must supply the lowering and
backend runtime. Unsupported overloads still need a named NotYet in that lowerer.

Owners distinguish StatsBase.isFile from Dirent.isFile. Register object receivers
as well when applicable, such as node:buffer.Buffer and node:perf_hooks.performance.
The census's declared_owner identifies the class or interface part of the name.

Adamic corrects false host axioms in host_axioms.json, without changing any
embedded package bytes. Each row pins the published file and line, its exact
replacement, observed Node behavior and ruling. Identical declarations on
separate lines have separate rows; unaccounted occurrences, changed lines and
already-corrected publications fail loading. CorrectNodeHostAxioms is shared by
the bundled checker view and admission-delta's prepared npm checker copy.

The table is a go:embed input, so buildcache.GoInputs includes it in GoBuild's
content key. Products using an already linked compiler include it through
compilerinputs.Lowering's internal directory. The admission compiler also keys
its revision SHA. The package manifest and tarball integrity tests still hold
only the untouched published bytes.
