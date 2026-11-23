'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const ts = require('typescript');
const { relative, findOwner, plan } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <applied tree>');
const tree = process.argv[2];
const adapterRef = 'refs/stage3/99-seats-build22-104-module-chain-arguments';
const { ModuleResolutionKind } = ts;

function snapshot(side) {
    return execFileSync('git', ['-C', tree, 'show', `${adapterRef}-${side}:${relative}`],
        { encoding: 'utf8', maxBuffer: 4 * 1024 * 1024 });
}

const baseline = snapshot('before');
const adapted = snapshot('after');

function compile(text) {
    const owner = findOwner(text).getText().replace(/^export /, '');
    const js = ts.transpileModule(owner, {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
    }).outputText;
    return new Function('Diagnostics', 'ModuleResolutionKind', 'chainDiagnosticMessages',
        'getEmitModuleResolutionKind', 'nodeModulesPathPart', 'mangleScopedPackageName',
        `${js}\nreturn createModuleNotFoundChain;`);
}

function run(compileOwner, scenario) {
    const trace = [];
    const sourceFile = {};
    const moduleReference = '@scope/pkg/subpath';
    const packageName = '@scope/pkg';
    const mode = ts.ModuleKind.ESNext;
    const host = {
        getResolvedModule(file, reference, resolutionMode) {
            assert.equal(file, sourceFile);
            trace.push(['getResolvedModule', reference, resolutionMode]);
            return scenario.resolved;
        },
        getCompilerOptions() {
            trace.push(['getCompilerOptions']);
            return { moduleResolution: ModuleResolutionKind[scenario.resolution] };
        },
        typesPackageExists(name) {
            trace.push(['typesPackageExists', name]);
            return scenario.typesPackageExists;
        },
        packageBundlesTypes(name) {
            trace.push(['packageBundlesTypes', name]);
            return scenario.packageBundlesTypes;
        },
    };
    const Diagnostics = new Proxy({}, {
        get(_target, name) {
            trace.push(['Diagnostics', name]);
            return name;
        },
    });
    const createChain = compileOwner(
        Diagnostics, ModuleResolutionKind,
        (details, message, ...args) => {
            trace.push(['chainDiagnosticMessages', details, message, ...args]);
            return { details, message, args };
        },
        options => {
            trace.push(['getEmitModuleResolutionKind', options.moduleResolution]);
            return options.moduleResolution;
        },
        '/node_modules/',
        name => {
            trace.push(['mangleScopedPackageName', name]);
            return 'scope__pkg';
        },
    );
    const { repopulateInfo, ...chain } = createChain(sourceFile, host, moduleReference, mode, packageName);
    return { chain, repopulateInfo: repopulateInfo(), trace };
}

const original = compile(baseline);
const direct = compile(adapted);
const missingAlternates = [
    { name: 'no resolved module', resolved: undefined },
    { name: 'resolved module without alternateResult', resolved: { resolvedModule: undefined } },
    { name: 'empty-string alternate', resolved: { resolvedModule: undefined, alternateResult: '' } },
];
const fallbacks = [
    { name: 'types package exists', typesPackageExists: true, packageBundlesTypes: false },
    { name: 'package bundles types', typesPackageExists: false, packageBundlesTypes: true },
    { name: 'neither package fallback', typesPackageExists: false, packageBundlesTypes: false },
];
const cases = missingAlternates.flatMap(({ name, resolved }) => fallbacks.map(fallback => ({
    ...fallback, name: `${name}, ${fallback.name}`, resolved,
})));
const regularPath = '/node_modules/pkg/index.d.ts';
const typesPath = '/node_modules/@types/pkg/index.d.ts';
cases.push({
    name: 'Node10 alternate', resolution: 'Node10',
    resolved: { resolvedModule: undefined, alternateResult: typesPath },
});
for (const resolution of ['Classic', 'Node16', 'NodeNext', 'Bundler', 'NodeJs']) {
    cases.push(
        {
            name: `${resolution} alternate without @types`, resolution,
            resolved: { resolvedModule: undefined, alternateResult: regularPath },
        },
        {
            name: `${resolution} alternate with @types`, resolution,
            resolved: { resolvedModule: undefined, alternateResult: typesPath },
        },
    );
}

for (const scenario of cases) {
    const expected = run(original, scenario);
    assert.deepStrictEqual(run(direct, scenario), expected, scenario.name);
    console.log(`PASS ${scenario.name}: ${JSON.stringify(expected)}`);
}

// Check replay after behavior so an output drift reports the case it breaks.
assert.equal(plan(baseline), adapted);
assert.equal(plan(adapted), adapted);
