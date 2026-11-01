#!/usr/bin/env node
"use strict";
// Adds propertyOf to 47's src/compiler/hostErrors.ts and changes no call site.
// Later adaptations read a property off an unknown value through it.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");

// The helper reads a property the way JavaScript does: null and undefined throw
// V8's own TypeError, and any other value, primitives included, is read through
// its object wrapper with the original value as the receiver, so a getter sees
// the same `this` a plain read would give it.
const helper = `
/** @internal */
export function propertyOf(value: unknown, key: string): unknown {
    if (value === null || value === undefined) {
        throw new TypeError(\`Cannot read properties of \${String(value)} (reading '\${key}')\`);
    }
    return Reflect.get(Object(value), key, value);
}
`;
// 47's two helpers, the only statements this adapter accepts before its own.
const expected = ["errorCode", "errorMessage"];

function plan(tree, ts) {
    assert.equal(ts.version, "6.0.3", `want stock typescript@6.0.3, got ${ts.version}`);
    const fileName = path.join(tree, "src/compiler/hostErrors.ts");
    assert(fs.existsSync(fileName), "src/compiler/hostErrors.ts is missing: adaptation 47 must run first");
    const text = fs.readFileSync(fileName, "utf8");
    const file = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(file.parseDiagnostics.length, 0, "cannot parse hostErrors.ts");
    const names = file.statements.map(statement => {
        assert(ts.isFunctionDeclaration(statement) && statement.name, "unreviewed statement in hostErrors.ts");
        const exported = statement.modifiers?.map(modifier => modifier.kind);
        assert.deepEqual(exported, [ts.SyntaxKind.ExportKeyword], `unreviewed modifiers on ${statement.name.text}`);
        return statement.name.text;
    });
    if (names.length === expected.length) {
        assert.deepEqual(names, expected, "unreviewed hostErrors.ts functions");
        assert(!text.includes("propertyOf"), "propertyOf named outside a reviewed declaration");
        const after = text + helper;
        const parsed = ts.createSourceFile(fileName, after, ts.ScriptTarget.Latest, true);
        assert.equal(parsed.parseDiagnostics.length, 0, "adapted syntax error");
        assert.deepEqual(parsed.statements.map(statement => statement.name.text), [...expected, "propertyOf"]);
        return { fileName, before: text, after };
    }
    // A second run finds the helper already appended, byte for byte, and writes nothing.
    assert.deepEqual(names, [...expected, "propertyOf"], "unreviewed hostErrors.ts functions");
    assert(text.endsWith(helper), "propertyOf already exists with different contents");
    return { fileName, before: text, after: text };
}

function adapt(tree, ts) {
    const change = plan(tree, ts);
    assert.equal(fs.readFileSync(change.fileName, "utf8"), change.before, "source changed while planning");
    if (change.after !== change.before) fs.writeFileSync(change.fileName, change.after);
    return { files: change.after === change.before ? 0 : 1, helperAdded: change.after !== change.before, callSites: 0 };
}

module.exports = { adapt, helper, plan };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
