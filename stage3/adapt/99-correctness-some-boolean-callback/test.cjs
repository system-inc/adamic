'use strict';
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');
const {adapt, callback} = require('./adapt.cjs');
const [pin] = require('./owners.json');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
const tree = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-correctness-some-boolean-callback';
const read = side => execFileSync('git', ['-C', tree, 'show', ref + '-' + side + ':' + pin.file], {encoding: 'utf8'});
const before = read('before');
const after = read('after');
const originalExpression = 'namedCapturingGroupsScopeStack.some(group => group?.has(tokenValue))';
const adaptedExpression = 'namedCapturingGroupsScopeStack.some(group => group !== undefined && group.has(tokenValue))';
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-some-'));
const scanner = path.join(scratch, pin.file);
let failures = 0;

function row(name, check) {
    try {
        const detail = check();
        console.log('PASS ' + name + (detail === undefined ? '' : ' ' + detail));
    } catch (error) {
        failures++;
        console.error('FAIL ' + name + ': ' + error.message);
    }
}

try {
    fs.mkdirSync(path.dirname(scanner), {recursive: true});
    row('control writes exact after snapshot and replay does not write', () => {
        fs.writeFileSync(scanner, before);
        assert.equal(adapt(scratch).writes, 1, 'expected one write');
        assert(fs.readFileSync(scanner, 'utf8') === after, 'unexpected output bytes');
        assert.equal(adapt(scratch).writes, 0, 'expected replay to write nothing');
        assert(fs.readFileSync(scanner, 'utf8') === after, 'unexpected replay bytes');
    });

    const drift = /drift or half-adapted owner/;
    const newline = before.includes('\r\n') ? '\r\n' : '\n';
    const duplicate = pin.before.replace(originalExpression, `(${originalExpression}, ${originalExpression})`);
    const cases = [
        ['owner changed away from site', before.replace(pin.before, () => pin.before.replace('export function', 'export  function')), drift],
        ['site expression changed', before.replace('group?.has(tokenValue)', 'group?.has(tokenValue) === true'), drift],
        ['expression twice', before.replace(pin.before, () => duplicate), drift],
        ['duplicate createScanner', before + newline + pin.before + newline, /missing or duplicate owner/],
        ['bare carriage return', before + '\r', /bare CR/],
        ['mixed line endings', before + (newline === '\n' ? '\r\n' : '\n'), /mixed line endings/],
    ];
    for (const [name, text, message] of cases) {
        row('refuses ' + name, () => {
            assert(text !== before, 'fixture did not change its input');
            fs.writeFileSync(scanner, text);
            assert.throws(() => adapt(scratch), message);
            assert(fs.readFileSync(scanner, 'utf8') === text, 'refusal wrote scanner.ts');
        });
    }

    row('public declaration emit unchanged', () => {
        const emit = text => ts.transpileDeclaration(text, {
            fileName: pin.file, compilerOptions: {stripInternal: true},
        }).outputText;
        const declarations = emit(before);
        assert(declarations.length > 0 && declarations === emit(after), 'public declaration bytes changed');
    });

    const stack = [undefined, new Set(['a']), undefined];
    const scenarios = [
        ['mixed stack with a', stack, 'a', true],
        ['mixed stack with b', stack, 'b', false],
        ['empty stack', [], 'a', false],
        ['undefined-only stack', [undefined], 'a', false],
    ];
    for (const [name, groups, tokenValue, expected] of scenarios) {
        row('witness ' + name, () => {
            const evaluate = (text, expression) => {
                const arrow = callback(text, expression);
                return groups.some(new Function('tokenValue', 'return (' + arrow + ')')(tokenValue));
            };
            const oldAnswer = evaluate(before, originalExpression);
            const newAnswer = evaluate(after, adaptedExpression);
            assert.equal(newAnswer, oldAnswer);
            assert.equal(oldAnswer, expected);
            return `before=${oldAnswer} after=${newAnswer}`;
        });
    }
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}
process.exitCode = Number(failures !== 0);
