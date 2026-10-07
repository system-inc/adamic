"use strict";
const sites = require("./handoff-sites.json");
function owners(ts, source, site) {
    const matches = [];
    function visit(n) {
        if (ts.isInterfaceDeclaration(n) && n.name.text === site.interface) {
            if (!n.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword)) throw new Error("handoff export drift");
            const internal = (ts.getJSDocTags(n) || []).some(t => t.tagName.text === "internal");
            if (internal === site.public) throw new Error("handoff visibility drift");
            matches.push(...n.members.filter(m => ts.isPropertySignature(m) && m.name.getText(source) === site.name));
        }
        ts.forEachChild(n, visit);
    }
    visit(source);
    if (matches.length !== 1 || !matches[0].questionToken || !matches[0].type) throw new Error("handoff owner occurrence drift: " + site.interface + "." + site.name);
    if (!!(ts.getCombinedModifierFlags(matches[0]) & ts.ModifierFlags.Readonly) !== site.readonly) throw new Error("handoff readonly modifier drift: " + site.interface + "." + site.name);
    return matches[0];
}
function plan(ts, file, text, check) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error("handoff parse failure: " + file);
    const edits = [];
    for (const site of sites.filter(s => s.file === file)) {
        const node = owners(ts, source, site).type, actual = node.getText(source);
        if (site.action === "decline") {
            if (actual !== site.before) throw new Error("deferred handoff type drift: " + site.interface);
            continue;
        }
        if (site.action !== "owner-union") throw new Error("unknown handoff action");
        if (actual === site.before) {
            if (check) throw new Error("handoff union missing: " + site.interface + "." + site.name);
            edits.push({at:node.getStart(source),end:node.end,text:site.after});
        } else if (actual !== site.after) throw new Error("handoff type drift: " + site.interface + "." + site.name);
    }
    for (const edit of edits.sort((a,b) => b.at-a.at)) text = text.slice(0,edit.at)+edit.text+text.slice(edit.end);
    return {text,contracts:edits.length};
}
function validate(ts, tree) {
    const fs = require("node:fs"), path = require("node:path"), crypto = require("node:crypto");
    for (const guard of require("./handoff-guards.json")) {
        const name = path.join(tree, "src/compiler", guard.file);
        const text = fs.readFileSync(name, "utf8"), source = ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);
        if (source.parseDiagnostics.length) throw new Error("handoff producer parse failure");
        const matches = [];
        function visit(n) {
            if (ts.isFunctionDeclaration(n) && n.name?.text === guard.function) matches.push(n);
            ts.forEachChild(n, visit);
        }
        visit(source);
        if (matches.length !== 1 || !matches[0].body) throw new Error("handoff producer occurrence drift");
        const js = ts.transpileModule(matches[0].getText(source), {compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;
        if (crypto.createHash("sha256").update(js).digest("hex") !== guard.runtime_sha256) throw new Error("handoff producer runtime drift: " + guard.function);
    }
}
module.exports = {plan,owners,validate};
