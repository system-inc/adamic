#!/usr/bin/env node
'use strict';
// #rv75m9k: writes road A's owners.json from scan.cjs's sites. A member is road A when every site reading it is a
// method signature, every declaration of it is @internal, and the tree never assigns it (a written member can't be a
// readonly field, and a writable one would meet "a method becomes a writable field" where a public host is seen as the
// internal one, so it stays a method and takes road B). Each is pinned whole: before is the method signature as the
// tree holds it, after is the readonly function-typed property of the same type, built from the signature's own text.
// An overload group becomes one property whose type literal holds every overload, in order. Interfaces merge, so a
// member is found in the one declaration of its interface that holds it. Anything else is refused and listed.
// Usage: NODE_PATH=<stage3/api/node_modules> node owners.cjs <tree> <sites.json> <owners.json>
const ts = require('typescript'), fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
const [tree, sitesPath, out] = process.argv.slice(2);
const sites = JSON.parse(fs.readFileSync(sitesPath, 'utf8'));
const written = new Set(JSON.parse(fs.readFileSync(sitesPath.replace(/\.json$/, '.writes.json'), 'utf8')));
const wanted = new Map(), refused = [];
for (const site of sites) {
    if (!site.internal || !site.kinds.every(kind => kind === 'MethodSignature')) continue;
    const [interfaceName, member] = site.member.split('.');
    for (const declaration of site.declarations) {
        const file = declaration.where.replace(/:\d+$/, '');
        const key = `${file}|${interfaceName}|${member}`;
        if (written.has(site.member)) { if (!wanted.has(key)) refused.push({ file, interface: interfaceName, member, why: 'the tree assigns it, so it stays a method (road B)' }); wanted.set(key, null); continue; }
        if (!wanted.has(key)) wanted.set(key, { file, interface: interfaceName, member });
    }
}
// A function type's text for one method signature: its type parameters, parameter list and return type, verbatim.
function functionType(text, source, signature) {
    const typeParameters = signature.typeParameters ? text.slice(signature.typeParameters.pos - 1, signature.typeParameters.end + 1) : '';
    const open = text.indexOf('(', signature.typeParameters ? signature.typeParameters.end + 1 : signature.name.end);
    const close = signature.parameters.end;
    if (text[close] !== ')') throw new Error(`unexpected parameter list end at ${signature.name.text}`);
    return { typeParameters, parameters: text.slice(open, close + 1), returnType: signature.type.getText(source) };
}
const owners = [];
for (const want of [...wanted.values()].filter(Boolean).sort((a, b) => (a.file + a.interface + a.member).localeCompare(b.file + b.interface + b.member))) {
    const text = fs.readFileSync(path.join(tree, want.file), 'utf8').replace(/\r\n/g, '\n');
    const source = ts.createSourceFile(want.file, text, ts.ScriptTarget.Latest, true);
    const holders = [];
    ts.forEachChild(source, function find(node) {
        if (ts.isInterfaceDeclaration(node) && node.name.text === want.interface && node.members.some(member => member.name && ts.isIdentifier(member.name) && member.name.text === want.member)) holders.push(node);
        ts.forEachChild(node, find);
    });
    if (holders.length !== 1) { refused.push({ ...want, why: `${holders.length} declarations of ${want.interface} hold ${want.member}` }); continue; }
    const members = holders[0].members.filter(member => member.name && ts.isIdentifier(member.name) && member.name.text === want.member);
    const indexes = members.map(member => holders[0].members.indexOf(member));
    if (indexes.some((index, at) => index !== indexes[0] + at)) { refused.push({ ...want, why: 'an overload group whose members aren\'t adjacent' }); continue; }
    const why = members.map(member => !ts.isMethodSignature(member) ? `a ${ts.SyntaxKind[member.kind]}` : !member.type ? 'no return type' : member.modifiers && member.modifiers.length ? 'modifiers' : '').find(Boolean);
    if (why) { refused.push({ ...want, why }); continue; }
    if (new Set(members.map(member => Boolean(member.questionToken))).size !== 1) { refused.push({ ...want, why: 'overloads that disagree on optional' }); continue; }
    const start = members[0].getStart(source), end = members.at(-1).end, before = text.slice(start, end);
    // Comments between overloads would be lost in the type literal: refuse rather than drop them.
    if (members.length > 1 && members.slice(1).some(member => /\/\/|\/\*/.test(text.slice(member.pos, member.getStart(source))))) { refused.push({ ...want, why: 'comments between overloads' }); continue; }
    const types = members.map(member => functionType(text, source, member));
    const terminator = before.endsWith(';') ? ';' : before.endsWith(',') ? ',' : '';
    const optional = members[0].questionToken ? '?' : '';
    const type = types.length === 1
        ? `${types[0].typeParameters}${types[0].parameters} => ${types[0].returnType}`
        : `{ ${types.map(signature => `${signature.typeParameters}${signature.parameters}: ${signature.returnType};`).join(' ')} }`;
    const after = `readonly ${want.member}${optional}: ${type}${terminator}`;
    owners.push({ ...want, overloads: members.length, before, after, beforeSha256: digest(before), afterSha256: digest(after) });
}
fs.writeFileSync(out, JSON.stringify(owners, null, 2) + '\n');
console.log(`owners: ${owners.length} members pinned, ${refused.length} refused`);
for (const owner of owners) console.log(`owner ${owner.file} ${owner.interface}.${owner.member}\n  before ${owner.before.replace(/\s+/g, ' ')}\n  after  ${owner.after.replace(/\s+/g, ' ')}`);
for (const refusal of refused) console.log(`refused ${refusal.file} ${refusal.interface}.${refusal.member}: ${refusal.why}`);
