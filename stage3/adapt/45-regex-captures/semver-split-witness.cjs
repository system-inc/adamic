'use strict';
const fs = require('node:fs');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const ts = require('typescript');
const ranges = ['>=1.0.0 <2.0.0', '>=1.0.0    <2.0.0', '>=1.0.0\t<2.0.0', '  >=1.0.0 <2.0.0  ', '1.0.0 || >=2.0.0 <3.0.0', '1.0.0 - 2.0.0', ''];
function witness(tree) {
  const source = fs.readFileSync(tree + '/src/compiler/semver.ts', 'utf8');
  const code = ts.transpileModule(source + '\nexport { parseRange };', {compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;
  const exports = {};
  vm.runInNewContext(code, {exports, require(name){assert.equal(name,'./_namespaces/ts.js');return ts;}});
  return ranges.map(range => JSON.stringify({range,parsed:exports.parseRange(range)})).join('\n')+'\n';
}
const before = witness(process.argv[2]);
const after = witness(process.argv[3]);
fs.writeFileSync(process.argv[4] + '/witness-before.jsonl', before);
fs.writeFileSync(process.argv[4] + '/witness-after.jsonl', after);
assert.equal(after, before);
console.log(before);
console.log('PASS: seven adapted parseRange observations are byte-identical');
