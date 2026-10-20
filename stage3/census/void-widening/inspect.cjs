'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('../../api/node_modules/typescript');
const {compiler, nodes} = require('../../adapt/47-host-errors/adapt.cjs');
const [treeArg, input, output] = process.argv.slice(2);
const tree = path.resolve(treeArg);
assert.equal(ts.version, '6.0.3');
const program = compiler(tree, ts), checker = program.getTypeChecker();
const sites = JSON.parse(fs.readFileSync(input));
const apiLines = fs.readFileSync(path.join(tree, 'built/local/typescript.d.ts'), 'utf8').split(/\r?\n/);
function publicReachability(d) {
    const sf = d.getSourceFile();
    let top = d;
    while (top.parent && top.parent !== sf) {
        if (ts.isFunctionLike(top.parent) && top.parent.body) return {reachable:false, reason:'Lexically local declaration inside a function body.'};
        top = top.parent;
    }
    if (/@internal\b/.test(top.getFullText(sf).slice(0, top.getStart(sf) - top.getFullStart()))) return {reachable:false, reason:'The enclosing exported declaration is @internal and omitted from the public API.'};
    if (!ts.getCombinedModifierFlags(top) || !(ts.getCombinedModifierFlags(top) & ts.ModifierFlags.Export)) return {reachable:false, reason:'The enclosing source declaration is not exported.'};
    const name = top.name?.getText(sf);
    const line = name ? apiLines.findIndex(s => s.includes('function ' + name + '(') || s.includes('function ' + name + '<')) : -1;
    return line >= 0 ? {reachable:true, file:'built/local/typescript.d.ts', line:line+1, sourceLine:apiLines[line]} : {reachable:null, reason:'Export requires an explicit emitted-declaration graph review.'};
}
function location(n) {
    const sf = n.getSourceFile(), p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
    return {file: path.relative(tree, sf.fileName), line: p.line + 1, column: p.character + 1};
}
for (const s of sites) {
    const sf = program.getSourceFile(path.join(tree, s.file));
    assert(sf, s.file);
    const at = sf.getPositionOfLineAndCharacter(s.line - 1, s.column - 1);
    const candidates = nodes(ts, sf, n => n.getStart(sf) === at && ts.isExpressionNode(n));
    assert(candidates.length, JSON.stringify(s));
    const n = candidates[0];
    s.sourceLine = sf.text.split(/\r?\n/)[s.line - 1];
    s.expression = n.getText(sf);
    s.syntaxKind = ts.SyntaxKind[n.kind];
    s.form = s.reason ? (s.reason.startsWith('a void value') ? 'return-of-void' : s.reason.match(/at the (.*?) site/)[1]) : s.form;
    let signatures;
    if (ts.isCallExpression(n)) {
        s.callee = n.expression.getText(sf);
        signatures = [checker.getResolvedSignature(n)].filter(Boolean);
    } else {
        s.callee = n.getText(sf);
        signatures = checker.getSignaturesOfType(checker.getTypeAtLocation(n), ts.SignatureKind.Call);
    }
    s.declarations = signatures.map(sig => {
        const d = sig.declaration;
        if (!d) return {returnType: checker.typeToString(checker.getReturnTypeOfSignature(sig)), declaration: null};
        const own = d.getSourceFile();
        const returning = d.body && ts.isBlock(d.body) ? nodes(ts, d.body, ts.isReturnStatement).filter(r => {
            let p = r.parent;
            while (p && p !== d && !ts.isFunctionLike(p)) p = p.parent;
            return p === d;
        }) : [];
        return {...location(d), returnType: checker.typeToString(checker.getReturnTypeOfSignature(sig)),
            declaredReturnType: d.type?.getText(own) ?? '<inferred>', text: d.getText(own),
            hasBody: !!d.body, bodyKind: d.body ? ts.SyntaxKind[d.body.kind] : null,
            returns: returning.map(r => ({...location(r), expression: r.expression?.getText(own) ?? null})),
            publicReachability: publicReachability(d)};
    });
    let actual = n;
    while (ts.isAsExpression(actual) || ts.isParenthesizedExpression(actual)) actual = actual.expression;
    if (actual !== n) s.underlyingDeclarations = checker.getSignaturesOfType(checker.getTypeAtLocation(actual), ts.SignatureKind.Call).map(sig => sig.declaration).filter(Boolean).map(d => ({...location(d), text:d.getText(d.getSourceFile()), publicReachability:publicReachability(d)}));
    let parent = n.parent;
    while (ts.isParenthesizedExpression(parent)) parent = parent.parent;
    if (ts.isCallExpression(parent)) {
        const signature = checker.getResolvedSignature(parent), declaration = signature?.declaration;
        s.receiver = {expression:parent.expression.getText(sf), declaredReturnType:signature ? checker.typeToString(checker.getReturnTypeOfSignature(signature)) : null, declaration:declaration ? {...location(declaration), publicReachability:publicReachability(declaration)} : null};
    }
    s.judgment = {kind: 'unclassified', reason: 'Declaration and public reachability review pending; no adaptation is credited.'};
}
const count = key => Object.fromEntries([...new Set(sites.map(key))].sort().map(k => [k, sites.filter(s => key(s) === k).length]));
fs.writeFileSync(output, JSON.stringify({mainSha: '165f427eb629dda1b27553bca3c01e0328edcaa2', date: '2026-10-20',
    typescriptVersion: ts.version, columns: 'one-based UTF-16', countsByForm: count(s => s.form), countsByFile: count(s => s.file),
    countsByJudgment: count(s => s.judgment.kind), sites}, null, 2) + '\n');
console.log(JSON.stringify({sites: sites.length, byForm: count(s => s.form)}));
