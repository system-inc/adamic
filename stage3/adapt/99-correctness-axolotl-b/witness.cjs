'use strict';
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');

const originalParseInt = globalThis.parseInt;
const numericReads = [];
let numeric;
try {
    globalThis.parseInt = function (value, radix) {
        numericReads.push({type: typeof value, absent: value === undefined});
        return originalParseInt(value, radix);
    };
    numeric = ts.parseCommandLine(['--maxNodeModuleJsDepth']);
}
finally { globalThis.parseInt = originalParseInt; }
assert(numericReads.some(read => read.absent));
assert(Number.isNaN(numeric.options.maxNodeModuleJsDepth));
assert(numeric.errors.some(error => error.code === 6044));

const args = ['--init', '--outFile', 'null'];
const parsed = ts.parseCommandLine(args);
assert.equal(parsed.errors.length, 0);
assert(Object.hasOwn(parsed.options, 'outFile'));
assert.equal(parsed.options.outFile, undefined);
const originalStringify = JSON.stringify;
const absentStringifyCalls = [];
let config;
try {
    JSON.stringify = function (value, ...rest) {
        const result = originalStringify.call(JSON, value, ...rest);
        if (value === undefined) absentStringifyCalls.push({inputType: typeof value, resultType: typeof result});
        return result;
    };
    config = ts.generateTSConfig(parsed.options, '\n');
}
finally { JSON.stringify = originalStringify; }
assert.equal(absentStringifyCalls.length, 1);
assert(config.includes('"outFile": undefined,'));
console.log(JSON.stringify({version: ts.version, D031: {args: ['--maxNodeModuleJsDepth'], numericReads,
    valueIsNaN: true, errors: numeric.errors.map(d => ({code: d.code, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}))},
    D036: {args, parseErrors: 0, ownOutFile: true, absentStringifyCalls, emittedLine: config.split('\n').find(line => line.includes('"outFile"'))}}, null, 2));
