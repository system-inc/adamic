"use strict";
const {createHash} = require("node:crypto");
const ledger = require("./nonempty-endings.json");
// Runtime hashes of the twelve reviewed AST function bodies preserve the complete
// private call graph, fresh nonempty factory returns, singleton calls and no writes.
// Stock emission erases earlier type-only adaptations and the required-read bang.
function validate(ts, sf) {
    const interfaces = sf.statements.filter(n => ts.isInterfaceDeclaration(n) && n.name.text === "ModuleSpecifierPreferences");
    const methods = interfaces.length === 1 ? interfaces[0].members.filter(n => ts.isMethodSignature(n) && n.name.getText(sf) === "getAllowedEndingsInPreferredOrder") : [];
    if (methods.length !== 1 || methods[0].type?.getText(sf) !== "ModuleSpecifierEnding[]") throw new Error("external preference interface drift");
    for (const row of ledger) {
        const found = [];
        function visit(n) {
            if (ts.isFunctionDeclaration(n) && n.name?.text === row.function && n.body) found.push(n);
            ts.forEachChild(n, visit);
        }
        visit(sf);
        if (found.length !== 1) throw new Error("nonempty ending owner drift: " + row.function);
        const js = ts.transpileModule(found[0].getText(sf), {compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,removeComments:true}}).outputText;
        if (createHash("sha256").update(js).digest("hex") !== row.runtimeSha256) throw new Error("nonempty ending invariant drift: " + row.function);
    }
}
module.exports = {validate};
