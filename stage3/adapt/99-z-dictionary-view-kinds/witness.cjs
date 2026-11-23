'use strict';
// node witness.cjs TREE: the adapted tree holds both after owners, and each after owner answers what its before owner
// answers, run in Node over a generated table.
//
// isJsonEqual: the pinned before and after bodies, each transpiled beside the tree's own core.ts equalOwnProperties,
// equateValues, every, isArray and hasProperty, run over every ordered pair of a table: dense arrays of lengths 0 to 3,
// records with and without index keys and "length", records keyed the way an index almost is ("00", "-0", "0.5",
// "1e0", " 0"), each nested one level, strings, numbers, booleans, null and undefined. Every answer compares by
// Object.is. The table must hold a pair whose answer is asymmetric, and the six answers README.md names are checked by
// value. The premise's control runs arrays with a hole or a key beside their indices, which the after body may answer
// differently, and must find that it does: the premise is what the after body leans on.
//
// loadModuleFromExports: the pinned before and after owners, each transpiled beside the tree's own
// moduleNameResolver.ts and core.ts helpers it reaches (loadModuleFromExportsOrImports included, so an empty array's
// walk is upstream's own), with trace and getLoadModuleFromTargetExportOrImport recording every call. Each kind of
// exports, false, 0, "", true, numbers, strings, arrays and records, runs at five subpaths with tracing on and off,
// and every trace, every target load and the result compare. The factory's own call isn't compared: upstream's
// getLoadModuleFromTargetExportOrImport (moduleNameResolver.ts:2751) only returns its nested closure, so a call
// that loads nothing has no effect a caller can see (J-5bzd5hk: an empty array's walk calls it, the guard doesn't). Both guarded conditions are
// also read alone over each kind: 2629's must answer the same for every kind that reaches it, and 2637's may differ
// only for the kinds README.md's table names, true, a nonzero number and an empty array, whose outcome is the same.
//
// WITNESS_MUTANT=symmetric swaps the argument order in isJsonEqual's array-and-record branches, and must fail.
// WITNESS_MUTANT=object-guard-inverted makes 2629's guard `!== "object"`, and must fail.
// WITNESS_COUNT=<a tsc tree> counts the premise in Node: README.md's "The premise's count".
const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const ts = require('typescript');
const {owner, pins} = require('./adapt.cjs');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 3, 'usage: node witness.cjs TREE');
let failed = 0;
const verdict = (ok, line) => { console.log(`witness: ${ok ? 'PASS' : 'FAIL'} ${line}`); if (!ok) failed++; };

const tree = path.resolve(process.argv[2]);
const pinOf = name => pins.owners.find(pin => pin.owner === name);
for (const pin of pins.owners)
    verdict(owner(fs.readFileSync(path.join(tree, pin.file), 'utf8'), pin).state === 'after', `${pin.file}: ${pin.owner} is the pinned after owner`);

// mutate TEXT OLD NEW: OLD, found exactly once, replaced.
function mutate(text, old, replacement) {
    assert.equal(text.split(old).length - 1, 1, `witness: mutant seam not found once: ${old}`);
    return text.replace(old, replacement);
}
let jsonAfter = pinOf('isJsonEqual').after, exportsAfter = pinOf('loadModuleFromExports').after;
if (process.env.WITNESS_MUTANT === 'symmetric') {
    jsonAfter = mutate(jsonAfter, 'isJsonEqual(a[index], right[String(index)])', 'isJsonEqual(right[String(index)], a[index])');
    jsonAfter = mutate(jsonAfter, 'isJsonEqual(left[key], b.length)', 'isJsonEqual(b.length, left[key])');
    jsonAfter = mutate(jsonAfter, 'isJsonEqual(left[key], b[Number(key)])', 'isJsonEqual(b[Number(key)], left[key])');
    console.log('witness: mutant symmetric planted');
}
else if (process.env.WITNESS_MUTANT === 'object-guard-inverted') {
    exportsAfter = mutate(exportsAfter, 'typeof scope.contents.packageJsonContent.exports === "object" && hasProperty(', 'typeof scope.contents.packageJsonContent.exports !== "object" && hasProperty(');
    console.log('witness: mutant object-guard-inverted planted');
}
else assert(!process.env.WITNESS_MUTANT, `witness: unknown mutant ${process.env.WITNESS_MUTANT}`);

// declarations: the named top-level functions with bodies and enums, read from one of the tree's files, export dropped.
function declarations(file, names) {
    const text = fs.readFileSync(path.join(tree, file), 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const found = new Map();
    for (const statement of source.statements) {
        const name = statement.name?.text;
        if (!names.includes(name) || !(ts.isFunctionDeclaration(statement) && statement.body || ts.isEnumDeclaration(statement))) continue;
        assert(!found.has(name), `witness: ${file} declares ${name} twice`);
        found.set(name, text.slice(statement.getStart(source), statement.end).replace(/^export /, ''));
    }
    for (const name of names) assert(found.has(name), `witness: ${file} declares no ${name}`);
    return names.map(name => found.get(name));
}
// compile: the texts transpiled into one script run with given's names bound, returning the named functions.
function compile(texts, given, names) {
    const script = ts.transpileModule(texts.map(text => text.replace(/^export /, '')).join('\r\n\r\n'), {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText;
    return new Function(...Object.keys(given), `${script.replace(/^export \{\};\s*$/m, '')}\nreturn {${names.join(', ')}};`)(...Object.values(given));
}

// isJsonEqual.
const jsonCore = declarations('src/compiler/core.ts', ['equalOwnProperties', 'equateValues', 'every', 'isArray', 'hasProperty']);
const jsonEqual = text => compile([...jsonCore, text], {hasOwnProperty: Object.prototype.hasOwnProperty}, ['isJsonEqual']).isJsonEqual;
const jsonBefore = jsonEqual(pinOf('isJsonEqual').before), jsonAfterBody = jsonEqual(jsonAfter);

const leaves = [undefined, null, true, false, 0, -0, 1, 5, NaN, '', 'a', '0', 'length'];
const elements = ['a', 0, [], {}, {length: 0}, {length: 5}, ['a'], {0: 'a'}];
const arrays = [[], ...elements.map(element => [element]), ...elements.flatMap(first => elements.map(second => [first, second])),
    ['a', 'a', 'a'], ['a', [], {length: 5}], [{length: 5}, {length: 5}, {length: 5}], [0, 0, 0], [[], [], []],
    [undefined], [null], [true], [5], [undefined, undefined]];
const indexRecord = array => Object.fromEntries(array.map((element, index) => [String(index), element]));
const records = [
    ...arrays.flatMap(array => [indexRecord(array), {...indexRecord(array), length: array.length}, {...indexRecord(array), length: array.length + 1}]),
    {'00': 'a'}, {'-0': 'a'}, {'0.5': 'a'}, {'1e0': 'a'}, {' 0': 'a'}, {'0 ': 'a'}, {'+0': 'a'}, {Length: 0}, {NaN: 1}, {Infinity: 1},
    {4294967294: 'a'}, {a: 1}, {0: 'a', a: 1}, {1: 'a'}, {0: undefined}, {length: undefined}, {length: '0'}, {a: {b: []}},
    {0: [], 1: {}}, {0: {length: 5}}, {0: [{length: 5}]}, {length: [], 0: 'a'},
];
const table = [...leaves, ...arrays, ...records];
const isRecord = value => typeof value === 'object' && value !== null && !Array.isArray(value);
const kinds = {pairs: 0, mixedArrayLeft: 0, mixedRecordLeft: 0, twoArrays: 0, twoRecords: 0, asymmetric: 0};
const differing = [];
for (const left of table) {
    for (const right of table) {
        kinds.pairs++;
        if (Array.isArray(left) && isRecord(right)) kinds.mixedArrayLeft++;
        if (isRecord(left) && Array.isArray(right)) kinds.mixedRecordLeft++;
        if (Array.isArray(left) && Array.isArray(right)) kinds.twoArrays++;
        if (isRecord(left) && isRecord(right)) kinds.twoRecords++;
        const before = jsonBefore(left, right);
        if (before !== jsonBefore(right, left)) kinds.asymmetric++;
        if (!Object.is(before, jsonAfterBody(left, right))) differing.push([left, right]);
    }
}
const show = value => value === undefined ? 'undefined' : Number.isNaN(value) ? 'NaN' : Object.is(value, -0) ? '-0' : JSON.stringify(value);
verdict(differing.length === 0, `isJsonEqual: ${table.length} values, ${kinds.pairs} ordered pairs (${kinds.twoArrays} two arrays, ${kinds.twoRecords} two records, ` +
    `${kinds.mixedArrayLeft} array then record, ${kinds.mixedRecordLeft} record then array), differing ${differing.length}` +
    (differing.length ? `, first ${differing.slice(0, 4).map(([left, right]) => `${show(left)} against ${show(right)}`).join('; ')}` : ''));
verdict(kinds.asymmetric > 0, `isJsonEqual: ${kinds.asymmetric} ordered pairs answer differently with their sides swapped, so the table sees the left-first order`);
const named = [[[], {}, true], [['x'], {0: 'x'}, true], [[], {length: 5}, true], [{length: 0}, [], true], [{length: 5}, [], false], [[{length: 5}], {0: []}, false]];
for (const [left, right, answer] of named) {
    verdict(jsonBefore(left, right) === answer && jsonAfterBody(left, right) === answer, `isJsonEqual: ${show(left)} against ${show(right)} is ${answer} before and after`);
}
// The premise's control: a hole or a key beside the indices is what the dense comparison can answer differently for.
const sparse = [[, 'a'], [, ], Object.assign(['a'], {x: 1}), Object.assign([], {x: 1})];
const controlTable = [...sparse, ...arrays.slice(0, 12), ...records.slice(0, 12), {x: 1}, {1: 'a'}];
let controlDiffering = 0;
for (const left of controlTable) for (const right of controlTable) if (!Object.is(jsonBefore(left, right), jsonAfterBody(left, right))) controlDiffering++;
verdict(controlDiffering > 0, `isJsonEqual: the premise's control, arrays with a hole or a key beside their indices, differs in ${controlDiffering} pairs, as the premise says it may`);

// loadModuleFromExports.
let log = [];
const host = {toJSON: () => '<host>'}, scope = {packageDirectory: '/p', contents: {packageJsonContent: {}}, toJSON: () => '<scope>'};
const given = {
    emptyArray: [], directorySeparator: '/', Diagnostics: new Proxy({}, {get: (_, key) => String(key)}),
    trace: (...args) => { log.push(['trace', ...args]); },
    getLoadModuleFromTargetExportOrImport: (extensions, state, cache, redirectedReference, moduleName, from, isImports) => {
        return (target, subpath, pattern, key) => {
            log.push(['load', target, subpath, pattern, key]);
            return {value: {loaded: [target, subpath, pattern, key]}};
        };
    },
};
const resolverTexts = [
    ...declarations('src/compiler/corePublic.ts', ['Comparison']),
    ...declarations('src/compiler/core.ts', ['hasProperty', 'getOwnKeys', 'every', 'some', 'isArray', 'startsWith', 'endsWith', 'filter', 'toSorted']),
    ...declarations('src/compiler/moduleNameResolver.ts', ['NodeResolutionFeatures', 'allKeysStartWithDot', 'noKeyStartsWithDot', 'toSearchResult', 'loadModuleFromExportsOrImports', 'hasOneAsterisk', 'comparePatternKeys']),
];
given.hasOwnProperty = Object.prototype.hasOwnProperty;
const resolver = text => compile([...resolverTexts, text], given, ['loadModuleFromExports']).loadModuleFromExports;
const exportsBefore = resolver(pinOf('loadModuleFromExports').before), exportsAfterBody = resolver(exportsAfter);
// conditionOf: the else-if condition in an owner whose text holds call, compiled as a function of scope.
function conditionOf(text, call) {
    const source = ts.createSourceFile('owner.ts', text, ts.ScriptTarget.Latest, true), found = [];
    ts.forEachChild(source, function visit(node) {
        if (ts.isIfStatement(node) && node.expression.getText(source).includes(call)) found.push(node.expression.getText(source));
        ts.forEachChild(node, visit);
    });
    assert.equal(found.length, 1, `witness: one condition calling ${call}`);
    return compile([...resolverTexts, `function condition(scope: any) { return ${found[0]}; }`], given, ['condition']).condition;
}
const conditions = {
    2629: [conditionOf(pinOf('loadModuleFromExports').before, 'hasProperty('), conditionOf(exportsAfter, 'hasProperty(')],
    2637: [conditionOf(pinOf('loadModuleFromExports').before, 'allKeysStartWithDot('), conditionOf(exportsAfter, 'allKeysStartWithDot(')],
};
const exportsKinds = [
    false, 0, '', null, undefined, true, 1, -1, 0.5, 'x', '.', './a.js', [], ['./a.js'], [''], [{import: './a.js'}], {},
    {'.': './i.js'}, {import: './i.js', require: './r.js'}, {'.': './i.js', './sub': './s.js'}, {'./sub': './s.js'},
    {'./*': './s/*.js'}, {'./sub/': './s/'}, {'./sub/*.js': './s/*.js'}, {'.': './i.js', import: './x.js'}, {length: './l.js'},
    {'./sub': './s.js', x: 1}, {'./sub/x.js': ['./a.js', './b.js']},
];
const subpaths = ['.', './sub', './sub/x.js', './other', './length'];
const serialize = value => JSON.stringify(value, (_, v) => v === undefined ? '<undefined>' : v);
let rows = 0, outcomeDiffering = [];
const run = (body, exports, subpath, traceEnabled) => {
    log = [];
    scope.contents.packageJsonContent = {exports};
    const result = body(scope, 7, subpath, {host, traceEnabled, features: -1}, undefined, undefined);
    return serialize({result, log});
};
for (const exports of exportsKinds) {
    for (const subpath of subpaths) {
        for (const traceEnabled of [true, false]) {
            rows++;
            const before = run(exportsBefore, exports, subpath, traceEnabled), after = run(exportsAfterBody, exports, subpath, traceEnabled);
            if (before !== after) outcomeDiffering.push(`${show(exports)} at ${subpath}${traceEnabled ? ' traced' : ''}: ${before} against ${after}`);
        }
    }
}
verdict(outcomeDiffering.length === 0, `loadModuleFromExports: ${exportsKinds.length} kinds of exports at ${subpaths.length} subpaths, traced and not, ${rows} rows, ` +
    `every trace, target load and result compared, differing ${outcomeDiffering.length}${outcomeDiffering.length ? `, first ${outcomeDiffering.slice(0, 3).join('; ')}` : ''}`);
// Which kinds reach each condition: falsy returns first, and at "." 2626 takes a string, an array and a record with
// no key starting with a dot.
const truthy = exportsKinds.filter(exports => exports);
const reaches2629 = truthy.filter(exports => !(typeof exports === 'string' || Array.isArray(exports) || isRecord(exports) && !Object.keys(exports).some(key => key.startsWith('.'))));
const read = (condition, exports) => { scope.contents.packageJsonContent = {exports}; return condition(scope); };
const conditionDiffering = kind => (kind === 2629 ? reaches2629 : truthy)
    .filter(exports => read(conditions[kind][0], exports) !== read(conditions[kind][1], exports));
verdict(conditionDiffering(2629).length === 0, `loadModuleFromExports 2629: the condition answers the same for every kind that reaches it (${reaches2629.map(show).join(', ')})`);
const differ2637 = conditionDiffering(2637), expected2637 = truthy.filter(exports => typeof exports === 'boolean' || typeof exports === 'number' || Array.isArray(exports) && exports.length === 0);
verdict(serialize(differ2637) === serialize(expected2637), `loadModuleFromExports 2637: the condition differs only for ${differ2637.map(show).join(', ')}, the kinds whose outcome the rows above show is the same`);

// The premise's count: an in-memory copy of stock 6.0.3's lib/typescript.js, compiled under its own path so it finds
// its lib files, with isJsonEqual's body opening on a count of what it receives, run over the tree's own projects.
if (process.env.WITNESS_COUNT) {
    const countTree = path.resolve(process.env.WITNESS_COUNT);
    const stock = require.resolve('typescript');
    const text = fs.readFileSync(stock, 'utf8');
    const opening = [...text.matchAll(/function isJsonEqual\d*\((\w+), (\w+)\) \{\r?\n/g)];
    assert.equal(opening.length, 1, 'witness: count: isJsonEqual opens once in the stock bundle');
    const [, first, second] = opening[0];
    const at = opening[0].index + opening[0][0].length;
    const patched = text.slice(0, at) + `  globalThis.__dictionaryViewKindsCount(${first}, ${second});\n` + text.slice(at);
    const count = {calls: 0, arrays: 0, holes: 0, extraKeys: 0, twoArrays: 0, mixed: 0, twoRecords: 0, differing: 0};
    const ownIndex = (array, key) => /^(0|[1-9]\d*)$/.test(key) && Number(key) < array.length;
    globalThis.__dictionaryViewKindsCount = (a, b) => {
        count.calls++;
        for (const side of [a, b]) {
            if (!Array.isArray(side)) continue;
            count.arrays++;
            for (let index = 0; index < side.length; index++) if (!Object.prototype.hasOwnProperty.call(side, index)) count.holes++;
            count.extraKeys += Object.keys(side).filter(key => !ownIndex(side, key)).length;
        }
        if (Array.isArray(a) && Array.isArray(b)) count.twoArrays++;
        else if (Array.isArray(a) && isRecord(b) || isRecord(a) && Array.isArray(b)) count.mixed++;
        else if (isRecord(a) && isRecord(b)) count.twoRecords++;
        if (!Object.is(jsonBefore(a, b), jsonAfterBody(a, b))) count.differing++;
    };
    const copy = new Module(stock, module);
    copy.filename = stock;
    copy.paths = Module._nodeModulePaths(path.dirname(stock));
    copy._compile(patched, stock);
    const counted = copy.exports;
    assert.notEqual(counted, ts, 'witness: count: the copy is its own module');
    const projects = fs.readdirSync(path.join(countTree, 'src')).sort()
        .map(name => path.join(countTree, 'src', name, 'tsconfig.json')).filter(name => fs.existsSync(name));
    const parse = name => counted.getParsedCommandLineOfConfigFile(name, {}, {
        ...counted.sys,
        onUnRecoverableConfigFileDiagnostic: diagnostic => { throw new Error(counted.flattenDiagnosticMessageText(diagnostic.messageText, '\n')); },
    });
    for (const project of projects) {
        const config = parse(project), again = parse(project);
        if (!config.fileNames.length) continue;
        const before = {...count};
        const compilerHost = counted.createCompilerHost(config.options);
        // Two programs from two parses of one tsconfig: the second reuses the first, comparing every option tsc reads
        // there (changesAffectModuleResolution, changesAffectingProgramStructure). Then two builders, the second from
        // the first's state, comparing the options that affect diagnostics, emit and declaration paths.
        const first = counted.createProgram({rootNames: config.fileNames, options: config.options, projectReferences: config.projectReferences, host: compilerHost});
        const second = counted.createProgram({rootNames: again.fileNames, options: again.options, projectReferences: again.projectReferences, host: compilerHost, oldProgram: first});
        const oldBuilder = counted.createSemanticDiagnosticsBuilderProgram(first, compilerHost);
        counted.createSemanticDiagnosticsBuilderProgram(second, compilerHost, oldBuilder);
        // The buildinfo round trip, where the old options are JSON.parse's own: a builder written out and read back.
        let written;
        oldBuilder.emitBuildInfo((name, data) => { written = {name, data}; });
        let roundTrip = 'no buildinfo written';
        if (written) {
            const fromBuildInfo = counted.readBuilderProgram(config.options, {useCaseSensitiveFileNames: () => counted.sys.useCaseSensitiveFileNames, getCurrentDirectory: () => counted.sys.getCurrentDirectory(), readFile: name => path.resolve(name) === path.resolve(written.name) ? written.data : undefined});
            if (fromBuildInfo) {
                counted.createSemanticDiagnosticsBuilderProgram(second, compilerHost, fromBuildInfo);
                roundTrip = 'buildinfo read back';
            }
            else roundTrip = 'buildinfo written, not read back';
        }
        const delta = Object.fromEntries(Object.entries(count).map(([key, value]) => [key, value - before[key]]));
        console.log(`witness: count: ${path.relative(countTree, project)}: ${config.fileNames.length} roots, ${roundTrip}, ${JSON.stringify(delta)}`);
    }
    verdict(count.calls > 0, `count: isJsonEqual was reached ${count.calls} times over ${projects.length} projects`);
    verdict(count.holes === 0 && count.extraKeys === 0, `count: ${count.arrays} arrays received, ${count.holes} holes, ${count.extraKeys} keys beside their indices (the premise holds at 0 and 0)`);
    verdict(count.differing === 0, `count: ${count.twoArrays} pairs of arrays, ${count.mixed} array and record pairs, ${count.twoRecords} pairs of records, ${count.differing} answered differently by the after body`);
}

console.log(failed === 0 ? 'witness: PASS' : `witness: FAIL (${failed})`);
process.exit(failed === 0 ? 0 : 1);
