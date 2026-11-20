#!/usr/bin/env node
'use strict';
// #ad161ms, Adamic's ruling: isPropertyName claims the five kinds its body accepts. The body returns true for an
// Identifier, a PrivateIdentifier, a StringLiteral, a NumericLiteral and a ComputedPropertyName, and the claim said
// PropertyName, which also holds NoSubstitutionTemplateLiteral and BigIntLiteral, two kinds the body never accepts.
// utilitiesPublic.ts's namespace import gains the two names it lacks, as type-only specifiers in the list's own order,
// and drops PropertyName when nothing else in the file names it. Type-only: the emitted JavaScript is compared before
// any write. The owner and the import are found through the stock AST, the owner's whole text is pinned before and
// after, a partly adapted file is refused, and a second run writes nothing. This directory sorts after
// 99-correctness-binding-views, which pins the same import's whole text; 99-truthful-generated-name-emit-node, after
// it, inserts into that import by name and pins none of it.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');

const file = 'src/compiler/utilitiesPublic.ts';
const namespaceModule = './_namespaces/ts.js';
const upstreamClaim = 'node is PropertyName';
const adaptedClaim = 'node is Identifier | PrivateIdentifier | StringLiteral | NumericLiteral | ComputedPropertyName';
// The claim's names the import lacks upstream and in every adapted tree before this one.
const added = ['ComputedPropertyName', 'NumericLiteral'];
const before = `export function isPropertyName(node: Node): ${upstreamClaim} {
    const kind = node.kind;
    return kind === SyntaxKind.Identifier
        || kind === SyntaxKind.PrivateIdentifier
        || kind === SyntaxKind.StringLiteral
        || kind === SyntaxKind.NumericLiteral
        || kind === SyntaxKind.ComputedPropertyName;
}`;
assert.equal(before.split(upstreamClaim).length, 2, 'ambiguous claim');

function parse(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${fileName}: parse error`);
    return source;
}

function transpile(text, removeComments) {
    return ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } }).outputText;
}

// The file's text as it should be, from its current text, with the claim written as `claim`. The mutant passes the
// upstream claim, so only the import moves.
function plan(fileName, raw, claim) {
    const after = before.replace(upstreamClaim, claim);
    const source = parse(fileName, raw);
    const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'isPropertyName');
    assert.equal(owners.length, 1, `${file}: isPropertyName count drift`);
    const [owner] = owners;
    const ownerStart = owner.getStart(source);
    const ownerText = raw.slice(ownerStart, owner.end).replaceAll('\r\n', '\n');
    assert([before, after].includes(ownerText), `${file}: isPropertyName drift or partial adaptation`);
    const adapted = ownerText === after && after !== before;

    const imports = source.statements.filter(node => ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)
        && node.moduleSpecifier.text === namespaceModule);
    assert.equal(imports.length, 1, `${file}: namespace import drift`);
    const bindings = imports[0].importClause?.namedBindings;
    assert(bindings && ts.isNamedImports(bindings), `${file}: expected a named namespace import`);
    const elements = [...bindings.elements];
    const named = name => elements.filter(element => element.name.text === name);

    // PropertyName as named anywhere but its import specifier and the claim.
    const claimName = owner.type && ts.isTypePredicateNode(owner.type) && owner.type.type && ts.isTypeReferenceNode(owner.type.type)
        ? owner.type.type.typeName : undefined;
    let otherNames = 0;
    (function visit(node) {
        if (ts.isIdentifier(node) && node.text === 'PropertyName' && !ts.isImportSpecifier(node.parent) && node !== claimName) otherNames++;
        ts.forEachChild(node, visit);
    })(source);
    const propertyName = named('PropertyName');
    assert(propertyName.length <= 1, `${file}: PropertyName imported twice`);

    if (adapted) {
        for (const name of added) {
            assert.equal(named(name).length, 1, `${file}: partly adapted import, ${name}`);
            assert.equal(named(name)[0].getText(source), `type ${name}`, `${file}: ${name} import drift`);
        }
        assert.equal(propertyName.length, otherNames > 0 ? 1 : 0, `${file}: partly adapted import, PropertyName`);
        return raw;
    }

    // Upstream claim: the import is as every earlier adaptation leaves it, or this file was partly adapted.
    for (const name of added) assert.equal(named(name).length, 0, `${file}: partly adapted import, ${name} without the claim`);
    assert.equal(propertyName.length, 1, `${file}: partly adapted import, PropertyName missing`);
    const edits = [];
    for (const name of added) {
        const next = elements.find(element => element.name.text.toLowerCase() > name.toLowerCase());
        assert(next, `${file}: no specifier sorts after ${name}`);
        const separator = raw.slice(next.getFullStart(), next.getStart(source));
        assert.match(separator, /^\r?\n[ \t]+$/, `${file}: expected ${next.name.text} on its own line`);
        edits.push({ at: next.getStart(source), end: next.getStart(source), text: `type ${name},${separator}` });
    }
    const keepsPropertyName = claim === upstreamClaim || otherNames > 0;
    if (!keepsPropertyName) {
        const index = elements.indexOf(propertyName[0]);
        assert(index > 0 && index < elements.length - 1, `${file}: PropertyName is not inside the import list`);
        assert.match(raw.slice(propertyName[0].getFullStart(), propertyName[0].getStart(source)), /^\r?\n[ \t]+$/,
            `${file}: expected PropertyName on its own line`);
        edits.push({ at: propertyName[0].getFullStart(), end: elements[index + 1].getFullStart(), text: '' });
    }
    const claimAt = ownerStart + raw.slice(ownerStart, owner.end).indexOf(upstreamClaim);
    edits.push({ at: claimAt, end: claimAt + upstreamClaim.length, text: claim });
    let output = raw;
    for (const edit of edits.sort((a, b) => b.at - a.at)) output = output.slice(0, edit.at) + edit.text + output.slice(edit.end);
    return output;
}

function rewrite(tree, claim = adaptedClaim) {
    const fileName = path.join(tree, file);
    const raw = fs.readFileSync(fileName, 'utf8');
    const output = plan(fileName, raw, claim);
    if (output === raw) return { files: 0, edits: 0, javascriptChanging: false };
    // The planned file must read as adapted to a second run, which then writes nothing.
    if (claim === adaptedClaim) assert.equal(plan(fileName, output, claim), output, 'invalid planned result');
    for (const removeComments of [false, true]) {
        assert.equal(transpile(output, removeComments), transpile(raw, removeComments), `${file}: JavaScript changed`);
    }
    assert.equal(fs.readFileSync(fileName, 'utf8'), raw, `${file}: changed while planning`);
    fs.writeFileSync(fileName, output);
    return { files: 1, edits: 1, javascriptChanging: false };
}

const adapt = tree => rewrite(tree);
module.exports = { adapt, rewrite, plan, upstreamClaim, adaptedClaim };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify({ typescript: ts.version, task: '#ad161ms', ...adapt(path.resolve(process.argv[2])) }));
}
