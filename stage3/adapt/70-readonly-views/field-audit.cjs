#!/usr/bin/env node
'use strict';
// Read-only inventory of field owners. Escapes are declines, not write proofs.
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error('expected stock TypeScript 6.0.3');
if (process.argv.length !== 5) throw new Error('usage: node field-audit.cjs <tree> <inventory.json.gz> <output.json.gz>');
const tree = path.resolve(process.argv[2]);
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(process.argv[3])));
const program = ts.createProgram(ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']), {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    noImplicitReturns: true, noFallthroughCasesInSwitch: true, verbatimModuleSyntax: true,
    erasableSyntaxOnly: true, allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
});
const checker = program.getTypeChecker();
function where(node) {
    const sf = node.getSourceFile(), p = sf.getLineAndCharacterOfPosition(node.getStart(sf));
    return `${path.relative(tree, sf.fileName)}:${p.line + 1}:${p.character + 1}`;
}
function at(sf, pos) {
    let result = sf;
    function visit(node) {
        ts.forEachChild(node, child => {
            if (child.getFullStart() <= pos && pos < child.end) { result = child; visit(child); }
        });
    }
    visit(sf);
    return result;
}
const owners = new Map();
const sites = [];
for (const site of inventory.sites) {
    const fields = [];
    for (const context of site.contexts.slice(0, 1)) {
        for (const field of context.fields) for (const declaration of field.declarations) {
            if (!declaration.where.startsWith('src/')) continue;
            const [file, line, column] = declaration.where.split(':');
            const sf = program.getSourceFile(path.join(tree, file));
            let node = at(sf, sf.getPositionOfLineAndCharacter(+line - 1, +column - 1));
            while (node && !ts.isPropertySignature(node)) node = node.parent;
            if (!node) throw new Error(`cannot resolve ${declaration.where}`);
            if (!owners.has(node)) owners.set(node, {
                where: where(node), name: field.name, api: declaration.api,
                parent: ts.isInterfaceDeclaration(node.parent) ? node.parent.name.text : ts.SyntaxKind[node.parent.kind],
                readonly: !!(ts.getCombinedModifierFlags(node) & ts.ModifierFlags.Readonly),
                references: 0, copies: [], writes: [], escapes: [],
            });
            fields.push(where(node));
        }
    }
    sites.push({ where: site.where, reason: site.reason, fields: [...new Set(fields)],
        contextual_owner: site.contexts[0]?.owner, target: site.contexts[0]?.target });
}
function assignment(node) {
    let current = node;
    while (current.parent) {
        const p = current.parent;
        if (ts.isParenthesizedExpression(p) || ts.isArrayLiteralExpression(p) || ts.isObjectLiteralExpression(p) ||
            ts.isPropertyAssignment(p) || ts.isSpreadAssignment(p) || ts.isSpreadElement(p)) { current = p; continue; }
        if (ts.isBinaryExpression(p) && p.left === current && p.operatorToken.kind >= ts.SyntaxKind.FirstAssignment &&
            p.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return p;
        if ((ts.isPrefixUnaryExpression(p) || ts.isPostfixUnaryExpression(p)) &&
            [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(p.operator)) return p;
        if (ts.isDeleteExpression(p) || ((ts.isForOfStatement(p) || ts.isForInStatement(p)) && p.initializer === current)) return p;
        return undefined;
    }
}
function copied(node) {
    let current = node;
    while (current.parent) {
        const p = current.parent;
        if (ts.isParenthesizedExpression(p) || ts.isNonNullExpression(p) ||
            (ts.isBinaryExpression(p) && [ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken,
                ts.SyntaxKind.QuestionQuestionToken].includes(p.operatorToken.kind)) ||
            (ts.isConditionalExpression(p) && p.condition !== current)) { current = p; continue; }
        return ts.isVariableDeclaration(p) && p.initializer === current &&
            (ts.isObjectBindingPattern(p.name) || ts.isArrayBindingPattern(p.name));
    }
    return false;
}
const cache = new Map();
function fields(type) {
    type = checker.getNonNullableType(type);
    if (cache.has(type)) return cache.get(type);
    const result = new Set();
    for (const property of checker.getPropertiesOfType(type)) {
        for (const declaration of property.declarations || []) if (owners.has(declaration)) result.add(declaration);
    }
    cache.set(type, [...result]);
    return cache.get(type);
}
for (const sf of program.getSourceFiles()) {
    if (!path.resolve(sf.fileName).startsWith(tree + path.sep + 'src' + path.sep)) continue;
    function visit(node) {
        if (ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) {
            const receiver = checker.getNonNullableType(checker.getTypeAtLocation(node.expression));
            const name = ts.isPropertyAccessExpression(node) ? node.name.text :
                ts.isStringLiteralLike(node.argumentExpression) ? node.argumentExpression.text : undefined;
            const property = name && checker.getPropertyOfType(receiver, name);
            const selected = name ? (property?.declarations || []).filter(d => owners.has(d)) : fields(receiver);
            const write = assignment(node);
            for (const declaration of selected) {
                const owner = owners.get(declaration);
                if (name) owner.references++;
                if (write) owner.writes.push({ where: where(write), expression: write.getText().slice(0, 500),
                    value: ts.isBinaryExpression(write) ? checker.typeToString(checker.getTypeAtLocation(write.right)) : undefined });
            }
        }
        if (ts.isExpressionNode(node) && node.parent?.name !== node) {
            const selected = fields(checker.getTypeAtLocation(node));
            if (selected.length) {
                if (copied(node)) for (const declaration of selected) owners.get(declaration).copies.push(where(node));
                else {
                    let target = checker.getContextualType(node);
                    if (ts.isAsExpression(node.parent) || ts.isTypeAssertionExpression(node.parent))
                        target = checker.getTypeFromTypeNode(node.parent.type);
                    if (target) {
                        const targetFields = new Set(fields(target));
                        for (const declaration of selected) if (!targetFields.has(declaration)) {
                            owners.get(declaration).escapes.push({ where: where(node), target: checker.typeToString(target) });
                        }
                    }
                }
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(sf);
}
const result = { typescript: ts.version, sites, owners: [...owners.values()],
    limit: 'Shallow field replacement audit. Array/Map/Set element mutation and parameter alias chains require separate owners. Escapes prevent readonly certification.' };
fs.writeFileSync(process.argv[4], zlib.gzipSync(JSON.stringify(result, null, 2) + '\n'));
console.log(JSON.stringify({ sites: sites.length, owners: result.owners.length,
    no_writes_or_escapes: result.owners.filter(o => !o.writes.length && !o.escapes.length).length }));
