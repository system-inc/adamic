'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { plan } = require('./adapt.cjs');
const pins = require('./pins.json');
const runtimePins = require('./runtime-pins.json');

const rule = pins.find(pin => pin.owner === 'isNullOrUndefined');
const source = 'function isNullOrUndefined(x: any) { return x === undefined || x === null; }\r\n';

test('changes only the pinned binding and replays without edits', () => {
    const result = plan(rule.file, source, [rule]);
    assert.equal(result.edits, 1);
    assert.equal(result.output, source.replace('x: any', 'x: unknown'));
    assert.equal(plan(rule.file, result.output, [rule]).edits, 0);
});

test('rejects a binding outside its two reviewed snapshots', () => {
    assert.throws(() => plan(rule.file, source.replace('x: any', 'x: object'), [rule]), /snapshot drift/);
});

test('keeps compareDataObjects with the upstream implementation', () => {
    assert.equal(runtimePins.some(rule => rule.owner === 'compareDataObjects'), false,
        'compareDataObjects stays deferred until primitive receiver and getter semantics are supported');
});

const fs = require('node:fs');
const { tmpdir } = require('node:os');
const path = require('node:path');
const bindingViews = require('../99-correctness-binding-views/diagnostic-engine.cjs');
const [bindingImport] = require('../99-correctness-core-multimap/diagnostic-owners.json');
const multimap = require('../99-correctness-core-multimap/adapt.cjs');
const tagsCache = require('../99-correctness-tags-cache-table/adapt.cjs');
const tagsPins = require('../99-correctness-tags-cache-table/owners.json');
const utilitiesFile = 'src/compiler/utilities.ts';

function addThirdName(text) {
    const anchor = '    isAdamicJsonObject,\n';
    assert.equal(text.split(anchor).length, 2);
    return text.replace(anchor, '    checkedAdamicJsonObject,\n' + anchor);
}

test('binding views accepts the widened namespace import', () => {
    for (const { before, after } of bindingImport.variants) {
        assert.equal(bindingViews.inspect(bindingImport, before).state, false);
        assert.equal(bindingViews.inspect(bindingImport, after).state, true);
    }
});

test('binding views refuses a third namespace specifier', () => {
    assert.throws(() => bindingViews.inspect(bindingImport, addThirdName(bindingImport.before)),
        /namespace import \.\/_namespaces\/ts\.js: owner drift or half adaptation/);
});

const utilitiesPins = multimap.pins.filter(pin => pin.file === utilitiesFile);
const multimapSource = utilitiesPins.map(pin => pin.before).join('\n\n');

test('core multimap accepts the widened namespace import', () => {
    const result = multimap.inspect(utilitiesFile, multimapSource);
    assert.equal(result.adapted, false);
    assert.equal(result.edits.length, 2);
});

test('core multimap refuses a third namespace specifier', () => {
    assert.throws(() => multimap.inspect(utilitiesFile, addThirdName(multimapSource)),
        /changed or half-adapted owner \.\/_namespaces\/ts\.js/);
});

function tagsFixture(t, utilitiesImport) {
    const root = fs.mkdtempSync(path.join(tmpdir(), 'adamic-tags-cache-'));
    t.after(() => fs.rmSync(root, { recursive: true, force: true }));
    for (const file of new Set(tagsPins.map(pin => pin.file))) {
        const owners = tagsPins.filter(pin => pin.file === file);
        let text = owners.map(pin => pin.before).join('\n\n');
        if (file === 'src/compiler/types.ts') {
            const [container, deletion] = owners;
            text = container.before + deletion.before + 'interface LocalsContainer {}\n';
        }
        if (file === 'src/compiler/parser.ts') {
            const [performanceImport, currentNode] = owners;
            text = performanceImport.before + '\nnamespace Parser {\n' + currentNode.before + '\n}';
        }
        if (file === utilitiesFile) text = text.replace(owners[0].before, utilitiesImport);
        const name = path.join(root, file);
        fs.mkdirSync(path.dirname(name), { recursive: true });
        fs.writeFileSync(name, text);
    }
    return root;
}

const tagsImport = tagsPins.find(pin => pin.file === utilitiesFile && pin.kind === 'ImportDeclaration').before;

test('tags cache accepts the widened namespace import', t => {
    assert.equal(tagsCache.adapt(tagsFixture(t, tagsImport)).files, 5);
});

test('tags cache refuses a third namespace specifier', t => {
    assert.throws(() => tagsCache.adapt(tagsFixture(t, addThirdName(tagsImport))),
        /src\/compiler\/utilities\.ts: \.\/_namespaces\/ts\.js: drift or half-adapted owner/);
});
