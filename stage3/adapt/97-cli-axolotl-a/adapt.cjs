'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// verifyCompilerOptions tests a paths substitution through a stored typeof, which never narrows subst (#qg1ry2t).
// The condition reads typeof subst itself; typeOfSubst stays for the else branch's message (README.md).
const file = 'src/compiler/program.ts';
const before = 'typeOfSubst === "string"';
const after = 'typeof subst === "string"';
// Each binding's references at the site, the declaration's name included: the else branch keeps one typeOfSubst.
const references = {before: {subst: 7, typeOfSubst: 3}, after: {subst: 8, typeOfSubst: 2}};

function nodes(source, test) {
    const found = [];
    (function visit(node) {
        if (test(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(source);
    return found;
}

function identifiers(node, name) {
    return nodes(node, n => ts.isIdentifier(n) && n.text === name).length;
}

function constDeclaration(statement, name) {
    if (!ts.isVariableStatement(statement)) return undefined;
    const list = statement.declarationList;
    if (!(list.flags & ts.NodeFlags.Const) || list.declarations.length !== 1) return undefined;
    const declaration = list.declarations[0];
    return ts.isIdentifier(declaration.name) && declaration.name.text === name && !declaration.type ? declaration : undefined;
}

// before: typeOfSubst === "string". after: typeof subst === "string". Anything else is not inventoried.
function conditionShape(condition, source) {
    if (!ts.isBinaryExpression(condition) || condition.operatorToken.kind !== ts.SyntaxKind.EqualsEqualsEqualsToken) return undefined;
    if (!ts.isStringLiteral(condition.right) || condition.right.text !== 'string') return undefined;
    const left = condition.left;
    const shape = ts.isIdentifier(left) && left.text === 'typeOfSubst' ? 'before'
        : ts.isTypeOfExpression(left) && ts.isIdentifier(left.expression) && left.expression.text === 'subst' ? 'after'
        : undefined;
    if (shape) assert.equal(condition.getText(source), shape === 'before' ? before : after, `the condition reads ${condition.getText(source)}, not an inventoried spelling`);
    return shape;
}

function locate(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse program.ts');
    const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'verifyCompilerOptions');
    assert.equal(functions.length, 1, 'verifyCompilerOptions moved');
    // Every loop over a paths entry's substitutions, anywhere in the file: there must be exactly one, in verifyCompilerOptions.
    const loops = nodes(source, n => ts.isForStatement(n) && ts.isBlock(n.statement) && n.statement.statements.length > 0 && constDeclaration(n.statement.statements[0], 'subst') !== undefined);
    assert.equal(loops.length, 1, `${loops.length} substitution loops, not one`);
    const loop = loops[0];
    let owner = loop.parent;
    while (owner && !ts.isFunctionLike(owner)) owner = owner.parent;
    assert.equal(owner, functions[0], 'the substitution loop left verifyCompilerOptions');
    const statements = loop.statement.statements;
    assert.equal(statements.length, 3, 'the substitution loop grew or shrank');
    const subst = constDeclaration(statements[0], 'subst');
    assert.equal(subst.initializer?.getText(source), 'options.paths[key]![i]', 'subst reads something else');
    const typeOfSubst = constDeclaration(statements[1], 'typeOfSubst');
    assert(typeOfSubst, 'const typeOfSubst moved');
    assert(typeOfSubst.initializer && ts.isTypeOfExpression(typeOfSubst.initializer) && ts.isIdentifier(typeOfSubst.initializer.expression) && typeOfSubst.initializer.expression.text === 'subst', `typeOfSubst is ${typeOfSubst.initializer?.getText(source)}, not typeof subst`);
    const branch = statements[2];
    assert(ts.isIfStatement(branch) && branch.elseStatement, 'the substitution test moved');
    const shape = conditionShape(branch.expression, source);
    assert(shape, `the condition is ${branch.expression.getText(source)}, not an inventoried shape`);
    // The else branch still reports typeOfSubst; without that read, the declaration this repair keeps would be dead.
    assert.equal(identifiers(branch.elseStatement, 'typeOfSubst'), 1, 'the else branch no longer reports typeOfSubst');
    // Either untouched or already adapted, never half: every reference must agree with the condition's shape.
    const counted = {subst: identifiers(loop.statement, 'subst'), typeOfSubst: identifiers(source, 'typeOfSubst')};
    assert.deepEqual(counted, references[shape], `program.ts is partly adapted: ${JSON.stringify(counted)} with a ${shape} condition`);
    assert.equal(nodes(source, n => ts.isBinaryExpression(n) && conditionShape(n, source) !== undefined).length, 1, 'another substitution test appeared');
    return {shape, condition: branch.expression, source};
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    const site = locate(fileName, text);
    const edits = site.shape === 'before' ? [{start: site.condition.getStart(site.source), end: site.condition.end, text: after}] : [];
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    assert.equal(locate(fileName, result).shape, 'after', 'the planned edit does not adapt the site');
    assert.equal(result.length - text.length, edits.length * (after.length - before.length), 'the planned edit touches other bytes');
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: true, behaviorChanging: false};
}
module.exports = {adapt, file, before, after, references};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
