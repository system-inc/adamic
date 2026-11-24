'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const parser = process.env.TSC_ADAPT_TYPESCRIPT || path.join(__dirname, '../../api/node_modules/typescript');
const ts = require(parser);

assert.equal(ts.version, '6.0.3');
const [tree, mode] = process.argv.slice(2);
assert(tree && (mode === 'printer' || mode === 'trace'), 'usage: signatures.cjs TREE printer|trace');

function find(source, predicate, label, count = 1) {
    const matches = [];
    function visit(node) {
        if (predicate(node)) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, count, label);
    return matches;
}

const compiler = path.resolve(tree, 'src/compiler');
if (mode === 'printer') {
    const file = path.join(compiler, 'emitter.ts');
    const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'emitter syntax');
    const [printer] = find(source, node => ts.isFunctionDeclaration(node) && node.name?.text === 'createPrinter', 'createPrinter owner');
    const local = name => find(printer.body, node =>
        ts.isFunctionDeclaration(node) && node.name?.text === name && node.parent === printer.body,
        `createPrinter.${name} must have one implementation`)[0];
    const writeList = local('writeList');
    const writeNode = local('writeNode');
    assert(writeList.body && writeNode.body, 'printer backing functions have bodies');
    assert.equal(writeList.parameters[1].name.getText(source), 'nodes');
    assert.equal(writeList.parameters[1].type.getText(source), 'NodeArray<T> | undefined', 'writeList.nodes');
    assert.equal(writeNode.parameters[1].type.getText(source), 'Node', 'writeNode.node');
    assert.equal(writeNode.parameters[2].type.getText(source), 'SourceFile | undefined', 'writeNode.sourceFile');
    console.log('PASS printer signatures');
} else {
    const project = path.resolve(tree, 'src/tsconfig-base.json');
    const config = ts.readConfigFile(project, ts.sys.readFile);
    assert(!config.error, 'read upstream TypeScript configuration');
    const {options, errors} = ts.convertCompilerOptionsFromJson(config.config.compilerOptions, path.dirname(project));
    assert.equal(errors.length, 0, 'upstream TypeScript compiler options');
    const typeRoots = path.resolve(path.dirname(require.resolve(parser)), '../../@types');
    const program = ts.createProgram([path.resolve(tree, 'src/tsc/tsc.ts')], {
        ...options, noEmit: true, composite: false, isolatedDeclarations: false, types: ['node'],
        typeRoots: [typeRoots],
    });
    const checker = program.getTypeChecker();
    const format = ts.TypeFormatFlags.NoTruncation;
    const signatureText = signature => checker.signatureToString(signature, undefined, format);
    const signatureOf = node => {
        const signatures = checker.getTypeAtLocation(node).getCallSignatures();
        assert.equal(signatures.length, 1, node.getText());
        const [signature] = signatures;
        return signature;
    };
    const parameter = (signature, index) => checker.getTypeOfSymbolAtLocation(signature.parameters[index], signature.getDeclaration());
    const typeText = type => checker.typeToString(type, undefined, format);
    const describe = (label, call) => {
        const resolved = checker.getResolvedSignature(call);
        const declaration = resolved.getDeclaration();
        console.log(label, call.getText());
        console.log('resolved', signatureText(resolved));
        console.log('declaration', path.relative(compiler, declaration.getSourceFile().fileName), declaration.getText());
        return resolved;
    };
    const source = file => {
        const result = program.getSourceFile(path.join(compiler, file));
        assert(result, file);
        return result;
    };
    const [reduce] = find(source('checker.ts'), node => ts.isCallExpression(node) &&
        node.expression.getText() === 'reduceLeft' && node.arguments[1]?.getText() === '(facts, t) => facts | getTypeFactsWorker(t, callerOnlyNeeds)', 'seeded reducer');
    const reduceResolved = describe('reducer', reduce);
    const reduceCallback = signatureOf(reduce.arguments[1]);
    const foundAccumulator = parameter(reduceCallback, 0);
    const expectedCallback = parameter(reduceResolved, 1).getCallSignatures()[0];
    const passedAccumulator = parameter(expectedCallback, 0);
    console.log('actual callback', signatureText(reduceCallback));
    console.log('slot callback', signatureText(expectedCallback));
    console.log('slot -> actual, actual -> slot', checker.isTypeAssignableTo(passedAccumulator, foundAccumulator), checker.isTypeAssignableTo(foundAccumulator, passedAccumulator));

    const [predicate] = find(source('transformers/destructuring.ts'), node => ts.isCallExpression(node) &&
        node.expression.getText() === 'Debug.assertNode' && ts.isConditionalExpression(node.arguments[1]) &&
        node.arguments[1].condition.getText() === 'createAssignmentCallback', 'conditional assertion');
    const predicateResolved = describe('predicate', predicate);
    const conditional = predicate.arguments[1];
    const branches = [conditional.whenTrue, conditional.whenFalse];
    const branchSignatures = branches.map(signatureOf).map(signatureText);
    branches.forEach((branch, i) => console.log(branch.getText(), branchSignatures[i]));
    const common = signatureText(signatureOf(conditional));
    console.log('common', common);
    console.log('test slot', typeText(parameter(predicateResolved, 1)));
    console.log('condition', typeText(checker.getTypeAtLocation(conditional.condition)));

    assert.equal(path.basename(reduceResolved.getDeclaration().getSourceFile().fileName), 'core.ts');
    assert.equal(reduceResolved.getDeclaration().getText(), 'export function reduceLeft<T, U>(array: readonly T[] | undefined, f: (memo: U, value: T, i: number) => U, initial: U, start?: number, count?: number): U;', 'seeded overload');
    assert.equal(typeText(foundAccumulator), 'TypeFacts.None', 'reducer accumulator');
    assert.equal(typeText(passedAccumulator), 'number', 'reducer slot accumulator');
    assert(checker.isTypeAssignableTo(passedAccumulator, foundAccumulator), 'stock slot-to-accumulator relation');
    assert(checker.isTypeAssignableTo(foundAccumulator, passedAccumulator), 'stock accumulator-to-slot relation');
    assert.equal(path.basename(predicateResolved.getDeclaration().getSourceFile().fileName), 'debug.ts');
    assert.equal(predicateResolved.getDeclaration().getText(), 'export function assertNode<T extends Node, U extends T>(node: T | undefined, test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts node is U;', 'assertNode overload');
    assert.deepEqual(branchSignatures, ['(node: Node): node is Identifier', '(node: Node): node is Expression'], 'predicate branches');
    assert.equal(common, '(node: Node): node is Expression', 'predicate common signature');
    assert.equal(typeText(parameter(predicateResolved, 1)), '(node: Node) => node is Expression', 'assertion slot');
    console.log('PASS held signature traces');
}
