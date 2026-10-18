// Independent source coverage/role audit plus deliberately corrupted ledgers.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const root = path.resolve(process.argv[2]);
const data = JSON.parse(fs.readFileSync(path.join(__dirname, 'inherited-key-global.json')));
const reviewed = JSON.parse(fs.readFileSync(path.join(__dirname, 'inherited-key-provenance.json')));
const original = JSON.parse(cp.execFileSync('git', ['show', 'origin/codex/tsc-census:stage3/census/data/files.json'], {cwd: path.resolve(__dirname, '../../..'), encoding: 'utf8'})).filter(f => !f.generated && f.file.endsWith('.ts'));
const oldSites = JSON.parse(cp.execFileSync('git', ['show', 'origin/codex/tsc-census:stage3/census/data/string_lookups.json'], {cwd: path.resolve(__dirname, '../../..'), encoding: 'utf8'}));
const expected = [];
const loops = [];
function unwrap(node) {
    while (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) || ts.isNonNullExpression(node)) node = node.expression;
    return node;
}
function literal(node) {
    node = unwrap(node);
    return ts.isStringLiteralLike(node) || ts.isNumericLiteral(node) || ts.isPrefixUnaryExpression(node) && ts.isNumericLiteral(node.operand);
}
for (const entry of original) {
    const text = fs.readFileSync(path.join(root, entry.file), 'utf8');
    assert.equal(crypto.createHash('sha256').update(text).digest('hex'), entry.sha256, 'Pinned source hash: ' + entry.file);
    const file = ts.createSourceFile(entry.file, text, ts.ScriptTarget.Latest, true);
    function location(node) {
        const start = node.getStart(file);
        const p = file.getLineAndCharacterOfPosition(start);
        return {file: entry.file, line: p.line + 1, column: p.character + 1, start, end: node.end};
    }
    function add(node, kind, key, role) {
        const l = location(node);
        const context = [];
        // Reconstruct contexts from parent pointers, independently of the scanner's stack.
        let child = node;
        for (let parent = node.parent; parent; child = parent, parent = parent.parent) {
            if (ts.isForInStatement(parent) && child === parent.statement) {
                const p = location(parent);
                context.unshift(p.file + ':' + p.line + ':' + p.column);
            }
        }
        expected.push({...l, id: l.file + ':' + l.line + ':' + l.column + '@' + l.end, kind, role, dynamic: !literal(key), for_in: context});
    }
    function visit(node) {
        if (ts.isForInStatement(node)) {
            const l = location(node);
            loops.push(l.file + ':' + l.line + ':' + l.column);
        }
        if (ts.isElementAccessExpression(node)) {
            const assignmentKind = ts.getAssignmentTargetKind(node);
            let outer = node;
            while (ts.isParenthesizedExpression(outer.parent) || ts.isNonNullExpression(outer.parent) || ts.isAsExpression(outer.parent)) outer = outer.parent;
            let role = assignmentKind === 2 ? 'read_write' : assignmentKind === 1 ? 'write' : 'read';
            if (ts.isDeleteExpression(outer.parent)) role = 'delete';
            if (assignmentKind === 1 && ts.isBinaryExpression(outer.parent) && outer.parent.operatorToken.kind !== ts.SyntaxKind.EqualsToken) role = 'read_write';
            add(node, 'element', node.argumentExpression, role);
        }
        if (ts.isCallExpression(node)) {
            const callee = unwrap(node.expression);
            if (ts.isIdentifier(callee) && ['hasProperty', 'getProperty'].includes(callee.text)) add(node, callee.text, node.arguments[1], callee.text === 'hasProperty' ? 'membership' : 'read');
            if (ts.isPropertyAccessExpression(callee) && callee.name.text === 'call' && callee.expression.getText(file).endsWith('hasOwnProperty')) add(node, 'hasOwnProperty.call', node.arguments[1], node.arguments[0] ? 'membership' : 'read');
            if (ts.isPropertyAccessExpression(callee) && callee.expression.getText(file) === 'Object' && callee.name.text === 'hasOwn') add(node, 'Object.hasOwn', node.arguments[1], 'membership');
        }
        if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.InKeyword) add(node, 'in', node.left, 'membership');
        ts.forEachChild(node, visit);
    }
    visit(file);
}
function check(ledger) {
    assert.equal(ledger.files.length, 77, 'Original file coverage');
    assert.equal(ledger.diagnostics.length, 0, 'Stock checker gate');
    const actual = new Map(ledger.sites.map(r => [r.id, r]));
    assert.equal(actual.size, ledger.sites.length, 'Duplicate site');
    assert.equal(actual.size, expected.length, 'Site coverage');
    let counted = 0;
    const counts = {};
    const stringCounts = {};
    const census = new Map();
    for (const e of expected) {
        const r = actual.get(e.id);
        assert(r, 'Missing site: ' + e.id);
        assert.equal(r.kind, e.kind, 'Operation kind: ' + e.id);
        assert.equal(r.role, e.role, 'Read/write role: ' + e.id);
        assert.equal(r.dynamic, e.dynamic, 'Literal exclusion: ' + e.id);
        assert.deepEqual(r.for_in, e.for_in, 'for-in context: ' + e.id);
        const eligible = (!['element', 'in'].includes(e.kind) || e.dynamic) && ['read', 'read_write', 'membership'].includes(e.role);
        assert.equal(r.counted, eligible, 'Counted flag: ' + e.id);
        assert(['fixed_set_excludes_prototype', 'user_input', 'unknown'].includes(r.classification), 'Class schema');
        if (r.classification_method === 'reviewed_source') {
            const o = reviewed[r.location_id];
            assert(o, 'Unreviewed provenance override');
            assert.equal(r.classification, o.classification, 'Reviewed class');
            assert.equal(r.evidence, o.evidence, 'Reviewed evidence');
            assert.deepEqual(r.input_sources, o.input_sources, 'Reviewed input source');
        } else {
            assert(r.classification === 'fixed_set_excludes_prototype' || r.classification === 'unknown', 'Unreviewed user provenance');
            assert.equal(r.classification_method, r.classification === 'unknown' ? 'unknown' : 'checker_key_domain', 'Class evidence method');
            if (r.classification === 'unknown') assert.deepEqual(r.input_sources, [], 'Unknown source must not be invented');
        }
        if (r.classification === 'user_input') assert(r.input_sources.length > 0, 'Missing input origin');
        if (r.census_index !== null) {
            assert(!census.has(r.census_index), 'Duplicate census link');
            census.set(r.census_index, r);
        }
        if (eligible) {
            counted++;
            counts[r.classification] = (counts[r.classification] || 0) + 1;
            if (r.string_capable) stringCounts[r.classification] = (stringCounts[r.classification] || 0) + 1;
        }
    }
    assert.equal(census.size, 68, 'Census coverage');
    oldSites.forEach((row, index) => {
        const got = census.get(index);
        assert(got, 'Missing census row ' + index);
        assert.equal(got.file, row.file);
        assert.equal(got.start, row.start);
        assert.equal(got.end, row.end);
    });
    assert.deepEqual(ledger.loops.map(l => l.id).sort(), [...loops].sort(), 'Loop coverage');
    assert.equal(counted, ledger.totals.all_dynamic_reads_and_membership_checks, 'Global total');
    assert.deepEqual(counts, ledger.totals.all_classes, 'Global classes');
    assert.deepEqual(stringCounts, ledger.totals.string_capable_classes, 'String classes');
    return counted;
}
console.log('PASS: source hashes, 77 files, ' + check(data) + ' counted sites, all 68 census spans, all 27 for-in loops, all 10 literal in tests, zero non-literal in tests');
function mutant(name, mutate) {
    const copy = structuredClone(data);
    mutate(copy);
    let caught = false;
    try { check(copy); } catch (error) { caught = true; console.log('CAUGHT ' + name + ': ' + error.message.split('\n')[0]); }
    assert(caught, 'Surviving mutant: ' + name);
}
mutant('omit a dynamic element read', d => { const i = d.sites.findIndex(r => r.location_id === 'src/compiler/core.ts:1279:44'); d.sites.splice(i, 1); });
mutant('omit a hasProperty call', d => { const i = d.sites.findIndex(r => r.location_id === 'src/compiler/factory/nodeFactory.ts:6138:17'); d.sites.splice(i, 1); });
mutant('drop for-in context while leaving totals unchanged', d => { d.sites.find(r => r.counted && r.for_in.length).for_in = []; });
mutant('omit in-operator enumeration', d => { d.sites = d.sites.filter(r => r.kind !== 'in'); });
mutant('count an assignment target as a read', d => { d.sites.find(r => r.census_index === 2).counted = true; });
mutant('invent a fixed classification for an unknown', d => { d.sites.find(r => r.classification === 'unknown' && r.counted).classification = 'fixed_set_excludes_prototype'; });
mutant('remove user-input origin', d => { d.sites.find(r => r.classification === 'user_input' && r.counted).input_sources = []; });
