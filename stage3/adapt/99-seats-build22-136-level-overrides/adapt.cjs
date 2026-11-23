'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const relativeFile = 'src/compiler/sys.ts';
const before = '{ ...defaultLevels, ...customLevels }';
const after = '{ Low: customLevels.Low ?? defaultLevels.Low, Medium: customLevels.Medium ?? defaultLevels.Medium, High: customLevels.High ?? defaultLevels.High }';
const producerSnapshot = `function getCustomLevels(baseVariable: string) {
        let customLevels: Partial<Levels> | undefined;
        setCustomLevel("Low");
        setCustomLevel("Medium");
        setCustomLevel("High");
        return customLevels;

        function setCustomLevel(level: keyof Levels) {
            const customLevel = getLevel(baseVariable, level);
            if (customLevel) {
                (customLevels || (customLevels = {}))[level] = Number(customLevel);
            }
        }
    }`;
const mergeSnapshot = `function getCustomPollingBasedLevels(baseVariable: string, defaultLevels: Levels) {
        const customLevels = getCustomLevels(baseVariable);
        return (pollingIntervalChanged || customLevels) &&
            createPollingIntervalBasedLevels(customLevels ? ${before} : defaultLevels);
    }`;
const levelsSnapshot = `interface Levels {
    Low: number;
    Medium: number;
    High: number;
}`;
const consumerSnapshot = `function createPollingIntervalBasedLevels(levels: Levels) {
    return {
        [PollingInterval.Low]: levels.Low,
        [PollingInterval.Medium]: levels.Medium,
        [PollingInterval.High]: levels.High,
    };
}`;
const defaultsSnapshot = 'defaultChunkLevels: Levels = { Low: 32, Medium: 64, High: 256 }';
const callSnapshots = [
    'getCustomPollingBasedLevels("TSC_WATCH_POLLINGCHUNKSIZE", defaultChunkLevels)',
    'getCustomPollingBasedLevels("TSC_WATCH_UNCHANGEDPOLLTHRESHOLDS", defaultChunkLevels)',
];

function one(nodes, predicate, name) {
    const matches = nodes.filter(predicate);
    assert.equal(matches.length, 1, `sys.ts: expected one ${name}`);
    return matches[0];
}

function functionNamed(statements, name) {
    const declaration = one(statements, node => ts.isFunctionDeclaration(node) && node.name?.text === name, name);
    assert(declaration.body, `sys.ts: expected a body for ${name}`);
    return declaration;
}

function inspect(text) {
    assert.equal(ts.version, '6.0.3');
    const source = ts.createSourceFile('sys.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'sys.ts: parse error');
    const spelling = node => node.getText(source).replace(/\r\n/g, '\n');
    const statements = source.statements;
    const owner = functionNamed(statements, 'setCustomPollingValues');
    const producer = functionNamed(owner.body.statements, 'getCustomLevels');
    assert.equal(spelling(producer), producerSnapshot, 'sys.ts: getCustomLevels producer changed');

    const levels = one(statements, node => ts.isInterfaceDeclaration(node) && node.name.text === 'Levels', 'Levels');
    assert.equal(spelling(levels), levelsSnapshot, 'sys.ts: Levels changed');
    const declarations = statements.filter(ts.isVariableStatement).flatMap(node => [...node.declarationList.declarations]);
    const defaults = one(declarations, node => ts.isIdentifier(node.name) && node.name.text === 'defaultChunkLevels', 'defaultChunkLevels');
    assert.equal(spelling(defaults), defaultsSnapshot, 'sys.ts: defaultChunkLevels changed');
    assert.equal(spelling(functionNamed(statements, 'createPollingIntervalBasedLevels')), consumerSnapshot,
        'sys.ts: createPollingIntervalBasedLevels changed');

    const merge = functionNamed(owner.body.statements, 'getCustomPollingBasedLevels');
    const calls = [];
    function visit(node) {
        if (ts.isIdentifier(node) && node.text === merge.name.text && node !== merge.name) {
            assert(ts.isCallExpression(node.parent) && node.parent.expression === node,
                'sys.ts: unexpected getCustomPollingBasedLevels reference');
            calls.push(spelling(node.parent));
        }
        ts.forEachChild(node, visit);
    }
    visit(owner.body);
    assert.deepEqual(calls, callSnapshots, 'sys.ts: getCustomPollingBasedLevels callers changed');

    const returned = one(merge.body.statements, ts.isReturnStatement, 'merge return');
    const expression = returned.expression;
    assert(expression && ts.isBinaryExpression(expression) && expression.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken);
    assert(ts.isCallExpression(expression.right) && expression.right.arguments.length === 1);
    const conditional = expression.right.arguments[0];
    assert(ts.isConditionalExpression(conditional) && ts.isObjectLiteralExpression(conditional.whenTrue));
    const literal = conditional.whenTrue;
    const adapted = spelling(literal) === after;
    assert.equal(spelling(merge), adapted ? mergeSnapshot.replace(before, after) : mergeSnapshot,
        'sys.ts: getCustomPollingBasedLevels merge changed');
    return { adapted, producer, conditional, literal, defaults };
}

function adapt(tree) {
    const file = path.join(tree, relativeFile);
    const text = fs.readFileSync(file, 'utf8');
    const { adapted, literal } = inspect(text);
    if (adapted) return { files: 0, edits: 0, javascriptChanging: true };
    const next = text.slice(0, literal.getStart()) + after + text.slice(literal.end);
    assert(inspect(next).adapted, 'sys.ts: invalid planned merge');
    assert.equal(fs.readFileSync(file, 'utf8'), text, 'sys.ts: changed while planning');
    fs.writeFileSync(file, next);
    return { files: 1, edits: 1, javascriptChanging: true };
}

module.exports = { adapt, inspect, relativeFile };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(process.argv[2]), null, 2));
}
