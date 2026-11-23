"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");
const pins = require("./pins.json");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
const tree = process.argv[2];
const plans = new Map();

function source(relative) {
    if (!plans.has(relative)) {
        const fileName = path.join(tree, relative);
        const file = ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
        assert.equal(file.parseDiagnostics.length, 0, fileName);
        plans.set(relative, { file, edits: [] });
    }
    return plans.get(relative);
}

function replace(plan, node, text) {
    plan.edits.push({ start: node.getStart(plan.file), end: node.end, text });
}

const owner = source(pins.owner);
const declarations = owner.file.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === pins.old);
assert.deepEqual(declarations.map(node => node.getText(owner.file)), pins.declarations);
assert(!owner.file.statements.some(node => ts.isFunctionDeclaration(node) && node.name?.text === pins.twin));
const end = declarations.at(-1).end;
const newline = owner.file.text.includes("\r\n") ? "\r\n" : "\n";
owner.edits.push({ start: end, end, text: [ "", "", ...pins.helper ].join(newline) });

let calls = 0;
for (const filePin of pins.files) {
    const plan = source(filePin.file);
    const { file } = plan;
    const candidates = nodes(ts, file, ts.isCallExpression);
    for (const callPin of filePin.calls) {
        const found = candidates.filter(node => node.getText(file) === callPin.text);
        assert.equal(found.length, callPin.count, `${filePin.file}: ${callPin.text}`);
        for (const call of found) {
            assert.equal(call.expression.getText(file), callPin.expression);
            assert(!call.typeArguments?.length);
            const name = ts.isPropertyAccessExpression(call.expression) ? call.expression.name : call.expression;
            assert(ts.isIdentifier(name) && name.text === pins.old);
            replace(plan, name, pins.twin);
            calls++;
        }
    }
    if (filePin.import) {
        const { text, module, keep } = filePin.import;
        const found = nodes(ts, file, node => ts.isImportSpecifier(node) && node.getText(file) === text
            && node.parent.parent.parent.moduleSpecifier.text === module);
        assert.equal(found.length, 1, `${filePin.file}: ${text} from ${module}`);
        const spelling = keep ? text + "," + (file.text.includes("\r\n") ? "\r\n" : "\n") + "    " + pins.twin : pins.twin;
        replace(plan, found[0], spelling);
    }
}

for (const plan of plans.values()) {
    let text = plan.file.text;
    let boundary = text.length;
    for (const edit of plan.edits.sort((a, b) => b.start - a.start)) {
        assert(edit.end <= boundary, `overlapping edits in ${plan.file.fileName}`);
        text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
        boundary = edit.start;
    }
    assert.equal(ts.createSourceFile(plan.file.fileName, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    plan.after = text;
}

for (const { file, after } of plans.values()) fs.writeFileSync(file.fileName, after);
console.log(JSON.stringify({ twin: pins.twin, files: plans.size, calls }));
