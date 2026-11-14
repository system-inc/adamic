# Fresh resolution host (#b8ffarm)

| Path exception | Counter | Reason and direct check |
| --- | --- | --- |
| getDirectories | 0 | Its resolver presence test/call are moduleNameResolver.ts:820,825, in getAutomaticTypeDirectiveNames, which resolveJSModule does not call. Direct check forwards the exact path with nodeSystem as receiver. |
| trace | 0 | tryResolveJSModuleWorker fixes options without traceResolution at moduleNameResolver.ts:1755; isTraceEnabled short-circuits at :124. Direct check forwards the exact string with nodeSystem as receiver. |
| getCurrentDirectory | 0 | Successful resolutions use Node10, features None and JavaScript-only extensions (:1750–1764); they have no paths/rootDirs/typeRoots or imports/self-name/export-map route. A failed retry passes extension mask 0 (:1897); both lookup passes are skipped (:3017,:3023). Direct check forwards zero arguments with nodeSystem as receiver. |

These three exceptions are expressly authorized by the task's rulings. They are
forwarder checks, not require-path proofs. A successful nodeSystem.require does
not read getCurrentDirectory. The fixture separately reports counters for its
successful package and relative-file resolutions and its failed exports-map
package. The complete conservative symbol-bound chain and branch exclusions
are retained with the evidence; the chain is not inferred from lane coverage.

This is one amendment on ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8. It
stacks after the require result annotation in adapter 99's same getNodeSystem
whole-owner pin. Adapter 93 also pins that owner, earlier in the pipeline; its
pins and transformation stay unchanged. The measured road is site 10 of
3ecc182ad4c79a8783e4b673c01d977d046e515e, stage3/census/price-sys-public/REPORT.md,
task #b8ffarm. Its present-and-undefined realpath slot is replaced here by
conditional property initialization preserving optional presence.

Every forwarder names its exact ModuleResolutionHost parameters, calls the
member through nodeSystem, and has no hidden rest (#2t3h708). No method value
is copied into a writable host field. A local type-only interface extending System exposes the two
optional extension methods trace and getGlobalTypingsCacheLocation without
claiming that either exists. The fresh host has a separate local interface of writable callback fields;
conditional initialization therefore never writes a method slot. Every optional
host member, including the boolean useCaseSensitiveFileNames, is conditionally
initialized on that fresh object. resolve-host-contract.cjs
checks all 128 optional presence combinations, exact parameters and receiver
syntax. The host is fresh on each require.

The production audit counts zero writes to the forwarded members of nodeSystem
after construction, and zero reads of forwarder identity or .length along the
resolver chain. Presence tests are inventoried separately in the evidence,
with file:line citations. A forwarder reads at call time, so the zero-write
condition is part of the value-equivalence argument.

The ordinary lane never enters nodeSystem.require's resolveJSModule path; its
forwarder counts are vacuous. The separate fixture is held to Node. It compares
parent and candidate output bytes with realpath present and absent, counts all
reachable forwarders, checks exact receiver/arguments for the three exceptions,
and makes a detached-receiver readFile mutant fail. resolve-host-mutant.cjs
restores the original nodeSystem argument and its original System annotation,
while retaining the sibling ModuleImportResult annotation.

Acceptance receipts belong on typescript/resolve-host-1612-evidence under
stage3/census/resolve-host-1612/: full single-file sys.ts census and restoring
mutant, zero stock diagnostics on 79 roots, adamic types, fresh full apply,
JavaScript scope, actual built Node fixture, byte-identical hereby dts outputs,
and the whole unfiltered lane. No compiler, sanction, filter or lane expectation
changes. Nothing is submitted to Loom.
