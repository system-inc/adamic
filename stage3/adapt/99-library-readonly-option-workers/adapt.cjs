#!/usr/bin/env node
'use strict';
// #6c4swvt's piece 4: commandLineParser.ts's serializeOptionBaseObject takes its options as a readonly record, so the
// WatchOptions serializeWatchOptions hands it (2792) is no longer a view that can write a TsConfigSourceFile where
// WatchOptions' own index reads a CompilerOptionsValue (Adamic's invariance refusal). Correctness traced the receiver
// (typescript/options-base-evidence 5a91be06): it enumerates and reads its options and writes only a fresh map. The
// lowering that reads a record through a readonly view is piece 4's (library/readonly-record-views). It's @internal.
//
// serializeCompilerOptions and handleOptionConfigDirTemplateSubstitution stay as upstream writes them: retyping them
// as well (J-cdcj9md) traded their refusals for two lowering gaps, a record seen as its own Readonly<> (2675, 2712)
// and core.ts's assign returning its fresh target as the readonly source's type (3313). Each is a piece of its own.
//
// It sorts after 99-correctness-axolotl-j, which pins serializeOptionBaseObject's whole owner by its hash, and after
// 81-readonly-option-values, whose casts sit inside it: both run first on upstream's signature, and this edits only
// the signature. A signature already adapted keeps its bytes, so a second run changes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const relative = 'src/compiler/commandLineParser.ts';
const fileName = path.join(tree, relative);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
// The worker's options parameter as upstream writes it and as this leaves it.
const workers = [
    { name: 'serializeOptionBaseObject', before: 'OptionsBase', after: 'Readonly<OptionsBase>' },
];
const edits = [], report = [];
for (const worker of workers) {
    const found = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.body && node.name?.text === worker.name);
    if (found.length !== 1) throw new Error(`expected one top-level ${worker.name} in ${relative}, found ${found.length}`);
    const declaration = found[0];
    const parameter = declaration.parameters[0];
    if (!parameter || !ts.isIdentifier(parameter.name) || parameter.name.text !== 'options' || !parameter.type)
        throw new Error(`${worker.name}'s first parameter isn't a typed options`);
    const now = parameter.type.getText(source);
    if (now !== worker.before && now !== worker.after)
        throw new Error(`${worker.name}'s options is ${JSON.stringify(now)}, want ${JSON.stringify(worker.before)}`);
    const parameterEdited = now === worker.before;
    if (parameterEdited) edits.push({ start: parameter.type.getStart(source), end: parameter.type.end, text: worker.after });
    let resultEdited = false;
    if (worker.result) {
        const annotated = declaration.type ? declaration.type.getText(source) : undefined;
        if (annotated !== undefined && annotated !== worker.result)
            throw new Error(`${worker.name}'s result is annotated ${JSON.stringify(annotated)}, want none or ${JSON.stringify(worker.result)}`);
        if (annotated === undefined) {
            // The annotation goes right after the parameter list's closing parenthesis.
            const close = declaration.parameters.end;
            const paren = text.indexOf(')', close);
            if (paren < 0 || paren > declaration.body.getStart(source) || text.slice(close, paren).trim() !== '' && text.slice(close, paren).trim() !== ',')
                throw new Error(`${worker.name}'s parameter list doesn't close where expected`);
            edits.push({ start: paren + 1, end: paren + 1, text: ': ' + worker.result });
            resultEdited = true;
        }
    }
    const at = source.getLineAndCharacterOfPosition(parameter.getStart(source));
    report.push({ owner: worker.name, where: `${relative}:${at.line + 1}:${at.character + 1}`, edited: parameterEdited || resultEdited });
}
if ([...new Set(report.map(edit => edit.edited))].length !== 1) throw new Error(`${relative} is partly adapted: ${JSON.stringify(report)}`);
if (edits.length) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    fs.writeFileSync(fileName, result);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#6c4swvt', certificate: "the compiled program's gate", edits: report }, null, 2));
