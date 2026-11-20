'use strict';
const assert = require('node:assert/strict');
const { createHash } = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');

const checker = 'src/compiler/checker.ts';
const types = 'src/compiler/types.ts';
const edits = [
    [checker, 'createCanonicalSignature', 'map(signature.typeParameters, tp =>', 'map<TypeParameter, Type>(signature.typeParameters, tp =>'],
    [checker, 'getTypeArguments', 'concatenate(type.target.outerTypeParameters, getEffectiveTypeArguments(',
        'concatenate(type.target.outerTypeParameters as readonly TypeParameter[] | undefined, getEffectiveTypeArguments('],
    [checker, 'getTypeFromClassOrInterfaceReference', 'concatenate(type.outerTypeParameters, fillMissingTypeArguments(',
        'concatenate(type.outerTypeParameters as readonly TypeParameter[] | undefined, fillMissingTypeArguments('],
    [checker, 'getTypeFromTypeAliasReference', 'let aliasTypeArguments: Type[] | undefined;',
        'let aliasTypeArguments: readonly Type[] | undefined;'],
    [types, 'ConditionalRoot', 'aliasTypeArguments?: Type[] | undefined;', 'aliasTypeArguments?: readonly Type[] | undefined;'],
];

function inspect(files) {
    const plans = [];
    for (const [file, text] of files) {
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, file);
        const candidates = edits.filter(([ownerFile]) => ownerFile === file);
        for (const [, name, before, after] of candidates) {
            const owners = [];
            ts.forEachChild(source, function visit(node) {
                if ((file === types ? ts.isInterfaceDeclaration(node) : ts.isFunctionDeclaration(node))
                    && node.name?.text === name) owners.push(node);
                ts.forEachChild(node, visit);
            });
            assert.equal(owners.length, 1, `${file}: expected one ${name}`);
            const start = owners[0].getStart(source);
            const original = text.slice(start, owners[0].end);
            const hash = createHash('sha256').update(original.replace(/\r\n/g, '\n')).digest('hex');
            plans.push({ file, name, start, end: owners[0].end, original, hash, before, after });
        }
    }
    assert.equal(plans.length, edits.length);
    return plans;
}

function adapt(files) {
    const pins = require('./type-argument-owners.json');
    const plans = inspect(files);
    for (const plan of plans) {
        const [beforeHash, afterHash] = pins[plan.name];
        assert(plan.hash === beforeHash || plan.hash === afterHash, `${plan.name}: owner text drift`);
        plan.adapted = plan.hash === afterHash;
    }
    const alreadyAdapted = plans.filter(plan => plan.adapted).length;
    assert(alreadyAdapted === 0 || alreadyAdapted === plans.length, 'partly adapted type argument views');
    if (alreadyAdapted) return;

    for (const plan of plans.sort((a, b) => b.start - a.start)) {
        assert.equal(plan.original.split(plan.before).length, 2, `${plan.name}: expected one rewrite site`);
        const text = files.get(plan.file);
        const owner = plan.original.replace(plan.before, () => plan.after);
        files.set(plan.file, text.slice(0, plan.start) + owner + text.slice(plan.end));
    }
    for (const { name, hash } of inspect(files)) assert.equal(hash, pins[name][1], `${name}: adapted owner text drift`);
}

module.exports = { adapt, inspect };
