'use strict';
const fs = require('node:fs'), path = require('node:path'), ts = require('typescript');
const rules = [
    ['src/compiler/debug.ts', 200, '        if ((Error as any).captureStackTrace) {'],
    ['src/compiler/debug.ts', 201, '            (Error as any).captureStackTrace(e, stackCrawlMark || fail);'],
    ['src/compiler/sys.ts', 75, '    if ((Error as any).stackTraceLimit < 100) { // Also tests that we won\'t set the property if it doesn\'t exist.'],
    ['src/compiler/sys.ts', 76, '        (Error as any).stackTraceLimit = 100;'],
];
function plan(text, file) {
    const selected = rules.filter(r => r[0] === file);
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const lines = text.split('\n'), edits = [];
    for (const [, pinnedLine, original] of selected) {
        const adapted = original.replace('(Error as any)', 'Error');
        const matches = lines.map((s, i) => [s.replace(/\r$/, ''), i + 1]).filter(([s]) => s === original || s === adapted);
        if (matches.length !== 1) throw Error('missing or duplicate Error site: ' + file + ':' + pinnedLine);
        const [current, line] = matches[0];
        let found = 0;
        function visit(n) {
            if (ts.isPropertyAccessExpression(n) && source.getLineAndCharacterOfPosition(n.getStart(source)).line + 1 === line) {
                const receiver = n.expression;
                if (receiver.getText(source) === (current === original ? '(Error as any)' : 'Error')) {
                    if (current === original && !(ts.isParenthesizedExpression(receiver) && ts.isAsExpression(receiver.expression) && receiver.expression.type.kind === ts.SyntaxKind.AnyKeyword)) throw Error('Error assertion AST drift');
                    let owner = n.parent;
                    while (owner && !ts.isFunctionDeclaration(owner)) owner = owner.parent;
                    if (owner?.name?.text !== (file.endsWith('debug.ts') ? 'fail' : 'setStackTraceLimit')) throw Error('unreviewed Error owner');
                    found++;
                    if (current === original) edits.push([receiver.getStart(source), receiver.end]);
                }
            }
            ts.forEachChild(n, visit);
        }
        visit(source);
        if (found !== 1) throw Error('missing or duplicate Error receiver');
    }
    for (const [start, end] of edits.sort((a,b) => b[0]-a[0])) text = text.slice(0,start) + 'Error' + text.slice(end);
    return {text, removed: edits.length};
}
function apply(tree) {
    if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
    const plans = [...new Set(rules.map(r => r[0]))].map(file => {
        const name = path.join(tree, file), before = fs.readFileSync(name, 'utf8');
        return {name, before, ...plan(before, file)};
    });
    for (const p of plans) if (fs.readFileSync(p.name, 'utf8') !== p.before) throw Error('concurrent Error source change');
    for (const p of plans) if (p.text !== p.before) fs.writeFileSync(p.name, p.text);
    console.log(JSON.stringify({class: 'standard runtime feature probes: Error', removed: plans.reduce((n,p) => n+p.removed,0)}));
}
module.exports = {plan, apply, rules};
