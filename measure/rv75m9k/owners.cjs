#!/usr/bin/env node
'use strict';
// #rv75m9k: writes road A's owners.json from scan.cjs's sites. A member is road A when every site reading it is a
// method signature and every declaration of it is @internal. Each is pinned whole: before is the method signature as
// the tree holds it, after is the readonly function-typed property of the same type, built from the signature's own
// text (type parameters, parameter list, return type), so nothing is retyped by hand. An overload group, a signature
// with no return type, or a member declared twice is refused and listed, never guessed at.
// Usage: NODE_PATH=<stage3/api/node_modules> node owners.cjs <tree> <sites.json> <owners.json>
const ts = require('typescript'), fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
const [tree, sitesPath, out] = process.argv.slice(2);
const sites = JSON.parse(fs.readFileSync(sitesPath, 'utf8'));
const wanted = new Map();
for (const site of sites) {
    if (!site.internal || !site.kinds.every(kind => kind === 'MethodSignature')) continue;
    for (const declaration of site.declarations) {
        const file = declaration.where.replace(/:\d+$/, '');
        const [interfaceName, member] = site.member.split('.');
        wanted.set(`${file}|${interfaceName}|${member}`, { file, interface: interfaceName, member });
    }
}
const owners = [], refused = [];
for (const want of [...wanted.values()].sort((a, b) => (a.file + a.interface + a.member).localeCompare(b.file + b.interface + b.member))) {
    const text = fs.readFileSync(path.join(tree, want.file), 'utf8').replace(/\r\n/g, '\n');
    const source = ts.createSourceFile(want.file, text, ts.ScriptTarget.Latest, true);
    const interfaces = [];
    ts.forEachChild(source, function find(node) {
        if (ts.isInterfaceDeclaration(node) && node.name.text === want.interface) interfaces.push(node);
        ts.forEachChild(node, find);
    });
    if (interfaces.length !== 1) { refused.push({ ...want, why: `${interfaces.length} interfaces named ${want.interface}` }); continue; }
    const members = interfaces[0].members.filter(member => member.name && ts.isIdentifier(member.name) && member.name.text === want.member);
    if (members.length !== 1) { refused.push({ ...want, why: `${members.length} members named ${want.member} (an overload group needs a ruling)` }); continue; }
    const member = members[0];
    if (!ts.isMethodSignature(member)) { refused.push({ ...want, why: `a ${ts.SyntaxKind[member.kind]}, not a method signature` }); continue; }
    if (!member.type) { refused.push({ ...want, why: 'no return type' }); continue; }
    const before = text.slice(member.getStart(source), member.end);
    // The leading JSDoc and modifiers stay where they are: getStart skips trivia, and a method signature has no
    // modifiers of its own here (checked), so the member's text starts at its name.
    if (member.modifiers && member.modifiers.length) { refused.push({ ...want, why: 'modifiers on the method signature' }); continue; }
    const typeParameters = member.typeParameters ? text.slice(member.typeParameters.pos - 1, member.typeParameters.end + 1) : '';
    const open = text.indexOf('(', member.typeParameters ? member.typeParameters.end + 1 : member.name.end);
    const close = member.parameters.end;
    if (text[close] !== ')') { refused.push({ ...want, why: 'unexpected parameter list end' }); continue; }
    const parameters = text.slice(open, close + 1);
    const returnType = member.type.getText(source);
    const terminator = before.endsWith(';') ? ';' : before.endsWith(',') ? ',' : '';
    const after = `readonly ${want.member}${member.questionToken ? '?' : ''}: ${typeParameters}${parameters} => ${returnType}${terminator}`;
    owners.push({ ...want, before, after, beforeSha256: digest(before), afterSha256: digest(after) });
}
fs.writeFileSync(out, JSON.stringify(owners, null, 2) + '\n');
console.log(`owners: ${owners.length} members pinned, ${refused.length} refused`);
for (const owner of owners) console.log(`owner ${owner.file} ${owner.interface}.${owner.member}\n  before ${owner.before.replace(/\s+/g, ' ')}\n  after  ${owner.after.replace(/\s+/g, ' ')}`);
for (const refusal of refused) console.log(`refused ${refusal.file} ${refusal.interface}.${refusal.member}: ${refusal.why}`);
