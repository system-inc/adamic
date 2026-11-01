'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');

function plan(tree, rule) {
    assert.equal(ts.version, '6.0.3');
    assert.equal(rule.replacement, 'unknown');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const file = program.getSourceFile(path.join(tree, rule.file));
    assert(file && !file.parseDiagnostics.length, 'cannot parse timer owner');
    const owners = nodes(ts, file, n => ts.isInterfaceDeclaration(n) && n.name.text === rule.owner);
    assert.equal(owners.length, 1, 'timer owner moved');
    const owner = owners[0];
    assert(!owner.modifiers?.length, 'timer state became public');
    assert.equal(owner.typeParameters?.length, 1);
    assert.equal(owner.typeParameters[0].getText(file), 'T extends BuilderProgram');
    const members = owner.members.filter(n => n.name?.getText(file) === rule.key);
    assert.equal(members.length, 1, 'timer slot moved');
    const member = members[0];
    assert(ts.isPropertySignature(member) && member.type && !member.questionToken && !member.modifiers?.length, 'unreviewed timer declaration');
    assert([ts.SyntaxKind.AnyKeyword, ts.SyntaxKind.UnknownKeyword].includes(member.type.kind), 'unreviewed timer type');
    const symbol = checker.getSymbolAtLocation(member.name);
    assert(symbol, 'timer symbol missing');
    const belongs = s => s === symbol || s?.declarations?.includes(member);
    function functionName(node) {
        for (let p = node.parent; p; p = p.parent) {
            if (ts.isFunctionDeclaration(p)) return p.name?.text;
        }
        throw Error('timer reference outside reviewed function');
    }
    const observed = [], labels = [];
    function callable(call, name, input) {
        assert.equal(call.expression.getText(call.getSourceFile()), 'hostWithWatch.' + name, 'timer host call moved');
        const signature = checker.getResolvedSignature(call);
        assert(signature, 'timer host signature missing');
        if (input) {
            assert.equal(signature.parameters.length, 1, 'timer cancellation contract moved');
            assert.equal(checker.getTypeOfSymbolAtLocation(signature.parameters[0], call).flags, ts.TypeFlags.Any, 'timer cancellation no longer accepts an opaque handle');
        }
        else assert.equal(checker.getReturnTypeOfSignature(signature).flags, ts.TypeFlags.Any, 'timer registration is no longer an untyped host boundary');
    }
    for (const source of program.getSourceFiles().filter(s => s.fileName.startsWith(path.join(tree, 'src/compiler') + path.sep))) {
        for (const name of nodes(ts, source, n => (ts.isIdentifier(n) || ts.isStringLiteral(n)) && n.text === rule.key)) {
            if (name === member.name) continue;
            const parent = name.parent;
            if (ts.isStringLiteral(name)) {
                assert(ts.isCallExpression(parent) && parent.arguments[2] === name, 'unreviewed string-key timer use');
                callable(parent, 'setTimeout', false);
                labels.push({function: functionName(name), expression: parent.getText(source)});
                continue;
            }
            let role, expression, inputType;
            if (ts.isPropertyAssignment(parent) && parent.name === name) {
                const contextual = checker.getContextualType(parent.parent);
                assert(belongs(contextual?.getProperty(rule.key)), 'unreviewed timer initialization owner');
                assert.equal(checker.getTypeAtLocation(parent.initializer).flags, ts.TypeFlags.Undefined, 'timer initialization is no longer undefined');
                role = 'initialize'; expression = parent;
            }
            else {
                assert(ts.isPropertyAccessExpression(parent) && parent.name === name && belongs(checker.getSymbolAtLocation(name)), 'unreviewed timer reference');
                const use = parent.parent;
                if (ts.isIfStatement(use) && use.expression === parent) {
                    role = 'truthiness'; expression = parent;
                }
                else if (ts.isCallExpression(use) && use.arguments.length === 1 && use.arguments[0] === parent) {
                    callable(use, 'clearTimeout', true);
                    role = 'cancel'; expression = use;
                }
                else if (ts.isBinaryExpression(use) && use.left === parent && use.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
                    expression = use;
                    if (ts.isCallExpression(use.right)) {
                        callable(use.right, 'setTimeout', false);
                        role = 'register'; inputType = checker.typeToString(checker.getTypeAtLocation(use.right));
                        assert.equal(checker.getTypeAtLocation(use.right).flags, ts.TypeFlags.Any);
                    }
                    else {
                        assert(ts.isIdentifier(use.right) && use.right.text === 'undefined' && checker.getTypeAtLocation(use.right).flags === ts.TypeFlags.Undefined, 'unreviewed timer reset');
                        role = 'reset';
                    }
                }
                else throw Error('unreviewed timer read or write');
            }
            const row = {role, function: functionName(name), expression: expression.getText(source)};
            if (inputType) row.inputType = inputType;
            observed.push(row);
        }
    }
    const order = rows => [...rows].sort((a, b) => JSON.stringify(a).localeCompare(JSON.stringify(b)));
    assert.deepEqual(order(observed), order(rule.uses), 'timer use inventory moved');
    assert.deepEqual(labels, rule.labels, 'timer registration label moved');
    const before = file.text;
    const after = before.slice(0, member.type.getStart(file)) + rule.replacement + before.slice(member.type.end);
    for (const removeComments of [false, true]) {
        const options = {fileName: file.fileName, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        assert.equal(ts.transpileModule(before, options).outputText, ts.transpileModule(after, options).outputText, 'timer annotation changed JavaScript');
    }
    return {name: file.fileName, before, after, uses: observed, public: false};
}
function apply(tree, rules) {
    assert.equal(rules.length, 1, 'reviewed timer class size moved');
    const result = plan(path.resolve(tree), rules[0]);
    assert.equal(fs.readFileSync(result.name, 'utf8'), result.before, 'timer source changed during planning');
    const files = Number(result.before !== result.after);
    if (files) fs.writeFileSync(result.name, result.after);
    const evidence = {class: 'opaque-timer-slot', files, removed: files, uses: result.uses.length, public: result.public, javascriptIdentical: true};
    console.log(JSON.stringify(evidence));
    return evidence;
}
module.exports = {plan, apply};
