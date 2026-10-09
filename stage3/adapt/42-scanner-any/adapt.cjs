'use strict';
// Concrete private scanner storage and diagnostic payloads; public API stays intact.
const fs = require('node:fs'), path = require('node:path'), ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
function plan(text) {
    const source = ts.createSourceFile('scanner.ts', text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw Error('scanner syntax changed');
    const edits = [], found = {map: 0, overload: 0, implementation: 0};
    function visit(n) {
        if (ts.isVariableDeclaration(n) && n.name.getText(source) === 'textToKeyword') {
            found.map++;
            if (n.type || !ts.isNewExpression(n.initializer) || n.initializer.expression.getText(source) !== 'Map' ||
                n.initializer.arguments?.length !== 1 || n.initializer.arguments[0].getText(source) !== 'Object.entries(textToKeywordObj)')
                throw Error('unreviewed keyword map initializer');
            const args = n.initializer.typeArguments;
            if (args) {
                if (args.length !== 2 || args[0].getText(source) !== 'string' || args[1].getText(source) !== 'KeywordSyntaxKind')
                    throw Error('unreviewed keyword map types');
            }
            else edits.push({start: n.initializer.expression.end, end: n.initializer.expression.end, value: '<string, KeywordSyntaxKind>'});
        }
        if (ts.isFunctionDeclaration(n) && n.name?.text === 'error' && n.parameters.length === 4) {
            if (!ts.isFunctionDeclaration(n.parent) && !ts.isBlock(n.parent)) throw Error('error owner changed');
            const owner = n.parent.parent;
            if (!ts.isFunctionDeclaration(owner) || owner.name?.text !== 'createScanner') throw Error('error moved out of createScanner');
            const key = n.body ? 'implementation' : 'overload'; found[key]++;
            const parameter = n.parameters[3];
            if (n.parameters.map(p => p.name.getText(source)).join(',') !== 'message,errPos,length,arg0' ||
                !parameter.questionToken || !parameter.type || n.type?.getText(source) !== 'void') throw Error('error signature changed');
            if (n.body && n.body.getText(source).replace(/\r\n/g, '\n') !== '{\n        if (onError) {\n            const oldPos = pos;\n            pos = errPos;\n            onError(message, length || 0, arg0);\n            pos = oldPos;\n        }\n    }') throw Error('error body changed');
            if (parameter.type.kind === ts.SyntaxKind.AnyKeyword)
                edits.push({start: parameter.type.getStart(source), end: parameter.type.end, value: 'string | number'});
            else if (parameter.type.getText(source) !== 'string | number') throw Error('unreviewed error payload');
        }
        ts.forEachChild(n, visit);
    }
    visit(source);
    if (Object.values(found).some(n => n !== 1)) throw Error('missing or duplicate scanner owners: ' + JSON.stringify(found));
    let result = text;
    for (const e of edits.sort((a,b) => b.start-a.start)) result = result.slice(0,e.start)+e.value+result.slice(e.end);
    return {text: result, edits: edits.length};
}
function apply(tree) {
    const file = path.join(tree, 'src/compiler/scanner.ts'), before = fs.readFileSync(file, 'utf8');
    const result = plan(before);
    if (fs.readFileSync(file, 'utf8') !== before) throw Error('concurrent scanner source edit');
    if (result.text !== before) fs.writeFileSync(file, result.text);
    console.log(JSON.stringify({adaptation: 42, edits: result.edits}));
}
module.exports = {plan, apply};
if (require.main === module) {
    if (process.argv.length !== 3) throw Error('usage: node adapt.cjs <tree>');
    apply(path.resolve(process.argv[2]));
}
