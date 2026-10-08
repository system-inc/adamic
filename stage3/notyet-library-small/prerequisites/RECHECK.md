# Prerequisite replay recheck

Merged views ffe428ab and checked non-null c41c0e06.

Every one of the 35 assigned sites produced a parsed census report. Signature disappearance alone is not certification of the whole declaration; named subsequent stops are retained below.

| Site | Requested signature reproduced | Named stops |
| --- | --- | --- |
| src/compiler/core.ts:1877:9 | no | none |
| src/compiler/core.ts:2433:11 | no | none |
| src/compiler/tracing.ts:194:19 | no | reading performance; node:fs.writeSync: JSON.stringify recursive array types; reading performance |
| src/compiler/sys.ts:1734:20 | no | none |
| src/compiler/sys.ts:1750:46 | yes | numeric coercion requiring dynamic ToPrimitive; numeric coercion requiring dynamic ToPrimitive |
| src/compiler/sys.ts:1751:22 | yes | numeric coercion requiring dynamic ToPrimitive; numeric coercion requiring dynamic ToPrimitive |
| src/compiler/scanner.ts:222:46 | no | none |
| src/compiler/scanner.ts:224:44 | yes | Object.entries on a shape not proven by a plain literal or its const binding |
| src/compiler/scanner.ts:4075:59 | yes | Object.entries on a shape not proven by a plain literal or its const binding |
| src/compiler/utilities.ts:6211:12 | no | regex replacement callback taking captures, offset or input |
| src/compiler/utilities.ts:6221:9 | no | none |
| src/compiler/utilities.ts:6252:12 | no | none |
| src/compiler/utilities.ts:8578:19 | yes | Object.assign on a shape not proven by a plain literal or its const binding; a function returning U \| undefined |
| src/compiler/utilities.ts:9598:12 | no | none |
| src/compiler/utilities.ts:9818:23 | yes | RegExp with a nonconstant pattern |
| src/compiler/utilities.ts:9930:34 | no | none |
| src/compiler/utilities.ts:10916:12 | no | none |
| src/compiler/factory/emitHelpers.ts:1470:11 | yes | a tagged template other than the intrinsic String.raw |
| src/compiler/factory/emitHelpers.ts:1478:11 | yes | a tagged template other than the intrinsic String.raw |
| src/compiler/parser.ts:8922:39 | no | a method call through a structural signature in a program with statics; use typeof the declaring class; a declaration directly in a case (wrap the case in a block); a BinaryExpression with a number and a number |
| src/compiler/parser.ts:10707:31 | yes | RegExp with a nonconstant pattern; reading result; reading result |
| src/compiler/commandLineParser.ts:2960:31 | yes | a cast the runtime can't check; JSON.stringify a union containing containers without runtime element metadata |
| src/compiler/commandLineParser.ts:4127:56 | yes | RegExp with a nonconstant pattern; a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions; an array of CanonicalKey |
| src/compiler/moduleNameResolver.ts:1262:25 | no | substr with a length argument |
| src/compiler/checker.ts:8943:45 | no | a method call through a structural signature in a program with statics; use typeof the declaring class; a method call through a structural signature in a program with statics; use typeof the declaring class; reading identifier |
| src/compiler/sourcemap.ts:380:64 | no | none |
| src/compiler/emitter.ts:1138:27 | yes | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| src/compiler/emitter.ts:4979:55 | no | none |
| src/compiler/program.ts:751:57 | no | a BinaryExpression with a number and a string; assigning to an Identifier; a method call through a structural signature in a program with statics; use typeof the declaring class |
| src/compiler/program.ts:2981:46 | no | a method call through a structural signature in a program with statics; use typeof the declaring class |
| src/compiler/builder.ts:1617:68 | no | none |
| src/compiler/watchPublic.ts:686:47 | yes | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata); JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata); a method call through a structural signature in a program with statics; use typeof the declaring class |
| src/compiler/watchPublic.ts:687:80 | yes | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata); JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata); a method call through a structural signature in a program with statics; use typeof the declaring class |
| src/compiler/tsbuildPublic.ts:666:29 | yes | a value of type ResolvedConfigFileName; a Map of ResolvedConfigFilePath; a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) |
| src/compiler/tsbuildPublic.ts:693:13 | yes | a value of type ResolvedConfigFileName; a Map of ResolvedConfigFilePath; a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) |

Observed: scanner.ts:222:46, sourcemap.ts:380:64, emitter.ts:4979:55 and utilities.ts:6221:9, 6252:12, 9598:12, 9930:34, 10916:12 have no findings. moduleNameResolver.ts:1262:25 clears and reaches substr with a length argument. core.ts:1877:9, core.ts:2433:11 and builder.ts:1617:68 remain clear.

utilities.ts:6211:12 reaches regex replacement callback taking captures, offset or input. It needs the getReplacement(match, offset, input) ABI. The original refusal is gone, but this root is not complete.

scanner.ts:224:44 and 4075:59 remain unproven Object.entries shapes; 4075 wraps its literal in as const. Both JSON object and container-union guards remain, and the tracing wrapper reaches JSON recursion. ResolvedConfigFileName uses a required never brand, which the incoming void phantom-brand rule does not admit. Numeric Date coercion still reaches its named stop.

Two initial replay batches failed before producing reports: an unsourced toolchain selected an unrelated go executable, then the merged loader required the new pinned @types/node dependency. Neither batch is counted as a clear site. After sourcing /workspace/adamic-tools/env.sh and npm ci --prefix stage3/api (3 packages in 531ms), all 35 reports parsed.


Validation: representative Node oracles and nine IR mutants PASS (33.150s); focused lower checks PASS (0.649s); CLI, flow, fresh, IR, native and JavaScript focused checks PASS (native typed-array runtime 11.792s). Constructor parameter-property compatibility oracle PASS (0.887s).

The mandatory counts update was attempted and fails (68.449s) on multiple integration dependencies, including unsupported host fixtures and graph-region invalid-pointer cleanup. Its first failure is library_method_values.a (detached Object.prototype.hasOwnProperty). No successful full counts refresh or fully green integration baseline is claimed. See counts.log. No new fixture is introduced in this prerequisite group.

Merge compatibility fixes after the merge: cmd/adamic/checks.go and cmd/adamic/tsgo.go retain both explanation protocols without duplicate argument cases; cmd/adamic/non_null_checks_test.go reuses the common driver; internal/lower/readiness.go records deferred initializer assertions; internal/lower/class_inheritance.go checks the AST kind before treating a parameter property as a property declaration. These preserve the two incoming mechanisms rather than changing another worker's lowering lesson. Runtime C changes were solely prerequisite conflict resolution.
