'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const child = require('node:child_process');
const os = require('node:os');
const { ts, options, createAudit } = require('./receiving-audit.cjs');
const directory = __dirname;
const repository = path.resolve(directory, '../../../..');
const ledger = JSON.parse(fs.readFileSync(path.join(directory, 'sites.json'), 'utf8'));
const mode = process.argv[2];
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
assert.equal(ts.version, '6.0.3');
const original = child.execFileSync('git', ['show', `${ledger.classification_ref}:${ledger.classification_path}`], { cwd: repository, maxBuffer: 4 * 1024 * 1024 });
assert.equal(digest(original), ledger.classification_sha256, 'classification pin');
const auditSource = child.execFileSync('git', ['show', `${ledger.classification_ref}:stage3/adapt/71-writable-views/language-decision/audit.cjs`], { cwd: repository });
assert.equal(digest(fs.readFileSync(path.join(directory, 'receiving-audit.cjs'))), digest(auditSource), 'receiver audit pin');
const selected = JSON.parse(original).records.filter(record => ['a', 'b'].includes(record.classification));
const rows = structuredClone(ledger.records);
if (mode === '--drop-row') rows.pop();
assert.deepEqual(rows.map(row => row.id), selected.map(row => row.id), 'exact 98-site coverage, ordered identities');
const sourceCounts = {}, proofCounts = {};
for (let index = 0; index < selected.length; index++) {
    const row = rows[index];
    for (const key of Object.keys(selected[index])) assert.deepEqual(row[key], selected[index][key], `site ${row.id}: ${key}`);
    assert.deepEqual(row.expected, { compiles: true, writable_view_runtime_checks: 0, classification: 'proven' });
    assert.equal(row.unresolved.length, 0);
    if (row.classification === 'a') assert.equal(row.writes.length, 0);
    else assert(row.writes.length && row.writes.every(write => write.compatible && !write.original_domain_is_generic_upper_bound));
    assert.deepEqual(row.proof_required.write_obligations, row.writes);
    assert.equal(row.proof_required.kind, row.classification === 'a' ? 'no-write' : 'all-writes-fit-original-domain');
    assert(ledger.fixtures.some(fixture => fixture.file === row.witness));
    sourceCounts[row.source_family] = (sourceCounts[row.source_family] || 0) + 1;
    proofCounts[row.proof_family] = (proofCounts[row.proof_family] || 0) + 1;
}
assert.deepEqual(sourceCounts, ledger.source_families);
assert.deepEqual(proofCounts, ledger.proof_families);
assert.deepEqual(ledger.counts, { no_write: rows.filter(row => row.classification === 'a').length, compatible_write: rows.filter(row => row.classification === 'b').length, total: rows.length });
// Optional reconstruction check ties coordinates and source excerpts to the adapted tree.
const tree = process.env.PROVEN_SOURCE_TREE;
if (tree) {
    for (const [file, hash] of Object.entries(ledger.source_sha256)) assert.equal(digest(fs.readFileSync(path.join(tree, file))), hash, file);
    for (const fixture of ledger.fixtures) for (const span of fixture.spans) {
        const lines = fs.readFileSync(path.join(tree, span.file), 'utf8').split(/\r?\n/);
        assert.equal(lines.slice(span.start - 1, span.end).join('\n') + '\n', span.text, `${fixture.file}: source span`);
    }
}
const selectors = {
    '01_truthiness.a': node => ts.isIdentifier(node) && node.text === 'node' && ts.isBinaryExpression(node.parent) && node.parent.left === node,
    '02_control.a': node => ts.isPropertyAccessExpression(node) && node.getText() === 'parentAccess.flowNode' && ts.isBinaryExpression(node.parent),
    '03_kind_reader.a': node => ts.isIdentifier(node) && node.text === 'name' && ts.isCallExpression(node.parent) && node.parent.expression.getText() === 'isPrivateIdentifier',
    '04_range_reader.a': node => ts.isIdentifier(node) && node.text === 'children' && ts.isCallExpression(node.parent),
    '05_diagnostic_related.a': node => ts.isIdentifier(node) && node.text === 'diagnostic' && ts.isCallExpression(node.parent) && node.parent.expression.getText() === 'forward',
    '06_modifiers_initialize.a': node => ts.isIdentifier(node) && node.text === 'node' && ts.isAsExpression(node.parent),
    '07_assert_clause.a': node => ts.isPropertyAccessExpression(node) && node.getText() === 'node.assertions' && ts.isAsExpression(node.parent),
    '08_property_initializer.a': node => ts.isIdentifier(node) && node.text === 'node' && ts.isAsExpression(node.parent),
    '09_static_modifiers.a': node => ts.isIdentifier(node) && node.text === 'node' && ts.isAsExpression(node.parent),
    '10_module_parent.a': node => ts.isPropertyAccessExpression(node) && node.getText() === 'node.name' && ts.isAsExpression(node.parent),
};
// Compiler workers can supply observations for every original site after checked writes land.
if (process.env.PROVEN_COMPILER_RESULTS) {
    const measured = JSON.parse(fs.readFileSync(process.env.PROVEN_COMPILER_RESULTS, 'utf8'));
    assert.match(measured.adamic_sha, /^[a-f0-9]{40}$/);
    assert.equal(measured.adapted_source_commit, ledger.adapted_source_commit);
    assert.deepEqual(measured.records.map(row => row.id), rows.map(row => row.id), 'compiler result coverage');
    measured.records.forEach((record, index) => {
        assert.equal(record.where, rows[index].where);
        assert.equal(record.compiles, true, `site ${record.id}: compile required`);
        assert.equal(record.writable_view_runtime_checks, 0, `site ${record.id}: static proof must elide checks`);
    });
}
const status = JSON.parse(fs.readFileSync(path.join(directory, 'status.json'), 'utf8'));
assert.deepEqual(status.map(row => row.file), ledger.fixtures.map(fixture => fixture.file), 'fixture status coverage');
const observations = [];
for (const fixture of ledger.fixtures) {
    let source = fs.readFileSync(path.join(directory, fixture.file), 'utf8');
    for (const span of fixture.spans) assert(source.includes(span.text), `${fixture.file}: original statements retained`);
    if (mode === '--bad-write' && fixture.file === '03_kind_reader.a') source = source.replace('    return node.kind ===', '    node.emitNode = undefined;\n    return node.kind ===');
    if (mode === '--node-output' && fixture.file === '03_kind_reader.a') source = source.replace('autoGenerate: 7', 'autoGenerate: 8');
    const root = '/proven-fixture', virtual = path.join(root, fixture.file + '.ts');
    const host = ts.createCompilerHost(options), originalGet = host.getSourceFile;
    host.getSourceFile = (file, version, onError, fresh) => file === virtual ? ts.createSourceFile(file, source + '\ndeclare const console: { log: (text: string) => void };\n', version, true) : originalGet(file, version, onError, fresh);
    const program = ts.createProgram([virtual], options, host);
    const diagnostics = ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length, 0, `${fixture.file}: ${diagnostics.map(diagnostic => ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')).join('\n')}`);
    const audit = createAudit(program, root), found = [];
    function visit(node) { if (selectors[fixture.file](node)) found.push(node); ts.forEachChild(node, visit); }
    visit(program.getSourceFile(virtual));
    assert(found.length, fixture.file);
    const expected = selected.find(row => row.id === fixture.representative_id).classification;
    for (const node of found) {
        const result = audit.inspect(node, audit.checker.getTypeAtLocation(node));
        if (mode === '--bad-write' && fixture.file === '03_kind_reader.a') {
            assert.equal(result.classification, 'c', JSON.stringify(result));
            console.log('site 420 receiver mutant: proven -> checked; emitNode = undefined is outside EmitNode & { autoGenerate: number }');
        }
        assert.equal(result.classification, expected, `${fixture.file}: proven contract changed: ${JSON.stringify(result)}`);
        observations.push({ file: fixture.file, classification: result.classification, writes: result.writes });
    }
    let nodeFile = path.join(directory, fixture.file);
    if (mode === '--node-output' && fixture.file === '03_kind_reader.a') {
        nodeFile = path.join(fs.mkdtempSync(path.join(os.tmpdir(), 'proven-output-mutant-')), fixture.file);
        fs.writeFileSync(nodeFile, source);
    }
    const run = child.spawnSync(process.execPath, ['--disable-warning=ExperimentalWarning', path.join(repository, 'oracle/node.mjs'), nodeFile]);
    const expectedNode = status.find(row => row.file === fixture.file).node;
    assert.deepEqual(run.stdout, Buffer.from(expectedNode.stdout), `${fixture.file}: Node stdout bytes`);
    assert.deepEqual(run.stderr, Buffer.from(expectedNode.stderr), `${fixture.file}: Node stderr bytes`);
    assert.equal(run.status, expectedNode.exit, `${fixture.file}: Node exit`);

}
const mutants = [];
if (!mode) for (const mutation of ['--bad-write', '--drop-row', '--node-output']) {
    const result = child.spawnSync(process.execPath, [__filename, mutation], { encoding: 'utf8' });
    assert.equal(result.status, 1, `${mutation} must fail`);
    assert(result.stderr.includes('AssertionError'), result.stderr);
    assert(result.stderr.includes(mutation === '--bad-write' ? '03_kind_reader.a: proven contract changed' : mutation === '--drop-row' ? 'exact 98-site coverage' : '03_kind_reader.a: Node stdout bytes'), result.stderr);
    mutants.push({ mutation, exit: result.status, stdout: result.stdout, caught_by: mutation === '--bad-write' ? 'receiver audit first proves class c, then rejects the original no-write expectation; zero stock TypeScript diagnostics' : mutation === '--drop-row' ? 'exact pinned identity comparison' : 'source Node stdout byte comparison; driver input autoGenerate 7 -> 8' });
}
if (!mode) {
    fs.writeFileSync(path.join(directory, 'validation.json'), JSON.stringify({ typescript: ts.version, sites: rows.length, source_families: Object.keys(sourceCounts).length, proof_families: proofCounts, fixtures: observations, mutants }, null, 2) + '\n');
    console.log(`PASS: 98 sites (91 no-write, 7 compatible-write), ${Object.keys(sourceCounts).length} source families, ${ledger.fixtures.length} witnesses; three mutants caught`);
}
