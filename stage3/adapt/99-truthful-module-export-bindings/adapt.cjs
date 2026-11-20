#!/usr/bin/env node
'use strict';
// Adaptation 99-truthful-module-export-bindings, #k8ed1f2, family 3's cause 10. Family 3 is the refusal "a value of
// type A seen as B, which can write X where Y is read". ExternalModuleInfo.exportedBindings is declared Identifier[][],
// but collectExternalModuleInfo stores ModuleExportNames in it: an `export * as` name, an export specifier's exported
// name, and the name an export specifier gives a function, any of which can be a StringLiteral. tsc's own code sees the
// array as ModuleExportName[][] where it adds. The declaration is made true (road 4): exportedBindings, the local that
// fills it and collectExportedVariableInfo's parameter become ModuleExportName[][], and the readers follow where tsc's
// own types demand it, module.ts getExports' bindingsSet and its arrayFrom type argument and system.ts getExports'
// exportedNames, so the cause's 7 sites clear. Adamic delegated the ruling on #k8ed1f2. README.md says why each slot,
// every reader followed, and every write into the array. Type-only: each edit is an annotation or a type argument, so
// the emitted JavaScript does not change. No declaration is public: ExternalModuleInfo is `@internal`,
// collectExportedVariableInfo is file-local and the rest are local to a function, so the published declaration files
// do not change and no api-sanctions.json is needed. Through the stock AST each slot is found by name, never by line,
// offset or whole-line text, and the edit replaces one type node's span. It refuses before any write unless every slot
// is found exactly once, holds the text it expects or exactly its adapted text, and is local, file-local or
// `@internal`, and a second run changes nothing. It sorts after every adaptation that edits these three files, the last
// of them 99-truthful-generated-name-emit-node, and after 99-correctness-receiver-transformers, which pins
// transformModule's and transformSystemModule's whole text, where both getExports sit, so each pin reads its
// declarations as it expects.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const utilities = 'src/compiler/transformers/utilities.ts';
const moduleTransform = 'src/compiler/transformers/module/module.ts';
const systemTransform = 'src/compiler/transformers/module/system.ts';
const bindings = { expected: 'Identifier[][]', adapted: 'ModuleExportName[][]' };
const declaredFlags = { const: ts.NodeFlags.Const, let: ts.NodeFlags.Let };
// Each slot, addressed by name. A slot is one of:
// - an interface's property, whose type changes (`interface`, `property`);
// - a function's named parameter's type (`function`, `parameter`);
// - a binding declared `const` or `let` in the function declaration nearest it (`within`, `variable`, `declared`), its
//   initializer starting `starts`, or `starts: null` where it has none; with `construct`, its initializer is a `new` of
//   that name with no arguments, and the slot is that expression's one type argument;
// - a call in the function declaration nearest it (`within`, `call`), by its callee and its first argument's text,
//   whose one type argument is the slot.
// `expected` is the slot's text as the adaptations before this one leave it, which is upstream 6.0.3's except for
// arrayFrom's type argument: upstream writes none, and 33-indexed-reads-emit adds `<Identifier>`. `adapted` is the text
// this one leaves. Any other text is not the slot the ruling covers.
const slots = [
    { file: utilities, interface: 'ExternalModuleInfo', property: 'exportedBindings', ...bindings },
    { file: utilities, within: 'collectExternalModuleInfo', variable: 'exportedBindings', declared: 'const',
        starts: '[]', ...bindings },
    { file: utilities, function: 'collectExportedVariableInfo', parameter: 'exportedBindings', ...bindings },
    // module.ts's getExports gathers the bindings in a set and returns it as an array: both widen with the bindings.
    { file: moduleTransform, within: 'getExports', variable: 'bindingsSet', declared: 'const', construct: 'Set',
        expected: 'Identifier', adapted: 'ModuleExportName' },
    { file: moduleTransform, within: 'getExports', call: 'arrayFrom', argument: 'bindingsSet',
        expected: 'Identifier', adapted: 'ModuleExportName' },
    // system.ts's getExports declares two exportedNames: this `let`, which addRange fills from the bindings, and a
    // `const` in its other branch that upstream already types ModuleExportName[].
    { file: systemTransform, within: 'getExports', variable: 'exportedNames', declared: 'let', starts: null,
        expected: 'Identifier[] | undefined', adapted: 'ModuleExportName[] | undefined' },
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
const name = slot => {
    if (slot.call) return `${slot.within} ${slot.call}(${slot.argument}) type argument`;
    if (slot.construct) return `${slot.within} ${slot.declared} ${slot.variable}'s new ${slot.construct} type argument`;
    if (slot.variable) return `${slot.within} ${slot.declared} ${slot.variable}`;
    if (slot.property) return `${slot.interface}.${slot.property}`;
    return `${slot.function} ${slot.parameter}`;
};
const functionLike = node => ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node) || ts.isArrowFunction(node)
    || ts.isMethodDeclaration(node) || ts.isConstructorDeclaration(node) || ts.isGetAccessorDeclaration(node)
    || ts.isSetAccessorDeclaration(node);
const enclosingFunction = node => {
    let outer = node.parent;
    while (outer && !functionLike(outer)) outer = outer.parent;
    return outer;
};
// The function declaration nearest a node, by name.
function inside(node, within) {
    const outer = enclosingFunction(node);
    return outer !== undefined && ts.isFunctionDeclaration(outer) && outer.name?.text === within;
}
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
    const statement = ts.isVariableDeclaration(declaration) ? declaration.parent.parent : declaration;
    if (statement.parent === source && ts.isExternalModule(source) && !exported(statement)) return 'file-local';
    throw new Error(`expected ${what} to be local to a function, file-local or @internal`);
}
// The slot's type node, and the declaration that owns it, for the visibility check.
function find(source, slot) {
    const what = `${name(slot)} in ${slot.file}`;
    if (slot.call) {
        const call = one(nodes(source, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression)
            && node.expression.text === slot.call && node.arguments.length > 0
            && node.arguments[0].getText(source) === slot.argument), `call ${what}`);
        if (!inside(call, slot.within)) throw new Error(`expected ${what} to sit in ${slot.within}`);
        // No type argument means 33-indexed-reads-emit has not run first: the slot refuses rather than add one here.
        if (call.typeArguments?.length !== 1)
            throw new Error(`expected ${what} to carry one type argument: ${call.getText(source)}`);
        return { call, type: call.typeArguments[0] };
    }
    if (slot.variable) {
        const declaration = one(nodes(source, node => ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)
            && node.name.text === slot.variable && inside(node, slot.within)
            && ts.isVariableDeclarationList(node.parent)
            && (node.parent.flags & ts.NodeFlags.BlockScoped) === declaredFlags[slot.declared]),
            `${slot.declared} binding ${what}`);
        const initializer = declaration.initializer;
        if (slot.construct) {
            // The binding's type is inferred from the `new`, so the type argument is the slot and no annotation may
            // sit beside it.
            if (declaration.type || !initializer || !ts.isNewExpression(initializer)
                || !ts.isIdentifier(initializer.expression) || initializer.expression.text !== slot.construct
                || (initializer.arguments?.length ?? 0) !== 0 || initializer.typeArguments?.length !== 1)
                throw new Error(`expected ${what} to be new ${slot.construct}<T>(): ${declaration.getText(source)}`);
            return { owner: declaration, type: initializer.typeArguments[0] };
        }
        const expected = slot.starts === null ? 'to have no initializer' : `to start ${slot.starts}`;
        if (slot.starts === null ? initializer !== undefined : !initializer?.getText(source).startsWith(slot.starts))
            throw new Error(`expected ${what} ${expected}: ${declaration.getText(source).split('\n')[0]}`);
        return { owner: declaration, type: declaration.type };
    }
    if (slot.property) {
        const container = one(nodes(source, node => ts.isInterfaceDeclaration(node)
            && node.name.text === slot.interface), `interface ${slot.interface} in ${slot.file}`);
        // One of the interface's own members. A call or index signature has no name, so the name is checked before
        // it is read.
        const member = one(container.members.filter(member => ts.isPropertySignature(member) && member.name
            && ts.isIdentifier(member.name) && member.name.text === slot.property), `property ${what}`);
        // A property whose type changes keeps no modifier and stays required, as upstream declares it.
        if (member.modifiers?.length || member.questionToken)
            throw new Error(`unexpected slot ${what}: ${member.getText(source)}`);
        return { owner: member, type: member.type };
    }
    const declaration = one(nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === slot.function),
        `function ${slot.function} in ${slot.file}`);
    const parameter = one(declaration.parameters.filter(parameter => ts.isIdentifier(parameter.name)
        && parameter.name.text === slot.parameter), `parameter ${what}`);
    return { owner: declaration, type: parameter.type };
}
// A slot whose type node is replaced whole: the expected text is edited, the adapted text is left as it is.
function retype(source, type, slot, record) {
    if (!type) throw new Error(`expected ${record.slot} at ${record.where} to declare its type`);
    const text = type.getText(source);
    if (text === slot.adapted) return { ...record, edited: false };
    if (text !== slot.expected) throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${text}`);
    return { ...record, edited: true, at: type.getStart(source), end: type.end, text: slot.adapted };
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
    const found = find(source, slot);
    // Where the slot is reported: its type node, or else the owning declaration, so a slot missing its type is named
    // before it refuses.
    const anchor = found.type ?? found.owner;
    const at = source.getLineAndCharacterOfPosition(anchor.getStart(source));
    const record = { slot: name(slot), where: `${slot.file}:${at.line + 1}:${at.character + 1}`, file: slot.file };
    // Found exactly once across the table too: two entries naming one node would be one slot read twice.
    const key = `${slot.file}:${anchor.getStart(source)}`;
    if (addressed.has(key)) throw new Error(`slot ${record.slot} at ${record.where} is already in the table`);
    addressed.add(key);
    // A call changes no declaration; its type argument erases.
    record.visibility = found.call ? `a call in ${slot.within}` : visibility(source, found.owner, record.slot);
    records.push(retype(source, found.type, slot, record));
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
