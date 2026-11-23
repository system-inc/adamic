'use strict';

const assert = require('node:assert/strict');
const { execFileSync, spawnSync } = require('node:child_process');
const { createHash } = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');
const ts = require('typescript');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
assert.equal(ts.version, '6.0.3');
const tree = path.resolve(process.argv[2]);
const before = 'refs/stage3/99-seats-build22-89-writer-lets-before';
const after = 'refs/stage3/99-seats-build22-89-writer-lets-after';
const utilities = 'src/compiler/utilities.ts';
const types = 'src/compiler/types.ts';
const readSnapshot = (ref, file) => execFileSync('git', ['-C', tree, 'show', ref + ':' + file],
    { encoding: 'utf8', maxBuffer: 4 * 1024 * 1024 });
const original = readSnapshot(before, utilities);
const originalTypes = readSnapshot(before, types);
const adapted = readSnapshot(after, utilities);
const eol = original.includes('\r\n') ? '\r\n' : '\n';
const digest = bytes => createHash('sha256').update(bytes).digest('hex');

function replaceOnce(text, old, replacement) {
    const start = text.indexOf(old);
    assert(start >= 0 && start === text.lastIndexOf(old), 'expected one match for ' + old);
    return text.slice(0, start) + replacement + text.slice(start + old.length);
}

function scratch(t) {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), 'writer-lets-'));
    t.after(() => fs.rmSync(root, { recursive: true, force: true }));
    fs.mkdirSync(path.dirname(path.join(root, utilities)), { recursive: true });
    fs.writeFileSync(path.join(root, types), originalTypes);
    return root;
}

function runAdapter(root) {
    const result = spawnSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), root],
        { encoding: 'utf8' });
    if (result.error) throw result.error;
    return result;
}

function fixtureCase(t, input, exit, reason) {
    const root = scratch(t);
    const file = path.join(root, utilities);
    fs.writeFileSync(file, input);
    const initialHash = digest(fs.readFileSync(file));
    const result = runAdapter(root);
    const output = fs.readFileSync(file);
    const unchanged = digest(output) === initialHash;
    if (exit === 1) {
        assert.deepEqual({ exit: result.status, unchanged }, { exit, unchanged: true },
            `adapter exit ${result.status}; ${unchanged ? 'unchanged' : 'wrote'}\n${result.stderr}`);
        assert.equal(output.toString('utf8'), input);
        if (reason) assert.match(result.stderr, reason);
        t.diagnostic('exit 1, bytes unchanged');
    } else {
        assert.equal(result.status, 0, result.stderr);
    }
    return { root, file, output: output.toString('utf8') };
}

const single = '    var str = "";';
const adaptedSingle = '    let str = "";';
const duplicate = '    var str: string;';
const reset = '    function reset(): void {';
const helper = '    function updateLineCountAndPosFor(s:';
const lastLocal = '    var hasTrailingComment = false;';
const resetCall = ['    reset();', '', '    return {', '        write,'].join(eol);

const cases = [
    ['duplicate var on original input', single, single + eol + duplicate,
        /utilities\.ts:\d+: conflicting binding of str/],
    ['duplicate var on adapted input', adaptedSingle, adaptedSingle + eol + duplicate,
        /utilities\.ts:\d+: conflicting binding of str/,
        () => adapted],
    ['helper named indent', reset, '    function indent(): void {}' + eol + reset,
        /utilities\.ts:\d+: conflicting binding of indent/],
    ['parameter named indent', 'createTextWriter(newLine: string)',
        'createTextWriter(newLine: string, indent: number)',
        /utilities\.ts:\d+: conflicting binding of indent/],
    ['var binding hoisted from a nested block', single,
        single + eol + '    if (true) { var { value: str } = { value: "" }; }',
        /utilities\.ts:\d+: conflicting binding of str/],
    ['direct lexical binding', single, single + eol + '    const str = "";',
        /utilities\.ts:\d+: conflicting binding of str/],
    ['changed reset value', '        lineCount = 0;', '        lineCount = 1;'],
    ['seventh local', lastLocal, lastLocal + eol + '    var seventh = false;'],
    ['early statement', helper, '    output = "early";' + eol + helper],
    ['moved reset', resetCall, resetCall.replace(eol, eol + '    output = "late";' + eol)],
    ['half applied', single, adaptedSingle],
];

for (const [label, old, replacement, reason, input = () => original] of cases) {
    test(label, t => fixtureCase(t, replaceOnce(input(), old, replacement), 1, reason));
}

test('nested functions and blocks can shadow str', t => {
    const shadows = [
        '    function ownParameter(str: string): void {}',
        '    function ownVar(): void { var str = ""; }',
        '    { let str = ""; }',
    ].join(eol);
    const { output } = fixtureCase(t, replaceOnce(original, single, single + eol + shadows), 0);
    assert(output.includes(adaptedSingle + eol + shadows));
});

function extract(text, file, names, predicate) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    return names.map(name => {
        const matches = source.statements.filter(node => predicate(node) && node.name?.text === name);
        assert.equal(matches.length, 1, name);
        return matches[0].getText(source).replaceAll('\r\n', '\n');
    }).join('\n\n');
}

const dependencies = `
interface WriterSymbol { name: string; }
function noop(): void {}
function getIndentSize(): number { return 4; }
function getIndentString(indent: number): string {
    let result = "";
    for (let i = 0; i < indent; i++) result += "    ";
    return result;
}
function computeLineStarts(text: string): number[] {
    const starts = [0];
    for (let i = 0; i < text.length; i++) {
        if (text.charCodeAt(i) === 10) starts.push(i + 1);
    }
    return starts;
}
function last(values: number[]): number { return values[values.length - 1]!; }
function isWhiteSpaceLike(value: number): boolean { return value === 32 || value === 9; }
`;

const fixture = `
function state(label: string, writer: EmitTextWriter): void {
    console.log(label + " text=<" + writer.getText() + "> pos=" + writer.getTextPos() +
        " line=" + writer.getLine() + " col=" + writer.getColumn() + " indent=" + writer.getIndent() +
        " start=" + writer.isAtStartOfLine() + " comment=" + writer.hasTrailingComment() +
        " whitespace=" + writer.hasTrailingWhitespace());
}
function piece(n: number): string { return "part" + String(n); }
const one = createTextWriter("\\n");
const two = createTextWriter("\\n");
state("one initial", one);
one.increaseIndent();
one.write(piece(1));
one.writeLine();
one.writeComment(piece(2));
two.rawWrite(piece(3) + "\\n" + piece(4) + " ");
state("one before", one);
state("two before", two);
one.clear();
state("one clear 1", one);
state("two still 1", two);
one.write(piece(5));
one.writeLine();
one.increaseIndent();
one.writeComment(piece(6));
one.clear();
state("one clear 2", one);
state("two still 2", two);
const a = createSingleLineStringWriter();
const b = createSingleLineStringWriter();
a.write(piece(7));
b.write(piece(8));
a.writeLine();
state("single a", a);
a.clear();
state("single a clear", a);
state("single b", b);
`;

function emit(text) {
    const interfaces = extract(originalTypes, types, ['SymbolWriter', 'EmitTextWriter'],
        ts.isInterfaceDeclaration).replace('symbol: Symbol', 'symbol: WriterSymbol');
    const factories = extract(text, utilities, ['createSingleLineStringWriter', 'createTextWriter'],
        ts.isFunctionDeclaration);
    const result = ts.transpileModule(dependencies + interfaces + '\n' + factories + fixture, {
        compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.deepEqual(result.diagnostics, []);
    return result.outputText;
}

function observe(javascript) {
    const lines = [];
    const context = { exports: {}, console: { log: line => lines.push(line) } };
    vm.runInNewContext(javascript, context, { timeout: 5000 });
    return lines;
}

test('clean input and byte-identical replay', t => {
    const { root, file, output } = fixtureCase(t, original, 0);
    assert.notEqual(digest(output), digest(original));
    assert.equal(output, adapted);
    const replay = runAdapter(root);
    assert.equal(replay.status, 0, replay.stderr);
    assert.equal(digest(fs.readFileSync(file)), digest(output));
    t.diagnostic('exit 0 on both runs, second run byte-identical');
});

test('interleaved writers preserve all ten Node observations', t => {
    const { output } = fixtureCase(t, original, 0);
    const expected = observe(emit(original));
    const adaptedJavascript = emit(output);
    const actual = observe(adaptedJavascript);
    assert.equal(expected.length, 10);
    assert.equal(actual.length, 10);
    for (const [index, line] of expected.entries()) {
        assert.equal(actual[index], line, `writer observation ${index + 1}: ${line}`);
    }
    t.diagnostic('Node observations SHA-256 ' + digest(expected.join('\n') + '\n'));
});
