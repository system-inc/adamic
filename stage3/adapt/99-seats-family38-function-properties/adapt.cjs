"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || "typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: adapt.cjs <adapted TypeScript tree>");

const sites = [
    ["moduleSpecifiers.ts", "ModuleSpecifierPreferences", "getAllowedEndingsInPreferredOrder",
        "getAllowedEndingsInPreferredOrder(syntaxImpliedNodeFormat?: ResolutionMode): ModuleSpecifierEnding[];"],
    ["types.ts", "EvaluationResolver", "evaluateElementAccessExpression",
        "evaluateElementAccessExpression(expr: ElementAccessExpression, location: Declaration | undefined): EvaluatorResult;"],
    ["types.ts", "EvaluationResolver", "evaluateEntityNameExpression",
        "evaluateEntityNameExpression(expr: EntityNameExpression, location: Declaration | undefined): EvaluatorResult;"],
];

const tree = path.resolve(process.argv[2], "src/compiler");
const plans = [];

for (const file of new Set(sites.map(([file]) => file))) {
    const filename = path.join(tree, file);
    const before = fs.readFileSync(filename, "utf8");
    const source = ts.createSourceFile(filename, before, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file + " must parse");
    const edits = [];

    for (const [, owner, name, method] of sites.filter(([siteFile]) => siteFile === file)) {
        const label = owner + "." + name;
        const owners = source.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === owner);
        assert.equal(owners.length, 1, label + " interface count");
        const members = owners[0].members.filter(node => node.name?.getText(source) === name);
        assert.equal(members.length, 1, label + " member count");
        const node = members[0];
        // Extract the original method's type so strictFunctionTypes still treats its arguments bivariantly.
        const propertyName = method.slice(0, method.indexOf("("));
        const property = propertyName + ": { " + name + method.slice(propertyName.length) + " }[" + JSON.stringify(name) + "];";
        const actual = node.getText(source);
        if (ts.isPropertySignature(node) && actual === property) continue;
        assert(ts.isMethodSignature(node), label + " must be a method signature");
        assert.equal(actual, method, label + " signature drift");
        edits.push({ start: node.getStart(source), end: node.end, property });
    }

    let after = before;
    for (const { start, end, property } of edits.sort((a, b) => b.start - a.start)) {
        after = after.slice(0, start) + property + after.slice(end);
    }
    if (edits.length) plans.push({ filename, after, edits: edits.length });
}

for (const { filename, after } of plans) fs.writeFileSync(filename, after);
console.log(JSON.stringify({ files: plans.length, edits: plans.reduce((total, plan) => total + plan.edits, 0) }));
