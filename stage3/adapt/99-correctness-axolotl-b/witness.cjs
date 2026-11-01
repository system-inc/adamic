'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 4, 'usage: node witness.cjs <before.ts> <after.ts>');
const beforeText = fs.readFileSync(process.argv[2], 'utf8');
const afterText = fs.readFileSync(process.argv[3], 'utf8');
const beforeCall = 'parseInt(args[i])';
const afterCall = 'parseInt(String(args[i]))';
assert.equal(beforeText.split(beforeCall).length - 1, 1);
assert.equal(afterText, beforeText.replace(beforeCall, afterCall));
// Exercise the stock library's entire parser with exactly the inventoried call
// replaced, preserving its real option declarations, validation and diagnostics.
const file = require.resolve('typescript');
const libraryText = fs.readFileSync(file, 'utf8');
assert.equal(libraryText.split(beforeCall).length - 1, 1);
const library = new Module(file, module);
library.filename = file;
library.paths = Module._nodeModulePaths(require('node:path').dirname(file));
library._compile(libraryText.replace(beforeCall, afterCall), file);
const repaired = library.exports;
const rows = [];
for (const value of [undefined, '12', '12.9', '0x10', 'abc', 'null']) {
    const args = ['--maxNodeModuleJsDepth', ...(value === undefined ? [] : [value])];
    const a = ts.parseCommandLine(args);
    const b = repaired.parseCommandLine(args);
    assert.deepEqual(b, a);
    if (value === undefined) {
        assert(Number.isNaN(b.options.maxNodeModuleJsDepth));
        assert.deepEqual(b.errors.map(d => d.code), [6044]);
    }
    else if (value === 'null') assert.equal(b.options.maxNodeModuleJsDepth, undefined);
    else assert(Object.is(b.options.maxNodeModuleJsDepth, parseInt(value)));
    const result = b.options.maxNodeModuleJsDepth;
    rows.push({args, value: Number.isNaN(result) ? 'NaN' : result === undefined ? 'undefined' : result,
        errors: b.errors.map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}))});
}
// "null" bypasses the numeric branch, but its two parseInt forms also agree.
assert(Object.is(parseInt('null'), parseInt(String('null'))));
const parsed = repaired.parseCommandLine(['--init', '--outFile', 'null']);
assert.equal(parsed.errors.length, 0);
assert(Object.hasOwn(parsed.options, 'outFile'));
assert.equal(parsed.options.outFile, undefined);
const config = repaired.generateTSConfig(parsed.options, '\n');
assert(config.includes('"outFile": undefined,'));
console.log(JSON.stringify({version: ts.version, D031: rows, D036: {unasserted: true,
    parseErrors: 0, emittedLine: config.split('\n').find(line => line.includes('"outFile"'))}}, null, 2));
