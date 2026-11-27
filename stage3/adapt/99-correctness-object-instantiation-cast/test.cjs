'use strict';
const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');
const {adapt} = require('./adapt.cjs');
const [pin] = require('./owners.json');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
const tree = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-correctness-object-instantiation-cast';
const read = side => execFileSync('git', ['-C', tree, 'show', ref + '-' + side + ':' + pin.file], {encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
const before = read('before');
const after = read('after');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-object-instantiation-'));
const checker = path.join(scratch, pin.file);
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
    fs.mkdirSync(path.dirname(checker), {recursive: true});
    row('control writes exact after snapshot and replay does not write', () => {
        fs.writeFileSync(checker, before);
        const result = adapt(scratch);
        assert.equal(result.writes, 1, 'expected one write');
        assert.equal(result.javascriptChanging, false);
        assert(fs.readFileSync(checker, 'utf8') === after, 'unexpected output bytes');
        assert.equal(adapt(scratch).writes, 0, 'expected replay to write nothing');
        assert(fs.readFileSync(checker, 'utf8') === after, 'unexpected replay bytes');
    });

    const drift = /drift or half-adapted owner/;
    const newline = before.includes('\r\n') ? '\r\n' : '\n';
    const cases = [
        ['owner changed away from site', before.replace(pin.before, () => pin.before.replace('function instantiateTypeWorker', 'function  instantiateTypeWorker')), drift],
        ['site expression changed', before.replace(pin.before, () => pin.before.replace('type as TypeReference | AnonymousType | MappedType', 'type as unknown as TypeReference | AnonymousType | MappedType')), drift],
        ['duplicate instantiateTypeWorker', before + newline + pin.before + newline, /missing or duplicate owner/],
        ['bare carriage return', before + '\r', /bare CR/],
        ['mixed line endings', before + (newline === '\n' ? '\r\n' : '\n'), /mixed line endings/],
    ];
    for (const [name, text, message] of cases) {
        row('refuses ' + name, () => {
            assert(text !== before, 'fixture did not change its input');
            fs.writeFileSync(checker, text);
            assert.throws(() => adapt(scratch), message);
            assert(fs.readFileSync(checker, 'utf8') === text, 'refusal wrote checker.ts');
        });
    }

    row('public declaration emit unchanged', () => {
        const emit = text => ts.transpileDeclaration(text, {
            fileName: pin.file, compilerOptions: {stripInternal: true},
        }).outputText;
        const declarations = emit(before);
        assert(declarations.length > 0 && declarations === emit(after), 'public declaration bytes changed');
    });

    row('JavaScript emit byte-identical', () => {
        const emit = text => ts.transpileModule(text, {
            fileName: pin.file,
            compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS},
        }).outputText;
        const javascript = emit(before);
        assert(javascript.length > 0 && javascript === emit(after), 'JavaScript bytes changed');
    });
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}
process.exitCode = Number(failures !== 0);
