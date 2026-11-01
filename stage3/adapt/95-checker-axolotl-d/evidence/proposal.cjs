'use strict';
// Evaluate the proposed absence branch in Node's existing build, without building
// or changing any compiler source or built file on disk.
const Module = require('node:module');
const assert = require('node:assert/strict');
const compile = Module.prototype._compile;
Module.prototype._compile = function(text, name) {
    if (name === process.env.OBSERVE_TREE + '/built/local/typescript.js') {
        const before = 'const constraint = getConstraintOfTypeParameter(relevantTypeParameter);';
        assert.equal(text.split(before).length - 1, 1);
        text = text.replace(before, `if (relevantTypeParameter === undefined) {
          throw new TypeError("Cannot read properties of undefined (reading 'resolvedBaseConstraint')");
        }
        ${before}`);
        const parentBefore = 'copy.parent = type.aliasSymbol ? type.aliasSymbol : type.symbol ? type.symbol : getSymbolAtLocation(copy.declarations[0].parent);';
        assert.equal(text.split(parentBefore).length - 1, 1);
        text = text.replace(parentBefore, `if (type.aliasSymbol) {
            copy.parent = type.aliasSymbol;
          } else if (type.symbol) {
            copy.parent = type.symbol;
          } else {
            const declaration = copy.declarations[0];
            if (declaration === undefined) {
              throw new TypeError("Cannot read properties of undefined (reading 'parent')");
            }
            copy.parent = getSymbolAtLocation(declaration.parent);
          }`);

    }
    return compile.call(this, text, name);
};
