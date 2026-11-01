'use strict';
// Reproduce the real protocol value that prevents the checked client from satisfying LanguageService.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const rules = require('./rules.json');
const [before, after, responseFile] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
const response = JSON.parse(fs.readFileSync(responseFile, 'utf8'));
const source = ts.createSourceFile('client.ts', fs.readFileSync(path.join(after, 'src/harness/client.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
const helpers = [];
(function visit(node) {
    if (ts.isFunctionDeclaration(node) && /^adamic(ArrayOf|Is)/.test(node.name?.text || '')) helpers.push(node.getText(source));
    ts.forEachChild(node, visit);
})(source);
const compiler = ts.createSourceFile('commandLineParser.ts', fs.readFileSync(path.join(after, 'src/compiler/commandLineParser.ts'), 'utf8'), ts.ScriptTarget.Latest, true);
const predicate = compiler.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'isAdamicJsonObject');
assert(predicate);
const emit = text => ts.transpileModule(text, {compilerOptions: {target: ts.ScriptTarget.ES2020}}).outputText;
const isAdamicJsonObject = Function(emit(predicate.getText(compiler).replace(/^export /, '')) + '; return isAdamicJsonObject;')();
const checks = Function('ts', emit(helpers.join('\n')) + '; return {adamicArrayOf, adamicIsCompletionEntryDetails};')({isAdamicJsonObject});
const rule = rules.find(row => row.owner === 'getCompletionEntryDetails' && row.file === 'src/harness/client.ts');
assert(rule);
const originalSource = fs.readFileSync(path.join(before, rule.file), 'utf8');
assert(originalSource.includes(rule.before));
const run = code => Function('response', 'Debug', ...Object.keys(checks), code)(response, {assert: condition => assert(condition)}, ...Object.values(checks));
const original = run(rule.before);
assert.equal(original.tags[0].text, 'p1');
assert.throws(() => run(rule.after), {name: 'TypeError', message: 'Expected completion details array'});
console.log(JSON.stringify({before: 'returns original protocol value', tagText: original.tags[0].text,
    tagTextType: typeof original.tags[0].text, after: 'throws TypeError',
    declaredServiceShape: 'JSDocTagInfo.text?: SymbolDisplayPart[]',
    actualProtocolShape: 'JSDocTagInfo.text?: string | SymbolDisplayPart[]',
    status: 'caller counterexample; adaptation is not ready'}, null, 2));
