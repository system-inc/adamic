#!/usr/bin/env node
'use strict';
// Adaptation 77, #5pstmvw: ObjectType's five resolved slots in src/compiler/types.ts, members, properties,
// callSignatures, constructSignatures and indexInfos, become readonly. ResolvedType redeclares all five required, so a
// ResolvedType (or a GenericType | ResolvedType) seen as ObjectType was a view that could write undefined where
// ResolvedType reads a value (21 refusals in checker.ts). With the slots readonly on ObjectType the view can't write
// them. ResolvedType keeps its own five writable, as upstream declares them, and every other interface in the family
// stays as upstream declares it. A new internal interface, ObjectTypeUnderConstruction, redeclares all five writable,
// and checker.ts names it only where a slot is written outside ResolvedType: createObjectType's local, which its own
// five writes go through, while createObjectType still returns ObjectType; the declarations of the two callers that
// write the slots of the type createObjectType gives them, getOrCreateTypeFromSignature and getTypeWithoutSignatures;
// and a type assertion at three writes through a narrower interface, resolveAnonymousTypeMembers' two, of
// callSignatures and constructSignatures after setStructuredTypeMembers, and resolveBaseTypesOfClass' reset of an
// InterfaceType's members to undefined, the hole Adamic named, which is refused at its own assertion rather than
// hidden. Returning the view from createObjectType made every caller's downcast a cast from the view, eight sites the
// meter refused (J-7zaedjy). A redeclaration on a family interface would make one interface's slot writable while its
// siblings' stay readonly, and tsc's subtype reduction never takes a readonly property for a subtype of a writable one:
// a writable AnonymousType slot stopped getObjectTypeInstantiation's AnonymousType | DeferredTypeReference |
// GenericType from reducing (TS2345 at its instantiations.set), and a writable members on InterfaceType and GenericType
// made an ObjectType seen as GenericType a readonly slot become writable. checker.ts imports the new name with its own
// `import type` from ./types.js, after its last import, and leaves the import from ./_namespaces/ts.js byte for byte as
// it was: later adaptations (99-correctness-core-multimap among them) pin that declaration's whole text and find it as
// the one import of that module. Every edit is a declaration, a cast's type, a type assertion or a type-only import, so
// the emitted JavaScript doesn't change. All five slots are internal and the new interface is internal, so the
// published declarations don't change either. README.md says what certifies it, the writes counted, and the mutant that
// brings the refusals back. The edits are found through the stock AST, never by a line: each refuses unless it finds
// exactly upstream's declarations or statements, and a second run changes nothing. It sorts after 76-truthful-casts, so
// every later adaptation and every consumer reads apply's tree with the slots readonly. 75-optional-widening and
// 76-truthful-casts edit none of these interfaces or functions, so the match expects upstream's shape.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);

// The five resolved slots and their types, as upstream 6.0.3 declares them on ResolvedType (required). ObjectType
// declares them optional, and 20-optional-declarations, which runs first, spells each optional one `T | undefined`, so
// that is the text this adaptation expects there and writes on every optional slot it adds. A slot whose type has moved
// is not the slot Adamic ruled on.
const slotTypes = {
    members: 'SymbolTable',
    properties: 'Symbol[]',
    callSignatures: 'readonly Signature[]',
    constructSignatures: 'readonly Signature[]',
    indexInfos: 'readonly IndexInfo[]',
};
// The slots ObjectType keeps writable: none.
const objectTypeWritable = [];
// The family's other interfaces, each with its heritage as upstream writes it. None declares any of the five slots, and
// none gets one here: a slot writable on one subtype of ObjectType and readonly on its siblings changes which of them
// tsc's subtype reduction keeps.
const untouched = {
    InterfaceType: 'extends ObjectType',
    TypeReference: 'extends ObjectType',
    GenericType: 'extends InterfaceType, TypeReference',
    AnonymousType: 'extends ObjectType',
};
// The callers that write the slots of the type createObjectType gives them: in the one function of each name, with that
// parameter list, the one declaration of that text, upstream's or with the cast already at its initializer's end.
const casted = [
    { function: 'getOrCreateTypeFromSignature', parameters: 'signature: Signature',
        declaration: 'const type = createObjectType(ObjectFlags.Anonymous | ObjectFlags.SingleSignatureType, signature.declaration?.symbol)' },
    { function: 'getTypeWithoutSignatures', parameters: 'type: Type',
        declaration: 'const result = createObjectType(ObjectFlags.Anonymous, type.symbol)' },
];
// The writes through a narrower interface that go through a type assertion to the view instead: in the one function of
// each name, with that parameter list, each statement once, upstream's text or already adapted.
const asserted = [
    { function: 'resolveAnonymousTypeMembers', parameters: 'type: AnonymousType',
        statements: [['callSignatures', 'getSignaturesOfSymbol(symbol)'], ['constructSignatures', 'constructSignatures']] },
    { function: 'resolveBaseTypesOfClass', parameters: 'type: InterfaceType', statements: [['members', 'undefined']] },
];
const view = 'ObjectTypeUnderConstruction';
const optionalType = slot => `${slotTypes[slot]} | undefined`;

function parse(relative) {
    const fileName = path.join(tree, relative);
    const text = fs.readFileSync(fileName, 'utf8');
    const newline = text.includes('\r\n') ? '\r\n' : '\n';
    return { relative, fileName, text, newline, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true) };
}

function where(file, node) {
    const at = file.source.getLineAndCharacterOfPosition(node.getStart(file.source));
    return `${file.relative}:${at.line + 1}:${at.character + 1}`;
}

// write applies insertions and replacements, last first, and refuses if the file moved underneath.
function write(file, edits) {
    let adapted = file.text;
    for (const edit of [...edits].sort((a, b) => b.start - a.start))
        adapted = adapted.slice(0, edit.start) + edit.text + adapted.slice(edit.end);
    if (adapted === file.text) return;
    if (fs.readFileSync(file.fileName, 'utf8') !== file.text) throw new Error(`source changed during adaptation: ${file.fileName}`);
    fs.writeFileSync(file.fileName, adapted);
}

function isReadonly(member) {
    return (ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly) !== 0;
}

// slotMembers is an interface's property signatures named for one of the five slots, by name, refusing a duplicate.
function slotMembers(file, declaration) {
    const found = new Map();
    for (const member of declaration.members) {
        if (!member.name || !ts.isIdentifier(member.name) || !Object.hasOwn(slotTypes, member.name.text)) continue;
        if (!ts.isPropertySignature(member) || found.has(member.name.text))
            throw new Error(`unexpected slot ${declaration.name.text}.${member.name.text}: ${member.getText(file.source)}`);
        found.set(member.name.text, member);
    }
    return found;
}

function heritage(file, declaration) {
    return (declaration.heritageClauses ?? []).map(clause => clause.getText(file.source)).join(' ');
}

// types.ts.
const types = parse('src/compiler/types.ts');
const interfaces = new Map();
for (const statement of types.source.statements) {
    if (!ts.isInterfaceDeclaration(statement)) continue;
    if (!interfaces.has(statement.name.text)) interfaces.set(statement.name.text, []);
    interfaces.get(statement.name.text).push(statement);
}
function only(name) {
    const found = interfaces.get(name) ?? [];
    if (found.length !== 1) throw new Error(`expected one interface ${name} in types.ts, found ${found.length}`);
    return found[0];
}
const typeEdits = [];
const slots = [];

// ObjectType: readonly before each slot.
const objectType = only('ObjectType');
if (heritage(types, objectType) !== 'extends Type') throw new Error(`unexpected heritage of ObjectType: ${heritage(types, objectType)}`);
const objectMembers = slotMembers(types, objectType);
for (const [slot, type] of Object.entries(slotTypes)) {
    const member = objectMembers.get(slot);
    if (!member || !member.type || member.type.getText(types.source) !== optionalType(slot) || !member.questionToken)
        throw new Error(`expected ObjectType.${slot}?: ${optionalType(slot)}, found ${member ? member.getText(types.source) : 'none'}`);
    if (objectTypeWritable.includes(slot)) {
        if (isReadonly(member)) throw new Error(`ObjectType.${slot} is readonly, but this adaptation keeps it writable`);
        slots.push({ slot: `ObjectType.${slot}`, where: where(types, member), readonly: false });
        continue;
    }
    // Idempotent: a slot already readonly keeps its bytes, so a second run changes nothing.
    slots.push({ slot: `ObjectType.${slot}`, where: where(types, member), readonly: true, edited: !isReadonly(member) });
    if (!isReadonly(member)) typeEdits.push({ start: member.getStart(types.source), end: member.getStart(types.source), text: 'readonly ' });
}

// ResolvedType redeclares all five, required and writable, as upstream declares them, and setStructuredTypeMembers
// writes through it.
const resolvedType = only('ResolvedType');
if (heritage(types, resolvedType) !== 'extends ObjectType, UnionOrIntersectionType')
    throw new Error(`unexpected heritage of ResolvedType: ${heritage(types, resolvedType)}`);
const resolvedMembers = slotMembers(types, resolvedType);
for (const [slot, type] of Object.entries(slotTypes)) {
    const member = resolvedMembers.get(slot);
    if (!member || !member.type || member.type.getText(types.source) !== type || member.questionToken || isReadonly(member))
        throw new Error(`expected ResolvedType.${slot}: ${type}, writable, found ${member ? member.getText(types.source) : 'none'}`);
    slots.push({ slot: `ResolvedType.${slot}`, where: where(types, member), readonly: false });
}

// The rest of the family declares none of the five, and keeps declaring none.
for (const [name, expected] of Object.entries(untouched)) {
    const declaration = only(name);
    if (heritage(types, declaration) !== expected || slotMembers(types, declaration).size !== 0)
        throw new Error(`expected ${name} to be declared ${expected} and to declare none of the five slots`);
}

// ObjectTypeUnderConstruction, right after ObjectType: createObjectType's view of the type it makes.
const viewText = [
    '', '',
    '/** @internal */',
    '// What createObjectType makes: an ObjectType whose resolved slots are still being written',
    `export interface ${view} extends ObjectType {`,
    ...Object.keys(slotTypes).map(slot => `    ${slot}?: ${optionalType(slot)};`),
    '}',
].join(types.newline);
const views = interfaces.get(view) ?? [];
const viewAt = objectType.getEnd();
if (views.length === 0) typeEdits.push({ start: viewAt, end: viewAt, text: viewText });
else if (views.length !== 1 || types.text.slice(viewAt, viewAt + viewText.length) !== viewText || views[0].getEnd() !== viewAt + viewText.length)
    throw new Error(`expected ${view} once in types.ts, right after ObjectType and as this adaptation writes it`);
slots.push({ slot: `${view}: ${Object.keys(slotTypes).join(', ')}`, where: views.length ? where(types, views[0]) : where(types, objectType),
    readonly: false, edited: views.length === 0 });

// checker.ts. functionNamed is the one function of a name, with exactly that parameter list, or a refusal.
const checker = parse('src/compiler/checker.ts');
function functionNamed(name, parameters) {
    const found = [];
    ts.forEachChild(checker.source, function find(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === name) found.push(node);
        ts.forEachChild(node, find);
    });
    if (found.length !== 1 || !found[0].body ||
        found[0].parameters.map(parameter => parameter.getText(checker.source)).join(', ') !== parameters)
        throw new Error(`expected one ${name}(${parameters}) with a body in checker.ts, found ${found.length}`);
    return found[0];
}
const checkerEdits = [];

// createObjectType keeps upstream's return annotation, ObjectType, and its local takes the view, so its own five writes
// go through it and `return type;` is an upcast. Upstream's
//     const type = createTypeWithSymbol(TypeFlags.Object, symbol!) as ObjectType;
// becomes
//     const type = createTypeWithSymbol(TypeFlags.Object, symbol!) as ObjectTypeUnderConstruction;
const creator = functionNamed('createObjectType', 'objectFlags: ObjectFlags, symbol?: Symbol');
if (creator.type?.getText(checker.source) !== 'ObjectType')
    throw new Error(`expected createObjectType to return ObjectType, found ${creator.type?.getText(checker.source)}`);
const creatorFirst = creator.body.statements[0];
const cast = creatorFirst && ts.isVariableStatement(creatorFirst) ? creatorFirst.declarationList.declarations[0]?.initializer : undefined;
const castType = cast && ts.isAsExpression(cast) ? cast.type.getText(checker.source) : undefined;
if (castType !== 'ObjectType' && castType !== view) throw new Error(`unexpected first statement of createObjectType: ${creatorFirst?.getText(checker.source)}`);
const expectedStatements = [
    `const type = createTypeWithSymbol(TypeFlags.Object, symbol!) as ${castType};`,
    'type.objectFlags = objectFlags;',
    ...Object.keys(slotTypes).map(slot => `type.${slot} = undefined;`),
    'return type;',
];
if (creator.body.statements.map(statement => statement.getText(checker.source)).join('\n') !== expectedStatements.join('\n'))
    throw new Error(`unexpected createObjectType: ${creator.getText(checker.source)}`);
if (castType === 'ObjectType') checkerEdits.push({ start: cast.type.getStart(checker.source), end: cast.type.getEnd(), text: view });

// The two writing callers: the declaration's initializer gains ` as ObjectTypeUnderConstruction` at its end, which the
// emitter erases, and nothing else in them changes. Upstream's
//     const type = createObjectType(ObjectFlags.Anonymous | ObjectFlags.SingleSignatureType, signature.declaration?.symbol);
//     const result = createObjectType(ObjectFlags.Anonymous, type.symbol);
// become
//     const type = createObjectType(ObjectFlags.Anonymous | ObjectFlags.SingleSignatureType, signature.declaration?.symbol) as ObjectTypeUnderConstruction;
//     const result = createObjectType(ObjectFlags.Anonymous, type.symbol) as ObjectTypeUnderConstruction;
const castedCallers = [];
for (const caller of casted) {
    const owner = functionNamed(caller.function, caller.parameters);
    const upstream = `${caller.declaration};`;
    const adapted = `${caller.declaration} as ${view};`;
    const found = [];
    ts.forEachChild(owner.body, function find(node) {
        if (ts.isVariableStatement(node) && [upstream, adapted].includes(node.getText(checker.source))) found.push(node);
        ts.forEachChild(node, find);
    });
    if (found.length !== 1) throw new Error(`expected \`${upstream}\` once in ${caller.function}, adapted or not, found ${found.length}`);
    const edited = found[0].getText(checker.source) === upstream;
    if (edited) {
        const initializer = found[0].declarationList.declarations[0].initializer;
        checkerEdits.push({ start: initializer.getEnd(), end: initializer.getEnd(), text: ` as ${view}` });
    }
    castedCallers.push({ function: caller.function, where: where(checker, found[0]), edited, statement: adapted });
}

// The asserted writes. Each receiver becomes a type assertion to the view, which the emitter erases, parentheses and
// all. resolveAnonymousTypeMembers' two, after setStructuredTypeMembers has resolved the type:
//     type.callSignatures = getSignaturesOfSymbol(symbol);
//     type.constructSignatures = constructSignatures;
// and resolveBaseTypesOfClass' reset, which really does make an InterfaceType's members undefined until they re-resolve:
//     type.members = undefined;
// become
//     (type as ObjectTypeUnderConstruction).callSignatures = getSignaturesOfSymbol(symbol);
//     (type as ObjectTypeUnderConstruction).constructSignatures = constructSignatures;
//     (type as ObjectTypeUnderConstruction).members = undefined;
const assertedWrites = [];
for (const owner of asserted) {
    const body = functionNamed(owner.function, owner.parameters).body;
    for (const [slot, value] of owner.statements) {
        const upstream = `type.${slot} = ${value};`;
        const adapted = `(type as ${view}).${slot} = ${value};`;
        const found = [];
        ts.forEachChild(body, function find(node) {
            if (ts.isExpressionStatement(node) && [upstream, adapted].includes(node.getText(checker.source))) found.push(node);
            ts.forEachChild(node, find);
        });
        if (found.length !== 1)
            throw new Error(`expected \`${upstream}\` once in ${owner.function}, adapted or not, found ${found.length}`);
        const statement = found[0];
        const edited = statement.getText(checker.source) === upstream;
        if (edited) {
            const receiver = statement.expression.left.expression;
            if (!ts.isIdentifier(receiver) || receiver.text !== 'type') throw new Error(`unexpected receiver in ${upstream}`);
            checkerEdits.push({ start: receiver.getStart(checker.source), end: receiver.getEnd(), text: `(type as ${view})` });
        }
        assertedWrites.push({ function: owner.function, slot, where: where(checker, statement), edited, statement: adapted });
    }
}

// The import: one `import type` of the view from ./types.js, which declares it, as the statement after checker.ts's last
// import. The one import from ./_namespaces/ts.js, which later adaptations pin whole, keeps its bytes; it must not
// name the view, and it must still be the only import of its module.
const importText = `import type { ${view} } from "./types.js";`;
const imports = checker.source.statements.filter(ts.isImportDeclaration);
const moduleOf = declaration => ts.isStringLiteral(declaration.moduleSpecifier) ? declaration.moduleSpecifier.text : undefined;
if (imports.filter(declaration => moduleOf(declaration) === './_namespaces/ts.js').length !== 1)
    throw new Error('expected checker.ts to import from ./_namespaces/ts.js once');
const own = imports.filter(declaration => moduleOf(declaration) === './types.js');
const naming = imports.filter(declaration => {
    const bindings = declaration.importClause?.namedBindings;
    return bindings && ts.isNamedImports(bindings) && bindings.elements.some(element => element.name.text === view);
});
let imported;
if (own.length === 0 && naming.length === 0) {
    const last = imports.at(-1);
    if (!last) throw new Error('expected checker.ts to import something');
    checkerEdits.push({ start: last.getEnd(), end: last.getEnd(), text: `${checker.newline}${importText}` });
    imported = { where: where(checker, last), edited: true };
}
else if (own.length === 1 && naming.length === 1 && naming[0] === own[0] && own[0].getText(checker.source) === importText && imports.at(-1) === own[0])
    imported = { where: where(checker, own[0]), edited: false };
else throw new Error(`expected checker.ts to import ${view} once, as this adaptation writes it, after its other imports`);

// Every check passes before either file is written.
write(types, typeEdits);
write(checker, checkerEdits);
console.log(JSON.stringify({ typescript: ts.version, task: '#5pstmvw', certificate: "the compiled program's gate", slots,
    createObjectType: { where: where(checker, creator), edited: castType === 'ObjectType', local: view, returns: 'ObjectType' },
    castedCallers, assertedWrites,
    import: { ...imported, statement: importText } }, null, 2));
