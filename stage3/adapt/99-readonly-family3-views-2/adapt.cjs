#!/usr/bin/env node
'use strict';
// Adaptation 99-readonly-family3-views-2, #k8ed1f2, family 3's causes 4 and 11. Family 3 is the refusal "a value of
// type A seen as B, which can write X where Y is read". At 4 properties in src/compiler/builder.ts, each only read
// through the view tsc hands it to, the property gains `readonly`, so 15 of the family's sites clear. Adamic delegated
// the ruling on #k8ed1f2: readonly where tsc never writes, the ruling 99-readonly-family3-views applied. Cause 4's two
// properties, on ReusableBuilderProgramState, already hold that adaptation's `ReadonlyMap<...>` and keep it; cause
// 11's two `next` properties, on ReusableRepopulateInfoChain and SerializedDiagnosticMessageChain, take
// `readonly ReusableDiagnosticMessageChain[]` with the modifier. README.md says why each slot, what was left out, and
// every write found to a touched property. Type-only: the emitted JavaScript does not change. Each declaration is
// `@internal`, so the published declaration files do not change and no api-sanctions.json is needed. Through the stock
// AST each slot is found by its interface or type alias and property name, never by line, offset or whole-line text;
// the edit inserts `readonly ` before the property's name and, where its type changes, replaces the type node's span.
// It refuses before any write unless every slot is found exactly once, holds the text it expects or exactly its
// adapted text, and is local, file-local or `@internal`, and a second run changes nothing. It sorts right after
// 99-readonly-family3-views, whose output it reads, and before 99-readonly-inferred-bindings, which edits another
// declaration in builder.ts by name.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const builder = 'src/compiler/builder.ts';
// Cause 4's type as 99-readonly-family3-views leaves it. Upstream's `Map<...>` here means that adaptation did not
// run first, and the slot refuses rather than make a writable map's slot readonly on its own.
const perFile = 'ReadonlyMap<Path, readonly ReusableDiagnostic[] | readonly Diagnostic[]>';
const chains = { before: 'ReusableDiagnosticMessageChain[]', after: 'readonly ReusableDiagnosticMessageChain[]' };
// Each slot, addressed by name, with the cause it clears: an interface's or a type alias's property (`interface` or
// `typeAlias`, `property`), optional or not as upstream declares it, which gains `readonly` and whose type goes from
// `before`, the text this adaptation finds, to `after`. Cause 4's type does not change. The alias's property is a
// member of the type literal inside it, the one SerializedDiagnosticMessageChain intersects with its Omit<...>.
const slots = [
    { cause: 4, file: builder, interface: 'ReusableBuilderProgramState', property: 'semanticDiagnosticsPerFile',
        optional: false, before: perFile, after: perFile },
    { cause: 4, file: builder, interface: 'ReusableBuilderProgramState', property: 'emitDiagnosticsPerFile',
        optional: true, before: `${perFile} | undefined`, after: `${perFile} | undefined` },
    { cause: 11, file: builder, interface: 'ReusableRepopulateInfoChain', property: 'next', optional: true, ...chains },
    { cause: 11, file: builder, typeAlias: 'SerializedDiagnosticMessageChain', property: 'next', optional: true,
        ...chains },
];
const nodes = (root, predicate) => {
    const found = [];
    ts.forEachChild(root, function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    return found;
};
const one = (found, what) => {
    if (found.length !== 1) throw new Error(`expected one ${what}, found ${found.length}`);
    return found[0];
};
const name = slot => `${slot.interface ?? slot.typeAlias}.${slot.property}`;
const functionLike = node => ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node)
    || ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) || ts.isGetAccessorDeclaration(node)
    || ts.isSetAccessorDeclaration(node);
const enclosingFunction = node => {
    let outer = node.parent;
    while (outer && !functionLike(outer)) outer = outer.parent;
    return outer;
};
// The declarations that can carry a JSDoc tag: an `@internal` on any of them, the slot's own or an enclosing one,
// keeps it out of the published declaration files.
const documented = node => ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)
    || ts.isFunctionDeclaration(node) || ts.isVariableDeclaration(node) || ts.isVariableStatement(node)
    || ts.isPropertySignature(node) || ts.isMethodSignature(node) || ts.isModuleDeclaration(node)
    || ts.isClassDeclaration(node);
const internal = node => ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal');
const exported = node => node.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword) ?? false;
// Why a declaration can't reach typescript.d.ts: it is local to a function, it or a declaration enclosing it carries
// `@internal`, or it is a top-level statement of a module that does not export it. Anything else refuses.
function visibility(source, declaration, what) {
    const outer = enclosingFunction(declaration);
    if (outer) return `local to ${outer.name && ts.isIdentifier(outer.name) ? outer.name.text : 'a function'}`;
    for (let node = declaration; node && !ts.isSourceFile(node); node = node.parent)
        if (documented(node) && internal(node)) return '@internal';
    if (declaration.parent === source && ts.isExternalModule(source) && !exported(declaration)) return 'file-local';
    throw new Error(`expected ${what} to be local to a function, file-local or @internal`);
}
// The slot's property: an interface's own member, or a member of a type literal inside the alias. A call or index
// signature has no name, so the name is checked before it is read.
function find(source, slot) {
    const what = `${name(slot)} in ${slot.file}`;
    const container = slot.interface
        ? one(nodes(source, node => ts.isInterfaceDeclaration(node) && node.name.text === slot.interface),
            `interface ${slot.interface} in ${slot.file}`)
        : one(nodes(source, node => ts.isTypeAliasDeclaration(node) && node.name.text === slot.typeAlias),
            `type ${slot.typeAlias} in ${slot.file}`);
    const members = slot.interface ? container.members : nodes(container, ts.isTypeElement);
    return one(members.filter(member => ts.isPropertySignature(member) && member.name && ts.isIdentifier(member.name)
        && member.name.text === slot.property), `property ${what}`);
}
const files = [...new Set(slots.map(slot => slot.file))].map(file => {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    return { file, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true) };
});
const records = [];
const addressed = new Set();
for (const slot of slots) {
    const { source } = files.find(file => file.file === slot.file);
    const member = find(source, slot);
    const at = source.getLineAndCharacterOfPosition(member.getStart(source));
    const record = { slot: name(slot), where: `${slot.file}:${at.line + 1}:${at.character + 1}`, file: slot.file };
    // Found exactly once across the table too: two entries naming one node would be one slot read twice.
    const key = `${slot.file}:${member.getStart(source)}`;
    if (addressed.has(key)) throw new Error(`slot ${record.slot} at ${record.where} is already in the table`);
    addressed.add(key);
    record.visibility = visibility(source, member, record.slot);
    if (!member.type) throw new Error(`expected ${record.slot} at ${record.where} to declare its type`);
    const modifiers = member.modifiers?.map(modifier => modifier.getText(source)) ?? [];
    const type = member.type.getText(source);
    // Optional as upstream declares it, so a property tsc made required or optional since is not the slot the ruling
    // covers.
    if (Boolean(member.questionToken) !== slot.optional)
        throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${member.getText(source)}`);
    if (modifiers.length === 1 && modifiers[0] === 'readonly' && type === slot.after)
        records.push({ ...record, edited: false });
    else if (modifiers.length === 0 && type === slot.before) {
        // `readonly ` before the name and the adapted type in the type node's span, as one edit from the name to the
        // type's end: what lies between, `?: ` or `: `, keeps its bytes.
        const start = member.name.getStart(source);
        const between = source.text.slice(start, member.type.getStart(source));
        records.push({ ...record, edited: true, at: start, end: member.type.end,
            text: `readonly ${between}${slot.after}` });
    }
    // Anything else, a half-adapted slot or text neither side expects, is not what the ruling covers.
    else throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${member.getText(source)}`);
}
// Each file's edits from the end, so offsets hold, and no two of them touching the same bytes.
const planned = files.map(({ file, fileName, text }) => {
    const edits = records.filter(record => record.edited && record.file === file).sort((a, b) => b.at - a.at);
    edits.forEach((edit, index) => {
        if (index > 0 && edit.end > edits[index - 1].at)
            throw new Error(`overlapping edits in ${file}: ${edit.slot} and ${edits[index - 1].slot}`);
    });
    let adapted = text;
    for (const edit of edits) adapted = adapted.slice(0, edit.at) + edit.text + adapted.slice(edit.end);
    return { fileName, text, adapted };
});
// Every slot passed: only now is anything written, each file once.
for (const { fileName, text, adapted } of planned) {
    if (adapted === text) continue;
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
for (const record of records)
    console.log(`${record.slot}: ${record.edited ? 'edited' : 'already adapted'} (${record.where}, ${record.visibility})`);
