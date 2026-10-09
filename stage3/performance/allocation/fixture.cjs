// A host allocation witness, executed as JavaScript through the observation runner.
// Exact totals: Node 13, Symbol 9, Type 12, Signature 5.
class FixtureNode { constructor() { globalThis.__allocation.construct('Node4'); this.value = 1; } }
class FixtureSymbol { constructor() { globalThis.__allocation.construct('Symbol4'); this.value = 2; } }
class FixtureType { constructor() { globalThis.__allocation.construct('Type3'); this.value = 3; } }
class FixtureSignature { constructor() { globalThis.__allocation.construct('Signature2'); this.value = 4; } }
let roots = [];
function make(constructor, count, keep) {
    for (let index = 0; index < count; index++) {
        const value = new constructor();
        if (index < keep) roots.push(value);
    }
}
exports.parse = () => { make(FixtureNode, 10, 6); make(FixtureSymbol, 4, 2); };
exports.bind = () => { make(FixtureNode, 2, 1); make(FixtureSymbol, 3, 1); make(FixtureType, 5, 3); make(FixtureSignature, 2, 1); };
exports.check = () => { make(FixtureNode, 1, 0); make(FixtureSymbol, 2, 1); make(FixtureType, 7, 4); make(FixtureSignature, 3, 1); };
exports.release = () => { roots = []; };
