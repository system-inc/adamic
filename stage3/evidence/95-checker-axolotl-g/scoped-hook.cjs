'use strict';
// Audit only: replace the selection in memory and trace only its new no-symbol exit.
const Module = require('node:module'), assert = require('node:assert/strict'), path = require('node:path');
const target = path.resolve(process.env.G_CHECKER || '/workspace/scratch/95d-before/built/local/typescript.js');
const compile = Module.prototype._compile;
Module.prototype._compile = function(text, file) {
    if (file === target) {
        const before = 'copy.parent = type.aliasSymbol ? type.aliasSymbol : type.symbol ? type.symbol : getSymbolAtLocation(copy.declarations[0].parent);';
        assert.equal(text.split(before).length - 1, 1);
        text = text.replace(before, `if (type.aliasSymbol) { copy.parent = type.aliasSymbol; }
          else if (type.symbol) { copy.parent = type.symbol; }
          else {
            const declaration = copy.declarations[0];
            if (declaration === undefined) {
              if (global.__95gEvents) global.__95gEvents.push(new Error('scoped no-symbol exit').stack);
              return undefined;
            }
            copy.parent = getSymbolAtLocation(declaration.parent);
          }`);
    }
    return compile.call(this, text, file);
};
