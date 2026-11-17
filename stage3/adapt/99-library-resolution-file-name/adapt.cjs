#!/usr/bin/env node
'use strict';
// #w743639's F: program.ts's ResolutionWithResolvedFileName gets readonly on resolvedFileName alone, so
// getResolvedModuleFromResolution, which returns a ResolvedModuleFull (resolvedFileName: string), is no longer a function
// whose result can have undefined written into resolvedFileName when it is used as resolveNamesReusingOldState's
// getResolved (program.ts, resolveModuleNamesReusingOldState). resolveNamesReusingOldState and its readers only read
// resolvedFileName off a resolution.
//
// packageId stays as 98-axolotl-d leaves it, widened to `?: PackageId | undefined` and writable: its member row says
// `readonly: false` on purpose, and this edit doesn't touch it. It runs after 98-axolotl-d, which edits this interface's
// packageId, so the directory sorts after 98 and the edit pins the interface as 98-axolotl-d leaves it. Ruled by Adamic
// on #w743639, certified by the compiled program's gate. A member already readonly keeps its bytes, so a second run
// changes nothing.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const relative = 'src/compiler/program.ts';
const fileName = path.join(tree, relative);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const interfaces = source.statements.filter(node => ts.isInterfaceDeclaration(node) && node.name.text === 'ResolutionWithResolvedFileName');
if (interfaces.length !== 1) throw new Error(`expected one ResolutionWithResolvedFileName in ${relative}, found ${interfaces.length}`);
const members = interfaces[0].members.map(member => member.getText(source));
// The interface as 98-axolotl-d leaves it: resolvedFileName as upstream writes it, packageId widened by 98-axolotl-d.
const want = ['resolvedFileName: string | undefined;', 'packageId?: PackageId | undefined;'];
if (JSON.stringify(members) !== JSON.stringify(want) && JSON.stringify(members) !== JSON.stringify(['readonly ' + want[0], want[1]]))
    throw new Error(`ResolutionWithResolvedFileName holds ${JSON.stringify(members)}, want 98-axolotl-d's ${JSON.stringify(want)}`);
const member = interfaces[0].members[0];
const at = source.getLineAndCharacterOfPosition(member.getStart(source));
const edited = members[0] === want[0];
if (edited) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, text.slice(0, member.getStart(source)) + 'readonly ' + text.slice(member.getStart(source)));
}
console.log(JSON.stringify({ typescript: ts.version, task: '#w743639', certificate: "the compiled program's gate",
    edits: [{ owner: 'ResolutionWithResolvedFileName.resolvedFileName', where: `${relative}:${at.line + 1}:${at.character + 1}`, edited }] }, null, 2));
