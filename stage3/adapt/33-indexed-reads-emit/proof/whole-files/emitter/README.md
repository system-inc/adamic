# emitter.ts at zero

All-code source-only latent census **3 -> 0**. Clean files **37 -> 38 of 78**.
This file was read whole for the previous waves; its complete integration diff
was reviewed before this follow-up. Those intervening edits are type-only.

Two original printer-hook reads become assertions. Both result objects are
returned directly by transformNodes, whose single return literal installs the
local isEmitNotificationEnabled function unconditionally. Neither private result
has that property modified before the printer's original read. The optional
TransformationResult interface is broader than this producer.

getBuildInfoText's original JSON.stringify result also becomes an assertion.
The host forwards program.getBuildInfo, installed by builder.ts; all three
builder branches return fresh version-bearing records without toJSON hooks.
emitBuildInfo's fallback is the same ordinary version record. Standard JSON
serialization of those compiler records returns a string or throws. The helper's
string contract therefore matches its actual producer domain. The call and
property entries are individually recorded in required-values.json, with parsed
addresses and one-line invariants; they are not misclassified as indexed reads.

All stock JavaScript bytes and adapter idempotence pass. A hook ! -> ?? 0 mutant
and a JSON-result ! -> ?? 0 mutant each fail the independent JavaScript comparison
and site contract. No reads move and no initializer assertions are introduced.
Default oracle: **106367 passing**, zero failing/pending,
**empty baseline diff**, 229.955 seconds. The mechanical API projection
permits only 20's 189 and 40's 28 lines and compares every other reference byte.

Reproduce with latent-file-census.sh /tmp/emit33-close-meter SOURCE OUTPUT,
verify.cjs /tmp/emit33-close-source-before-emitter SOURCE, and the default oracle
on /tmp/emit33-close-tree. SOURCE is /tmp/emit33-close-meter-source. Mutant
expressions are transform.isEmitNotificationEnabled and JSON.stringify(buildInfo).
