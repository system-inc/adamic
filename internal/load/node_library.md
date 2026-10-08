The Node host embeds the exact published @types/node 25.3.3 package and its
locked undici-types 7.18.2 dependency in internal/load/node_types. No npm install
or Adamic checkout is needed to compile node:* imports. The declarations are
served unchanged beside TypeScript's libraries through the bundled filesystem,
at bundled:///node/node_modules/@types/node. Caller @types/node packages are
ignored. An incomplete compiler reports that its bundled declarations are
missing and asks the user to reinstall Adamic, without a repository path.

node_types_manifest.json records the npm URLs, package-lock SHA512 integrity
values and every published file's SHA256. The test-only published tarballs are
independently held to those integrity values, and TestEmbeddedNodeTypesIntegrity
compares every embedded byte with those archives. The test needs no registry
network access or stage3/api installation.

stage3/api remains for stage 3's pinned TypeScript and its adaptation tools.
Historical reports naming its installation describe the previous loader.
Ordinary Adamic sources retain the existing prelude.

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
