'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: adapt.cjs <tree>');
function once(text, before, after) {
    assert.equal(text.split(before).length - 1 + text.split(after).length - 1, 1, 'source site drift: ' + before);
    return text.includes(before) ? text.replace(before, after) : text;
}
function unchanged(before, after) {
    for (const removeComments of [false, true]) {
        const compilerOptions = {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments};
        assert.equal(ts.transpileModule(after, {compilerOptions}).outputText,
            ts.transpileModule(before, {compilerOptions}).outputText, 'runtime JavaScript changed');
    }
}
const tree = path.resolve(process.argv[2]);
const plans = ['core.ts', 'performanceCore.ts', 'tracing.ts'].map(name => {
    const file = path.join(tree, 'src/compiler', name), before = fs.readFileSync(file, 'utf8');
    let after = before;
    if (name === 'core.ts') {
        after = after.replace('declare const process: (NodeJS.Process & { browser?: unknown }) | undefined;', 'declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;');
        assert(after.includes('declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;'), 'process declaration drift');
        // Explicit any now lands in 40; accept its erased cast or the older 65 spelling.
        assert(after.includes('(process as any).browser') || after.includes('process.browser'), 'browser probe drift');
    }
    if (name === 'performanceCore.ts') {
        after = once(after, 'Partial<typeof import("perf_hooks")>', '{ performance?: Performance | undefined }');
        if (!after.includes('declare const require:')) after += '\ndeclare const require: (specifier: "perf_hooks") => any;\n';
    }
    if (name === 'tracing.ts') {
        after = once(after, 'let fs: typeof import("fs");', 'let fs: { existsSync(path: string): boolean; mkdirSync(path: string, options: { recursive: boolean }): string | undefined; openSync(path: string, flags: string): number; writeSync(fd: number, data: string): number; closeSync(fd: number): void; writeFileSync(path: string, data: string): void; };');
        if (!after.includes('declare const require:')) after += '\ndeclare const require: (specifier: "fs") => any;\ndeclare const process: { pid: number };\n';
    }
    unchanged(before, after);
    return {file, before, after};
});
for (const p of plans) assert.equal(fs.readFileSync(p.file, 'utf8'), p.before);
for (const p of plans) if (p.before !== p.after) fs.writeFileSync(p.file, p.after);
console.log(JSON.stringify({files: plans.map(p => ({file: path.relative(tree,p.file), changed:p.before!==p.after})), javascript_identical:true}));
module.exports = { unchanged };
