'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const { compiler, nodes } = require('../47-host-errors/adapt.cjs');
assert.equal(ts.version, '6.0.3');
function adapt(tree) {
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const plans = [], evidence = [];
    function source(relative) {
        return program.getSourceFile(path.join(tree, relative)) || ts.createSourceFile(path.join(tree, relative), fs.readFileSync(path.join(tree, relative), 'utf8'), ts.ScriptTarget.Latest, true);
    }
    function one(root, predicate, label) {
        const found = nodes(ts, root, predicate);
        assert.equal(found.length, 1, label);
        return found[0];
    }
    function change(file, node, before, after) {
        assert([before, after].includes(node.getText(file)), 'unreviewed type: ' + node.getText(file));
        if (node.getText(file) === after) return;
        const location = file.getLineAndCharacterOfPosition(node.getStart(file));
        evidence.push({file:path.relative(tree,file.fileName),line:location.line+1,column:location.character+1,before,after});
        plans.push({file,start:node.getStart(file),end:node.end,text:after});
    }
    const sys = source('src/compiler/sys.ts');
    const system = one(sys, n => ts.isInterfaceDeclaration(n) && n.name.text === 'System', 'System');
    const exit = one(system, n => ts.isMethodSignature(n) && n.name.getText(sys) === 'exit', 'System.exit');
    const signature = checker.getSignatureFromDeclaration(exit);
    assert(signature, 'unresolved System.exit');
    const type = checker.getReturnTypeOfSignature(signature);
    const valid = type.flags === ts.TypeFlags.Void || (type.isUnion() && type.types.length === 2 && type.types.some(t => t.flags === ts.TypeFlags.Number) && type.types.some(t => t.flags === ts.TypeFlags.Undefined));
    assert(valid, 'System.exit must be exactly number | undefined (or the original void)');
    change(sys, exit.type, 'void', 'number | undefined');
    const nodeExit = one(sys, n => ts.isMethodDeclaration(n) && n.name.getText(sys) === 'exit', 'nodeSystem.exit');
    assert.equal(nodeExit.body.getText(sys).replace(/\r\n/g,'\n'), '{\n                disableCPUProfiler(() => process.exit(exitCode));\n            }', 'Node exit implementation drift');
    change(sys, nodeExit.type, 'void', 'undefined');
    const core = source('src/compiler/core.ts');
    const noop = one(core, n => ts.isFunctionDeclaration(n) && n.name?.text === 'noop', 'noop');
    assert.equal(noop.body.statements.length, 0, 'noop must return undefined');
    change(core, noop.type, 'void', 'undefined');
    const cli = source('src/compiler/executeCommandLine.ts');
    const execute = one(cli, n => ts.isFunctionDeclaration(n) && n.name?.text === 'executeCommandLine', 'executeCommandLine');
    const oldReturn = 'void | SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>';
    change(cli, execute.type, oldReturn, oldReturn.replace('void', 'number | undefined'));
    const exitCalls = nodes(ts, cli, n => ts.isReturnStatement(n) && n.expression && ts.isCallExpression(n.expression) && n.expression.expression.getText(cli) === 'sys.exit');
    assert.equal(exitCalls.length, 20, 'exit return census drift');
    for (const call of exitCalls) assert.equal(checker.getResolvedSignature(call.expression)?.declaration, exit, 'exit does not resolve to System.exit');
    for (const relative of ['src/harness/fakesHosts.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts']) {
        const file = source(relative);
        const method = one(file, n => ts.isMethodDeclaration(n) && n.name.getText(file) === 'exit', relative + ' exit');
        assert.equal(method.body.statements.length,2,'throwing exit drift');
        assert(ts.isThrowStatement(method.body.statements[1]),'exit must throw');
        change(file, method.type, 'void', 'number | undefined');
    }
    const file = source('src/compiler/checker.ts');
    const marker = one(file, n => ts.isVariableDeclaration(n) && n.name.getText(file) === 'outofbandVarianceMarkerHandler', 'marker');
    const symbol = checker.getSymbolAtLocation(marker.name);
    const references = nodes(ts, file, n => ts.isIdentifier(n) && n.text === 'outofbandVarianceMarkerHandler' && checker.getSymbolAtLocation(n) === symbol);
    assert.equal(references.length, 15, 'marker reference drift');
    const assignments = references.filter(n => ts.isBinaryExpression(n.parent) && n.parent.left === n && n.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken).map(n=>n.parent.right.getText(file).replace(/\r\n/g,'\n'));
    assert.deepEqual(assignments, [
        'onlyUnreliable => {\n                    propagatingVarianceFlags |= onlyUnreliable ? RelationComparisonResult.ReportsUnreliable : RelationComparisonResult.ReportsUnmeasurable;\n                    return originalHandler!(onlyUnreliable);\n                }',
        'originalHandler',
        'onlyUnreliable => onlyUnreliable ? unreliable = true : unmeasurable = true',
        'oldHandler',
    ], 'marker writers drift');
    const callback = one(marker.type, ts.isFunctionTypeNode, 'marker callback');
    change(file, callback.type, 'void', 'boolean');
    // Validate every result before touching any file. Type annotations alone must erase identically.
    const outputs = [];
    for (const file of new Set(plans.map(p=>p.file))) {
        let text = file.text;
        for (const edit of plans.filter(p=>p.file===file).sort((a,b)=>b.start-a.start)) text=text.slice(0,edit.start)+edit.text+text.slice(edit.end);
        for (const removeComments of [false,true]) {
            const compilerOptions = {target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};
            assert.equal(ts.transpileModule(text,{fileName:file.fileName,compilerOptions}).outputText,ts.transpileModule(file.text,{fileName:file.fileName,compilerOptions}).outputText,'runtime JavaScript changed: '+file.fileName);
        }
        outputs.push([file.fileName,text]);
    }
    for (const [file,text] of outputs) fs.writeFileSync(file,text);
    return {files:outputs.length,edits:evidence,javascript_identical:true,marker_references:references.map(n=>({line:file.getLineAndCharacterOfPosition(n.getStart(file)).line+1,parent:ts.SyntaxKind[n.parent.kind]}))};
}
module.exports = {adapt};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
