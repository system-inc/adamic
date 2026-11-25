'use strict';
// node test.cjs TREE: TREE is a tree stage3/apply.sh wrote. The test reads this adaptation's own snapshots, its -before
// and -after refs, never the final tree, since a later adapter may edit visitorPublic.ts elsewhere. Each check prints
// `--- PASS: <name>` or `--- FAIL: <name> (<why>)`, and any failure exits 1.
//   TestSixEditsInPlace reads the -after visitorPublic.ts by its syntax tree, without the adapter's code: no
//     Debug.assert* call is left in visitArrayWorker, each of its four sites holds its ungated check with that site's
//     neighbours, each of visitNodes' two returned arrays is checked by the statement right before its return, the
//     `nodes === undefined` return is untouched, and with the six edits undone the file is the -before file byte for
//     byte.
//   TestTscWords runs debug.ts's own assert, fail, assertEachNode and assertNode at AssertionLevel.Normal beside the
//     before owners, and the after owners at None: on an element that fails test at each of the four sites, both throw
//     the same message having called test on the same elements. debug.ts alone is read from the applied tree, not a
//     snapshot: the words asked for are the built tsc's, and 99-correctness-core-debug-messages, later in apply
//     order, sets them.
//   TestOrderOfChecks runs the after owners at None on six visits that pass every element, and compares the order
//     test is called in with the order stock calls it at Normal, followed by the returned array once more, which is
//     visitNodes' own loop. The result is stock's at None, element for element.
//   TestVisitNodesLoops plants a test that passes a node the first time and fails it after, which only the visitNodes
//     loops can catch: both returns throw tsc's words, and stock at Normal is quiet.
//   TestAdapterReplay runs adapt.cjs's plan on the -before snapshot and requires the -after bytes, a second run that
//     changes nothing, and refusals of a drifted owner, a half-adapted file and drifted debug.ts words.
// A reviewer's mutation of the output, applied to the -after ref (one of the six edits dropped, or a check moved after
// its return), turns TestSixEditsInPlace and TestOrderOfChecks red.
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const {plan, pins} = require('./adapt.cjs');

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node test.cjs TREE');
const tree = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-apple-visit-nodes-guard';
const visitorFile = 'src/compiler/visitorPublic.ts';
const debugFile = 'src/compiler/debug.ts';
const read = (side, file) => execFileSync('git', ['-C', tree, 'show', `${ref}-${side}:${file}`], {encoding: 'utf8', maxBuffer: 64 * 1024 * 1024});
const before = read('before', visitorFile), after = read('after', visitorFile);
const debugBefore = read('before', debugFile), debugAfter = read('after', debugFile);
const debugBuilt = fs.readFileSync(path.join(tree, debugFile), 'utf8');

let failed = 0;
function check(name, body) {
    try {
        body();
        console.log(`--- PASS: ${name}`);
    }
    catch (error) {
        failed++;
        console.log(`--- FAIL: ${name} (${String(error?.message).split('\n').join(' | ')})`);
    }
}

const parse = (text, fileName) => {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${fileName}: parse error`);
    return source;
};
function implementation(source, name) {
    const found = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(found.length, 1, name);
    return found[0];
}
function descendants(root, predicate) {
    const found = [];
    (function visit(node) {
        if (node !== root && ts.isFunctionLike(node)) return;
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(root);
    return found;
}
const sibling = (statement, offset) => statement.parent.statements[statement.parent.statements.indexOf(statement) + offset];

// The checks' expected spelling, written here rather than read from the adapter.
const eachText = '"False expression: Unexpected node.\\r\\nVerbose Debug Information: Node array did not pass test."';
const oneText = subject => '`False expression: Unexpected node.\\r\\nVerbose Debug Information: Node ${Debug.formatSyntaxKind(' + subject + '.kind)} did not pass test.`';
const isFail = (source, statement, message) => ts.isExpressionStatement(statement) && ts.isCallExpression(statement.expression) &&
    statement.expression.expression.getText(source) === 'Debug.fail' && statement.expression.arguments.length === 1 &&
    statement.expression.arguments[0].getText(source) === message;
function isEachGuard(source, statement, subject) {
    if (!statement || !ts.isIfStatement(statement) || statement.elseStatement || statement.expression.getText(source) !== 'test !== undefined') return false;
    const block = statement.thenStatement;
    if (!ts.isBlock(block) || block.statements.length !== 1) return false;
    const loop = block.statements[0];
    if (!ts.isForOfStatement(loop) || loop.awaitModifier || !ts.isVariableDeclarationList(loop.initializer)) return false;
    const list = loop.initializer;
    if (!(list.flags & ts.NodeFlags.Const) || list.declarations.length !== 1 || list.declarations[0].name.getText(source) !== 'testedNode' ||
        list.declarations[0].type || list.declarations[0].initializer) return false;
    if (!ts.isIdentifier(loop.expression) || loop.expression.text !== subject) return false;
    if (!ts.isBlock(loop.statement) || loop.statement.statements.length !== 1) return false;
    const test = loop.statement.statements[0];
    return ts.isIfStatement(test) && !test.elseStatement && test.expression.getText(source) === '!test(testedNode)' && isFail(source, test.thenStatement, eachText);
}
const isOneGuard = (source, statement, subject) => !!statement && ts.isIfStatement(statement) && !statement.elseStatement &&
    statement.expression.getText(source) === `test !== undefined && !test(${subject})` && isFail(source, statement.thenStatement, oneText(subject));

check('TestSixEditsInPlace', () => {
    const source = parse(after, visitorFile);
    const original = parse(before, visitorFile);
    // visitArrayWorker: the four sites, in source order, each where its assertion stood.
    const worker = implementation(source, 'visitArrayWorker');
    assert.equal(descendants(worker.body, node => ts.isPropertyAccessExpression(node) && node.getText(source).startsWith('Debug.assert')).length, 0,
        'visitArrayWorker still calls a Debug assertion');
    const guards = descendants(worker.body, node => ts.isIfStatement(node) && node.expression.getText(source).startsWith('test !== undefined'));
    const sites = [['each', 'updated', 'Debug.assertEachNode(updated, test);'], ['one', 'visitedNode', 'Debug.assertNode(visitedNode, test);'],
        ['one', 'visited', 'Debug.assertNode(visited, test);'], ['each', 'nodes', 'Debug.assertEachNode(nodes, test);']];
    assert.equal(guards.length, 4, `visitArrayWorker holds ${guards.length} checks, not 4`);
    sites.forEach(([kind, subject], index) => assert((kind === 'each' ? isEachGuard : isOneGuard)(source, guards[index], subject),
        `visitArrayWorker's check ${index + 1} is not the ${kind} check of ${subject}: ${guards[index].getText(source).split('\n')[0]}`));
    assert.equal(sibling(guards[0], -1)?.getText(source), 'updated = nodes.slice(0, i);', 'the prefix check does not follow the slice');
    assert.equal(sibling(guards[0], 1), undefined, 'the prefix check is not the last statement of its block');
    assert.equal(sibling(guards[1], 1)?.getText(source), 'updated.push(visitedNode);', 'the visitedNode check is not right before its push');
    assert.equal(sibling(guards[2], 1)?.getText(source), 'updated.push(visited);', 'the visited check is not right before its push');
    const workerReturn = sibling(guards[3], 1);
    assert(workerReturn && ts.isReturnStatement(workerReturn) && workerReturn.getText(source) === 'return nodes;' && worker.body.statements.at(-1) === workerReturn,
        'visitArrayWorker\'s check of nodes is not right before its last statement, `return nodes;`');
    let workerUndone = worker.getText(source);
    const workerStart = worker.getStart(source);
    for (let index = guards.length - 1; index >= 0; index--) {
        const start = guards[index].getStart(source) - workerStart, end = guards[index].end - workerStart;
        workerUndone = workerUndone.slice(0, start) + sites[index][2] + workerUndone.slice(end);
    }
    assert.equal(workerUndone, implementation(original, 'visitArrayWorker').getText(original), 'visitArrayWorker differs from -before beyond its four sites');

    // visitNodes: each returned array is checked by the statement right before its return.
    const visitNodes = implementation(source, 'visitNodes');
    const returns = descendants(visitNodes.body, ts.isReturnStatement);
    assert.deepEqual(returns.map(statement => statement.getText(source)), ['return nodes;', 'return updatedArray;', 'return nodes;'], 'visitNodes\' returns');
    const [undefinedReturn, updatedReturn, finalReturn] = returns;
    assert(undefinedReturn.parent.statements.length === 1 && undefinedReturn.parent.parent.expression?.getText(source) === 'nodes === undefined',
        'the `nodes === undefined` return is not alone in its branch');
    assert(isEachGuard(source, sibling(updatedReturn, -1), 'updatedArray'), '`return updatedArray` is not right after its check');
    assert.equal(sibling(updatedReturn, -2)?.getText(source), 'setTextRangePosEnd(updatedArray, pos, end);', 'the updatedArray check does not follow setTextRangePosEnd');
    assert.equal(visitNodes.body.statements.at(-1), finalReturn, 'the final `return nodes` is not visitNodes\' last statement');
    assert(isEachGuard(source, sibling(finalReturn, -1), 'nodes'), 'the final `return nodes` is not right after its check');
    const loops = descendants(visitNodes.body, node => ts.isIfStatement(node) && node.expression.getText(source).startsWith('test !== undefined'));
    assert.equal(loops.length, 2, `visitNodes holds ${loops.length} checks, not 2`);
    let visitNodesUndone = visitNodes.getText(source);
    const visitNodesStart = visitNodes.getStart(source);
    for (const statement of [finalReturn, updatedReturn]) {
        const guard = sibling(statement, -1);
        visitNodesUndone = visitNodesUndone.slice(0, guard.getStart(source) - visitNodesStart) + visitNodesUndone.slice(statement.getStart(source) - visitNodesStart);
    }
    assert.equal(visitNodesUndone, implementation(original, 'visitNodes').getText(original), 'visitNodes differs from -before beyond its two checks');

    // The file: with both owners undone it is -before byte for byte, so nothing else moved.
    const owners = [[worker, workerUndone], [visitNodes, visitNodesUndone]].sort((left, right) => right[0].getStart(source) - left[0].getStart(source));
    let fileUndone = after;
    for (const [owner, undone] of owners) fileUndone = fileUndone.slice(0, owner.getStart(source)) + undone + fileUndone.slice(owner.end);
    assert.equal(fileUndone, before, `${visitorFile} differs from -before outside the two owners`);
    assert.equal(debugAfter, debugBefore, `${debugFile} changed`);
});

// The harness: the two owners of one snapshot run in their own context, beside debug.ts's own assertion functions from
// the applied tree, at a level the case sets. formatSyntaxKind is the one stand-in, the same on both sides.
const debugSource = parse(debugBuilt, debugFile);
const debugNamespace = debugSource.statements.find(node => ts.isModuleDeclaration(node) && node.name.text === 'Debug');
const debugCode = ['assert', 'fail', 'assertEachNode', 'assertNode'].map(name => {
    const owners = debugNamespace.body.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(owners.length, 1, `Debug.${name}`);
    return owners[0].getText(debugSource).replace(/^export /, '');
}).join('\n');
function debugApi() {
    const context = vm.createContext({
        level: 0,
        AssertionLevel: {None: 0, Normal: 1, Aggressive: 2, VeryAggressive: 3},
        formatSyntaxKind: kind => `SyntaxKind(${kind})`,
        getFunctionName: func => func.name,
        every: (array, callback) => array === undefined || array.every(callback),
    });
    vm.runInContext('function shouldAssertFunction(level) { return globalThis.level >= level; }\n' +
        ts.transpileModule(debugCode, {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText, context);
    const api = vm.runInContext('({assert, fail, assertEachNode, assertNode})', context);
    return {setLevel: value => { context.level = value; }, Debug: {...api, formatSyntaxKind: context.formatSyntaxKind}};
}
function owners(text) {
    const source = parse(text, visitorFile);
    const code = ['visitArrayWorker', 'visitNodes'].map(name => implementation(source, name).getText(source).replace(/^export /, '')).join('\n');
    const debug = debugApi();
    const context = vm.createContext({
        Debug: debug.Debug,
        isArray: Array.isArray,
        factory: {createNodeArray: (elements, hasTrailingComma) => Object.assign(elements.slice(), {pos: -1, end: -1, hasTrailingComma})},
        setTextRangePosEnd: (range, pos, end) => Object.assign(range, {pos, end}),
    });
    vm.runInContext(ts.transpileModule(code, {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText, context);
    return {visitNodes: vm.runInContext('visitNodes', context), setLevel: debug.setLevel};
}
const stock = owners(before), adapted = owners(after);
const node = name => ({name, kind: name.length, pos: -1, end: -1});
const nodeArray = names => Object.assign(names.map(node), {pos: 0, end: 10, hasTrailingComma: false});
// run visits one case on one side at one level, recording each element test is called on and what it threw.
function run(side, level, visit, predicate) {
    side.setLevel(level);
    const nodes = nodeArray(visit.nodes);
    const calls = [];
    const test = predicate === undefined ? undefined : function isWanted(element) {
        calls.push(element.name);
        return predicate(element);
    };
    try {
        const result = side.visitNodes(nodes, element => visit.visitor(element, node), test, visit.start, visit.count);
        // Array.from, not map: an array the owners made belongs to their context's realm, and deepStrictEqual compares
        // prototypes.
        return {calls, result: Array.from(result, element => element.name), same: result === nodes, range: [result.pos, result.end, result.hasTrailingComma]};
    }
    catch (error) {
        return {calls, threw: error.message};
    }
}
const replace = (target, by) => (element, make) => element.name === target ? by(make) : element;
const visits = {
    unchanged: {nodes: ['a', 'b'], visitor: element => element, calls: ['a', 'b', 'a', 'b']},
    replaced: {nodes: ['a', 'b', 'c'], visitor: replace('b', make => make('z')), calls: ['a', 'z', 'c', 'a', 'z', 'c']},
    spread: {nodes: ['a', 'b', 'c'], visitor: replace('b', make => [make('x'), make('y')]), calls: ['a', 'x', 'y', 'c', 'a', 'x', 'y', 'c']},
    removed: {nodes: ['a', 'b', 'c'], visitor: replace('b', () => undefined), calls: ['a', 'c', 'a', 'c']},
    fragment: {nodes: ['a', 'b', 'c'], visitor: element => element, start: 1, count: 1, calls: ['b', 'b']},
    empty: {nodes: [], visitor: element => element, calls: []},
};

check('TestOrderOfChecks', () => {
    for (const [name, visit] of Object.entries(visits)) {
        const plain = run(stock, 0, visit, () => true), live = run(stock, 1, visit, () => true), ours = run(adapted, 0, visit, () => true);
        for (const result of [plain, live, ours]) assert.equal(result.threw, undefined, `${name}: threw ${result.threw}`);
        assert.deepEqual(plain.calls, [], `${name}: stock at None called test`);
        assert.deepEqual({...ours, calls: undefined}, {...plain, calls: undefined}, `${name}: the result is not stock's`);
        assert.deepEqual(ours.calls, [...live.calls, ...live.result], `${name}: test's calls are not stock's at Normal and then the returned array`);
        assert.deepEqual(ours.calls, visit.calls, `${name}: test was called on ${JSON.stringify(ours.calls)}`);
        const untested = run(adapted, 0, visit, undefined);
        assert.deepEqual({...untested, calls: undefined}, {...plain, calls: undefined}, `${name}: without test, the result is not stock's`);
        assert.deepEqual(untested.calls, [], `${name}: without test, test was called`);
    }
    assert.equal(stock.visitNodes(undefined, element => element, () => true), undefined);
    assert.equal(adapted.visitNodes(undefined, element => element, () => { throw new Error('test called on undefined nodes'); }), undefined);
});

check('TestTscWords', () => {
    const wanted = element => element.name !== 'bad';
    const sites = {
        'the prefix, assertEachNode(updated)': {nodes: ['bad', 'b'], visitor: replace('b', make => make('z'))},
        'a spread element, assertNode(visitedNode)': {nodes: ['a', 'b'], visitor: replace('b', make => [make('bad')])},
        'a replaced element, assertNode(visited)': {nodes: ['a', 'b'], visitor: replace('b', make => make('bad'))},
        'the unchanged array, assertEachNode(nodes)': {nodes: ['a', 'bad'], visitor: element => element},
    };
    for (const [name, visit] of Object.entries(sites)) {
        const plain = run(stock, 0, visit, wanted), live = run(stock, 1, visit, wanted), ours = run(adapted, 0, visit, wanted);
        assert.equal(plain.threw, undefined, `${name}: stock at None threw`);
        assert.match(live.threw ?? '', /^Debug Failure\. False expression: Unexpected node\.\r\nVerbose Debug Information: Node /, `${name}: stock at Normal: ${live.threw}`);
        assert.equal(ours.threw, live.threw, `${name}: the adapted message is not tsc's`);
        assert.deepEqual(ours.calls, live.calls, `${name}: test was called on ${JSON.stringify(ours.calls)}, stock at Normal on ${JSON.stringify(live.calls)}`);
    }
});

check('TestVisitNodesLoops', () => {
    const once = () => {
        const seen = new Set();
        return element => !seen.has(element.name) && Boolean(seen.add(element.name));
    };
    const words = run(stock, 1, {nodes: ['a', 'bad'], visitor: element => element}, element => element.name !== 'bad').threw;
    for (const [name, visit, calls] of [['the final return nodes', visits.unchanged, ['a', 'b', 'a']], ['return updatedArray', visits.replaced, ['a', 'z', 'c', 'a']]]) {
        const live = run(stock, 1, visit, once());
        assert.equal(live.threw, undefined, `${name}: stock at Normal threw ${live.threw}`);
        const ours = run(adapted, 0, visit, once());
        assert.equal(ours.threw, words, `${name}: ${ours.threw}`);
        assert.deepEqual(ours.calls, calls, `${name}: test was called on ${JSON.stringify(ours.calls)}`);
    }
});

check('TestAdapterReplay', () => {
    assert.equal(plan(before, debugBefore), after, 'the adapter on -before is not -after');
    assert.equal(plan(after, debugAfter), after, 'a second run changes the file');
    const refuses = (text, debug, pattern, label) => assert.throws(() => plan(text, debug), pattern, label);
    const once = (text, from, to) => {
        assert.equal(text.split(from).length, 2, `one ${from}`);
        return text.replace(from, to);
    };
    const between = once(before, 'setTextRangePosEnd(updatedArray, pos, end);\r\n', 'setTextRangePosEnd(updatedArray, pos, end);\r\n        void 0;\r\n');
    refuses(between, debugBefore, /visitNodes: unreviewed owner text/, 'a statement between setTextRangePosEnd and the return');
    const workerAfter = pins.owners.visitArrayWorker.after, workerBefore = pins.owners.visitArrayWorker.before;
    refuses(once(before, workerBefore, workerAfter), debugBefore, /one owner adapted and the other not/, 'a half-adapted file');
    refuses(before, once(debugBefore, '() => `Node array did not pass test', '() => `Nodes did not pass test'), /no longer says/, 'drifted debug.ts words');
    refuses(once(before, 'Debug.assertNode(visited, test);', 'Debug.assertNode(visited, test, "message");'), debugBefore, /unreviewed owner text/, 'a drifted assertion');
});

console.log(failed === 0 ? 'test.cjs: PASS' : `test.cjs: FAIL (${failed})`);
process.exit(failed === 0 ? 0 : 1);
