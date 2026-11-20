#!/usr/bin/env node
'use strict';
// #ad161ms part 2, Adamic's ruling: a fork fix. isPropertyNameLiteral returns true for an Identifier, a StringLiteral, a
// NoSubstitutionTemplateLiteral and a NumericLiteral, and its claim said PropertyNameLiteral, which also holds a
// BigIntLiteral and a JsxNamespacedName. Three callers took the false side for a ComputedPropertyName and read
// `name.expression`, undefined for a bigint name, so stock tsc 6.0.3 throws on `class C { @dec 1n = 0; }`. The claim now
// names the three kinds, and the three callers give a bigint name its property key, the decimal digits a `[1n]` key's
// __propKey gives, as a string literal:
// - utilities.ts: isPropertyNameLiteral claims `Identifier | StringLiteralLike | NumericLiteral` (type-only);
// - esDecorators.ts: partialTransformClassElement's name chain and visitReferencedPropertyName gain a BigIntLiteral branch;
// - namedEvaluation.ts: getAssignedNameOfPropertyName gains the same branch.
// Both transformer files import isBigIntLiteral and parsePseudoBigInt, each in its list's own order. JavaScript-changing,
// only where a bigint name reaches these branches, which crashed before (the whole-oracle count is in the README). Each
// owner is found through the stock AST and pinned whole, upstream or adapted; a partly adapted tree is refused, and a
// second run writes nothing. `rewrite(tree, { callers: false })` is the mutant: the claim alone, the callers left.
// This directory sorts after 99-correctness-receiver-transformers, which pins transformESDecorators and esDecorators.ts's
// namespace import whole by hash and would refuse them edited; no directory after this one touches these owners.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');

const utilities = 'src/compiler/utilities.ts';
const esDecorators = 'src/compiler/transformers/esDecorators.ts';
const namedEvaluation = 'src/compiler/transformers/namedEvaluation.ts';
const namespaceModule = '../_namespaces/ts.js';
// The names each transformer file imports for its new branches, absent upstream and in every adapted tree before this.
const added = ['isBigIntLiteral', 'parsePseudoBigInt'];

const claimBefore = `export function isPropertyNameLiteral(node: Node): node is PropertyNameLiteral {
    switch (node.kind) {
        case SyntaxKind.Identifier:
        case SyntaxKind.StringLiteral:
        case SyntaxKind.NoSubstitutionTemplateLiteral:
        case SyntaxKind.NumericLiteral:
            return true;
        default:
            return false;
    }
}`;
const claimAfter = claimBefore.replace('node is PropertyNameLiteral', 'node is Identifier | StringLiteralLike | NumericLiteral');

const decoratedNameBefore = `if (isIdentifier(member.name) || isPrivateIdentifier(member.name)) {
                propertyName = { computed: false, name: member.name };
            }
            else if (isPropertyNameLiteral(member.name)) {
                propertyName = { computed: true, name: factory.createStringLiteralFromNode(member.name) };
            }
            else {
                const expression = member.name.expression;
                if (isPropertyNameLiteral(expression) && !isIdentifier(expression)) {
                    propertyName = { computed: true, name: factory.createStringLiteralFromNode(expression) };
                }
                else {
                    enterName();
                    ({ referencedName, name } = visitReferencedPropertyName(member.name));
                    propertyName = { computed: true, name: referencedName };
                    exitName();
                }
            }`;
const decoratedNameAfter = decoratedNameBefore.replace(`
            else {
                const expression`, `
            else if (isBigIntLiteral(member.name)) {
                propertyName = { computed: true, name: factory.createStringLiteral(parsePseudoBigInt(member.name.text)) };
            }
            else {
                const expression`);

const referencedNameBefore = `function visitReferencedPropertyName(node: PropertyName) {
        if (isPropertyNameLiteral(node) || isPrivateIdentifier(node)) {
            const referencedName = factory.createStringLiteralFromNode(node);
            const name = visitNode(node, visitor, isPropertyName);
            return { referencedName, name };
        }

        if (isPropertyNameLiteral(node.expression) && !isIdentifier(node.expression)) {
            const referencedName = factory.createStringLiteralFromNode(node.expression);
            const name = visitNode(node, visitor, isPropertyName);
            return { referencedName, name };
        }

        const referencedName = factory.getGeneratedNameForNode(node);
        hoistVariableDeclaration(referencedName);

        const key = emitHelpers().createPropKeyHelper(visitNode(node.expression, visitor, isExpression));
        const assignment = factory.createAssignment(referencedName, key);
        const name = factory.updateComputedPropertyName(node, injectPendingExpressions(assignment));
        return { referencedName, name };
    }`;
const referencedNameAfter = referencedNameBefore.replace(`

        if (isPropertyNameLiteral(node.expression)`, `

        if (isBigIntLiteral(node)) {
            const referencedName = factory.createStringLiteral(parsePseudoBigInt(node.text));
            return { referencedName, name: node };
        }

        if (isPropertyNameLiteral(node.expression)`);

const assignedNameBefore = `function getAssignedNameOfPropertyName(context: TransformationContext, name: PropertyName, assignedNameText: string | undefined) {
    const { factory } = context;
    if (assignedNameText !== undefined) {
        const assignedName = factory.createStringLiteral(assignedNameText);
        return { assignedName, name };
    }

    if (isPropertyNameLiteral(name) || isPrivateIdentifier(name)) {
        const assignedName = factory.createStringLiteralFromNode(name);
        return { assignedName, name };
    }

    if (isPropertyNameLiteral(name.expression) && !isIdentifier(name.expression)) {
        const assignedName = factory.createStringLiteralFromNode(name.expression);
        return { assignedName, name };
    }

    const assignedName = factory.getGeneratedNameForNode(name);
    context.hoistVariableDeclaration(assignedName);

    const key = context.getEmitHelperFactory().createPropKeyHelper(name.expression);
    const assignment = factory.createAssignment(assignedName, key);
    const updatedName = factory.updateComputedPropertyName(name, assignment);
    return { assignedName, name: updatedName };
}`;
const assignedNameAfter = assignedNameBefore.replace(`

    if (isPropertyNameLiteral(name.expression)`, `

    if (isBigIntLiteral(name)) {
        const assignedName = factory.createStringLiteral(parsePseudoBigInt(name.text));
        return { assignedName, name };
    }

    if (isPropertyNameLiteral(name.expression)`);

// Each owner: its file, the chain of function names that holds it, and, for the decorated name's if statement, the
// condition that opens it. `caller` marks the edits the mutant leaves out.
const pins = [
    { file: utilities, functions: ['isPropertyNameLiteral'], before: claimBefore, after: claimAfter, caller: false },
    {
        file: esDecorators, functions: ['transformESDecorators', 'partialTransformClassElement'],
        condition: 'isIdentifier(member.name) || isPrivateIdentifier(member.name)', before: decoratedNameBefore, after: decoratedNameAfter, caller: true,
    },
    { file: esDecorators, functions: ['transformESDecorators', 'visitReferencedPropertyName'], before: referencedNameBefore, after: referencedNameAfter, caller: true },
    { file: namedEvaluation, functions: ['getAssignedNameOfPropertyName'], before: assignedNameBefore, after: assignedNameAfter, caller: true },
];
for (const pin of pins) assert.notEqual(pin.before, pin.after, `${pin.functions.at(-1)}: the edit didn't apply to its own pin`);
const importers = [esDecorators, namedEvaluation];

function parse(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${fileName}: parse error`);
    return source;
}

// Every node under root that the predicate takes.
function find(root, predicate) {
    const found = [];
    (function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(root);
    return found;
}

// The function declarations naming `functions`, each the next's ancestor, the first at the file's top level.
function owner(file, source, functions) {
    let scope = source;
    for (const [index, name] of functions.entries()) {
        const found = find(scope, node => ts.isFunctionDeclaration(node) && node.name?.text === name
            && (index > 0 || node.parent === source));
        assert.equal(found.length, 1, `${file}: ${functions.slice(0, index + 1).join('.')} count drift`);
        [scope] = found;
    }
    return scope;
}

// A file's state and its planned edits: each pin read as 'before' or 'after', its import names as present or absent.
function read(file, text, source, callers) {
    const states = [];
    const edits = [];
    for (const pin of pins.filter(pin => pin.file === file)) {
        let node = owner(file, source, pin.functions);
        if (pin.condition) {
            const statements = find(node, child => ts.isIfStatement(child) && child.expression.getText(source) === pin.condition);
            assert.equal(statements.length, 1, `${file}: ${pin.functions.at(-1)}'s name chain count drift`);
            [node] = statements;
        }
        const start = node.getStart(source);
        const current = text.slice(start, node.end);
        assert([pin.before, pin.after].includes(current), `${file}: ${pin.functions.at(-1)} drift or partial adaptation`);
        const state = current === pin.after ? 'after' : 'before';
        states.push({ name: pin.functions.at(-1), state });
        if (state === 'before' && (callers || !pin.caller)) edits.push({ at: start, end: node.end, text: pin.after });
    }
    if (importers.includes(file)) {
        const imports = source.statements.filter(node => ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)
            && node.moduleSpecifier.text === namespaceModule);
        assert.equal(imports.length, 1, `${file}: namespace import drift`);
        const bindings = imports[0].importClause?.namedBindings;
        assert(bindings && ts.isNamedImports(bindings), `${file}: expected a named namespace import`);
        const elements = [...bindings.elements];
        for (const name of added) {
            const named = elements.filter(element => element.name.text === name);
            assert(named.length <= 1, `${file}: ${name} imported twice`);
            if (named.length) {
                assert.equal(named[0].getText(source), name, `${file}: ${name} import drift`);
                states.push({ name: `import ${name}`, state: 'after' });
                continue;
            }
            states.push({ name: `import ${name}`, state: 'before' });
            if (!callers) continue;
            const next = elements.find(element => element.name.text.toLowerCase() > name.toLowerCase());
            assert(next, `${file}: no specifier sorts after ${name}`);
            const separator = text.slice(next.getFullStart(), next.getStart(source));
            assert.match(separator, /^\n[ \t]+$/, `${file}: expected ${next.name.text} on its own line`);
            edits.push({ at: next.getStart(source), end: next.getStart(source), text: `${name},${separator}`, name });
        }
    }
    return { states, edits };
}

// One file as it should be, from its raw text: its owners' and import names' states, and its planned output.
function planFile(file, raw, callers) {
    assert(!/\r(?!\n)/.test(raw), `${file}: bare CR`);
    assert(!(raw.includes('\r\n') && /(?<!\r)\n/.test(raw)), `${file}: mixed line endings`);
    const crlf = raw.includes('\r\n');
    const text = raw.replaceAll('\r\n', '\n');
    const { states, edits } = read(file, text, parse(file, text), callers);
    let output = text;
    // Back to front; two names at one place go in so the earlier name ends up first.
    edits.sort((a, b) => b.at - a.at || (b.name ?? '').localeCompare(a.name ?? ''));
    for (const edit of edits) output = output.slice(0, edit.at) + edit.text + output.slice(edit.end);
    if (output !== text) parse(file, output);
    return { states: states.map(state => ({ file, ...state })), output: crlf ? output.replaceAll('\n', '\r\n') : output, edits: edits.length };
}

// The tree's files as they should be, from their raw texts by file. Every owner and import name must read upstream,
// or every one adapted; the mutant (callers: false) plans the claim's edit alone.
function plan(raws, callers = true) {
    const plans = Object.entries(raws).map(([file, raw]) => ({ file, raw, ...planFile(file, raw, callers) }));
    const states = plans.flatMap(entry => entry.states);
    const adapted = states.filter(state => state.state === 'after');
    assert(adapted.length === 0 || adapted.length === states.length,
        `partly adapted: ${states.map(state => `${state.file} ${state.name} ${state.state}`).join(', ')}`);
    return plans;
}

function rewrite(tree, { callers = true } = {}) {
    const files = [...new Set(pins.map(pin => pin.file))];
    const raws = Object.fromEntries(files.map(file => [file, fs.readFileSync(path.join(tree, file), 'utf8')]));
    const plans = plan(raws, callers);
    const changed = plans.filter(entry => entry.output !== entry.raw);
    if (!changed.length) return { files: 0, edits: 0, javascriptChanging: true };
    // The planned files must read as adapted to a second run, which then plans nothing.
    if (callers) {
        const second = plan(Object.fromEntries(plans.map(entry => [entry.file, entry.output])));
        assert(second.every(entry => entry.output === entry.raw), 'invalid planned result');
    }
    for (const file of files) assert.equal(fs.readFileSync(path.join(tree, file), 'utf8'), raws[file], `${file}: changed while planning`);
    for (const entry of changed) fs.writeFileSync(path.join(tree, entry.file), entry.output);
    return { files: changed.length, edits: changed.reduce((sum, entry) => sum + entry.edits, 0), javascriptChanging: true };
}

const adapt = tree => rewrite(tree);
module.exports = { adapt, rewrite, plan, pins };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify({ typescript: ts.version, task: '#ad161ms', ...adapt(path.resolve(process.argv[2])) }));
}
