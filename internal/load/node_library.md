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

Embedding unit checkpoint 58fcc362, based on area/library 09ad6d53:

- @types/node 25.3.3: 89 published files, including all versioned declarations.
- undici-types 7.18.2: 46 published files, the lockfile's transitive dependency.
- The compiler embeds the packages; published archives and the conflicting
  @types/node 24.0.0 archive remain test-only. Both packages' MIT notices,
  copyright and license text are in THIRD_PARTY_NOTICES.md.
- The offline integrity test first validates each published archive's SHA512
  against the recorded npm integrity, then compares SHA256 for every embedded
  file and rejects additions or omissions. No caller package is a fallback.
- Node declarations, their nested imports and the dependency resolve through
  the same cached bundled filesystem view as TypeScript's own libraries.
  TypeScript libraries and other project type roots retain their selection.

Observed checkpoint commands (all output written to /tmp/embedded-node-*.log):

    go test ./internal/load -run '^(TestEmbeddedNodeTypesIntegrity|TestNodeLibrary)' -count=1 -timeout=10m
    ok internal/load 2.582s

    go test ./internal/oracle -run '^TestEmbeddedNodeTypesPortable$' -count=1 -timeout=10m
    ok internal/oracle 1.863s

Each portability fixture is copied to a fresh directory outside the checkout;
the compiler runs in a separate process with that actual working directory.
The outside case has no node_modules. The project case installs the unchanged
published @types/node 24.0.0 archive. Both compile and compare stdout, stderr and
exit status with Node on native under sanitizers and on the JavaScript backend.
The helper checks that no caller declaration source entered the checked program.

The byte mutant replaces only fs.d.ts's final newline with a space. It remains
valid declaration syntax. An isolated Go overlay embeds the changed file;
TestEmbeddedNodeTypesIntegrity exits 1 with:

    node_library_integrity_test.go:107: embedded byte mismatch: @types/node/fs.d.ts

This is a hash-check kill, not a Go or TypeScript syntax failure. The published
archive and the production vendor files remain unchanged.

Completed broad checks before the final isolated oracle:

    go test ./internal/load ./internal/lower -count=1 -timeout=30m
    load 32.928s; lower 97.670s; both ok

    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=2 -timeout=30m -args -update-counts
    ok internal/oracle 174.751s

    go vet ./internal/load ./internal/lower ./internal/oracle
    exit 0

Linux counts add 2/2/1/4/2/0 for each new fixture. The directory-system row changes
from 2992/2992/5328/4464/1767/0 to 3000/3000/5344/4476/1773/0, reflecting its directory
enumeration; no other existing row changes. No macOS counts were regenerated.

The full configured command runs without concurrent builds:

    ADAMIC_ORACLE_WASI=1 go test -v ./internal/oracle -count=1 -parallel=2 -timeout=30m

WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot and
GOPROXY='https://proxy.golang.org|direct' are set. Final result:

    ok internal/oracle 1222.109s
    zero failing tests

Both new fixtures pass the native comparison, WASI comparison, WASI emission
and recorded-count checks. The separate portability workers compare both native
and JavaScript against Node from the foreign working directories. This gate was
run on Linux; macOS was not run.

internal/load's Go tests now need neither stage3/api nor any installed Node
package. The new worktree has no stage3/api/node_modules. stage3/api and its lock
are retained for the stage 3 TypeScript 6.0.3 adaptation and stock-check tools:
lane/check.py and normalize-api.cjs, parser and scanner drivers, census and
triage scripts, and ledger/checker-259's stock checker measurements. Their
separate npm package remains useful; historical installation reports are kept.

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0; Linux; nproc=5 with a four-CPU
quota. Initial setup initialized the pinned submodules in 304.755s, then its
cache-warm build overlapped edits and failed on the intermediate source. The
completed retry reports node 0.066s, Go 0.074s, markdown dependencies 0.183s,
submodules 0.254s, clang 0.475s, Go build 87.302s, cache warm 87.585s and done
87.656s. The printed environment file is /workspace/adamic-tools/env.sh.

Only codex/embedded-node-types is pushed. No main push, force push, rebase or PR.
