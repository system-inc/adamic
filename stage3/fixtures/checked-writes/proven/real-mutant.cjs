'use strict';
// Change the real receiver in an in-memory host. The adapted source on disk stays intact.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const { ts, options, createAudit } = require('./receiving-audit.cjs');
const tree = path.resolve(process.argv[2]);
const ledger = JSON.parse(fs.readFileSync(path.join(__dirname, 'sites.json'), 'utf8'));
const site = ledger.records.find(record => record.id === 420);
const file = path.join(tree, 'src/compiler/factory/nodeTests.ts');
const source = fs.readFileSync(file, 'utf8');
const needle = 'export function isPrivateIdentifier(node: Node): node is PrivateIdentifier {';
assert.equal(source.split(needle).length, 2);
const changed = source.replace(needle, needle + '\n    node.emitNode = undefined;');
const roots = ts.sys.readDirectory(path.join(tree, 'src'), ['.ts'], ['**/lib/**']);
const observations = [];
for (const mutant of [false, true]) {
    const host = ts.createCompilerHost(options), originalGet = host.getSourceFile;
    host.getSourceFile = (name, version, onError, fresh) => mutant && path.resolve(name) === file ? ts.createSourceFile(name, changed, version, true) : originalGet(name, version, onError, fresh);
    const program = ts.createProgram(roots, options, host), audit = createAudit(program, tree);
    const located = audit.locate(site);
    assert(located.node, located.failure || 'site not located');
    const result = audit.inspect(located.node, located.original);
    const diagnostics = program.getSemanticDiagnostics(program.getSourceFile(file)).map(diagnostic => ({ code: diagnostic.code, text: ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n') }));
    let memberProofs = [];
    if (!mutant) assert.equal(result.classification, 'a', JSON.stringify(result));
    else {
        // The historical audit declines unions even when every member rejects the value.
        assert.equal(result.classification, 'd', JSON.stringify(result));
        assert(result.writes.some(write => write.path === 'emitNode' && write.definitely_outside && !write.compatible));
        const nodeTests = program.getSourceFile(file);
        let writtenValue;
        function find(node) {
            if (ts.isBinaryExpression(node) && node.left.getText(nodeTests) === 'node.emitNode' && node.operatorToken.kind === ts.SyntaxKind.EqualsToken) writtenValue = node.right;
            ts.forEachChild(node, find);
        }
        find(nodeTests);
        assert(writtenValue);
        const writtenType = audit.checker.getTypeAtLocation(writtenValue);
        assert(located.original.isUnion());
        memberProofs = located.original.types.map(member => {
            const field = member.getProperty('emitNode');
            assert(field);
            const domain = audit.checker.getTypeOfSymbolAtLocation(field, located.node);
            const compatible = audit.checker.isTypeAssignableTo(writtenType, domain);
            assert.equal(compatible, false, 'out-of-type value must fail every original union member');
            return { member: audit.type(member), domain: audit.type(domain), value: audit.type(writtenType), compatible };
        });
        assert.equal(memberProofs.length, 2);
    }
    observations.push({ mutant, site: site.id, where: site.where, receiver: 'src/compiler/factory/nodeTests.ts:322', routing: mutant ? 'checked' : 'proven', member_proofs: memberProofs, diagnostics, ...result });
}
assert.deepEqual(observations[1].diagnostics, observations[0].diagnostics, 'mutant must not be killed by the TypeScript checker');
fs.writeFileSync(path.join(__dirname, 'real-mutant.json'), JSON.stringify({ source_commit: ledger.adapted_source_commit, mutation: 'add node.emitNode = undefined before the return in isPrivateIdentifier', observations }, null, 2) + '\n');
console.log('PASS: real site 420 moves proven -> checked (both original union members reject undefined); TypeScript diagnostics unchanged');
if (process.argv[3] === '--expect-proven') assert.equal(observations[1].routing, 'proven', 'original proven-site contract rejects the real out-of-type write');
