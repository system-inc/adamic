import fs from 'node:fs';
import vm from 'node:vm';
const root = process.argv[2];
if (!root) { throw new Error('provide the pinned test262 checkout'); }
const harness = fs.readFileSync(root + '/harness/sta.js', 'utf8') + '\n' + fs.readFileSync(root + '/harness/assert.js', 'utf8');
for (const name of ['array-contract-expand.js', 'array-expand-contract.js', 'array-expand.js', 'head-const-fresh-binding-per-iteration.js']) {
 const path = root + '/test/language/statements/for-of/' + name;
 for (const strict of [false, true]) {
  vm.runInNewContext((strict ? '"use strict";\n' : '') + harness + '\n' + fs.readFileSync(path, 'utf8'), {}, { filename: path });
  console.log(name + ': ' + (strict ? 'strict' : 'default') + ' passed');
 }
}
