'use strict';
const registry = require('./configurations.json');
const {spawnSync} = require('node:child_process');
const probes = new Map();
function probe(binary, spec) {
    const key = JSON.stringify([binary, spec.probe, spec.expect]);
    if (!probes.has(key)) {
        const result = spawnSync(binary, ['-p', spec.probe], {encoding:'utf8',timeout:10000});
        const actual = ((result.stdout || '') + (result.stderr || '')).replace(/\n$/, '');
        probes.set(key, result.error || result.status !== 0 || actual !== spec.expect
            ? new Error(`configuration probe ${JSON.stringify(binary)} printed ${JSON.stringify(actual)}; expected ${JSON.stringify(spec.expect)} (execution: ${result.error || result.status})`) : null);
    }
    if (probes.get(key)) throw probes.get(key);
}
function nodeFor(row, number) {
    const configuration = row.configuration;
    if (configuration == null) return process.execPath;
    const spec = registry[configuration.name];
    if (!Object.hasOwn(registry, configuration.name)) throw Error(`row ${number}: unknown host configuration ${JSON.stringify(configuration.name)}`);
    if (['node', 'configure', 'ruling'].some(key => configuration[key] !== spec[key])) throw Error(`row ${number}: invalid metadata for host configuration ${configuration.name}`);
    const binary = process.env[spec.environment];
    if (!binary) console.log(`SKIP: row ${number} needs ${configuration.name}; set ${spec.environment}`);
    if (binary) probe(binary, spec);
    return binary;
}
module.exports = { nodeFor };
