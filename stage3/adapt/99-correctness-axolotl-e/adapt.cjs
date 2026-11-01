#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const before = "const heritageClauses = factory.createNodeArray(map(input.heritageClauses, clause => {\n                        if (clause.token === SyntaxKind.ExtendsKeyword) {\n                            const oldDiag = getSymbolAccessibilityDiagnostic;\n                            getSymbolAccessibilityDiagnostic = createGetSymbolAccessibilityDiagnosticForNode(clause.types[0]);\n                            const newClause = factory.updateHeritageClause(clause, map(clause.types, t => factory.updateExpressionWithTypeArguments(t, newId, visitNodes(t.typeArguments, visitDeclarationSubtree, isTypeNode))));\n                            getSymbolAccessibilityDiagnostic = oldDiag;\n                            return newClause;\n                        }\n                        return factory.updateHeritageClause(clause, visitNodes(factory.createNodeArray(filter(clause.types, t => isEntityNameExpression(t.expression) || t.expression.kind === SyntaxKind.NullKeyword)), visitDeclarationSubtree, isExpressionWithTypeArguments));\n                    }));";
const after = "const heritageClauses = factory.createNodeArray(mapDefined(input.heritageClauses, clause => {\n                        if (clause.token === SyntaxKind.ExtendsKeyword) {\n                            const firstType = clause.types[0];\n                            if (!firstType) return undefined;\n                            const oldDiag = getSymbolAccessibilityDiagnostic;\n                            getSymbolAccessibilityDiagnostic = createGetSymbolAccessibilityDiagnosticForNode(firstType);\n                            const newClause = factory.updateHeritageClause(clause, map(clause.types, t => factory.updateExpressionWithTypeArguments(t, newId, visitNodes(t.typeArguments, visitDeclarationSubtree, isTypeNode))));\n                            getSymbolAccessibilityDiagnostic = oldDiag;\n                            return newClause;\n                        }\n                        return factory.updateHeritageClause(clause, visitNodes(factory.createNodeArray(filter(clause.types, t => isEntityNameExpression(t.expression) || t.expression.kind === SyntaxKind.NullKeyword)), visitDeclarationSubtree, isExpressionWithTypeArguments));\n                    }));";
function adapt(tree, ts) {
    if (ts.version !== "6.0.3") throw new Error(`want stock typescript@6.0.3, got ${ts.version}`);
    const file = path.join(tree, "src/compiler/transformers/declarations.ts");
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error("cannot parse declarations.ts");
    const functions = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "transformTopLevelDeclaration") functions.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (functions.length !== 1) throw new Error("want one transformTopLevelDeclaration");
    const fn = functions[0];
    const candidates = [];
    function find(node) {
        if (ts.isVariableDeclaration(node) && node.name.getText(source) === "heritageClauses" &&
            node.initializer?.getText(source).startsWith("factory.createNodeArray(")) candidates.push(node.parent.parent);
        ts.forEachChild(node, find);
    }
    find(fn);
    // The second heritageClauses local uses transformHeritageClauses instead.
    if (candidates.length !== 1) throw new Error("unreviewed heritage mapping count");
    const statement = candidates[0];
    const actual = statement.getText(source).replace(/\r\n/g, "\n");
    if (actual !== before && actual !== after) throw new Error("unreviewed or half-adapted heritage mapping");
    if (actual === after) return { files: 0, edits: 0, javascriptChanging: true };
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    const result = text.slice(0, statement.getStart(source)) + after.replace(/\n/g, newline) + text.slice(statement.end);
    if (ts.createSourceFile(file, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length) throw new Error("adapted syntax error");
    if (fs.readFileSync(file, "utf8") !== text) throw new Error("declarations.ts changed while planning");
    fs.writeFileSync(file, result);
    return { files: 1, edits: 1, javascriptChanging: true };
}
module.exports = { adapt };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript"))));
    } catch (error) { console.error(error.message); process.exitCode = 1; }
}
