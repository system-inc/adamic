'use strict';
const fs = require('node:fs');
const vm = require('node:vm');
const ts = require('typescript');
const assert = require('node:assert/strict');
const file = process.argv[2];
const text = fs.readFileSync(file, 'utf8');
const start = text.indexOf('if (sys && sys.getEnvironmentVariable) {');
assert(start >= 0 && start === text.lastIndexOf('if (sys && sys.getEnvironmentVariable) {'));
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const owner = source.statements.find(n => n.getStart(source) === start);
assert(ts.isIfStatement(owner));
const code = ts.transpileModule('(globalThis as any).__envPresence1965(sys);\n' + owner.getText(source), {compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
for (const sys of [undefined, {}, {getEnvironmentVariable(){return '';}}]) {
    vm.runInNewContext(code, {sys, setCustomPollingValues(){}, Debug:{setAssertionLevel(){}}, AssertionLevel:{Normal:1,None:0}});
}
const rows = fs.readFileSync(process.env.ENV_PRESENCE_COUNTS, 'utf8').trim().split('\n').map(JSON.parse);
assert.equal(rows.length, 3);
assert.equal(rows.filter(r=>!r.truthy).length,1);
assert.equal(rows.filter(r=>r.truthy&&!r.function).length,1);
assert.equal(rows.filter(r=>r.truthy&&r.function).length,1);
console.log('PASS: actual original condition with the counting insertion observes undefined, missing-method and valid-method controls');
