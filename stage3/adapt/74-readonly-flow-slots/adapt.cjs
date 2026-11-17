#!/usr/bin/env node
'use strict';
// Adaptation 74, #5pstmvw: a flow node's node and antecedent in src/compiler/types.ts become readonly on FlowNodeBase and
// on every arm of the FlowNode union, except FlowLabel's antecedent, which a label writes as it gathers its antecedents.
// tsc reads a union's property as read only when any arm's is, so a flow node seen as FlowNode, or as a union of some of
// its arms, is a view that can no longer write one arm's node or antecedent into another arm that reads it narrower
// (39 refusals at the scan of 03c140ec: binder.ts 24, checker.ts 15). Every flow node gets its node from createFlowNode
// at construction, but for one: binder.ts's bindContainer makes a FlowStart with no node and assigns the function on the
// next statement, through currentFlow, typed FlowNode. This adaptation passes that node to createFlowNode instead, which
// makes the same object with the same keys and values, so the emitted JavaScript changes by that statement pair and
// nothing else, as Adamic ruled on #5pstmvw. README.md says what certifies it, what the count of writes found, and the
// mutant that brings the refusals back. Both slots are internal, so the published declarations don't change. The edits
// are found through the stock AST, never by a line: each refuses unless it finds exactly upstream's declaration or
// statements, and a second run changes nothing. It sorts after 72-readonly-diagnostic-slots and before
// 75-optional-widening, so every later adaptation and every consumer reads apply's tree with the slots readonly.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);

// Each flow interface's node and antecedent, as upstream 6.0.3 declares them. A slot whose type has moved is not the slot
// Adamic ruled on. writable names the one slot that stays writable.
const flowSlots = {
    FlowNodeBase: { node: 'unknown', antecedent: 'FlowNode | FlowNode[] | undefined' },
    FlowUnreachable: { node: 'undefined', antecedent: 'undefined' },
    FlowStart: { node: 'FunctionExpression | ArrowFunction | MethodDeclaration | GetAccessorDeclaration | SetAccessorDeclaration | undefined', antecedent: 'undefined' },
    FlowLabel: { node: 'undefined', antecedent: 'FlowNode[] | undefined', writable: 'antecedent' },
    FlowAssignment: { node: 'Expression | VariableDeclaration | BindingElement', antecedent: 'FlowNode' },
    FlowCall: { node: 'CallExpression', antecedent: 'FlowNode' },
    FlowCondition: { node: 'Expression', antecedent: 'FlowNode' },
    FlowSwitchClause: { node: 'FlowSwitchClauseData', antecedent: 'FlowNode' },
    FlowArrayMutation: { node: 'CallExpression | BinaryExpression', antecedent: 'FlowNode' },
    FlowReduceLabel: { node: 'FlowReduceLabelData', antecedent: 'FlowNode' },
};

function parse(relative) {
    const fileName = path.join(tree, relative);
    const text = fs.readFileSync(fileName, 'utf8');
    return { relative, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true) };
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

// types.ts: readonly on every node and antecedent but FlowLabel's antecedent.
const types = parse('src/compiler/types.ts');
const interfaces = new Map();
ts.forEachChild(types.source, function find(node) {
    if (ts.isInterfaceDeclaration(node) && Object.hasOwn(flowSlots, node.name.text)) {
        if (interfaces.has(node.name.text)) throw new Error(`expected one ${node.name.text} in types.ts, found more`);
        interfaces.set(node.name.text, node);
    }
    ts.forEachChild(node, find);
});
const typeEdits = [];
const slots = [];
for (const [name, expected] of Object.entries(flowSlots)) {
    const declaration = interfaces.get(name);
    if (!declaration) throw new Error(`expected interface ${name} in types.ts`);
    for (const slot of ['node', 'antecedent']) {
        const members = declaration.members.filter(member => ts.isPropertySignature(member) && ts.isIdentifier(member.name) && member.name.text === slot);
        if (members.length !== 1) throw new Error(`expected one property signature ${name}.${slot}, found ${members.length}`);
        const member = members[0];
        if (!member.type || member.type.getText(types.source) !== expected[slot] || member.questionToken)
            throw new Error(`unexpected slot ${name}.${slot}: ${member.getText(types.source)}`);
        const readonly = (ts.getCombinedModifierFlags(member) & ts.ModifierFlags.Readonly) !== 0;
        if (expected.writable === slot) {
            if (readonly) throw new Error(`${name}.${slot} must stay writable: a label writes it`);
            slots.push({ slot: `${name}.${slot}`, where: where(types, member), readonly: false });
            continue;
        }
        // Idempotent: a slot already readonly keeps its bytes, so a second run changes nothing.
        slots.push({ slot: `${name}.${slot}`, where: where(types, member), readonly: true, edited: !readonly });
        if (!readonly) typeEdits.push({ start: member.getStart(types.source), end: member.getStart(types.source), text: 'readonly ' });
    }
}

// binder.ts: the FlowStart takes its node at construction. Upstream's two statements,
//     currentFlow = createFlowNode(FlowFlags.Start, /*node*/ undefined, /*antecedent*/ undefined);
//     if (containerFlags & (ContainerFlags.IsFunctionExpression | ContainerFlags.IsObjectLiteralOrClassExpressionMethodOrAccessor)) {
//         currentFlow.node = node as FunctionExpression | ArrowFunction | MethodDeclaration | GetAccessorDeclaration | SetAccessorDeclaration;
//     }
// become one, with the same condition choosing the node createFlowNode is given:
//     currentFlow = createFlowNode(FlowFlags.Start, /*node*/ containerFlags & (...) ? node as ... : undefined, /*antecedent*/ undefined);
// createFlowNode reads nothing of the node between (Debug.attachFlowNodeDebugInfo sets a prototype, or defines lazy
// getters, when debug info is on), so the object, its keys and its values are the same once the pair has run.
const binder = parse('src/compiler/binder.ts');
const startCall = 'createFlowNode(FlowFlags.Start, /*node*/ undefined, /*antecedent*/ undefined)';
const condition = 'containerFlags & (ContainerFlags.IsFunctionExpression | ContainerFlags.IsObjectLiteralOrClassExpressionMethodOrAccessor)';
const cast = 'node as FunctionExpression | ArrowFunction | MethodDeclaration | GetAccessorDeclaration | SetAccessorDeclaration';
const adaptedCall = `createFlowNode(FlowFlags.Start, /*node*/ ${condition} ? ${cast} : undefined, /*antecedent*/ undefined)`;
const pairs = [];
const already = [];
ts.forEachChild(binder.source, function find(node) {
    if (ts.isBlock(node)) {
        node.statements.forEach((statement, index) => {
            if (!ts.isExpressionStatement(statement) || !ts.isBinaryExpression(statement.expression)) return;
            const assignment = statement.expression;
            if (assignment.operatorToken.kind !== ts.SyntaxKind.EqualsToken || !ts.isIdentifier(assignment.left) || assignment.left.text !== 'currentFlow') return;
            const right = assignment.right.getText(binder.source);
            if (right === adaptedCall) already.push(statement);
            if (right !== startCall) return;
            const next = node.statements[index + 1];
            if (!next || !ts.isIfStatement(next) || next.elseStatement || next.expression.getText(binder.source) !== condition) return;
            const body = ts.isBlock(next.thenStatement) ? next.thenStatement.statements : [next.thenStatement];
            if (body.length !== 1 || body[0].getText(binder.source) !== `currentFlow.node = ${cast};`) return;
            pairs.push({ statement, next });
        });
    }
    ts.forEachChild(node, find);
});
const binderEdits = [];
let moved;
if (pairs.length === 1 && already.length === 0) {
    const { statement, next } = pairs[0];
    moved = { where: where(binder, statement), edited: true };
    binderEdits.push({ start: statement.getStart(binder.source), end: statement.getEnd(), text: `currentFlow = ${adaptedCall};` });
    // The if statement goes whole, with the line it stood on.
    const lineStart = binder.text.lastIndexOf('\n', next.getStart(binder.source)) + 1;
    const lineEnd = binder.text.indexOf('\n', next.getEnd());
    binderEdits.push({ start: lineStart, end: lineEnd + 1, text: '' });
}
else if (pairs.length === 0 && already.length === 1) moved = { where: where(binder, already[0]), edited: false };
else throw new Error(`expected bindContainer's FlowStart pair once in binder.ts, found ${pairs.length} pairs and ${already.length} adapted statements`);

// Every check passes before either file is written.
write(types, typeEdits);
write(binder, binderEdits);
console.log(JSON.stringify({ typescript: ts.version, task: '#5pstmvw', certificate: "the compiled program's gate", slots,
    flowStart: { ...moved, statement: `currentFlow = ${adaptedCall};` } }, null, 2));
