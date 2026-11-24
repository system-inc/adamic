'use strict';
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const {adapt} = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <applied tree>');
const tree = process.argv[2];
const snapshotRef = 'refs/stage3/99-seats-build22-63-splice-truncations';
const directory = 'src/compiler';
const fixtures = ['program.ts', 'sys.ts'];
const scratch = fs.mkdtempSync('/tmp/adamic-splice-truncations-');
fs.mkdirSync(path.join(scratch, directory), {recursive: true});
const file = name => path.join(scratch, directory, name);
const baseline = path.join(scratch, 'original');

function readRef(when, name) {
    return execFileSync('git', ['-C', tree, 'show', `${snapshotRef}-${when}:${directory}/${name}`]);
}

function restoreFixtures() {
    for (const name of fixtures) {
        fs.copyFileSync(path.join(baseline, directory, name), file(name));
    }
}
function snapshot() {
    return fixtures.map(name => ({
        text: fs.readFileSync(file(name)),
        modified: fs.statSync(file(name), {bigint: true}).mtimeNs,
    }));
}

try {
    fs.mkdirSync(path.join(baseline, directory), {recursive: true});
    for (const name of [...fixtures, 'core.ts']) {
        fs.writeFileSync(path.join(baseline, directory, name), readRef('before', name));
    }

    const guards = [
        {
            file: 'program.ts',
            owner: 'computeCommonSourceDirectoryOfFilenames',
            before: /for \(let i = 0; i < n; i\+\+\)/g,
            after: 'for (let i = 0; i < n + 2; i++)',
        },
        {
            file: 'sys.ts',
            owner: 'pollWatchedFileQueue',
            before: /definedValueCopyToIndex < pollIndex(?=\) \{\r?\n +\/\/ There are holes)/g,
            after: 'definedValueCopyToIndex < pollIndex + 2',
        },
    ];
    for (const guard of guards) {
        restoreFixtures();
        const text = fs.readFileSync(file(guard.file), 'utf8');
        assert.equal([...text.matchAll(guard.before)].length, 1);
        fs.writeFileSync(file(guard.file), text.replace(guard.before, guard.after));
        const before = snapshot();
        assert.throws(() => adapt(scratch), {message: `unreviewed owner: ${guard.owner}`});
        assert.deepEqual(snapshot(), before);
        console.log(`PASS ${guard.owner}: changed guard rejected without writes`);
    }
    restoreFixtures();
    assert.deepEqual(adapt(scratch), {files: 2, edits: 2, javascriptChanging: true});
    const applied = snapshot();
    assert.deepEqual(adapt(scratch), {files: 0, edits: 0, javascriptChanging: true});
    assert.deepEqual(snapshot(), applied);
    console.log('PASS two statement edits and a zero-write replay');
    checkEmittedOwners();
    checkArrayTruncations();
    for (const name of fixtures) {
        assert.deepEqual(fs.readFileSync(file(name)), readRef('after', name), `recorded output: ${name}`);
    }
    console.log('PASS both scratch outputs match the pipeline after ref byte for byte');
} finally {
    fs.rmSync(scratch, {recursive: true, force: true});
}

function compareTruncations(original, keep) {
    const assigned = original.slice();
    const spliced = original.slice();
    assigned.length = keep;
    spliced.splice(keep);
    assert.equal(assigned.length, keep);
    assert.equal(spliced.length, keep);
    for (let i = 0; i < original.length; i++) {
        assert.equal(i in assigned, i in spliced, `index ${i} presence`);
        assert.equal(assigned[i], spliced[i], `index ${i} value`);
    }
}

function checkArrayTruncations() {
    const dense = [10, 20, 30, 40];
    for (const keep of [0, 2, dense.length]) {
        compareTruncations(dense, keep);
        console.log(`PASS dense length ${dense.length} to ${keep}: length, index presence and values agree`);
    }
    const holey = [10, , 30, 40, , 60];
    assert.equal(1 in holey, false);
    assert.equal(4 in holey, false);
    compareTruncations(holey, 3);
    console.log('PASS holey length 6 to 3: length, index presence and values agree with holes on both sides');
}

function readOwner(root, name, owner) {
    const source = ts.createSourceFile(name, fs.readFileSync(path.join(root, directory, name), 'utf8'),
        ts.ScriptTarget.ES2020, true);
    assert.equal(source.parseDiagnostics.length, 0);
    const matches = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === owner && node.body);
    assert.equal(matches.length, 1, owner);
    const declaration = matches[0];
    const start = declaration.modifiers?.at(-1).end ?? declaration.getStart(source);
    return source.text.slice(start, declaration.end);
}

function loadOwner(root, name, owner, dependencies, helpers = []) {
    const declarations = [...helpers, owner].map(symbol => readOwner(root, name, symbol));
    const {outputText, diagnostics} = ts.transpileModule(declarations.join('\n'), {
        compilerOptions: {target: ts.ScriptTarget.ES2020},
        reportDiagnostics: true,
    });
    assert.deepEqual(diagnostics, []);
    return new Function(...Object.keys(dependencies), `"use strict";\n${outputText}\nreturn ${owner};`)(
        ...Object.values(dependencies));
}

function observeArray(values, extent) {
    return {
        length: values.length,
        indices: Array.from({length: extent}, (_, index) => [index in values, values[index]]),
    };
}

function checkDirectoryOwners() {
    const forEach = loadOwner(baseline, 'core.ts', 'forEach', {});
    function run(root, paths) {
        const components = paths.map(segments => segments.slice());
        let next = 0;
        let reusedFirst = false;
        const walk = loadOwner(root, 'program.ts', 'computeCommonSourceDirectoryOfFilenames', {
            forEach,
            // Supply normalized components directly and retain the first array to observe aliasing.
            getNormalizedPathComponents: () => components[next++],
            getPathFromPathComponents: array => {
                reusedFirst = array === components[0];
                return array[0] + array.slice(1).join('/');
            },
        });
        const result = walk(paths.map(segments => segments.join('/')), 'c:/repo', name => name.toLowerCase());
        return {result, reusedFirst, components: components.map((array, i) => observeArray(array, paths[i].length))};
    }

    const cases = [
        ['different roots', ['c:/', 'repo', 'first.ts'], ['d:/', 'repo', 'second.ts'], '', false],
        ['root-only prefix', ['c:/', 'repo', 'first.ts'], ['c:/', 'other', 'second.ts'], 'c:/', true],
        ['non-root mismatch truncates the first array', ['c:/', 'repo', 'src', 'a', 'first.ts'],
            ['c:/', 'repo', 'src', 'b', 'second.ts'], 'c:/repo/src', true],
        ['second path is shorter', ['c:/', 'repo', 'src', 'first.ts'],
            ['c:/', 'repo', 'second.ts'], 'c:/repo', true],
        ['first path is shorter', ['c:/', 'repo', 'first.ts'],
            ['c:/', 'repo', 'src', 'second.ts'], 'c:/repo', true],
    ];
    for (const [label, first, second, expected, reuse] of cases) {
        const original = run(baseline, [first, second]);
        const emitted = run(scratch, [first, second]);
        assert.equal(original.result, expected, label);
        assert.equal(original.reusedFirst, reuse, label);
        assert.equal(emitted.result, original.result, label);
        assert.deepEqual(emitted, original, label);
        console.log(`PASS emitted directory owner: ${label}`);
    }
}

function checkPollingOwners() {
    const watcher = (fileName, isClosed = false) => ({fileName, isClosed, mtime: new Date(1)});
    function run(root, sparse) {
        const queue = [watcher('closed', true), watcher('first'), undefined, watcher('second'), undefined];
        if (sparse) {
            delete queue[2];
            delete queue[4];
        }
        const extent = queue.length;
        const before = observeArray(queue, extent);
        const visits = [];
        const poll = loadOwner(root, 'sys.ts', 'pollWatchedFileQueue', {
            missingFileModifiedTime: new Date(0),
        }, ['getModifiedTime', 'onWatchedFileStat']);
        // An unchanged mtime keeps event dispatch out of this queue test.
        const next = poll({getModifiedTime: () => new Date(1)}, queue, 0, extent, (item, index, changed) => {
            visits.push([item.fileName, index, changed, 0 in queue, queue[0]]);
        });
        return {next, before, visits, after: observeArray(queue, extent)};
    }

    for (const sparse of [false, true]) {
        const label = sparse ? 'sparse queue compaction' : 'closed watcher cleared to undefined';
        const original = run(baseline, sparse);
        const emitted = run(scratch, sparse);
        assert.equal(original.before.indices[2][0], !sparse, label);
        assert.equal(original.before.indices[4][0], !sparse, label);
        assert.deepEqual(original.visits[0], ['first', 1, false, true, undefined], label);
        assert.equal(original.next, 0, label);
        assert.equal(original.after.length, 2, label);
        assert.deepEqual(original.after.indices, [
            [true, watcher('first')], [true, watcher('second')],
            [false, undefined], [false, undefined], [false, undefined],
        ], label);
        assert.deepEqual(emitted, original, label);
        console.log(`PASS emitted polling owner: ${label}`);
    }
}

function checkEmittedOwners() {
    assert.equal(ts.version, '6.0.3');
    checkDirectoryOwners();
    checkPollingOwners();
    checkZeroCuts();
    checkRetainedHole();
    checkShortQueues();
    checkDirectoryShapes();
}

function checkZeroCuts() {
    function run(root, initiallyClosed) {
        const events = [];
        function watcher(fileName) {
            const watch = {
                fileName, isClosed: initiallyClosed, mtime: new Date(1),
                callback(name, event, time) {
                    events.push([name, event, time.getTime()]);
                    watch.isClosed = true;
                },
            };
            return watch;
        }
        const queue = [watcher('last'), initiallyClosed ? watcher('already closed') : undefined, undefined];
        delete queue[2];
        const extent = queue.length;
        const poll = loadOwner(root, 'sys.ts', 'pollWatchedFileQueue', {
            missingFileModifiedTime: new Date(0),
            FileWatcherEventKind: {Created: 0, Changed: 1, Deleted: 2},
        }, ['getModifiedTime', 'onWatchedFileStat', 'getFileWatcherEventKind']);
        const next = poll({getModifiedTime: () => new Date(2)}, queue, 0, extent);
        return {next, events, after: observeArray(queue, extent)};
    }

    for (const {initiallyClosed, label, events} of [
        {initiallyClosed: false, label: 'callback closes the sole watcher: truncate at zero', events: [['last', 1, 2]]},
        {initiallyClosed: true, label: 'all watchers already closed: truncate at zero', events: []},
    ]) {
        const original = run(baseline, initiallyClosed);
        assert.deepEqual(original, {
            next: 0,
            events,
            after: {length: 0, indices: [[false, undefined], [false, undefined], [false, undefined]]},
        }, label);
        assert.deepEqual(run(scratch, initiallyClosed), original, label);
        console.log(`PASS emitted polling owner: ${label}`);
    }
}

function checkRetainedHole() {
    const watcher = fileName => ({fileName, isClosed: false, mtime: new Date(1)});
    function run(root) {
        const queue = [, watcher('previous'), undefined, watcher('next')];
        const extent = queue.length;
        const poll = loadOwner(root, 'sys.ts', 'pollWatchedFileQueue', {
            missingFileModifiedTime: new Date(0),
        }, ['getModifiedTime', 'onWatchedFileStat']);
        const next = poll({getModifiedTime: () => new Date(1)}, queue, 2, 1);
        return {next, after: observeArray(queue, extent)};
    }

    const label = 'resumed poll preserves an existing prefix hole';
    const original = run(baseline);
    assert.deepEqual(original, {
        next: 0,
        after: {
            length: 3,
            indices: [[false, undefined], [true, watcher('previous')], [true, watcher('next')], [false, undefined]],
        },
    }, label);
    assert.deepEqual(run(scratch), original, label);
    console.log('PASS resumed poll retains a genuine hole before the cut at three');
}

function checkShortQueues() {
    const kinds = ['live', 'hole', 'undefined', 'closed', 'closes'];
    const load = root => loadOwner(root, 'sys.ts', 'pollWatchedFileQueue', {
        missingFileModifiedTime: new Date(0),
        FileWatcherEventKind: {Created: 0, Changed: 1, Deleted: 2},
    }, ['getModifiedTime', 'onWatchedFileStat', 'getFileWatcherEventKind']);
    const original = load(baseline);
    const emitted = load(scratch);

    function observe(poll, shape, start, chunk) {
        const queue = new Array(shape.length);
        const watchers = [];
        const callbacks = [];
        for (const [index, kind] of shape.entries()) {
            if (kind === 'hole') continue;
            if (kind === 'undefined') {
                queue[index] = undefined;
                continue;
            }
            const watcher = {
                fileName: String(index),
                isClosed: kind === 'closed',
                mtime: new Date(1),
                callback(name, event, time) {
                    callbacks.push([name, event, time.getTime()]);
                    if (kind === 'closes') watcher.isClosed = true;
                },
            };
            queue[index] = watcher;
            watchers.push(watcher);
        }
        const host = {getModifiedTime: name => new Date(shape[Number(name)] === 'closes' ? 2 : 1)};
        const next = poll(host, queue, start, chunk);
        const after = observeArray(queue, shape.length);
        after.indices = after.indices.map(([present, value]) => [present, value && {
            watcher: watchers.indexOf(value),
            fileName: value.fileName,
            isClosed: value.isClosed,
            mtime: value.mtime.getTime(),
        }]);
        return {next, after, callbacks};
    }

    let checked = 0;
    for (let length = 0; length <= 4; length++) {
        for (let encoding = 0; encoding < kinds.length ** length; encoding++) {
            const shape = Array.from({length}, (_, index) =>
                kinds[Math.floor(encoding / kinds.length ** index) % kinds.length]);
            for (let start = 0; start < Math.max(length, 1); start++) {
                for (let chunk = 0; chunk <= length + 1; chunk++) {
                    assert.deepEqual(observe(emitted, shape, start, chunk), observe(original, shape, start, chunk),
                        `short queue [${shape.join(', ')}], start ${start}, chunk ${chunk}`);
                    checked++;
                }
            }
        }
    }
    console.log(`PASS bounded emitted queue sweep: ${checked} cases`);
}

function checkDirectoryShapes() {
    const forEach = loadOwner(baseline, 'core.ts', 'forEach', {});
    let current;
    const load = root => loadOwner(root, 'program.ts', 'computeCommonSourceDirectoryOfFilenames', {
        forEach,
        getNormalizedPathComponents: () => current.components[current.next++],
        getPathFromPathComponents: array => {
            current.reusedFirst = array === current.components[0];
            return array[0] + array.slice(1).join('/');
        },
    });
    const original = load(baseline);
    const emitted = load(scratch);

    function observe(walk, paths, fileNames) {
        current = {components: paths.map(path => path.slice()), next: 0, reusedFirst: false};
        const result = walk(fileNames, 'c:/repo', name => name.toLowerCase());
        return {
            result,
            reusedFirst: current.reusedFirst,
            components: current.components.map((array, index) => observeArray(array, paths[index].length)),
        };
    }

    const paths = [];
    const shortPaths = [];
    let prefixes = [[]];
    for (let depth = 0; depth <= 3; depth++) {
        for (const root of ['c:/', 'd:/']) {
            for (const prefix of prefixes) {
                const path = [root, ...prefix, 'f.ts'];
                paths.push(path);
                if (depth <= 2) shortPaths.push(path);
            }
        }
        prefixes = prefixes.flatMap(prefix => ['a', 'A', 'b'].map(name => [...prefix, name]));
    }

    let checked = 0;
    function compare(paths) {
        const fileNames = paths.map(([root, ...parts]) => root + parts.join('/'));
        assert.deepEqual(observe(emitted, paths, fileNames), observe(original, paths, fileNames),
            `directory sweep [${fileNames.join(', ')}]`);
        checked++;
    }

    for (const first of paths) compare([first]);
    for (const first of paths) {
        for (const second of paths) compare([first, second]);
    }
    for (const first of shortPaths) {
        for (const second of shortPaths) {
            for (const third of shortPaths) compare([first, second, third]);
        }
    }
    console.log(`PASS bounded emitted directory sweep: ${checked} cases`);
}
