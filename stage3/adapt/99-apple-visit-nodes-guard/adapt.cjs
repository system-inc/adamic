'use strict';
// #ytagqs5: visitorPublic.ts's element tests run whenever test is given, not only at AssertionLevel.Normal. Upstream's
// visitArrayWorker checks each element it keeps with Debug.assertEachNode and Debug.assertNode, which do nothing below
// Normal, and tsc runs at None. Each of the four becomes an ungated check that fails in tsc's own words, the message
// Debug.assert composes from those assertions' words, and visitNodes gains the same element loop as the statement right
// before `return updatedArray` and before its final `return nodes`, so every array it returns was checked against test
// in the statement before the return. Adamic ruled the four sites on #ytagqs5 and covered the two visitNodes loops
// (J-79earng, J-ezkhwzs: 0 failures over the lane oracle, emit byte-identical, cost inside the noise).
// The adapter finds the four Debug.assert* call statements in visitArrayWorker and the three returns in visitNodes by
// the syntax tree, asserts their count and shape, and edits by those nodes alone. Each owner is pinned whole, before
// and after; anything else is refused, as is one owner adapted and the other not, and a second application writes
// nothing. debug.ts's words are read by their nodes, so a message that drifts from tsc's refuses the apply.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const pins = require('./pins.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
const newline = '\r\n';
const debugFile = 'src/compiler/debug.ts';

// The failure messages, spelled as TypeScript source: Debug.assert's `False expression: ${message}`, its
// "\r\nVerbose Debug Information: " joint, assertEachNode's and assertNode's default message and verbose text, as the
// built tsc says them. At this adapter's place in apply order, debug.ts's verbose texts still end with upstream's
// ` '${getFunctionName(test!)}'.`, which 99-correctness-core-debug-messages drops later for `.`; test.cjs holds these
// messages to the applied tree's debug.ts.
const eachMessage = '"False expression: Unexpected node.\\r\\nVerbose Debug Information: Node array did not pass test."';
const oneMessage = subject => '`False expression: Unexpected node.\\r\\nVerbose Debug Information: Node ${Debug.formatSyntaxKind(' +
    subject + '.kind)} did not pass test.`';
// debug.ts's words that compose those messages, each once in its owner inside namespace Debug, read in both the
// upstream and the later adapted spelling.
const debugWords = {
    assert: [
        'message = message ? `False expression: ${message}` : "False expression.";',
        'message += "\\r\\nVerbose Debug Information: " + (typeof verboseDebugInfo === "string" ? verboseDebugInfo : verboseDebugInfo());',
        'fail(message, stackCrawlMark || assert);',
    ],
    fail: ['const e = new Error(message ? `Debug Failure. ${message}` : "Debug Failure.");', 'throw e;'],
    assertEachNode: ['message || "Unexpected node.",', '() => `Node array did not pass test'],
    assertNode: ['node !== undefined && (test === undefined || test(node)),', 'message || "Unexpected node.",',
        '() => `Node ${formatSyntaxKind(node?.kind)} did not pass test'],
    formatSyntaxKind: [],
};

// eachGuard is the loop that stands where an assertEachNode stood, or right before a return; indent is the column its
// first line starts at.
function eachGuard(subject, indent) {
    return [
        'if (test !== undefined) {',
        `${indent}    for (const testedNode of ${subject}) {`,
        `${indent}        if (!test(testedNode)) Debug.fail(${eachMessage});`,
        `${indent}    }`,
        `${indent}}`,
    ].join(newline);
}
const oneGuard = subject => `if (test !== undefined && !test(${subject})) Debug.fail(${oneMessage(subject)});`;

function parse(text, fileName) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${fileName}: parse error`);
    return source;
}

function implementation(source, name) {
    const found = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(found.length, 1, `${name}: missing or duplicate implementation`);
    return found[0];
}

// state reads an owner against its pins: before, after, or refused.
function state(source, name) {
    const owner = implementation(source, name);
    const spelling = owner.getText(source);
    const pin = pins.owners[name];
    const found = spelling === pin.before ? 'before' : spelling === pin.after ? 'after' : undefined;
    assert(found, `${name}: unreviewed owner text`);
    assert.equal(digest(spelling), pin[`${found}Sha256`], `${name}: owner hash drift`);
    return {owner, state: found};
}

function descendants(root, predicate, stopAtFunctions = true) {
    const found = [];
    (function visit(node) {
        if (stopAtFunctions && node !== root && ts.isFunctionLike(node)) return;
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(root);
    return found;
}

// indentOf is the run of spaces from a statement's line start to the statement; anything else there is refused.
function indentOf(source, node) {
    const start = node.getStart(source);
    const line = source.getLineAndCharacterOfPosition(start).line;
    const indent = source.text.slice(source.getPositionOfLineAndCharacter(line, 0), start);
    assert.match(indent, /^ *$/, `${node.getText(source)}: not alone on its line`);
    return indent;
}

const previousStatement = statement => {
    const statements = statement.parent.statements;
    const index = statements.indexOf(statement);
    return index > 0 ? statements[index - 1] : undefined;
};

// locate finds the six edit points in a before file: visitArrayWorker's four assertion statements, in source order
// assertEachNode(updated), assertNode(visitedNode), assertNode(visited), assertEachNode(nodes), and visitNodes'
// `return updatedArray` and final `return nodes`. Its counts and neighbours are the shape this edit was measured on.
function locate(source) {
    const worker = implementation(source, 'visitArrayWorker');
    const visitNodes = implementation(source, 'visitNodes');
    for (const owner of [worker, visitNodes]) {
        assert.equal(descendants(owner, node => ts.isIdentifier(node) && node.text === 'testedNode', false).length, 0,
            `${owner.name.text}: already names testedNode`);
        assert.deepEqual(owner.parameters.map(parameter => parameter.name.getText(source)), ['nodes', 'visitor', 'test', 'start', 'count'],
            `${owner.name.text}: parameters`);
    }
    const debugCalls = descendants(worker.body, node => ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) &&
        ts.isIdentifier(node.expression.expression) && node.expression.expression.text === 'Debug');
    const assertions = debugCalls.map(call => {
        assert(ts.isExpressionStatement(call.parent), 'visitArrayWorker: a Debug call that is not its own statement');
        assert(ts.isBlock(call.parent.parent), 'visitArrayWorker: a Debug statement outside a block');
        assert.equal(call.arguments.length, 2, 'visitArrayWorker: a Debug call with other than two arguments');
        assert(call.arguments.every(ts.isIdentifier), 'visitArrayWorker: a Debug call whose arguments are not names');
        assert.equal(call.arguments[1].text, 'test', 'visitArrayWorker: a Debug call whose test is not the parameter');
        return {statement: call.parent, assertion: call.expression.name.text, subject: call.arguments[0].text};
    });
    assert.deepEqual(assertions.map(site => `${site.assertion}(${site.subject})`),
        ['assertEachNode(updated)', 'assertNode(visitedNode)', 'assertNode(visited)', 'assertEachNode(nodes)'],
        'visitArrayWorker: the four assertions');
    const workerTests = descendants(worker.body, node => ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'test');
    assert.equal(workerTests.length, 0, 'visitArrayWorker: calls test outside its assertions');
    const workerLast = worker.body.statements.at(-1);
    assert(ts.isReturnStatement(workerLast) && workerLast.expression?.getText(source) === 'nodes', 'visitArrayWorker: its last statement');
    assert.equal(previousStatement(workerLast), assertions[3].statement, 'visitArrayWorker: assertEachNode(nodes) is not right before its return');

    const returns = descendants(visitNodes.body, ts.isReturnStatement);
    assert.deepEqual(returns.map(statement => statement.expression?.getText(source)), ['nodes', 'updatedArray', 'nodes'], 'visitNodes: its returns');
    const [undefinedReturn, updatedReturn, finalReturn] = returns;
    const undefinedGuard = undefinedReturn.parent.parent;
    assert(ts.isBlock(undefinedReturn.parent) && undefinedReturn.parent.statements.length === 1 && ts.isIfStatement(undefinedGuard) &&
        undefinedGuard.expression.getText(source) === 'nodes === undefined', 'visitNodes: the first return is not alone under `nodes === undefined`');
    assert(ts.isBlock(updatedReturn.parent), 'visitNodes: `return updatedArray` is not in a block');
    assert.equal(previousStatement(updatedReturn)?.getText(source), 'setTextRangePosEnd(updatedArray, pos, end);',
        'visitNodes: `return updatedArray` does not follow setTextRangePosEnd');
    assert.equal(visitNodes.body.statements.at(-1), finalReturn, 'visitNodes: the final `return nodes` is not its last statement');
    const beforeFinal = previousStatement(finalReturn);
    assert(beforeFinal && ts.isIfStatement(beforeFinal) && beforeFinal.thenStatement.statements?.includes(updatedReturn),
        'visitNodes: the final `return nodes` does not follow the `return updatedArray` branch');
    return {
        replacements: assertions.map(site => ({
            statement: site.statement,
            text: site.assertion === 'assertEachNode' ? eachGuard(site.subject, indentOf(source, site.statement)) : oneGuard(site.subject),
        })),
        insertions: [updatedReturn, finalReturn].map(statement => {
            const indent = indentOf(source, statement);
            return {statement, text: eachGuard(statement.expression.getText(source), indent) + newline + indent};
        }),
    };
}

function checkDebugWords(text) {
    const source = parse(text, debugFile);
    const namespaces = source.statements.filter(node => ts.isModuleDeclaration(node) && node.name.text === 'Debug' && node.body && ts.isModuleBlock(node.body));
    assert.equal(namespaces.length, 1, `${debugFile}: namespace Debug`);
    for (const [name, words] of Object.entries(debugWords)) {
        const owners = namespaces[0].body.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
        assert.equal(owners.length, 1, `${debugFile}: Debug.${name}`);
        assert(owners[0].modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword), `${debugFile}: Debug.${name} is not exported`);
        const spelling = owners[0].getText(source);
        for (const word of words) assert.equal(spelling.split(word).length - 1, 1, `${debugFile}: Debug.${name} no longer says ${word}`);
    }
}

function plan(text, debugText) {
    assert.equal(ts.version, '6.0.3');
    checkDebugWords(debugText);
    const source = parse(text, pins.file);
    const states = ['visitArrayWorker', 'visitNodes'].map(name => state(source, name).state);
    assert.equal(states[0], states[1], `${pins.file}: one owner adapted and the other not`);
    if (states[0] === 'after') return text;
    const {replacements, insertions} = locate(source);
    const edits = [
        ...replacements.map(edit => ({start: edit.statement.getStart(source), end: edit.statement.end, text: edit.text})),
        ...insertions.map(edit => ({start: edit.statement.getStart(source), end: edit.statement.getStart(source), text: edit.text})),
    ].sort((left, right) => right.start - left.start);
    assert.equal(edits.length, 6);
    let changed = text;
    for (const edit of edits) changed = changed.slice(0, edit.start) + edit.text + changed.slice(edit.end);
    const adapted = parse(changed, pins.file);
    for (const name of ['visitArrayWorker', 'visitNodes']) assert.equal(state(adapted, name).state, 'after', `${name}: invalid planned owner`);
    // The JavaScript changes only inside the two owners: the file's emit is the before file's emit with each owner's
    // before JavaScript replaced by its after JavaScript, with and without comments.
    for (const removeComments of [true, false]) {
        const emit = input => ts.transpileModule(input, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
        const alone = input => emit(input).replace(/^"use strict";\r?\n/, '').trim();
        const before = emit(text);
        let expected = before;
        for (const name of ['visitArrayWorker', 'visitNodes']) {
            const ownerBefore = alone(pins.owners[name].before), ownerAfter = alone(pins.owners[name].after);
            assert.equal(before.split(ownerBefore).length - 1, 1, `${name}: before owner's JavaScript not found once`);
            expected = expected.replace(ownerBefore, ownerAfter);
        }
        assert.equal(emit(changed), expected, `${pins.file}: unreviewed JavaScript change`);
    }
    // No declaration moves: visitArrayWorker is unexported, and visitNodes' implementation body is not its declaration.
    for (const stripInternal of [true, false]) {
        const emit = input => ts.transpileDeclaration(input, {fileName: pins.file, compilerOptions: {stripInternal}}).outputText;
        assert.equal(emit(changed), emit(text), 'declarations changed');
    }
    return changed;
}

function adapt(tree) {
    const file = path.join(tree, pins.file);
    const before = fs.readFileSync(file, 'utf8');
    const after = plan(before, fs.readFileSync(path.join(tree, debugFile), 'utf8'));
    assert.equal(fs.readFileSync(file, 'utf8'), before, `${pins.file}: changed while planning`);
    if (after !== before) fs.writeFileSync(file, after);
    return {files: Number(after !== before), edits: after !== before ? 6 : 0, javascriptChanging: true};
}

module.exports = {adapt, plan, state, locate, checkDebugWords, eachGuard, oneGuard, pins, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
