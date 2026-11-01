'use strict';
// Runs appendSourceMap from each built compiler named on the command line (its lib/typescript.js) on two maps:
// a plain sourcesContent array, and 33's getter, which returns ['present'] for the first two property accesses
// and [] after. Prints, per compiler, how many times sourcesContent was read and the merged map's sourcesContent.
// Usage: node witness.cjs <label>=<path/to/typescript.js> ...
const assert = require('node:assert/strict');

// One mapping, generated 0:0 to source 0 at 0:0, so appendSourceMap adds source 0 and reaches the guard once.
function map() {
    return {version: 3, file: 'in.js', sources: ['a.ts'], names: [], mappings: 'AAAA'};
}

function run(ts, raw) {
    const host = {getCurrentDirectory: () => '/', getCanonicalFileName: name => name, useCaseSensitiveFileNames: () => true};
    const generator = ts.createSourceMapGenerator(host, 'out.js', '', '/', {});
    generator.appendSourceMap(0, 0, raw, '/in.js.map');
    return generator.toJSON();
}

function plain(ts) {
    const json = run(ts, {...map(), sourcesContent: ['present']});
    return {sourcesContent: JSON.stringify(json.sourcesContent)};
}

function getter(ts) {
    let reads = 0;
    // The getter sits on the map object itself, so each property access of sourcesContent is counted.
    const raw = map();
    Object.defineProperty(raw, 'sourcesContent', {enumerable: true, get() { reads++; return reads <= 2 ? ['present'] : []; }});
    const json = run(ts, raw);
    return {sourcesContent_property_reads: reads, sourcesContent: JSON.stringify(json.sourcesContent)};
}

const results = {};
for (const argument of process.argv.slice(2)) {
    const [label, file] = argument.split('=');
    assert(label && file, 'usage: node witness.cjs <label>=<typescript.js> ...');
    const ts = require(file);
    assert.equal(typeof ts.createSourceMapGenerator, 'function', `${file} does not expose createSourceMapGenerator`);
    results[label] = {version: ts.version, plain: plain(ts), getter: getter(ts)};
}
console.log(JSON.stringify(results, null, 2));
