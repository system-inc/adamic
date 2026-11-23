'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const record = 'Record<string, SourceMapRange | undefined>';
const mergeBefore = [
    'function mergeTokenSourceMapRanges(sourceRanges: (SourceMapRange | undefined)[], destRanges: (SourceMapRange | undefined)[]): (SourceMapRange | undefined)[] {',
    '    if (!destRanges) destRanges = [];',
    '    for (const key in sourceRanges) {',
    '        destRanges[key] = sourceRanges[key];',
    '    }',
    '    return destRanges;',
    '}',
].join('\n');
const mergeAfter = mergeBefore
    .replaceAll('(SourceMapRange | undefined)[]', record)
    .replace('destRanges = [];', 'destRanges = {};');

const files = {
    'types.ts': [
        ['EmitNode', ts.isInterfaceDeclaration, ts.isPropertySignature,
            'tokenSourceMapRanges?: (SourceMapRange | undefined)[];',
            'tokenSourceMapRanges?: ' + record + ';'],
    ],
    'factory/emitNode.ts': [
        ['getTokenSourceMapRange', ts.isFunctionDeclaration, ts.isBlock,
            '{\n    return node.emitNode?.tokenSourceMapRanges?.[token];\n}',
            '{\n    const t = node.emitNode?.tokenSourceMapRanges;\n    return t === undefined ? undefined : t[String(token)];\n}'],
        ['setTokenSourceMapRange', ts.isFunctionDeclaration, ts.isVariableStatement,
            'const tokenSourceMapRanges = emitNode.tokenSourceMapRanges ?? (emitNode.tokenSourceMapRanges = []);',
            'const tokenSourceMapRanges = emitNode.tokenSourceMapRanges ?? (emitNode.tokenSourceMapRanges = {});'],
        ['setTokenSourceMapRange', ts.isFunctionDeclaration, ts.isExpressionStatement,
            'tokenSourceMapRanges[token] = range;',
            'tokenSourceMapRanges[String(token)] = range;'],
    ],
    'emitter.ts': [
        ['emitTokenWithSourceMap', ts.isFunctionDeclaration, ts.isElementAccessExpression,
            'emitNode.tokenSourceMapRanges[token]', 'emitNode.tokenSourceMapRanges[String(token)]'],
    ],
    'factory/nodeFactory.ts': [
        ['mergeTokenSourceMapRanges', ts.isFunctionDeclaration, ts.isFunctionDeclaration,
            mergeBefore, mergeAfter],
    ],
};

function findNodes(root, predicate) {
    const matches = [];
    function visit(node) {
        if (predicate(node)) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return matches;
}

function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw Error(file + ': cannot parse source');
    return source;
}

function plan(file, text, pins) {
    const source = parse(file, text);
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    const edits = [];
    for (const [name, ownerKind, nodeKind, before, after] of pins) {
        const owners = findNodes(source, node => ownerKind(node) && node.name?.text === name);
        if (owners.length !== 1) throw Error(file + ': expected one owner ' + name);
        const candidates = findNodes(owners[0], node => {
            if (!nodeKind(node)) return false;
            const current = node.getText(source).replaceAll('\r\n', '\n');
            return current === before || current === after;
        });
        if (candidates.length !== 1) throw Error(file + ': changed or duplicate snapshot: ' + before);
        const node = candidates[0];
        if (node.getText(source).replaceAll('\r\n', '\n') === before) {
            edits.push({ start: node.getStart(source), end: node.end, replacement: after.replaceAll('\n', newline) });
        }
    }
    let result = text;
    for (const { start, end, replacement } of edits.sort((a, b) => b.start - a.start)) {
        result = result.slice(0, start) + replacement + result.slice(end);
    }
    parse(file, result);
    return { result, count: edits.length };
}

function main(tree) {
    if (ts.version !== '6.0.3') throw Error('expected stock TypeScript 6.0.3');
    if (!tree || process.argv.length !== 3) throw Error('usage: node adapt.cjs TREE');
    const plans = Object.entries(files).map(([file, pins]) => {
        const target = path.resolve(tree, 'src/compiler', file);
        const original = fs.readFileSync(target, 'utf8');
        return { target, original, ...plan(file, original, pins) };
    });
    const count = plans.reduce((sum, item) => sum + item.count, 0);
    const expected = Object.values(files).reduce((sum, pins) => sum + pins.length, 0);
    if (count !== 0 && count !== expected) throw Error('partially adapted token source-map table');
    for (const { target, original } of plans) {
        if (fs.readFileSync(target, 'utf8') !== original) throw Error(target + ': changed during planning');
    }
    for (const { target, result, count: edits } of plans) {
        if (edits) fs.writeFileSync(target, result);
    }
    console.log(JSON.stringify({ files: plans.filter(item => item.count).length, edits: count }));
}

main(process.argv[2]);
