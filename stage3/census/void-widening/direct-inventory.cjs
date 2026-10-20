'use strict';
// Independent stock-checker inventory of direct callable relations and void returns.
const fs = require('node:fs'), path = require('node:path');
const ts = require('../../api/node_modules/typescript');
const {compiler} = require('../../adapt/47-host-errors/adapt.cjs');
const tree = path.resolve(process.argv[2]), program = compiler(tree, ts), checker = program.getTypeChecker();
const rows = [];
function signatures(type) { return type ? checker.getSignaturesOfType(type, ts.SignatureKind.Call) : []; }
function wider(type) {
    return signatures(type).some(s => checker.getReturnTypeOfSignature(s).flags !== ts.TypeFlags.Void) ||
        !!type?.isUnion() && type.types.some(wider);
}
for (const sf of program.getSourceFiles().filter(s => s.fileName.startsWith(tree + '/src/compiler/') && !s.isDeclarationFile)) {
    function visit(n) {
        if (ts.isTypeNode(n)) return;
        if (ts.isExpressionNode(n) && !ts.isParenthesizedExpression(n)) {
            const source = checker.getTypeAtLocation(n);
            let site = n;
            while (ts.isParenthesizedExpression(site.parent)) site = site.parent;
            let target = checker.getContextualType(n), form;
            const p = site.parent;
            if (ts.isReturnStatement(p) || ts.isArrowFunction(p) && p.body === site) {
                let fn = p;
                while (fn && !ts.isFunctionLike(fn)) fn = fn.parent;
                if (fn) target = checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(fn));
                if (source.flags === ts.TypeFlags.Void && target && target.flags !== ts.TypeFlags.Void) form = 'return-of-void';
            }
            const calls = signatures(source);
            if (!form && calls.length && calls.every(s => checker.getReturnTypeOfSignature(s).flags === ts.TypeFlags.Void) && wider(target)) form = 'direct relation';
            if (form) {
                const loc = sf.getLineAndCharacterOfPosition(n.getStart(sf));
                rows.push({file: path.relative(tree, sf.fileName), line: loc.line + 1, column: loc.character + 1,
                    form, expression: n.getText(sf), sourceType: checker.typeToString(source), targetType: checker.typeToString(target)});
            }
        }
        ts.forEachChild(n, visit);
    }
    visit(sf);
}
fs.writeFileSync(process.argv[3], JSON.stringify(rows, null, 2) + '\n');
console.log(JSON.stringify({sites: rows.length}));
