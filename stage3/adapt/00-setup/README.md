# 00 setup

Runs exactly upstream Herebyfile.mjs's generate-diagnostics command. It creates
both diagnosticInformationMap.generated.ts and diagnosticMessages.generated.json.
Upstream's generator compares existing bytes before writing, so this is idempotent.
No compiler source is rewritten. There are no type decisions or declined sites;
using the stock compiler API here would add a dependency to a dependency-free
upstream generation step without improving its correctness.

Adamic needs the generated compiler module before checking the complete tree.
Counts against the pristine tree prepared by the same upstream step: zero files,
zero added lines, zero removed lines. Generated outputs are build preparation,
not an adaptation, and are included in the measurement snapshot for later edits.
