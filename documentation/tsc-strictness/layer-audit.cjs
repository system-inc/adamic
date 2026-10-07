// Read-only audit of the second survey's locations, quantile ranks and syntax census.
// Usage: TSC_SURVEY_TYPESCRIPT=/scratch/node_modules/typescript/lib/typescript.js
// node docs/tsc-strictness/layer-audit.cjs BASELINE_DIAGNOSTICS_JSON COMPILER_ROOT
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.TSC_SURVEY_TYPESCRIPT);
assert.equal(ts.version, '6.0.3');
const root = path.resolve(process.argv[3]);
const exported = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
assert.equal(exported.roots, 77);
const rows = exported.diagnostics.map(d => {
    const m = /^(.*):(\d+):(\d+): error TS(\d+): ([\s\S]*)$/.exec(d);
    assert(m);
    return {file: path.relative(root, m[1]).split(path.sep).join('/'), line: +m[2], column: +m[3], code: +m[4], message: m[5]};
});
const files = new Map();
function location(row) {
    let file = files.get(row.file);
    if (!file) {
        file = ts.createSourceFile(row.file, fs.readFileSync(path.join(root, row.file), 'utf8'), ts.ScriptTarget.Latest, true);
        files.set(row.file, file);
    }
    const pos = file.getPositionOfLineAndCharacter(row.line - 1, row.column - 1);
    let node = file;
    function visit(child) {
        if (child.getFullStart() <= pos && pos < child.end) {
            node = child;
            ts.forEachChild(child, visit);
        }
    }
    visit(file);
    return {file, node, pos};
}
const population = rows.filter(r => r.code === 2345).sort((a, b) => (a.file < b.file ? -1 : a.file > b.file ? 1 : 0) || a.line - b.line || a.column - b.column);
assert.equal(population.length, 693);
const sample = JSON.parse(fs.readFileSync(path.join(__dirname, 'argument-sample.json'), 'utf8'));
assert.equal(sample.length, 100);
const patterns = {}, judgments = {};
for (let k = 0; k < sample.length; k++) {
    const row = sample[k], rank = Math.floor((k + 0.5) * population.length / 100);
    assert.equal(row.id, `A${String(k + 1).padStart(3, '0')}`);
    assert.equal(row.rank, rank);
    assert.equal(row.population, population.length);
    for (const key of ['file', 'line', 'column', 'code', 'message']) assert.deepEqual(row[key], population[rank][key]);
    const {file} = location(row);
    assert.equal(row.source, file.text.split(/\r?\n/)[row.line - 1]);
    assert(row.rationale && ['M', 'C', 'U'].includes(row.judgment));
    patterns[row.pattern] = (patterns[row.pattern] || 0) + 1;
    judgments[row.judgment] = (judgments[row.judgment] || 0) + 1;
}
const syntax = JSON.parse(fs.readFileSync(path.join(__dirname, 'syntax-census.json'), 'utf8'));
const syntaxRows = rows.filter(r => r.code === 1294);
assert.equal(syntaxRows.length, 180);
assert.equal(syntax.diagnostics.length, syntaxRows.length);
const counts = {};
let constEnums = 0;
for (let k = 0; k < syntaxRows.length; k++) {
    const row = syntaxRows[k], saved = syntax.diagnostics[k];
    for (const key of ['file', 'line', 'column', 'code', 'message']) assert.deepEqual(saved[key], row[key]);
    const {file, node} = location(row);
    let declaration = node;
    while (declaration && !ts.isEnumDeclaration(declaration) && !ts.isModuleDeclaration(declaration) && !ts.isParameter(declaration)) declaration = declaration.parent;
    assert(declaration);
    const kind = ts.isEnumDeclaration(declaration) ? 'enum' : ts.isModuleDeclaration(declaration) ? 'namespace' : 'parameter property';
    assert.equal(saved.kind, kind);
    assert.equal(saved.source, file.text.split(/\r?\n/)[row.line - 1]);
    assert.equal(saved.name, declaration.name.getText(file));
    if (kind === 'parameter property') assert(ts.isParameterPropertyDeclaration(declaration, declaration.parent));
    const isConst = kind === 'enum' && !!declaration.modifiers?.some(m => m.kind === ts.SyntaxKind.ConstKeyword);
    assert.equal(saved.const, isConst);
    constEnums += +isConst;
    counts[kind] = (counts[kind] || 0) + 1;
}
assert.deepEqual(counts, syntax.counts);
assert.equal(constEnums, syntax.const_enums);
const enumProgram = ts.createProgram([...new Set(syntaxRows.map(r => path.join(root, r.file)))], {
    target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler
});
const checker = enumProgram.getTypeChecker(), enumValues = {};
for (const saved of syntax.diagnostics.filter(r => r.kind === 'enum')) {
    const file = enumProgram.getSourceFile(path.join(root, saved.file));
    let declaration;
    function visit(node) {
        if (ts.isEnumDeclaration(node) && node.name.text === saved.name && file.getLineAndCharacterOfPosition(node.name.getStart(file)).line + 1 === saved.line) declaration = node;
        ts.forEachChild(node, visit);
    }
    visit(file);
    assert(declaration);
    const values = declaration.members.map(member => checker.getConstantValue(member));
    const category = values.every(value => typeof value === 'number') ? 'numeric constants' : values.every(value => typeof value === 'string') ? 'string constants' : values.some(value => value === undefined) ? 'nonconstant or unresolved' : 'mixed constants';
    assert.equal(saved.values, category);
    enumValues[category] = (enumValues[category] || 0) + 1;
}
assert.deepEqual(enumValues, syntax.enum_values);
const document = fs.readFileSync(path.join(__dirname, '..', 'tsc-strictness-next-layer.md'), 'utf8');
for (const [pattern, count] of Object.entries(patterns)) assert(document.includes(`| ${pattern} | ${count} |`));
console.log('PASS: 100 argument locations, chains, ranks, source lines and classifications audited');
console.log('patterns', JSON.stringify(patterns), 'judgments', JSON.stringify(judgments));
console.log('PASS: all 180 syntax findings classified by AST', JSON.stringify(counts), 'const enums', constEnums);
console.log("PASS: all 164 enum value families re-resolved", JSON.stringify(enumValues));
