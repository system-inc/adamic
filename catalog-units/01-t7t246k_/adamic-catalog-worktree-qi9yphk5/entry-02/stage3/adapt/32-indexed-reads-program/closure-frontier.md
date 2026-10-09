# Current frontier after public method restoration

The unchanged latent run 0 all-code meter observes **310 -> 317 findings** on
area/stage3 06f89a04 plus the restoration. Whole original files at zero are
**55 -> 54 of 78**. Only **watch.ts leaves zero (0 -> 2)**. program.ts is
10 -> 13 and watchPublic.ts is 10 -> 12; both were already nonzero. The exact
seven returned findings and their reasons are in proof/method-restoration/census.json.

Both watch findings are declined because public returned hosts and injected
CreateProgram callbacks can observe own keys holding undefined. Node witnesses
show omission changes in, hasOwnProperty, keys, spread, for...in and Object.assign.
The original method declarations remain, and host literals remain byte-identical.

All remaining per-file findings are in closure-frontier.json. The original
required-read ledger and historical five-code counts remain unchanged. Tracing
and performanceCore still need Node host bindings; RawSourceMap still requires
partition 33's coupled private-owner fix. No guessed source shim is added.
