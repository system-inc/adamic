'use strict';
const Module = require('node:module'), assert = require('node:assert/strict');
const compile = Module.prototype._compile;
Module.prototype._compile = function(text, file) {
    if (file === '/workspace/scratch/95d-before/built/local/typescript.js') {
        const before = 'const declarationList = mapDefined(infos, (i) => i.declaration);';
        assert.equal(text.split(before).length - 1, 1);
        text = text.replace(before, before + '\n        if (!declarationList.length) return undefined;');
    }
    return compile.call(this, text, file);
};
