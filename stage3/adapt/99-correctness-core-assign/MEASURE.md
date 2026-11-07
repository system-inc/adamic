# assign measurement, task b5w3ycg

Measurement checkpoint. No assign adaptation has been built. Runtime and
spelling results remain pending. The baseline full apply completed before
this report-only directory was introduced. No after apply is claimed.

The complete detached_own.go was read. At the refusal scan, arg has generic
T narrowed past undefined, and T extends object supplies no known native
ir.Object representation. The adapted hasProperty(map: object, key: string)
uses the own-property-only opaque parameter path and requires ir.Object at
that call boundary, rejecting an unknown or different representation.

All production assign calls resolve to core.assign: seven calls in
compiler/commandLineParser.ts and one in server/scriptInfo.ts, with no
services call. The server arguments are FormatCodeSettings, with no string
index signature; optionMapToObject's declared return type is object.
The remaining nonempty compiler argument types have string index signatures:
CompilerOptions, WatchOptions, OptionsBase, and MapLike<string[]>.
There is also one harness call. Complete argument types are in callers.json.

No adapt.cjs or owners/sites JSON pins assign's owner. Adaptation 32's two
assign mentions own indexed reads of its returned value, not this function.

The required whole-lane observer is isolated in a scratch copy. Seed checks
passed for plain, null-prototype and cross-realm objects, inherited enumerable
keys, own/inherited accessors and symbols, with zero getter invocation.
Runtime measurement will precede all spelling trials.
