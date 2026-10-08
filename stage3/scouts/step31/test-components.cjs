"use strict";
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const ts = require(process.env.STEP31_TYPESCRIPT);
const {observer} = require('./component-dump.cjs');
const root = path.join(__dirname, 'fixtures');
const request = JSON.parse(fs.readFileSync(path.join(root, 'components.json'), 'utf8'));
for (const project of request.projects) for (const file of project.files)
    assert.equal(file.text, fs.readFileSync(path.join(root, file.path), 'utf8'));
const compiler = observer(ts, path.dirname(process.env.STEP31_TYPESCRIPT));
for (const mode of ['checker', 'emitter']) {
    const output = request.projects.flatMap(project => compiler.observe(project, mode));
    const bytes = output.map(row => JSON.stringify(row)).join('\n') + '\n';
    assert.equal(bytes, fs.readFileSync(path.join(root, mode + '-golden.jsonl'), 'utf8'));
    const fresh = observer(ts, path.dirname(process.env.STEP31_TYPESCRIPT), false);
    for (const project of [...request.projects].reverse()) {
        const reversed = {...project, files: [...project.files].reverse()};
        assert.deepEqual(compiler.observe(reversed, mode), fresh.observe(project, mode));
    }
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'step31-official-host-'));
    for (const project of request.projects) {
        const names = project.files.map(file => {
            const name = path.join(directory, file.path.replace(/\.a$/, '.ts'));
            fs.writeFileSync(name, file.text);
            return name;
        });
        const options = ts.convertCompilerOptionsFromJson(project.options, directory).options;
        const program = ts.createProgram(names, options, ts.createCompilerHost(options));
        const rows = compiler.observe(project, mode);
        if (mode === 'checker') {
            const expected = ts.getPreEmitDiagnostics(program);
            const actual = rows.flatMap(row => row.diagnostics || []);
            assert.deepEqual(actual.map(value => [value.code, value.start, value.length]),
                expected.map(value => [value.code, value.start, value.length]));
            assert.deepEqual(actual.map(value => typeof value.message === 'string' ? value.message : value.message.message),
                expected.map(value => typeof value.messageText === 'string' ? value.messageText : value.messageText.messageText));
        } else {
            const expected = [];
            const result = program.emit(undefined, (name, text, bom) => expected.push({path: path.relative(directory, name), text, bom}));
            expected.sort((a, b) => a.path < b.path ? -1 : a.path > b.path ? 1 : 0);
            assert.deepEqual(rows.filter(row => row.record === 'output').map(({path, text, bom}) => ({path, text, bom})), expected);
            assert.equal(rows[1].emitSkipped, result.emitSkipped);
        }
    }
}
const checkerRows = compiler.observe(request.projects[0], 'checker');
const diagnostics = checkerRows.flatMap(row => row.diagnostics || []);
assert.deepEqual(diagnostics.map(value => [value.code, value.start, value.length]), [[2322, 122, 6], [2345, 157, 7]]);
assert.equal(diagnostics[0].message.next[0].next[0].message, "Type 'string' is not assignable to type 'number'.");
assert.deepEqual(checkerRows.find(row => row.path === 'checker-values.a').diagnostics, []);
const emitted = compiler.observe(request.projects[1], 'emitter').filter(row => row.record === 'output');
assert.deepEqual(emitted.map(row => row.path), ['emitter-text.d.ts', 'emitter-text.d.ts.map', 'emitter-text.js', 'emitter-text.js.map']);
assert.ok(emitted.find(row => row.path === 'emitter-text.js').text.includes('// retained comment\n'));
assert.ok(emitted.find(row => row.path === 'emitter-text.js').text.includes('exports.value = 41 + 1;'));
const relatedProject = structuredClone(request.projects[0]);
relatedProject.files[0].text += 'needNumber();\n';
const relatedDiagnostic = compiler.observe(relatedProject, 'checker').flatMap(row => row.diagnostics || []).find(value => value.code === 2554);
assert.equal(relatedDiagnostic.related[0].file, 'checker-values.a');
assert.ok(relatedDiagnostic.related[0].start !== null && relatedDiagnostic.related[0].length > 0);
const unicodeProject = {id: 'unicode-jsonl', options: {target: 'es2020'}, files: [{path: 'separator.a', text: '// separator: \u2028\nexport const text = "\u2028";'}]};
const unicodeWire = compiler.observe(unicodeProject, 'emitter').map(row => JSON.stringify(row)).join('\n') + '\n';
assert.ok(unicodeWire.includes('\u2028'));
assert.equal(unicodeWire.trimEnd().split('\n').map(JSON.parse).filter(row => row.record === 'output').length, 1);
const malformed = structuredClone(request.projects[0]);
malformed.files[0].path = '../escape.ts';
assert.throws(() => compiler.observe(malformed, 'checker'), /relative virtual/);
console.log('checker chains, exact spans, clean-file records, imports, emitter JS/declarations/maps, fresh libraries and official compiler host pass');
