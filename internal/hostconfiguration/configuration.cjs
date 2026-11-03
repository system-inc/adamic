'use strict';
const registry = require('./configurations.json');
function nodeFor(row, number) {
    const configuration = row.configuration;
    if (configuration == null) return process.execPath;
    const spec = registry[configuration.name];
    if (!Object.hasOwn(registry, configuration.name)) throw Error(`row ${number}: unknown host configuration ${JSON.stringify(configuration.name)}`);
    if (['node', 'configure', 'ruling'].some(key => configuration[key] !== spec[key])) throw Error(`row ${number}: invalid metadata for host configuration ${configuration.name}`);
    const binary = process.env[spec.environment];
    if (!binary) console.log(`SKIP: row ${number} needs ${configuration.name}; set ${spec.environment}`);
    return binary;
}
module.exports = { nodeFor };
