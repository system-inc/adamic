"use strict";
const fs = require("node:fs"), path = require("node:path"), ts = require("typescript");
if (ts.version !== "6.0.3") throw Error("adaptation 44 requires stock TypeScript 6.0.3");
const tree = path.resolve(process.argv[2]);
const rules = require("./rules.json");
const plans = [];
for (const file of [...new Set(rules.map(r => r.file))]) {
    const name = path.join(tree, file), before = fs.readFileSync(name, "utf8");
    const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true);
    const predicate = require("./predicate.json");
    const owners = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === predicate.name);
    if (owners.length !== 1 || owners[0].getText(source).replaceAll("\r\n", "\n") !== predicate.text) throw Error("protected predicate changed: S02");
    let text = before, changed = 0;
    for (const r of rules.filter(r => r.file === file)) {
        const newline = text.includes("\r\n") ? "\r\n" : "\n";
        const a = r.before.replace(/\r?\n/g, newline), b = r.after.replace(/\r?\n/g, newline);
        if (text.split(b).length === 2 && !text.replace(b, "").includes(a)) continue;
        if (text.split(a).length !== 2 || text.includes(b)) throw Error("unreviewed site: " + r.id);
        text = text.replace(a, b); changed++;
    }
    if (ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length) throw Error("adapted syntax: " + file);
    // All rules must erase exactly, including the namespace emitter's formatting.
    const options = {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext};
    if (ts.transpileModule(before, {compilerOptions: options}).outputText !== ts.transpileModule(text, {compilerOptions: options}).outputText) throw Error("runtime emission changed: " + file);
    plans.push({name, before, text, changed});
}
for (const p of plans) if (fs.readFileSync(p.name, "utf8") !== p.before) throw Error("concurrent source edit");
for (const p of plans) if (p.text !== p.before) fs.writeFileSync(p.name, p.text);
console.log(JSON.stringify({adaptation: 44, sites: plans.reduce((n, p) => n + p.changed, 0)}));
