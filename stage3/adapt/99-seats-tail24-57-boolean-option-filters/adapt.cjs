"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const file = "src/compiler/commandLineParser.ts";
const sourceExpression = "option.allowConfigDirTemplateSubstitution || (!option.isCommandLineOnly && option.isFilePath)";
const newExpression = `(${sourceExpression}) ? true : false`;
const targets = [
    ["configDirTemplateSubstitutionOptions", "optionDeclarations"],
    ["configDirTemplateSubstitutionWatchOptions", "optionsForWatch"],
];

function locate(text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, "invalid option source");
    return targets.map(([name, table]) => {
        const statements = source.statements.filter(statement =>
            ts.isVariableStatement(statement) &&
            statement.declarationList.declarations.some(declaration => declaration.name.getText(source) === name));
        assert.equal(statements.length, 1, `missing or duplicate option statement: ${name}`);
        const statement = statements[0];
        const before = [
            `const ${name}: readonly CommandLineOption[] = ${table}.filter(`,
            `    option => ${sourceExpression},`,
            ");",
        ].join("\r\n");
        const after = before.replace(sourceExpression, () => newExpression);
        const current = statement.getText(source);
        assert(current === before || current === after, `option statement drift: ${name}`);
        const declaration = statement.declarationList.declarations[0];
        assert(ts.isCallExpression(declaration.initializer));
        const arrow = declaration.initializer.arguments[0];
        assert(ts.isArrowFunction(arrow));
        return { source, arrow, statement, before, after, current };
    });
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const filePath = path.join(tree, file);
    const text = fs.readFileSync(filePath, "utf8");
    const sites = locate(text);
    const edits = sites.filter(site => site.current === site.before);
    assert(edits.length === 0 || edits.length === sites.length, "partially adapted option filters");
    if (edits.length) {
        let revised = text;
        for (const site of edits.reverse()) {
            revised = revised.slice(0, site.arrow.body.getStart(site.source)) +
                newExpression + revised.slice(site.arrow.body.end);
        }
        locate(revised);
        assert.equal(fs.readFileSync(filePath, "utf8"), text, "source changed during planning");
        fs.writeFileSync(filePath, revised);
    }
    return { files: Number(edits.length > 0), edits: edits.length };
}

module.exports = { adapt, file, locate };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
