'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
assert.equal(process.argv.length, 4, 'usage: node d036-post99g.cjs <99g tree> <result>');
const tree = path.resolve(process.argv[2]);
const source = fs.readFileSync(path.join(tree, 'src/compiler/commandLineParser.ts'), 'utf8');
assert(source.includes('const formattedValue = hasProperty(options, setting) && value === undefined ? "null" : formatValueOrArray(setting, (value ?? defaultValue) as PresetValue);'));
assert(source.includes('return formatSingleValue(value, map);'));
const ts = require(path.join(tree, 'built/local/typescript.js'));
// A normal paths map retains its declared string-array values. A non-enumerable
// serialization hook is a legal JavaScript object property. This uses the
// exported internal generateTSConfig API, not a command-line argument.
const paths = { '@app/*': ['./src/*'] };
let hookCalls = 0;
Object.defineProperty(paths, 'toJSON', {value() { hookCalls++; return undefined; }});
assert.deepEqual(Object.keys(paths), ['@app/*']);
const options = { paths };
assert(Object.hasOwn(options, 'paths'));
assert.notEqual(options.paths, undefined);
const stringify = JSON.stringify;
const observed = [];
let config;
try {
    JSON.stringify = function(value, ...rest) {
        const result = stringify.call(JSON, value, ...rest);
        if (value === paths) observed.push({inputType: typeof value, resultType: typeof result, stack: new Error().stack});
        return result;
    };
    config = ts.generateTSConfig(options, '\n');
}
finally { JSON.stringify = stringify; }
assert.equal(hookCalls, 1);
assert.equal(observed.length, 1);
assert.equal(observed[0].resultType, 'undefined');
for (const caller of ['formatSingleValue', 'formatValueOrArray', 'emitOption', 'generateTSConfig']) assert(observed[0].stack.includes(caller));
assert(config.includes('"paths": undefined,'));
const plugin = { name: 'example', toJSON() { return undefined; } };
const arrayObserved = [];
let pluginConfig;
try {
    JSON.stringify = function(value, ...rest) {
        const result = stringify.call(JSON, value, ...rest);
        if (value === plugin) arrayObserved.push({resultType: typeof result, stack: new Error().stack});
        return result;
    };
    pluginConfig = ts.generateTSConfig({plugins: [plugin]}, '\n');
}
finally { JSON.stringify = stringify; }
assert.equal(arrayObserved.length, 1);
assert.equal(arrayObserved[0].resultType, 'undefined');
assert(pluginConfig.includes('"plugins": [],'));
const result = {version: ts.version, lane: 'exported internal API', option: 'paths', ownOptionValueType: typeof options.paths,
    guardBypassedBecauseValueIsDefined: true, toJSONCalls: hookCalls, stringify: observed,
    emittedLine: config.split('\n').find(line => line.includes('"paths"')), assertionDeclined: true, arrayCallback: {observed: arrayObserved, emittedLine: pluginConfig.split('\n').find(line => line.includes('"plugins"'))}};
fs.writeFileSync(path.resolve(process.argv[3]), stringify(result, null, 2) + '\n');
console.log(stringify(result, null, 2));
