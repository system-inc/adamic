"use strict";
const sites = require("./closure-sites.json");
function plan(ts, file, text, check) {
    const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (sf.parseDiagnostics.length) throw new Error(`closure parse failure: ${file}`);
    const edits = [];
    for (const site of sites.filter(s => s.file === file)) {
        const matches = [];
        function visit(n) {
            if (ts.isCallExpression(n) && n.expression.getText(sf) === site.call) {
                let owner = n.parent;
                while (owner && !(ts.isFunctionDeclaration(owner) && owner.name)) owner = owner.parent;
                if (owner && owner.name.text === site.function) matches.push(n);
            }
            ts.forEachChild(n, visit);
        }
        visit(sf);
        if (matches.length !== 1) throw new Error(`closure occurrence drift: ${file}:${site.function}:${site.call}`);
        const n = matches[0];
        if (JSON.stringify(n.arguments.map(a => a.getText(sf))) !== JSON.stringify(site.arguments)) throw new Error(`closure argument drift: ${file}:${site.function}`);
        if (n.typeArguments) {
            if (JSON.stringify(n.typeArguments.map(t => t.getText(sf))) !== JSON.stringify(site.typeArguments)) throw new Error(`closure type drift: ${file}:${site.function}`);
        }
        else {
            if (check) throw new Error(`closure types missing: ${file}:${site.function}`);
            edits.push({at: n.expression.end, text: `<${site.typeArguments.join(", ")}>`});
        }
    }
    for (const e of edits.sort((a,b) => b.at-a.at)) text = text.slice(0,e.at)+e.text+text.slice(e.at);
    return {text, contracts: edits.length};
}
module.exports = {plan};
