#!/usr/bin/env node
'use strict';
// #rv75m9k: every method read as a value in a tree's src/compiler, found with the stock checker, each with the shape of
// its use, where its method is declared, and whether that declaration is @internal. It prints a summary, the road each
// site takes (README's A, B or C), every road-A member's declaration text, and the program's own diagnostic count.
// Usage: NODE_PATH=<stage3/api/node_modules> node scan.cjs <tree> <names.json> <out.json>
const ts = require('typescript');
const path = require('node:path');
const fs = require('node:fs');

const [tree, namesPath, out] = process.argv.slice(2);
const names = new Set(JSON.parse(fs.readFileSync(namesPath, 'utf8')));
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.getParsedCommandLineOfConfigFile(configPath, {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: (diagnostic) => { throw new Error(ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')); } });
config.options.typeRoots = [path.join(process.env.NODE_PATH.split(':')[0], '@types')];
config.options.noEmit = true;
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
const diagnostics = ts.getPreEmitDiagnostics(program);

function unparenthesized(node) {
    while (node.parent && ts.isParenthesizedExpression(node.parent)) node = node.parent;
    return node;
}
function isCalled(node) {
    const outer = unparenthesized(node);
    const parent = outer.parent;
    if (ts.isCallExpression(parent) && parent.expression === outer) return true;
    if (ts.isNonNullExpression(parent)) {
        const asserted = unparenthesized(parent);
        if (ts.isCallExpression(asserted.parent) && asserted.parent.expression === asserted) return true;
    }
    return ts.isTaggedTemplateExpression(parent) && parent.tag === outer;
}
function leadingText(node) {
    return node.getFullText().slice(0, node.getStart() - node.getFullStart());
}
function isInternal(declaration) {
    if (/@internal/.test(leadingText(declaration))) return true;
    const owner = declaration.parent;
    return Boolean(owner && (ts.isInterfaceDeclaration(owner) || ts.isClassDeclaration(owner)) && /@internal/.test(leadingText(owner)));
}
function useOf(node) {
    const outer = unparenthesized(node);
    const parent = outer.parent;
    if (ts.isPropertyAccessExpression(parent) && parent.expression === outer)
        return ['bind', 'call', 'apply'].includes(parent.name.text) ? 'bind/call/apply' : 'member of it read';
    if (ts.isCallExpression(parent) || ts.isNewExpression(parent)) return 'argument';
    if (ts.isVariableDeclaration(parent)) return 'variable initializer';
    if (ts.isPropertyAssignment(parent) || ts.isShorthandPropertyAssignment(parent)) return 'object literal field';
    if (ts.isBinaryExpression(parent)) {
        const operator = parent.operatorToken.kind;
        if (operator === ts.SyntaxKind.EqualsToken) return parent.left === outer ? 'write' : 'assigned';
        if (operator === ts.SyntaxKind.AmpersandAmpersandToken && parent.left === outer) return 'presence';
        return 'logical operand (' + ts.tokenToString(operator) + ')';
    }
    if (ts.isConditionalExpression(parent)) return parent.condition === outer ? 'presence' : 'conditional branch';
    if (ts.isIfStatement(parent) || ts.isPrefixUnaryExpression(parent)) return 'presence';
    if (ts.isReturnStatement(parent) || ts.isArrowFunction(parent)) return 'returned';
    return ts.SyntaxKind[parent.kind];
}

const sites = [], writes = new Set();
for (const sourceFile of program.getSourceFiles()) {
    if (!path.resolve(sourceFile.fileName).startsWith(path.resolve(tree, 'src/compiler') + path.sep)) continue;
    const visit = (node) => {
        if (ts.isPropertyAccessExpression(node) && names.has(node.name.text) && !isCalled(node)) {
            let symbol = checker.getSymbolAtLocation(node.name);
            if (symbol && symbol.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
            const use = useOf(node);
            // A member the tree assigns can't become a readonly field, so road A leaves it a method (owners.cjs).
            if (symbol && symbol.flags & ts.SymbolFlags.Method && use === 'write')
                for (const declaration of symbol.declarations || []) if (declaration.parent.name) writes.add(`${declaration.parent.name.text}.${node.name.text}`);
            if (symbol && symbol.flags & ts.SymbolFlags.Method && !['write', 'presence', 'bind/call/apply'].includes(use)) {
                const declarations = symbol.declarations || [];
                const where = sourceFile.getLineAndCharacterOfPosition(node.getStart());
                sites.push({
                    site: `${path.relative(tree, sourceFile.fileName)}:${where.line + 1}:${where.character + 1}`,
                    member: `${declarations[0].parent.name ? declarations[0].parent.name.text : '?'}.${node.name.text}`,
                    kinds: declarations.map((declaration) => ts.SyntaxKind[declaration.kind]),
                    internal: declarations.every(isInternal),
                    declarations: declarations.map((declaration) => {
                        const at = declaration.getSourceFile().getLineAndCharacterOfPosition(declaration.getStart());
                        return { where: `${path.relative(tree, declaration.getSourceFile().fileName)}:${at.line + 1}`, text: declaration.getText().replace(/\s+/g, ' ') };
                    }),
                    use,
                    text: unparenthesized(node).parent.getText().slice(0, 140).replace(/\s+/g, ' '),
                });
            }
        }
        ts.forEachChild(node, visit);
    };
    visit(sourceFile);
}
fs.writeFileSync(out, JSON.stringify(sites, null, 1));
fs.writeFileSync(out.replace(/\.json$/, '.writes.json'), JSON.stringify([...writes].sort(), null, 1));
console.log('written members: ' + [...writes].sort().join(' '));
const count = (key) => Object.entries(sites.reduce((counts, site) => ({ ...counts, [key(site)]: (counts[key(site)] || 0) + 1 }), {})).sort((a, b) => b[1] - a[1]);
console.log(`scan: ${config.fileNames.length} roots, ${diagnostics.length} stock diagnostics, ${sites.length} method value reads under ${names.size} names`);
for (const diagnostic of diagnostics.slice(0, 10)) console.log('diagnostic: ' + ts.formatDiagnostic(diagnostic, { getCanonicalFileName: (f) => f, getCurrentDirectory: () => tree, getNewLine: () => '\n' }).trim());
console.log('by road: ' + JSON.stringify(count((site) => (site.kinds.every((kind) => kind === 'MethodSignature') && site.internal ? 'A' : 'B or C'))));
console.log('by use: ' + JSON.stringify(count((site) => site.use)));
const members = new Map();
for (const site of sites) if (site.internal && site.kinds.every((kind) => kind === 'MethodSignature')) members.set(site.member, site.declarations);
for (const [member, declarations] of [...members].sort())
    for (const declaration of declarations) console.log(`A ${member} ${declaration.where} ${declaration.text}`);
