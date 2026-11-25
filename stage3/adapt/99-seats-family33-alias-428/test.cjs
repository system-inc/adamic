'use strict';

const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const api = path.resolve(__dirname, '../../api/node_modules');
const ts = require(path.join(api, 'typescript'));

assert.equal(process.argv.length, 3, 'usage: node test.cjs <applied tree>');
assert.equal(ts.version, '6.0.3');
const tree = process.argv[2];
const file = 'src/compiler/transformers/utilities.ts';
const snapshot = 'refs/stage3/99-seats-family33-alias-428';
const childLimits = {timeout: 60_000, maxBuffer: 1024 * 1024};
function readRef(side) {
    return execFileSync('git', ['-C', tree, 'show', `${snapshot}-${side}:${file}`], childLimits);
}

const before = readRef('before');
const after = readRef('after');

function checkOutput() {
    const scratch = fs.mkdtempSync('/tmp/adamic-alias-428-');
    try {
        const input = path.join(scratch, file);
        fs.mkdirSync(path.dirname(input), {recursive: true});
        fs.writeFileSync(input, before);
        const adapt = () => execFileSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], {
            ...childLimits,
            encoding: 'utf8',
            env: {...process.env, NODE_PATH: api},
        });
        assert.equal(adapt().trim(), 'applied family 33 alias at 428');
        assert.deepEqual(fs.readFileSync(input), after, 'adapter output differs from the -after ref');
        const applied = fs.statSync(input, {bigint: true});
        assert.equal(adapt().trim(), 'family 33 alias at 428 already applied');
        const replayed = fs.statSync(input, {bigint: true});
        assert.equal(replayed.mtimeNs, applied.mtimeNs, 'replay changed modification time');
        assert.equal(replayed.ctimeNs, applied.ctimeNs, 'replay changed inode metadata');
        assert.deepEqual(fs.readFileSync(input), after, 'replay changed file bytes');
    } finally {
        fs.rmSync(scratch, {recursive: true, force: true});
    }
    console.log('PASS adapter output matches -after and replay performs no write');
}

function checkCut() {
    const original = before.toString('utf8');
    const indent = '                ';
    const conditional = indent + 'const baseName = isMemberName(node) && node !== name ? IdentifierNameMap.toKey(node) : `(generated@${getNodeId(node)})`;';
    const parts = original.split(conditional);
    assert.equal(parts.length, 2, 'expected one original conditional');
    const newline = parts[1].startsWith('\r\n') ? '\r\n' : '\n';
    const alias = indent + 'const nodeForId = node;';
    const replacement = conditional.replace('getNodeId(node)', 'getNodeId(nodeForId)');
    const expected = parts[0] + alias + newline + replacement + parts[1];
    assert.deepEqual(after, Buffer.from(expected), 'snapshot delta exceeds the alias and its one read');
    console.log('PASS snapshot cut contains only the alias and its one read');
}

function loadClass(bytes) {
    const source = ts.createSourceFile(file, bytes.toString('utf8'), ts.ScriptTarget.Latest, true);
    assert.deepEqual(source.parseDiagnostics, []);
    const owners = source.statements.filter(node => ts.isClassDeclaration(node) && node.name?.text === 'IdentifierNameMap');
    assert.equal(owners.length, 1, 'expected one IdentifierNameMap class');
    const owner = owners[0];
    const start = owner.modifiers?.at(-1).end ?? owner.getStart(source);
    const {outputText, diagnostics} = ts.transpileModule(
        source.text.slice(start, owner.end),
        {compilerOptions: {target: ts.ScriptTarget.ES2022}, reportDiagnostics: true},
    );
    assert.deepEqual(diagnostics, []);
    return outputText;
}

const GeneratedIdentifierFlags = {KindMask: 7, Auto: 1, Node: 4};

function generated(text, flags, isPrivate = false) {
    return {
        text,
        generated: true,
        member: true,
        private: isPrivate,
        emitNode: {autoGenerate: {flags, id: 71, prefix: '[', suffix: ']'}},
    };
}

function drive(output, name, generatedNode, ids) {
    const calls = [];
    const generatedNameCalls = [];
    const nodeIdCalls = [];
    const record = (label, read) => (...args) => {
        calls.push([label, ...args]);
        return read(...args);
    };
    const stubs = {
        GeneratedIdentifierFlags,
        isGeneratedPrivateIdentifier: record('isGeneratedPrivateIdentifier', value => value.generated && value.private),
        isGeneratedIdentifier: record('isGeneratedIdentifier', value => value.generated && !value.private),
        getNodeForGeneratedName: record('getNodeForGeneratedName', value => {
            generatedNameCalls.push(value);
            return generatedNode;
        }),
        isMemberName: record('isMemberName', value => value.member),
        getNodeId: record('getNodeId', value => {
            nodeIdCalls.push(value);
            if (!ids.has(value)) ids.set(value, ids.size + 1);
            return ids.get(value);
        }),
        formatGeneratedName(privateName, prefix, key, suffix, keyOf) {
            calls.push(['formatGeneratedName', privateName, prefix, key, suffix, keyOf.name]);
            return prefix + key + suffix;
        },
        isPrivateIdentifier: record('isPrivateIdentifier', value => value.private),
        idText: record('idText', value => value.text),
    };
    const toKey = new Function(...Object.keys(stubs), `"use strict";\n${output}\nreturn name => IdentifierNameMap.toKey(name);`)(
        ...Object.values(stubs),
    );
    const key = toKey(name);
    return {key, calls, generatedNameCalls, nodeIdCalls};
}

function checkDrives() {
    const original = loadClass(before);
    const adapted = loadClass(after);
    const self = generated('self', GeneratedIdentifierFlags.Node);
    const nonMember = {text: 'nonmember', member: false};
    const rows = [
        ['recursive member-name arm', generated('#generatedPrivate', GeneratedIdentifierFlags.Node, true),
            {text: 'member', member: true}, 1, [], '[member]'],
        ['generated node is the name itself', self, self, 1, [self], '[(generated@1)]'],
        ['getNodeId arm for a non-member node', generated('generated', GeneratedIdentifierFlags.Node),
            nonMember, 1, [nonMember], '[(generated@1)]'],
        ['auto-kind generated name', generated('auto', GeneratedIdentifierFlags.Auto), undefined, 0, [], '[(auto@71)]'],
        ['private identifier', {text: '#private', private: true}, undefined, 0, [], 'private'],
        ['plain identifier', {text: 'plain'}, undefined, 0, [], 'plain'],
    ];
    for (const [label, name, generatedNode, expectedNameCalls, expectedIds, expectedKey] of rows) {
        const ids = new Map();
        const base = drive(original, name, generatedNode, ids);
        const head = drive(adapted, name, generatedNode, ids);
        assert.equal(base.generatedNameCalls.length, expectedNameCalls, label);
        assert.equal(head.generatedNameCalls.length, base.generatedNameCalls.length, label);
        assert.equal(base.nodeIdCalls.length, expectedIds.length, label);
        assert.equal(head.nodeIdCalls.length, base.nodeIdCalls.length, label);
        for (const [index, node] of base.nodeIdCalls.entries()) {
            assert.strictEqual(node, expectedIds[index], `${label}: original getNodeId argument`);
            assert.strictEqual(head.nodeIdCalls[index], node, `${label}: getNodeId argument identity`);
        }
        assert.equal(base.key, expectedKey, label);
        assert.equal(head.key, base.key, label);
        assert.deepEqual(head.calls, base.calls, `${label}: call log`);
        console.log(`PASS toKey ${label}`);
    }
}

function withOuterRead(bytes, position) {
    const text = bytes.toString('utf8');
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    const lines = text.split(newline);
    function anchor(line) {
        const index = lines.indexOf(line);
        assert(index >= 0 && index === lines.lastIndexOf(line), 'expected one anchor: ' + line);
        return index;
    }

    const owner = anchor('export class IdentifierNameMap<V> {');
    assert.equal(lines[owner - 1], '/** @internal */');
    lines.splice(owner - 1, 0, 'const nodeForId = 1;', '');
    const indent = '                ';
    if (position === 'before') {
        const initializer = anchor(indent + 'const node = getNodeForGeneratedName(name);');
        lines.splice(initializer, 0, indent + 'void nodeForId;');
    } else {
        const argument = position === 'replay' ? 'nodeForId' : 'node';
        const conditional = anchor(indent + 'const baseName = isMemberName(node) && node !== name ? IdentifierNameMap.toKey(node) : `(generated@${getNodeId(' + argument + ')})`;');
        lines.splice(conditional + 1, 0, indent + 'if (autoGenerate.id === 72 && String(nodeForId) === "1") return "outer:" + baseName;');
    }
    return Buffer.from(lines.join(newline));
}

function checkOuterReads() {
    const rows = [
        ['outer read after the insertion', before, 'after'],
        ['outer read before the insertion', before, 'before'],
        ['outer read on replay', after, 'replay'],
    ];
    for (const [label, snapshotBytes, position] of rows) {
        const scratch = fs.mkdtempSync('/tmp/adamic-alias-428-drift-');
        try {
            const input = path.join(scratch, file);
            const bytes = withOuterRead(snapshotBytes, position);
            fs.mkdirSync(path.dirname(input), {recursive: true});
            fs.writeFileSync(input, bytes);
            const modified = fs.statSync(input, {bigint: true}).mtimeNs;
            assert.throws(() => execFileSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), scratch], {
                ...childLimits,
                encoding: 'utf8',
                env: {...process.env, NODE_PATH: api},
                stdio: ['ignore', 'pipe', 'pipe'],
            }), error => {
                assert(Number.isInteger(error.status) && error.status > 0, label + ': expected a nonzero exit');
                assert.match(error.stderr, /unexpected nodeForId at \d+:\d+/, label);
                return true;
            }, label + ': expected refusal');
            assert.deepEqual(fs.readFileSync(input), bytes, label + ': source bytes changed');
            assert.equal(fs.statSync(input, {bigint: true}).mtimeNs, modified, label + ': modification time changed');
        } finally {
            fs.rmSync(scratch, {recursive: true, force: true});
        }
        console.log('PASS rejects ' + label + ' without writing');
    }
}

checkOutput();
// Check behavior before the exact cut so a wrong binding fails at its observed argument.
checkDrives();
checkCut();
checkOuterReads();
