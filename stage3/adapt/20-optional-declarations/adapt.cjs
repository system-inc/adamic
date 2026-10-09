#!/usr/bin/env node
'use strict';
// Tooling only: rejected programs never reach Adamic lowering.
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
const tree = path.resolve(process.argv[2] || '');
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const ts = process.env.TSC_ADAPT_TYPESCRIPT
    ? require(process.env.TSC_ADAPT_TYPESCRIPT)
    : (() => {
        try { return require('typescript'); }
        catch (error) {
            if (error.code !== 'MODULE_NOT_FOUND') throw error;
            return createRequire(path.join(tree, 'package.json'))('typescript');
        }
    })();
if (ts.version !== '6.0.3') throw new Error(`expected TypeScript 6.0.3, got ${ts.version}`);
const directory = path.join(tree, 'src/compiler');
const roots = ts.sys.readDirectory(directory, ['.ts']);
const owned = new Set(roots.map(name => path.resolve(name)));
const options = {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    erasableSyntaxOnly: false,
    verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
};
const codes = new Set([2412, 2375, 2379]);
const forbidden = ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess;
const parts = type => type.isUnion() ? type.types : [type];
const undefinedPart = type => !!(type.flags & ts.TypeFlags.Undefined);
function at(node, position) {
    let result = node;
    ts.forEachChild(node, child => {
        if (child.pos <= position && position < child.end) result = at(child, position);
    });
    return result;
}
function indexed(node) {
    if (ts.isElementAccessExpression(node)) return true;
    let found = false;
    ts.forEachChild(node, child => { if (indexed(child)) found = true; });
    return found;
}
function unproven(checker, node) {
    if (!indexed(node)) return false;
    return !(ts.isConditionalExpression(node) &&
        [node.whenTrue, node.whenFalse].some(branch => undefinedPart(checker.getTypeAtLocation(branch))));
}
function record(diagnostic) {
    const file = diagnostic.file;
    const point = file && file.getLineAndCharacterOfPosition(diagnostic.start || 0);
    return { file: file && path.relative(tree, file.fileName), line: point && point.line + 1,
        column: point && point.character + 1, code: diagnostic.code,
        message: ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'),
        // Stable across type insertions, including columns and nested owners.
        source: file && file.text.slice(diagnostic.start, diagnostic.start + diagnostic.length) };
}
const identity = row => JSON.stringify([row.file, row.line, row.code, row.message, row.source]);
function ownerPath(owner) {
    const result = [];
    for (let node = owner; node; node = node.parent) {
        if (node.name && !(ts.isModuleDeclaration(node) && node.name.getText(owner.getSourceFile()) === 'ts')) result.unshift(node.name.getText(owner.getSourceFile()));
    }
    return result;
}
const declarations = [];
const resolvedFindings = [];
let initiallyDeclined;
let before;
let final;
let declined;
const declinedOwners = new Map();
for (;;) {
    const program = ts.createProgram(roots, options);
    const checker = program.getTypeChecker();
    const diagnostics = ts.getPreEmitDiagnostics(program);
    process.stderr.write(`recheck: ${diagnostics.length} diagnostics, ${declarations.length} declarations edited\n`);
    if (!before) before = diagnostics.map(record);
    const selected = new Set();
    const evidence = new Map();
    const excluded = [];
    function select(symbol, value, finding) {
        if (!symbol || !(symbol.flags & ts.SymbolFlags.Optional) || !value || !symbol.declarations?.length) return false;
        const values = parts(value);
        if (!values.some(undefinedPart)) return;
        const owners = symbol.declarations;
        for (const owner of owners) {
            if (!(ts.isPropertySignature(owner) || ts.isPropertyDeclaration(owner) || ts.isMethodSignature(owner)) ||
                !owner.questionToken || !owner.type || !owned.has(path.resolve(owner.getSourceFile().fileName))) return;
            if (ts.isMethodSignature(owner)) {
                const file = owner.getSourceFile();
                const point = file.getLineAndCharacterOfPosition(owner.getStart(file));
                declinedOwners.set(`${file.fileName}:${owner.pos}`, { file: path.relative(tree, file.fileName),
                    line: point.line + 1, name: owner.name.getText(file),
                    reason: 'optional method syntax cannot add a callable undefined union without changing method variance and public declaration shape' });
                return;
            }
            const target = ts.isMethodSignature(owner)
                ? checker.getNonNullableType(checker.getTypeOfSymbolAtLocation(symbol, owner))
                : checker.getTypeFromTypeNode(owner.type);
            if ((target.flags & forbidden) || checker.isTypeAssignableTo(checker.getUndefinedType(), target)) return;
            if (values.some(part => !undefinedPart(part) &&
                ((part.flags & forbidden) || !checker.isTypeAssignableTo(part, target)))) return;
        }
        for (const owner of owners) { selected.add(owner); evidence.set(owner, finding); }
        resolvedFindings.push({ ...finding, owners: owners.map(owner => {
            const file = owner.getSourceFile();
            const point = file.getLineAndCharacterOfPosition(owner.getStart(file));
            return { file: path.relative(tree, file.fileName), line: point.line + 1,
                column: point.character + 1, name: owner.name.getText(file) };
        }) });
        return true;
    }
    for (const diagnostic of diagnostics.filter(d => codes.has(d.code))) {
        const finding = record(diagnostic);
        let reason = 'expanded chain has no eligible owned optional declaration with compatible present-undefined evidence';
        if (!diagnostic.file) { excluded.push({ ...finding, reason }); continue; }
        if (diagnostic.code !== 2412 && !finding.message.includes("Type 'undefined' is not assignable to type")) {
            excluded.push({ ...finding, reason: 'expanded chain is not a present-undefined optional value mismatch' });
            continue;
        }
        if (finding.message.includes('could be instantiated')) {
            excluded.push({ ...finding, reason: 'arbitrary generic specialization' }); continue;
        }
        let resolved = false;
        const node = at(diagnostic.file, diagnostic.start);
        if (diagnostic.code === 2412) {
            for (let n = node; n; n = n.parent) {
                if (!ts.isBinaryExpression(n)) continue;
                if (n.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isPropertyAccessExpression(n.left)) {
                    const receiver = checker.getTypeAtLocation(n.left.expression);
                    if (receiver.flags & forbidden) reason = 'generic or unresolved receiver contract';
                    else if (unproven(checker, n.right)) reason = 'unproven indexed read, not optionality evidence';
                    else resolved = select(checker.getSymbolAtLocation(n.left.name), checker.getTypeAtLocation(n.right), finding) || resolved;
                }
                break;
            }
        } else {
            for (let n = node; n; n = n.parent) {
                let expression = n;
                if (ts.isReturnStatement(n)) expression = n.expression;
                if (ts.isVariableDeclaration(n)) expression = n.initializer;
                if (ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.EqualsToken) expression = n.right;
                if (!expression || !ts.isExpressionNode(expression)) continue;
                const target = checker.getContextualType(expression);
                if (!target || (target.flags & forbidden)) continue;
                const properties = checker.getPropertiesOfType(target);
                if (!properties.length) continue;
                const source = checker.getTypeAtLocation(expression);
                for (const property of properties) {
                    const actual = checker.getPropertyOfType(source, property.name);
                    if (!actual) continue;
                    let value = checker.getTypeOfSymbolAtLocation(actual, expression);
                    const declaration = actual.declarations?.[0];
                    if (declaration && ts.isPropertyAssignment(declaration) && unproven(checker, declaration.initializer)) continue;
                    if (actual.flags & ts.SymbolFlags.Optional) {
                        if (actual.declarations?.length !== 1 || !declaration.type) continue;
                        value = checker.getTypeFromTypeNode(declaration.type);
                    }
                    resolved = select(property, value, finding) || resolved;
                }
                break;
            }
        }
        if (!resolved) excluded.push({ ...finding, reason });
    }
    // These two builder assignments already supply present-undefined evidence.
    // An earlier incompatible member can hide their optional diagnostic until
    // indexed-read adaptations run. Resolve the reviewed fields individually
    // through the same selector, rather than depending on diagnostic ordering.
    const builder = program.getSourceFile(path.join(directory, 'builder.ts'));
    if (builder) {
        let assignments = 0;
        function visitBuilder(node) {
            if (ts.isFunctionDeclaration(node) && node.name?.text === 'createBuilderProgramUsingIncrementalBuildInfo') {
                function visitAssignment(n) {
                    if (ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
                        ts.isIdentifier(n.left) && n.left.text === 'state' && ts.isObjectLiteralExpression(n.right)) {
                        assignments++;
                        const target = checker.getTypeAtLocation(n.left);
                        for (const name of ['outSignature', 'hasErrors', 'emitSignatures']) {
                            const field = n.right.properties.find(p => ts.isPropertyAssignment(p) && p.name.getText(builder) === name);
                            if (!field) continue;
                            const property = checker.getPropertyOfType(target, name);
                            const owners = property?.declarations || [];
                            if (owners.length !== 1 || !ts.isInterfaceDeclaration(owners[0].parent) ||
                                owners[0].parent.name.text !== 'ReusableBuilderProgramState') throw new Error('builder optional owner drift: ' + name);
                            if (unproven(checker, field.initializer)) throw new Error('builder optional evidence became indexed: ' + name);
                            const point = builder.getLineAndCharacterOfPosition(field.getStart(builder));
                            const finding = { file: 'src/compiler/builder.ts', line: point.line + 1,
                                column: point.character + 1, code: null, source: field.getText(builder),
                                resolution: 'reviewed present-undefined assignment, independent of first diagnostic member' };
                            const value = checker.getTypeAtLocation(field.initializer);
                            const annotation = checker.getTypeFromTypeNode(owners[0].type);
                            if (!parts(value).some(undefinedPart) || parts(value).some(part => !undefinedPart(part) &&
                                ((part.flags & forbidden) || !checker.isTypeAssignableTo(part, annotation)))) {
                                throw new Error('builder optional value contract drift: ' + name);
                            }
                            if (!checker.isTypeAssignableTo(checker.getUndefinedType(), annotation) && !select(property, value, finding)) {
                                throw new Error('builder optional owner is ineligible: ' + name);
                            }
                        }
                    }
                    ts.forEachChild(n, visitAssignment);
                }
                visitAssignment(node.body);
            } else ts.forEachChild(node, visitBuilder);
        }
        visitBuilder(builder);
        if (assignments !== 2) throw new Error('builder optional assignment count drift: ' + assignments);
    }
    // One object slot can be viewed through several inherited declarations.
    // Its proven present-undefined write must be truthful in every optional view.
    const interfaces = [];
    for (const file of program.getSourceFiles()) {
        if (!owned.has(path.resolve(file.fileName))) continue;
        function visit(node) {
            if (ts.isInterfaceDeclaration(node)) interfaces.push(node);
            ts.forEachChild(node, visit);
        }
        visit(file);
    }
    const bases = interfaces.map(node => checker.getBaseTypes(checker.getTypeAtLocation(node)) || []);
    for (const owner of selected) {
        if (!ts.isInterfaceDeclaration(owner.parent)) continue;
        const name = owner.name.getText(owner.getSourceFile());
        const finding = { ...evidence.get(owner), resolution: 'same present-undefined slot through an inherited optional declaration', from: ownerPath(owner) };
        function inherit(property) {
            if (property?.declarations?.every(declaration => selected.has(declaration))) return;
            select(property, checker.getUndefinedType(), finding);
        }
        const parents = checker.getBaseTypes(checker.getTypeAtLocation(owner.parent)) || [];
        for (const base of parents) inherit(checker.getPropertyOfType(base, name));
        for (const group of bases) {
            const properties = group.map(base => checker.getPropertyOfType(base, name));
            if (properties.some(property => property?.declarations?.includes(owner))) {
                for (const property of properties) inherit(property);
            }
        }
    }
    if (!initiallyDeclined) initiallyDeclined = excluded;
    if (!selected.size) { final = diagnostics.map(record); declined = excluded; break; }
    const edits = new Map();
    for (const owner of selected) {
        const file = owner.getSourceFile();
        const point = file.getLineAndCharacterOfPosition(owner.getStart(file));
        declarations.push({ file: path.relative(tree, file.fileName), line: point.line + 1,
            column: point.character + 1, name: owner.name.getText(file), owner_path: ownerPath(owner), before: owner.type.getText(file) });
        if (!edits.has(file)) edits.set(file, []);
        const wrap = ts.isFunctionTypeNode(owner.type) || ts.isConstructorTypeNode(owner.type) || ts.isConditionalTypeNode(owner.type);
        if (wrap) edits.get(file).push({ position: owner.type.getStart(file), text: '(', owner: owner.pos });
        edits.get(file).push({ position: owner.type.end, text: (wrap ? ')' : '') + ' | undefined', owner: owner.pos });
    }
    for (const [file, insertions] of edits) {
        insertions.sort((a, b) => b.position - a.position || a.owner - b.owner);
        let text = file.text;
        for (const insertion of insertions) text = text.slice(0, insertion.position) + insertion.text + text.slice(insertion.end ?? insertion.position);
        fs.writeFileSync(file.fileName, text);
    }
}
const prior = new Set(before.map(identity));
const counts = rows => {
    const result = {};
    for (const row of rows) result[row.code] = (result[row.code] || 0) + 1;
    return result;
};
process.stdout.write(JSON.stringify({ typescript: ts.version, declarations_changed: declarations.length,
    declarations, resolved_findings: resolvedFindings, initially_declined: initiallyDeclined, before: counts(before), after: counts(final),
    newly_exposed_consumers: final.filter(row => !prior.has(identity(row))), declined, declined_owners: [...declinedOwners.values()] }, null, 2) + '\n');
