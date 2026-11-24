'use strict';

const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

if (ts.version !== '6.0.3' || process.argv.length !== 3) {
    throw new Error('usage: adapt.cjs <pinned adapted tree>, with TypeScript 6.0.3');
}

const tree = process.argv[2];
const slots = {
    'builder.ts': [
        ['hasSameKeys', 'map1', 'ReadonlyCollection<string> | undefined', 'ReadonlyCollection<K> | undefined'],
        ['hasSameKeys', 'map2', 'ReadonlyCollection<string> | undefined', 'ReadonlyCollection<K> | undefined'],
    ],
    'checker.ts': [
        ['getEmitResolver', 'sourceFile', 'SourceFile', 'SourceFile | undefined'],
        ['getEmitResolver', 'cancellationToken', 'CancellationToken', 'CancellationToken | undefined'],
        ['getDiagnostics', 'sourceFile', 'SourceFile', 'SourceFile | undefined'],
        ['getDiagnostics', 'ct', 'CancellationToken', 'CancellationToken | undefined'],
        ['getDiagnosticsWorker', 'sourceFile', 'SourceFile', 'SourceFile | undefined'],
        ['canReuseTypeNodeAnnotation', 'symbol', 'Symbol', 'Symbol | undefined'],
        ['getJsxFactoryEntity', 'location', 'Node', 'Node | undefined'],
        ['getJsxFragmentFactoryEntity', 'location', 'Node', 'Node | undefined'],
    ],
    'emitter.ts': [
        ['writeList', 'nodes', 'NodeArray<T>', 'NodeArray<T> | undefined'],
    ],
    'factory/emitHelpers.ts': [
        ['createAwaiterHelper', 'parameters', 'readonly ParameterDeclaration[]', 'readonly ParameterDeclaration[] | undefined'],
    ],
    'factory/nodeFactory.ts': [
        ['updateBinaryExpression', 'operator', 'BinaryOperatorToken', 'BinaryOperator | BinaryOperatorToken'],
        ['updateYieldExpression', 'expression', 'Expression', 'Expression | undefined'],
        ['updateJSDocTypeLiteral', 'isArrayType', 'boolean', 'boolean | undefined'],
        ['copyPrologue', 'visitor', '(node: Node) => VisitResult<Node>', '(node: Node) => VisitResult<Node | undefined>'],
        ['copyCustomPrologue', 'visitor', '(node: Node) => VisitResult<Node>', '(node: Node) => VisitResult<Node | undefined>', 3],
        ['replaceModifiers', 'modifiers', 'readonly Modifier[] | ModifierFlags', 'readonly Modifier[] | ModifierFlags | undefined', 2],
        ['replaceDecoratorsAndModifiers', 'modifiers', 'readonly ModifierLike[]', 'readonly ModifierLike[] | undefined'],
        ['replaceDecoratorsAndModifiers', 'modifierArray', 'readonly ModifierLike[]', 'readonly ModifierLike[] | undefined'],
    ],
    'factory/parenthesizerRules.ts': [
        ['getParenthesizeLeftSideOfBinaryForOperator', 'operatorKind', 'BinaryOperator', 'SyntaxKind'],
        ['getParenthesizeRightSideOfBinaryForOperator', 'operatorKind', 'BinaryOperator', 'SyntaxKind'],
    ],
};

const overloads = [
    'function writeNode(hint: EmitHint, node: TypeNode, sourceFile: undefined, output: EmitTextWriter): void;',
    'function writeNode(hint: EmitHint, node: Node, sourceFile: SourceFile, output: EmitTextWriter): void;',
];

function plan(relative, entries) {
    const file = path.join(tree, 'src/compiler', relative);
    const before = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    const nodes = [];
    function visit(node) {
        nodes.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    const edits = [];
    function expect(matches, count, label) {
        if (matches.length !== count) throw new Error(`${relative}: expected ${count} ${label}, found ${matches.length}`);
        return matches;
    }
    function edit(node, oldText, newText) {
        const actual = node.getText(source);
        if (actual !== oldText && actual !== newText) {
            throw new Error(`${relative}: expected ${JSON.stringify(oldText)}, saw ${JSON.stringify(actual)}`);
        }
        if (actual !== newText) edits.push([node.getStart(source), node.end, newText]);
    }
    for (const [owner, member, oldType, newType, count = 1] of entries) {
        const matches = nodes.filter(node =>
            (ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node)) && node.name?.getText(source) === owner)
            .flatMap(node => node.parameters.filter(parameter => parameter.name.getText(source) === member));
        for (const parameter of expect(matches, count, `${owner}.${member}`)) edit(parameter.type, oldType, newType);
    }

    if (relative === 'builder.ts') {
        const [owner] = expect(nodes.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'hasSameKeys'), 1, 'hasSameKeys');
        if (owner.typeParameters) {
            if (owner.typeParameters.length !== 1 || owner.typeParameters[0].getText(source) !== 'K') throw new Error('hasSameKeys type parameter drift');
        } else {
            edits.push([owner.name.end, owner.name.end, '<K>']);
        }
    }
    if (relative === 'emitter.ts') {
        const candidates = nodes.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'writeNode');
        expect(candidates.filter(node => node.body), 1, 'writeNode implementation');
        const signatures = candidates.filter(node => !node.body);
        if (signatures.length) {
            expect(signatures, 2, 'writeNode overloads');
            const [comment] = expect(signatures[0].jsDoc || [], 1, 'writeNode overload JSDoc');
            const oldComment = [
                '/**',
                '     * If `sourceFile` is `undefined`, `node` must be a synthesized `TypeNode`.',
                '     */',
            ].join(before.includes('\r\n') ? '\r\n' : '\n');
            edit(comment, oldComment, oldComment.replace(/[^\r\n]/g, ''));
            signatures.forEach((node, index) => edit(node, overloads[index], ''));
        }
    }
    if (relative === 'factory/nodeFactory.ts') {
        const [arrow] = expect(nodes.filter(node => ts.isArrowFunction(node) && node.body.getText(source) === 'setEmitFlags(createPropertyAccessChain(expression, questionDotToken, name), EmitFlags.NoIndentation)'), 1, 'no-indentation forwarder');
        const [name] = expect(arrow.parameters.filter(node => node.name.getText(source) === 'name'), 1, 'forwarded name');
        edit(name.type, 'string', 'string | MemberName');
    }
    if (relative === 'factory/parenthesizerRules.ts') {
        for (const name of ['binaryLeftOperandParenthesizerCache', 'binaryRightOperandParenthesizerCache']) {
            const [declaration] = expect(nodes.filter(node => ts.isVariableDeclaration(node) && node.name.getText(source) === name), 1, name);
            edit(declaration.type, 'Map<BinaryOperator, (node: Expression) => Expression> | undefined', 'Map<SyntaxKind, (node: Expression) => Expression> | undefined');
        }
    }

    let after = before;
    let boundary = before.length;
    for (const [start, end, replacement] of edits.sort((a, b) => b[0] - a[0])) {
        if (end > boundary) throw new Error(`overlapping edits in ${relative}`);
        after = after.slice(0, start) + replacement + after.slice(end);
        boundary = start;
    }
    return {file, before, after, edits: edits.length};
}

const plans = Object.entries(slots).map(([file, entries]) => plan(file, entries));
for (const {file, before, after, edits} of plans) {
    if (before !== after) fs.writeFileSync(file, after);
    console.log(`${path.relative(tree, file)}: ${edits} edits`);
}
