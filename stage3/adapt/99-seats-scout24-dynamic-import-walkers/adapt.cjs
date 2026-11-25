"use strict";

const pins = {
    "owner": "src/compiler/utilities.ts",
    "name": "forEachDynamicImportOrRequireCall",
    "value": "forEachDynamicImportOrRequireCallExpression",
    "original": "export function forEachDynamicImportOrRequireCall<IncludeTypeSpaceImports extends boolean, RequireStringLiteralLikeArgument extends boolean>(\r\n    file: SourceFile,\r\n    includeTypeSpaceImports: IncludeTypeSpaceImports,\r\n    requireStringLiteralLikeArgument: RequireStringLiteralLikeArgument,\r\n    cb: (node: CallExpression | (IncludeTypeSpaceImports extends false ? never : JSDocImportTag | ImportTypeNode), argument: RequireStringLiteralLikeArgument extends true ? StringLiteralLike : Expression) => void,\r\n): void {\r\n    const isJavaScriptFile = isInJSFile(file);\r\n    const r = /import|require/g;\r\n    while (r.exec(file.text) !== null) { // eslint-disable-line no-restricted-syntax\r\n        const node = getNodeAtPosition(file, r.lastIndex, /*includeJSDoc*/ includeTypeSpaceImports);\r\n        if (isJavaScriptFile && isRequireCall(node, requireStringLiteralLikeArgument)) {\r\n            cb(node, node.arguments[0] as RequireStringLiteralLikeArgument extends true ? StringLiteralLike : Expression);\r\n        }\r\n        else if (isImportCall(node) && node.arguments.length >= 1 && (!requireStringLiteralLikeArgument || isStringLiteralLike(node.arguments[0]!))) {\r\n            cb(node, node.arguments[0] as RequireStringLiteralLikeArgument extends true ? StringLiteralLike : Expression);\r\n        }\r\n        else if (includeTypeSpaceImports && isLiteralImportTypeNode(node)) {\r\n            (cb as (node: CallExpression | JSDocImportTag | ImportTypeNode, argument: StringLiteralLike) => void)(node, node.argument.literal);\r\n        }\r\n        else if (includeTypeSpaceImports && isJSDocImportTag(node)) {\r\n            const moduleNameExpr = getExternalModuleName(node);\r\n            if (moduleNameExpr && isStringLiteral(moduleNameExpr) && moduleNameExpr.text) {\r\n                (cb as (node: CallExpression | JSDocImportTag | ImportTypeNode, argument: StringLiteralLike) => void)(node, moduleNameExpr);\r\n            }\r\n        }\r\n    }\r\n}",
    "next": "export function forEachDynamicImportOrRequireCall(\r\n    file: SourceFile,\r\n    includeTypeSpaceImports: true,\r\n    requireStringLiteralLikeArgument: true,\r\n    cb: (node: CallExpression | JSDocImportTag | ImportTypeNode, argument: StringLiteralLike) => void,\r\n): void {\r\n    const isJavaScriptFile = isInJSFile(file);\r\n    const r = /import|require/g;\r\n    while (r.exec(file.text) !== null) { // eslint-disable-line no-restricted-syntax\r\n        const node = getNodeAtPosition(file, r.lastIndex, /*includeJSDoc*/ true);\r\n        if (isJavaScriptFile && isRequireCall(node, /*requireStringLiteralLikeArgument*/ false) && isStringLiteralLike(node.arguments[0]!)) {\r\n            cb(node, node.arguments[0]);\r\n        }\r\n        else if (isImportCall(node) && node.arguments.length >= 1 && isStringLiteralLike(node.arguments[0]!)) {\r\n            cb(node, node.arguments[0]);\r\n        }\r\n        else if (isLiteralImportTypeNode(node)) {\r\n            cb(node, node.argument.literal);\r\n        }\r\n        else if (isJSDocImportTag(node)) {\r\n            const moduleNameExpr = getExternalModuleName(node);\r\n            if (moduleNameExpr && isStringLiteral(moduleNameExpr) && moduleNameExpr.text) {\r\n                cb(node, moduleNameExpr);\r\n            }\r\n        }\r\n    }\r\n}\r\n\r\n/** @internal */\r\nexport function forEachDynamicImportOrRequireCallExpression(\r\n    file: SourceFile,\r\n    cb: (node: CallExpression, argument: Expression) => void,\r\n): void {\r\n    const isJavaScriptFile = isInJSFile(file);\r\n    const r = /import|require/g;\r\n    while (r.exec(file.text) !== null) { // eslint-disable-line no-restricted-syntax\r\n        const node = getNodeAtPosition(file, r.lastIndex, /*includeJSDoc*/ false);\r\n        if (isJavaScriptFile && isRequireCall(node, /*requireStringLiteralLikeArgument*/ false)) {\r\n            cb(node, node.arguments[0]!);\r\n        }\r\n        else if (isImportCall(node) && node.arguments.length >= 1) {\r\n            cb(node, node.arguments[0]!);\r\n        }\r\n    }\r\n}",
    "callers": [
        {
            "file": "src/compiler/transformers/module/module.ts",
            "before": "forEachDynamicImportOrRequireCall(node, /*includeTypeSpaceImports*/ false, /*requireStringLiteralLikeArgument*/ false, node => {\r\n                if (!isStringLiteralLike(node.arguments[0]!) || shouldRewriteModuleSpecifier(node.arguments[0].text, compilerOptions)) {\r\n                    importsAndRequiresToRewriteOrShim = append(importsAndRequiresToRewriteOrShim, node);\r\n                }\r\n            })",
            "after": "forEachDynamicImportOrRequireCallExpression(node, node => {\r\n                if (!isStringLiteralLike(node.arguments[0]!) || shouldRewriteModuleSpecifier(node.arguments[0].text, compilerOptions)) {\r\n                    importsAndRequiresToRewriteOrShim = append(importsAndRequiresToRewriteOrShim, node);\r\n                }\r\n            })",
            "module": "../../_namespaces/ts.js"
        },
        {
            "file": "src/compiler/transformers/module/esnextAnd2015.ts",
            "before": "forEachDynamicImportOrRequireCall(node, /*includeTypeSpaceImports*/ false, /*requireStringLiteralLikeArgument*/ false, node => {\r\n                    if (!isStringLiteralLike(node.arguments[0]!) || shouldRewriteModuleSpecifier(node.arguments[0].text, compilerOptions)) {\r\n                        importsAndRequiresToRewriteOrShim = append(importsAndRequiresToRewriteOrShim, node);\r\n                    }\r\n                })",
            "after": "forEachDynamicImportOrRequireCallExpression(node, node => {\r\n                    if (!isStringLiteralLike(node.arguments[0]!) || shouldRewriteModuleSpecifier(node.arguments[0].text, compilerOptions)) {\r\n                        importsAndRequiresToRewriteOrShim = append(importsAndRequiresToRewriteOrShim, node);\r\n                    }\r\n                })",
            "module": "../../_namespaces/ts.js"
        }
    ]
};

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
const root = process.argv[2];
function source(file) {
    const filename = path.join(root, file);
    const text = fs.readFileSync(filename, "utf8");
    const ast = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
    assert.equal(ast.parseDiagnostics.length, 0, file);
    const edits = [];
    return { ast, edits, replace(node, replacement) {
        edits.push({ start: node.getStart(ast), end: node.end, replacement });
    } };
}

function only(ast, match) {
    const hits = [];
    function visit(node) {
        if (match(node)) hits.push(node);
        ts.forEachChild(node, visit);
    }
    visit(ast);
    assert.equal(hits.length, 1, ast.fileName);
    return hits[0];
}

const owner = source(pins.owner);
const declaration = only(owner.ast, node =>
    ts.isFunctionDeclaration(node) && node.name?.text === pins.name);
assert.equal(declaration.getText(owner.ast), pins.original);
assert(!owner.ast.statements.some(node => node.name?.text === pins.value));
owner.replace(declaration, pins.next);

const plans = [owner];
for (const pin of pins.callers) {
    const plan = source(pin.file);
    const call = only(plan.ast, node =>
        ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === pins.name);
    assert.equal(call.getText(plan.ast), pin.before);
    const imported = only(plan.ast, node =>
        ts.isImportSpecifier(node) && node.name.text === pins.name &&
        node.parent.parent.parent.moduleSpecifier.text === pin.module);
    assert.equal(imported.getText(plan.ast), pins.name);
    plan.replace(call, pin.after);
    plan.replace(imported, pins.value);
    plans.push(plan);
}

for (const plan of plans) {
    const { ast, edits } = plan;
    let text = ast.text;
    for (const { start, end, replacement } of edits.sort((a, b) => b.start - a.start)) {
        text = text.slice(0, start) + replacement + text.slice(end);
    }
    assert.equal(ts.createSourceFile(ast.fileName, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    plan.text = text;
}
for (const { ast, text } of plans) fs.writeFileSync(ast.fileName, text);
console.log(JSON.stringify({ files: plans.length, redirectedCalls: pins.callers.length }));
