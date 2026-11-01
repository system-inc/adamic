'use strict';
// Scratch-only runtime instrumentation. No compiler source is changed.
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const counts = {}, samples = {}, patched = [];
global.__95d = (site, value, detail) => {
    const state = value === undefined ? 'absent' : 'present';
    const key = site + ':' + state;
    counts[key] = (counts[key] || 0) + 1;
    if (!samples[key]) samples[key] = [];
    if (samples[key].length < (state === 'absent' ? 100 : 3)) {
        samples[key].push({test: global.__95dTest, detail, stack: new Error().stack});
    }
    return value;
};
const compile = Module.prototype._compile;
Module.prototype._compile = function(text, filename) {
    if (filename.startsWith(process.env.OBSERVE_TREE + '/')) {
        const replacements = [
            ['reference.flowNode = file.endFlowNode;', 'reference.flowNode = global.__95d("12034", file, {file: file && file.fileName, declarations: symbol.declarations.length}).endFlowNode;'],
            ['const constraint = getConstraintOfTypeParameter(relevantTypeParameter);', 'const constraint = getConstraintOfTypeParameter(global.__95d("42849", relevantTypeParameter, {file: node.getSourceFile().fileName, text: node.parent.getText(), position: typeArgumentPosition, arity: typeParameters.length}));'],
            ['getSymbolAtLocation(copy.declarations[0].parent)', 'getSymbolAtLocation(global.__95d("50153", copy.declarations[0], {infos: infos.length, declarations: copy.declarations.length}).parent)'],
        ];
        for (const [before, after] of replacements) {
            if (text.includes(before)) {
                const hits = text.split(before).length - 1;
                if (hits !== 1) throw new Error('instrumentation occurrence drift: ' + before);
                text = text.replace(before, after);
                patched.push({filename, site: after.match(/"(\d+)"/)[1]});
            }
        }
    }
    return compile.call(this, text, filename);
};
const load = Module._load;
Module._load = function(name, ...args) {
    const value = load.call(this, name, ...args);
    if (name === 'mocha' && value.Runnable && !value.Runnable.prototype.__95dPatched) {
        const run = value.Runnable.prototype.run;
        value.Runnable.prototype.run = function(...args) {
            global.__95dTest = this.fullTitle();
            return run.apply(this, args);
        };
        value.Runnable.prototype.__95dPatched = true;
    }
    return value;
};
process.on('exit', () => {
    if (process.env.OBSERVE_OUT) fs.writeFileSync(path.join(process.env.OBSERVE_OUT, process.pid + '.json'), JSON.stringify({counts, samples, patched}, null, 2));
});
