"use strict";
const fs = require("node:fs");
const output = process.env.ENV_PRESENCE_COUNTS;
if (!output) throw new Error("ENV_PRESENCE_COUNTS required");
globalThis.__envPresence1965 = function(sys) {
    const row = {pid: process.pid, truthy: !!sys, function: !!sys && typeof sys.getEnvironmentVariable === "function"};
    fs.appendFileSync(output, JSON.stringify(row) + "\n");
};
// Compiler bundles also execute in upstream's explicit VM sandboxes.
const vm = require('node:vm');
const expose = sandbox => {
    if (sandbox && !Object.hasOwn(sandbox, '__envPresence1965'))
        Object.defineProperty(sandbox, '__envPresence1965', {value: globalThis.__envPresence1965, configurable: true});
    return sandbox;
};
for (const name of ['createContext', 'runInNewContext']) {
    const original = vm[name];
    vm[name] = name === 'createContext'
        ? function(sandbox = {}, ...args) { return original.call(this, expose(sandbox), ...args); }
        : function(code, sandbox = {}, ...args) { return original.call(this, code, expose(sandbox), ...args); };
}
