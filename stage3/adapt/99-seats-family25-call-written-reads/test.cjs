'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const ts = require('typescript');
const { edits } = require('./adapt.cjs');

test('README lists the scoped exposed stops', () => {
    const readme = fs.readFileSync(path.join(__dirname, 'README.md'), 'utf8');
    const reason = 'a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables';
    for (const location of ['core.ts:406:47', 'core.ts:409:48']) {
        const rows = readme.split('\n').filter(line => line.startsWith('| `' + location + '` |'));
        assert.equal(rows.length, 1, location);
        assert.deepEqual(rows[0].split('|').slice(1, -1).map(value => value.trim()), [
            '`' + location + '`',
            'Narrowed union with an uncheckable object tag',
            '`getEffectiveTypeParameterDeclarations`',
            reason,
        ], location);
    }
});

test('assignment position preserves the result and getter order', () => {
    assert.equal(ts.version, '6.0.3');
    const [, before, after] = edits['checker.ts'].find(([name]) => name === 'assignment position');
    const source = `
interface SymbolState { readonly lastAssignmentPos: number | undefined }
function isSymbolAssignedDefinitely(symbol: SymbolState) {
    if (symbol.lastAssignmentPos !== undefined) {
        return symbol.lastAssignmentPos < 0;
    }
    return isSymbolAssigned(symbol) && symbol.lastAssignmentPos !== undefined && symbol.lastAssignmentPos < 0;
}
`;
    assert.equal(source.split(before).length, 2);
    const adapted = source.replace(before, after);

    function observe(input, positions, assigned) {
        const calls = [];
        let next = 0;
        const symbol = {
            get lastAssignmentPos() {
                assert(next < positions.length, 'unexpected extra property read');
                calls.push('read');
                return positions[next++];
            },
        };
        const isSymbolAssigned = receiver => {
            assert.equal(receiver, symbol);
            calls.push('call');
            return assigned;
        };
        const { outputText } = ts.transpileModule(input, { compilerOptions: { target: ts.ScriptTarget.ES2020 } });
        const check = new Function('isSymbolAssigned', outputText + '\nreturn isSymbolAssignedDefinitely;');
        return { result: check(isSymbolAssigned)(symbol), calls };
    }

    for (const input of [source, adapted]) {
        assert.deepEqual(observe(input, [undefined, -1, 0], true), {
            result: false,
            calls: ['read', 'call', 'read', 'read'],
        });
        assert.deepEqual(observe(input, [undefined, -1, -1], true), {
            result: true,
            calls: ['read', 'call', 'read', 'read'],
        });
        assert.deepEqual(observe(input, [undefined], false), {
            result: false,
            calls: ['read', 'call'],
        });
    }
});
