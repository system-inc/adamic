"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

const sites = [
    ["tryLoadModuleUsingPathsIfEligible", ["baseUrl", "paths"]],
    ["tryLoadModuleUsingBaseUrl", ["baseUrl"]],
];
const held = "heldCompilerOptions";

function adapt(tree) {
    const file = path.join(tree, "src/compiler/moduleNameResolver.ts");
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    assert.equal(nodes(ts, source, node => ts.isIdentifier(node) && node.text === held).length, 0,
        `${held} must be fresh in moduleNameResolver.ts`);

    const edits = sites.map(([owner, names]) => {
        const functions = source.statements.filter(node =>
            ts.isFunctionDeclaration(node) && node.name?.text === owner);
        assert.equal(functions.length, 1, owner);
        const statements = functions[0].body.statements.filter(node =>
            ts.isVariableStatement(node) && node.declarationList.declarations.some(declaration =>
                ts.isObjectBindingPattern(declaration.name) &&
                declaration.initializer?.getText(source) === "state.compilerOptions"));
        assert.equal(statements.length, 1, owner);
        const statement = statements[0];
        assert.equal(statement.getText(source), `const { ${names.join(", ")} } = state.compilerOptions;`, owner);

        const reads = names.map(name => `    const ${name} = ${held}.${name};`);
        return {
            start: statement.getStart(source),
            end: statement.end,
            replacement: [`const ${held} = state.compilerOptions;`, ...reads].join(newline),
        };
    });

    let result = text;
    for (const { start, end, replacement } of edits.sort((left, right) => right.start - left.start)) {
        result = result.slice(0, start) + replacement + result.slice(end);
    }
    fs.writeFileSync(file, result);
}

if (require.main === module) adapt(process.argv[2]);
module.exports = { adapt };
