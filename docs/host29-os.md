# Step 29: Node os surface used by TypeScript

Base: origin/area/library `71f916872667150037677b4b975772f165e47898`.
Upstream inventory: TypeScript 6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
Reference runtime: Node 24.19.0 on Linux. Branch: codex/host29-os.

The exhaustive src inventory is recorded with source snippets in
[host29-os-inventory.json](host29-os-inventory.json). It contains these ten uses:

| File | Lines | Member |
|---|---|---|
| src/compiler/sys.ts | 1489 | platform() |
| src/compiler/sys.ts | 1525 | EOL |
| src/tsserver/server.ts | 57 | platform() |
| src/tsserver/nodeServer.ts | 629, 654 | homedir probe and call |
| src/tsserver/nodeServer.ts | 632, 657 | tmpdir() |
| src/testRunner/parallel/host.ts | 130, 136, 182 | EOL |

No cpus or availableParallelism call appears in src. The inventory excludes a
documentation import example, an import inside a test input string, and the
compiler's built-in module name list.

## Implementation and limits

Both os and node:os use the unchanged embedded @types/node declarations.
Default, namespace and named imports, aliases, and stable callable function
values are admitted. EOL string operations now lower as ordinary string
operations. The implemented homedir feature guard reads a known, always-present
function and evaluates its right operand.

Native homedir ports Node's fixed PATH_MAX binding buffer and libuv's HOME and
effective-user passwd lookup. Empty HOME stays empty. Unset HOME uses
getpwuid_r with EINTR retry and ERANGE buffer growth. Oversized HOME throws a
catchable Node SystemError, including ERR_SYSTEM_ERROR and uv_os_homedir details;
uncaught formatting includes the code. Node and libuv sources and licenses are
credited in THIRD_PARTY_NOTICES.md. The JavaScript backend invokes Node's real
homedir, platform and tmpdir functions.

All four families agree with Node on native and JavaScript. WASI EOL and tmpdir
agree with Node. WASI platform is refused at compile time because WASI provides
no Node host platform; homedir is refused because WASI provides no effective-user
account database for the HOME-unset fallback. Target refusal tests enforce both.

The full tsserver fallback `(os.homedir && os.homedir()) || ...` still encounters
the base compiler's string truthy-or refusal. Its homedir guard and calls are
supported; this unit does not claim that the full server source compiles.
A regression test explicitly keeps that compiler gap refused.

The declaration-derived refusal census covers all 19 other runtime members:
arch, availableParallelism, constants, cpus, devNull, endianness, freemem,
getPriority, hostname, loadavg, machine, networkInterfaces, release, setPriority,
totalmem, type, uptime, userInfo and version. Each reports that it is outside the
pinned TypeScript src inventory. No runtime placeholder is exposed for them.

## Fixtures, mutants and checks

Each family has one registered oracle fixture. Native and JavaScript comparisons
cover default and aliased imports, function identity, empty and unset HOME,
passwd fallback, Unicode and invalid UTF-8, oversized HOME caught and uncaught,
temporary-directory precedence, root paths and trailing-slash behavior.

The four native generated-C mutants replace EOL, platform and homedir with a
wrong string, or append an extra slash to tmpdir. Every mutant exits successfully
with empty stderr and clean ASan, UBSan and leak checks. Each is caught solely
because its stdout differs from Node.

Focused tests passed in internal/load, internal/lower, internal/flow,
internal/native, internal/javascript and internal/oracle; flow had no matching
test names and JavaScript has no package test files. The oracle comparisons and
WASI tests exercise the affected backends. go vet passed for all six packages.
The Linux TestCountsAreRecorded refresh passed. Run output is outside the
checkout, in /tmp/host29-os-*.log.

## Linux counts before and after

The recorded fixture count increased from 635 to 639. New rows had no baseline:

| New fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| node_os_eol.a | 4 | 4 | 0 | 8 | 3 | 0 |
| node_os_platform.a | 2 | 2 | 0 | 5 | 2 | 0 |
| node_os_homedir.a | 8 | 8 | 0 | 13 | 3 | 0 |
| node_os_tmpdir.a | 6 | 6 | 0 | 9 | 3 | 0 |

The checkout-enumerating node_fs_directory_system.a row changed from
3220/3220/5812/4829/1938/0 to 3236/3236/5844/4853/1950/0
(allocations/frees/retains/releases/peak/regions). All other existing rows stayed
unchanged. The authoritative table is internal/oracle/counts.md.
