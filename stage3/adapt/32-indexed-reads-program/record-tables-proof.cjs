"use strict";
// Independent literal port for the emitter comparison, without invoking the planner.
function expected(text) {
    const replacements = [
        ["    clear,\r\n", "", 1],
        ["    type MappedType,", "    type MapLike,\r\n    type MappedType,", 1],
        ["const visitedTypes: Type[] = []; // Sparse array from id to type", "const visitedTypes: MapLike<Type> = {}; // Record from id to type", 1],
        ["const visitedSymbols: Symbol[] = []; // Sparse array from id to symbol", "const visitedSymbols: MapLike<Symbol> = {}; // Record from id to symbol", 1],
        ["clear(visitedTypes);", "for (const key in visitedTypes) delete visitedTypes[key];", 2],
        ["clear(visitedSymbols);", "for (const key in visitedSymbols) delete visitedSymbols[key];", 2],
    ];
    for (const [before, after, count] of replacements) {
        if (text.split(before).length !== count + 1) throw new Error("independent record table port drift: " + before);
        text = text.replaceAll(before, after);
    }
    return text;
}
module.exports = { expected };
