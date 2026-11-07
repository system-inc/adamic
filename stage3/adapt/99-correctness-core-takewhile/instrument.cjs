'use strict';
const fs = require('node:fs'), ts = require('typescript'), crypto = require('node:crypto');
const file = process.argv[2] + '/src/compiler/core.ts', before = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true);
const found = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'takeWhile' && n.body);
if (found.length !== 1) throw Error('takeWhile implementation count drift');
const node = found[0], body = node.body.getText(source);
if (!body.includes('return array.slice(0, index) as U[];')) throw Error('original takeWhile body drift');
const signature = 'function(array: readonly T[] | undefined, predicate: (element: T) => element is U): U[] | undefined ';
const proposed = '{ if (array !== undefined) { const result: U[] = []; const len = array.length; let index = 0; while (index < len) { const element = array[index]!; if (!predicate(element)) break; result.push(element); index++; } return result; } }';
const replacement = '{ return (globalThis as any).__takeWhileCensus(array, predicate, ' + signature + body + ', ' + signature + proposed + '); }';
const after = before.slice(0, node.body.getStart(source)) + replacement + before.slice(node.body.end);
if (before.split('\n').length !== after.split('\n').length) throw Error('line count changed');
fs.writeFileSync(file, after);
const sha = s => crypto.createHash('sha256').update(s).digest('hex');
console.log(JSON.stringify({ file, beforeSha256: sha(before), afterSha256: sha(after), originalBody: body,
    replacement, lineCountPreserved: true, measurementOnly: true }, null, 2));
