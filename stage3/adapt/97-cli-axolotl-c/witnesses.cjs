'use strict';
const fs = require('node:fs');
const assert = require('node:assert/strict');
const ts = require('typescript');

// Runs close()'s clearing of sourceFilesCache, lifted out of each watchPublic.ts and transpiled, on the same cache (README.md).
// usage: node witnesses.cjs <watchPublic.ts before> <watchPublic.ts after>
function clearing(fileName) {
    const source = ts.createSourceFile(fileName, fs.readFileSync(fileName, 'utf8'), ts.ScriptTarget.Latest, true);
    const found = (test, what) => {
        const matches = [];
        (function visit(node) {
            if (test(node)) matches.push(node);
            ts.forEachChild(node, visit);
        })(source);
        assert.equal(matches.length, 1, `${fileName}: ${matches.length} ${what}, not one`);
        return matches[0];
    };
    const call = found(n => ts.isCallExpression(n) && ts.isIdentifier(n.expression) && n.expression.text === 'clearMap' && ts.isIdentifier(n.arguments[0]) && n.arguments[0].text === 'sourceFilesCache', 'clearMap calls over sourceFilesCache');
    let owner = call.parent;
    while (owner && !ts.isFunctionDeclaration(owner)) owner = owner.parent;
    assert.equal(owner?.name?.text, 'close', `${fileName}: the clearMap over sourceFilesCache left close()`);
    const callback = call.arguments[1].getText(source);
    // The adapted callback calls a predicate, so both of the cache's predicates travel along; the original calls neither.
    const predicates = ['isFileMissingOnHost', 'isFilePresenceUnknownOnHost'].map(name => found(n => ts.isFunctionDeclaration(n) && n.name?.text === name, `${name} functions`).getText(source));
    const javascript = ts.transpileModule(`function lift() {\n${predicates.join('\n')}\nreturn ${callback};\n}`, {compilerOptions: {target: ts.ScriptTarget.ES2020}}).outputText;
    return {callback, onDeleteValue: new Function(`${javascript}\nreturn lift();`)()};
}

function watcher(name, closes) {
    closes[name] = 0;
    return {name, close: () => closes[name]++};
}

// Every shape sourceFilesCache holds: present with a watcher, unknown with a watcher, a present entry whose version
// nextSourceFileVersion set to false (it keeps sourceFile and fileWatcher), and missing (false, no watcher).
function cache(closes) {
    return [
        ['present', {version: '1', sourceFile: {fileName: 'present.ts'}, fileWatcher: watcher('present', closes)}],
        ['unknown', {version: false, fileWatcher: watcher('unknown', closes)}],
        ['invalidated', {version: false, sourceFile: {fileName: 'invalidated.ts'}, fileWatcher: watcher('invalidated', closes)}],
        ['missing', false],
    ];
}

function describe(entry) {
    if (entry === false) return false;
    return {version: entry.version, sourceFile: entry.sourceFile?.fileName ?? null, fileWatcher: 'fileWatcher' in entry ? entry.fileWatcher?.name ?? '<undefined>' : '<absent>'};
}

function run(onDeleteValue) {
    const closes = {};
    const entries = cache(closes);
    let beforeClear;
    // The state clearMap leaves just before it empties the map, read as the map's own clear() starts.
    class Cache extends Map {
        clear() {
            beforeClear = Object.fromEntries([...this].map(([key, entry]) => [key, describe(entry)]));
            super.clear();
        }
    }
    const map = new Cache(entries);
    // Stock TypeScript 6.0.3's own clearMap, the one close() calls.
    ts.clearMap(map, onDeleteValue);
    return {closes, beforeClear, size: map.size};
}

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 4, 'usage: node witnesses.cjs <watchPublic.ts before> <watchPublic.ts after>');
const before = clearing(process.argv[2]);
const after = clearing(process.argv[3]);
const results = {before: run(before.onDeleteValue), after: run(after.onDeleteValue)};
console.log(JSON.stringify({callbacks: {before: before.callback, after: after.callback}, results}, null, 2));
for (const [side, result] of Object.entries(results)) {
    assert.deepEqual(result.closes, {present: 1, unknown: 1, invalidated: 1}, `${side}: a watcher was not closed exactly once`);
    assert.equal(result.size, 0, `${side}: the cache did not end empty`);
}
// The one difference the repair makes: the present entry keeps its watcher instead of a written undefined.
const differences = [];
for (const key of Object.keys(results.before.beforeClear)) {
    const was = results.before.beforeClear[key];
    const is = results.after.beforeClear[key];
    if (was === false || is === false) {
        if (was !== is) differences.push(`${key}`);
        continue;
    }
    for (const field of Object.keys(was)) if (was[field] !== is[field]) differences.push(`${key}.${field}: ${was[field]} -> ${is[field]}`);
}
console.log('differences before clear:', differences);
assert.deepEqual(differences, ['present.fileWatcher: <undefined> -> present'], 'the adapted clearing differs otherwise than in the guarded field');
console.log('every watcher closed once and the cache empty, before and after; the only difference is the present entry\'s fileWatcher');
