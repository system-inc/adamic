#!/usr/bin/env node
'use strict';
// #rv75m9k road B, Adamic's two counts for the five arrows, checker-resolved over all of upstream's src (v6.0.3):
// 1. writes to System's directoryExists and createDirectory, and to getCanonicalFileName on a CompilerHost-like value,
//    plus every call of changeCompilerHostLikeToUseCache with the type of the host it's handed;
// 2. every use of createModuleResolutionCache's getCanonicalFileName parameter, followed through each callee it's
//    passed to, flagging any use that isn't a call or a pass-through: identity, length, name, a store.
// Usage: node arrow-counts.cjs <TypeScript checkout>
const ts = require('/Users/kirkouimet/Projects/ahra/node_modules/typescript');
const path = require('node:path'), fs = require('node:fs');
const root = process.argv[2];
const files = [];
(function walk(directory) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
        const full = path.join(directory, entry.name);
        if (entry.isDirectory()) { if (entry.name !== 'lib' && entry.name !== 'node_modules') walk(full); }
        else if (full.endsWith('.ts') && !full.endsWith('.d.ts')) files.push(full);
    }
})(path.join(root, 'src'));
const program = ts.createProgram(files, { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS, strict: true, noEmit: true, types: ['node'], typeRoots: ['/Users/kirkouimet/Projects/ahra/node_modules/@types'], skipLibCheck: true });
const checker = program.getTypeChecker();
const rel = node => { const s = node.getSourceFile(); const p = s.getLineAndCharacterOfPosition(node.getStart()); return `${path.relative(root, s.fileName)}:${p.line + 1}:${p.character + 1}`; };
const owners = symbol => (symbol && symbol.declarations || []).map(d => d.parent && d.parent.name ? d.parent.name.text : '?');

console.log(`files: ${files.length}`);
// Count 1: writes.
const writes = [];
for (const source of program.getSourceFiles()) {
    if (!source.fileName.startsWith(path.join(root, 'src'))) continue;
    (function visit(node) {
        let target;
        if (ts.isBinaryExpression(node) && node.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && node.operatorToken.kind <= ts.SyntaxKind.LastAssignment) target = node.left;
        if ((ts.isPrefixUnaryExpression(node) || ts.isPostfixUnaryExpression(node)) && [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(node.operator)) target = node.operand;
        if (target && (ts.isPropertyAccessExpression(target) || ts.isElementAccessExpression(target))) {
            const name = ts.isPropertyAccessExpression(target) ? target.name.text : (target.argumentExpression && ts.isStringLiteral(target.argumentExpression) ? target.argumentExpression.text : '[dynamic]');
            if (['directoryExists', 'createDirectory', 'getCanonicalFileName', '[dynamic]'].includes(name)) {
                const symbol = ts.isPropertyAccessExpression(target) ? checker.getSymbolAtLocation(target.name) : undefined;
                if (name !== '[dynamic]' || /host|sys/i.test(target.expression.getText()))
                    writes.push(`${rel(target)} ${target.getText().slice(0, 60)} declared in ${JSON.stringify(owners(symbol))} receiver ${checker.typeToString(checker.getTypeAtLocation(target.expression)).slice(0, 80)}`);
            }
        }
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'defineProperty' || (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === 'defineProperty'))
            writes.push(`${rel(node)} defineProperty ${node.getText().slice(0, 100)}`);
        ts.forEachChild(node, visit);
    })(source);
}
console.log(`== count 1: ${writes.length} writes to directoryExists, createDirectory or getCanonicalFileName (and defineProperty calls)`);
for (const write of writes) console.log('write ' + write);
const cacheCalls = [];
for (const source of program.getSourceFiles()) {
    if (!source.fileName.startsWith(path.join(root, 'src'))) continue;
    (function visit(node) {
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'changeCompilerHostLikeToUseCache')
            cacheCalls.push(`${rel(node)} host ${node.arguments[0].getText()} : ${checker.typeToString(checker.getTypeAtLocation(node.arguments[0])).slice(0, 100)}`);
        ts.forEachChild(node, visit);
    })(source);
}
console.log(`== changeCompilerHostLikeToUseCache calls: ${cacheCalls.length}`);
for (const call of cacheCalls) console.log('cache-call ' + call);

// Count 2: follow a parameter's value.
const flags = [], seen = new Set();
function follow(parameter, label) {
    const key = rel(parameter);
    if (seen.has(key)) return;
    seen.add(key);
    const symbol = checker.getSymbolAtLocation(parameter.name);
    const fn = parameter.parent;
    if (!fn.body) { flags.push(`${label}: ${key} has no body (an overload or declaration)`); return; }
    (function visit(node) {
        if (ts.isIdentifier(node) && node !== parameter.name && checker.getSymbolAtLocation(node) === symbol) {
            let outer = node; while (ts.isParenthesizedExpression(outer.parent)) outer = outer.parent;
            const parent = outer.parent;
            if (ts.isCallExpression(parent) && parent.expression === outer) console.log(`use ${rel(node)} ${label}: called`);
            else if (ts.isCallExpression(parent) || ts.isNewExpression(parent)) {
                const index = parent.arguments.indexOf(outer);
                const signature = checker.getResolvedSignature(parent);
                const declaration = signature && signature.declaration;
                const target = declaration && declaration.parameters && declaration.parameters[Math.min(index, declaration.parameters.length - 1)];
                if (target && ts.isParameter(target)) { console.log(`use ${rel(node)} ${label}: passed to ${declaration.name ? declaration.name.getText() : '(anonymous)'}(${target.name.getText()})`); follow(target, `${label} > ${target.name.getText()}`); }
                else flags.push(`${rel(node)} ${label}: passed to an unresolved callee ${parent.expression.getText().slice(0, 60)}`);
            }
            else if (ts.isShorthandPropertyAssignment(parent) || ts.isPropertyAssignment(parent)) flags.push(`${rel(node)} ${label}: stored in an object field ${parent.getText().slice(0, 60)}`);
            else if (ts.isVariableDeclaration(parent)) flags.push(`${rel(node)} ${label}: stored in ${parent.name.getText()}`);
            else flags.push(`${rel(node)} ${label}: ${ts.SyntaxKind[parent.kind]} ${parent.getText().slice(0, 80)}`);
        }
        ts.forEachChild(node, visit);
    })(fn.body);
}
for (const source of program.getSourceFiles()) {
    (function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name && ['createModuleResolutionCache', 'createTypeReferenceDirectiveResolutionCache'].includes(node.name.text) && node.body) {
            const parameter = node.parameters.find(p => p.name.getText() === 'getCanonicalFileName');
            if (parameter) follow(parameter, node.name.text + '(getCanonicalFileName)');
        }
        ts.forEachChild(node, visit);
    })(source);
}
console.log(`== count 2: ${flags.length} uses that aren't a call or a pass-through`);
for (const flag of flags) console.log('flag ' + flag);
