// Review probes for TypeScript 6.0.3. No source adaptation is performed.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const { execFileSync } = require('node:child_process');
const root = process.argv[2];
assert(root, 'usage: node probe.cjs <TypeScript v6.0.3 checkout>');
assert.equal(execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim(),
    '050880ce59e30b356b686bd3144efe24f875ebc8');
const parser = fs.readFileSync(path.join(root, 'src/compiler/parser.ts'), 'utf8');
const initializer = parser.split(/\r?\n/).find(line => line.includes('const result = new RegExp(`'));
assert(initializer);
const matcher = vm.runInNewContext(initializer.trim().replace('const result = ', '').replace(/;$/, ''), { name: 'path' });
const expression = parser.split(/\r?\n/).find(line => line.includes('const value = matchResult[2] || matchResult[3]'));
assert(expression);
for (const [input, expected] of [[" path='x'", 'x'], [' path=""', ''], [" path=''", undefined]]) {
    const matchResult = matcher.exec(input);
    assert(matchResult);
    const value = vm.runInNewContext(expression.trim().replace('const value = ', '').replace(/;$/, ''), { matchResult });
    assert.equal(value, expected);
    console.log(`${JSON.stringify(input)}: group2=${JSON.stringify(matchResult[2])}, group3=${JSON.stringify(matchResult[3])}, value=${JSON.stringify(value)}`);
    if (expected === undefined) {
        assert.throws(() => value.length, TypeError);
        console.log('single-quoted empty capture: value.length throws TypeError');
    }
}
// Mutant of debug.ts's mandatory name capture: a successful alternative bypasses it.
const mutant = /^(?:function\s+([\w$]+)\s*\(|anonymous)$/;
const match = mutant.exec('anonymous');
assert(match);
assert.equal(match[1], undefined);
console.log('alternative mutant: successful match has group1=undefined; mandatory-capture rule declines');
// Prove the counterexample assertion can fail when its expected observation is corrupted.
assert.throws(() => assert.equal(match[1], 'anonymous'), assert.AssertionError);
console.log('counterexample observation mutant: wrong expected group value caught by assertion');
const semver = fs.readFileSync(path.join(root, 'src/compiler/semver.ts'), 'utf8');
const rangeDeclaration = semver.split(/\r?\n/).find(line => line.startsWith('const rangeRegExp = '));
assert(rangeDeclaration);
const range = vm.runInNewContext(rangeDeclaration.replace('const rangeRegExp = ', '').replace(/;$/, ''));
assert.equal(range.exec('1.2.3')[1], undefined);
assert(semver.includes('case undefined:'));
console.log('semver plain comparator: operator absent; upstream has case undefined');
console.log('PASS: review probes only; no census, oracle, or idempotence claim');
