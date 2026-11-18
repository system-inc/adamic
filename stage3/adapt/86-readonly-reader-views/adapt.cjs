#!/usr/bin/env node
'use strict';
// Adaptation 86, #38kdd4d: five fields tsc only reads are said readonly where a value whose own field is readonly is
// handed to them, so each handoff stops being a view that could write the field (family 4, "whose readonly field
// becomes writable"). Adamic ruled it (Nov 18) in adaptation 72's shape. Every edited declaration is internal, so no
// line of the published typescript.d.ts moves. It's type-only:
// - moduleSpecifiers.ts and program.ts: getModuleSpecifierPreferences' importingSourceFile, getDefaultResolutionModeForFile's
//   file, and the two program.ts workers it hands that file to, each a Pick of SourceFile it only reads, take it
//   Readonly, as FutureSourceFile's fileName is readonly;
// - types.ts: the four PrivateIdentifier declarations esDecorators.ts hands to a callback taking a readonly name say
//   their name is readonly too;
// - builderState.ts: canReuseOldState reads only oldState.referencedMap, so it takes a Readonly<BuilderState>;
// - types.ts: Program's internal redirectTargetsMap and getCanonicalFileName are readonly, as a TypeCheckerHost read as a
//   Program declares the first readonly and the second a method;
// - types.ts: SourceFileMayBeEmittedHost's getCanonicalFileName is readonly, as EmitHost declares it a method.
// Through the stock AST each edit is found by its owner and member and its exact upstream text, refused before any write
// if anything has moved, and left as it is on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
// Each edit: a parameter's type (function, parameter) or an interface member (interface, member), with its text as
// upstream 6.0.3 writes it and as this adaptation leaves it.
const edits = [
    { file: 'src/compiler/moduleSpecifiers.ts', function: 'getModuleSpecifierPreferences', parameter: 'importingSourceFile',
        upstream: 'Pick<SourceFile, "fileName" | "impliedNodeFormat">', adapted: 'Readonly<Pick<SourceFile, "fileName" | "impliedNodeFormat">>' },
    { file: 'src/compiler/moduleSpecifiers.ts', function: 'getDefaultResolutionModeForFile', parameter: 'file',
        upstream: 'Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">', adapted: 'Readonly<Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">>' },
    // getDefaultResolutionModeForFile hands its file on to these two, which only read it, so the readonly view goes on.
    { file: 'src/compiler/program.ts', function: 'getDefaultResolutionModeForFileWorker', parameter: 'sourceFile',
        upstream: 'Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">', adapted: 'Readonly<Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">>' },
    { file: 'src/compiler/program.ts', function: 'getImpliedNodeFormatForEmitWorker', parameter: 'sourceFile',
        upstream: 'Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">', adapted: 'Readonly<Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">>' },
    { file: 'src/compiler/builderState.ts', function: 'canReuseOldState', parameter: 'oldState',
        upstream: 'BuilderState | undefined', adapted: 'Readonly<BuilderState> | undefined' },
    ...['PrivateIdentifierPropertyDeclaration', 'PrivateIdentifierMethodDeclaration', 'PrivateIdentifierGetAccessorDeclaration', 'PrivateIdentifierSetAccessorDeclaration']
        .map(name => ({ file: 'src/compiler/types.ts', interface: name, member: 'name', upstream: 'name: PrivateIdentifier;', adapted: 'readonly name: PrivateIdentifier;' })),
    { file: 'src/compiler/types.ts', interface: 'Program', member: 'redirectTargetsMap',
        upstream: 'redirectTargetsMap: MultiMap<Path, string>;', adapted: 'readonly redirectTargetsMap: MultiMap<Path, string>;' },
    // checker.ts reads a TypeCheckerHost as a Program, whose getCanonicalFileName is a method on the host's side.
    { file: 'src/compiler/types.ts', interface: 'Program', member: 'getCanonicalFileName',
        upstream: 'getCanonicalFileName: GetCanonicalFileName;', adapted: 'readonly getCanonicalFileName: GetCanonicalFileName;' },
    { file: 'src/compiler/types.ts', interface: 'SourceFileMayBeEmittedHost', member: 'getCanonicalFileName',
        upstream: 'getCanonicalFileName: GetCanonicalFileName;', adapted: 'readonly getCanonicalFileName: GetCanonicalFileName;' },
];
const report = [];
// Every file is read and every edit found before anything is written, so a refusal leaves the tree as it was.
const writes = [];
for (const file of [...new Set(edits.map(edit => edit.file))]) {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const changes = [];
    for (const edit of edits.filter(edit => edit.file === file)) {
        const owners = [];
        ts.forEachChild(source, function find(node) {
            if (edit.function && ts.isFunctionDeclaration(node) && node.name && node.name.text === edit.function) owners.push(node);
            if (edit.interface && ts.isInterfaceDeclaration(node) && node.name.text === edit.interface) owners.push(node);
            ts.forEachChild(node, find);
        });
        const label = `${edit.function || edit.interface}.${edit.parameter || edit.member}`;
        let target;
        if (edit.function) {
            if (owners.length !== 1) throw new Error(`expected exactly one ${edit.function} in ${file}, found ${owners.length}`);
            const parameters = owners[0].parameters.filter(parameter => ts.isIdentifier(parameter.name) && parameter.name.text === edit.parameter);
            if (parameters.length !== 1 || !parameters[0].type) throw new Error(`expected one typed parameter ${label}`);
            target = parameters[0].type;
        }
        else {
            // An interface may be declared more than once (types.ts merges two Programs), so the member is looked for
            // across every declaration of it, and must be declared exactly once.
            const members = owners.flatMap(owner => owner.members.filter(member => member.name && ts.isIdentifier(member.name) && member.name.text === edit.member));
            if (owners.length === 0 || members.length !== 1 || !ts.isPropertySignature(members[0])) throw new Error(`expected one property ${label} across ${owners.length} declarations`);
            target = members[0];
        }
        // A property's text starts after its doc comment, so the comment and the @internal tag keep their bytes.
        const written = target.getText(source);
        const where = source.getLineAndCharacterOfPosition(target.getStart(source));
        const row = { edit: label, where: `${file}:${where.line + 1}:${where.character + 1}` };
        if (written === edit.adapted) {
            report.push({ ...row, edited: false });
            continue;
        }
        if (written !== edit.upstream) throw new Error(`unexpected text at ${label}: ${written}`);
        changes.push({ at: target.getStart(source), end: target.end, adapted: edit.adapted });
        report.push({ ...row, edited: true });
    }
    let adapted = text;
    for (const change of changes.sort((a, b) => b.at - a.at)) adapted = adapted.slice(0, change.at) + change.adapted + adapted.slice(change.end);
    if (adapted !== text) writes.push({ fileName, text, adapted });
}
for (const write of writes) {
    if (fs.readFileSync(write.fileName, 'utf8') !== write.text) throw new Error(`source changed during adaptation: ${write.fileName}`);
    fs.writeFileSync(write.fileName, write.adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#38kdd4d', ruling: "Adamic, Nov 18, in adaptation 72's shape", edits: report }, null, 2));
