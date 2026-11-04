'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const sites = [
    {
        "file": "builder.ts",
        "owner": "createBuilderProgramUsingIncrementalBuildInfo",
        "before": "emitSignatures!.set(\n                    key,\n                    !isString(value[1]) && !value[1].length ?\n                        // File signature is emit signature but differs in map\n                        [emitSignatures!.get(key)! as string] :\n                        value[1],\n                );",
        "after": "emitSignatures!.set(\n                    key,\n                    !isString(value[1]) && !value[1].length ?\n                        // File signature is emit signature but differs in map\n                        [((signature: EmitSignature | undefined) => {\n                            Debug.assert(typeof signature === \"string\");\n                            return signature;\n                        })(emitSignatures!.get(key))] :\n                        value[1],\n                );"
    },
    {
        "file": "checker.ts",
        "owner": "compareTypes",
        "before": "const c = compareComparableValues((t1 as LiteralType).value as string, (t2 as LiteralType).value as string);",
        "after": "const value1 = (t1 as LiteralType).value;\n            const value2 = (t2 as LiteralType).value;\n            Debug.assert(typeof value1 === \"string\" && typeof value2 === \"string\");\n            const c = compareComparableValues(value1, value2);"
    },
    {
        "file": "checker.ts",
        "owner": "compareTypes",
        "before": "const c = compareComparableValues((t1 as LiteralType).value as number, (t2 as LiteralType).value as number);",
        "after": "const value1 = (t1 as LiteralType).value;\n            const value2 = (t2 as LiteralType).value;\n            Debug.assert(typeof value1 === \"number\" && typeof value2 === \"number\");\n            const c = compareComparableValues(value1, value2);"
    },
    {
        "file": "commandLineParser.ts",
        "owner": "serializeOptionBaseObject",
        "before": "result.set(name, getRelativePathFromFile(pathOptions.configFilePath, getNormalizedAbsolutePath(value as string, getDirectoryPath(pathOptions.configFilePath)), getCanonicalFileName!));",
        "after": "Debug.assert(typeof value === \"string\");\n                        result.set(name, getRelativePathFromFile(pathOptions.configFilePath, getNormalizedAbsolutePath(value, getDirectoryPath(pathOptions.configFilePath)), getCanonicalFileName!));"
    },
    {
        "file": "commandLineParser.ts",
        "owner": "convertToOptionValueWithAbsolutePaths",
        "before": "return toAbsolutePath(value as string);",
        "after": "Debug.assert(typeof value === \"string\");\n            return toAbsolutePath(value);"
    },
    {
        "file": "commandLineParser.ts",
        "owner": "getConfigFileSpecs",
        "before": "excludeSpecs = filter([outDir, declarationDir], d => !!d) as string[];",
        "after": "excludeSpecs = filter([outDir, declarationDir], (d): d is string => !!d);"
    },
    {
        "file": "commandLineParser.ts",
        "owner": "convertJsonOption",
        "before": "return convertJsonOptionOfCustomType(opt as CommandLineOptionOfCustomType, value as string, errors, valueExpression, sourceFile);",
        "after": "return typeof value === \"string\" ? convertJsonOptionOfCustomType(opt as CommandLineOptionOfCustomType, value, errors, valueExpression, sourceFile) : undefined;"
    },
    {
        "file": "program.ts",
        "owner": "resolveTypeReferenceDirectiveNamesReusingOldState",
        "before": "return resolveNamesReusingOldState({\n            entries: typeDirectiveNames,\n            containingFile,\n            containingSourceFile,\n            redirectedReference: containingSourceFile && getRedirectReferenceForResolution(containingSourceFile),\n            nameAndModeGetter: typeReferenceResolutionNameAndModeGetter,\n            resolutionWorker: resolveTypeReferenceDirectiveNamesWorker,\n            getResolutionFromOldProgram: (name, mode) =>\n                containingSourceFile ?\n                    oldProgram?.getResolvedTypeReferenceDirective(containingSourceFile, name, mode) :\n                    oldProgram?.getAutomaticTypeDirectiveResolutions()?.get(name, mode),\n            getResolved: getResolvedTypeReferenceDirectiveFromResolution,\n            canReuseResolutionsInFile: () =>\n                containingSourceFile ?\n                    containingSourceFile === oldProgram?.getSourceFile(containingSourceFile.fileName) && !hasInvalidatedResolutions(containingSourceFile.path) :\n                    !hasInvalidatedResolutions(toPath(containingFile as string)),\n        });",
        "after": "return resolveNamesReusingOldState({\n            entries: typeDirectiveNames,\n            containingFile,\n            containingSourceFile,\n            redirectedReference: containingSourceFile && getRedirectReferenceForResolution(containingSourceFile),\n            nameAndModeGetter: typeReferenceResolutionNameAndModeGetter,\n            resolutionWorker: resolveTypeReferenceDirectiveNamesWorker,\n            getResolutionFromOldProgram: (name, mode) =>\n                containingSourceFile ?\n                    oldProgram?.getResolvedTypeReferenceDirective(containingSourceFile, name, mode) :\n                    oldProgram?.getAutomaticTypeDirectiveResolutions()?.get(name, mode),\n            getResolved: getResolvedTypeReferenceDirectiveFromResolution,\n            canReuseResolutionsInFile: () =>\n                typeof containingFile !== \"string\" ?\n                    containingFile === oldProgram?.getSourceFile(containingFile.fileName) && !hasInvalidatedResolutions(containingFile.path) :\n                    !hasInvalidatedResolutions(toPath(containingFile)),\n        });"
    },
    {
        "file": "sys.ts",
        "owner": "invokeCallbacks",
        "before": "(existing as string[]).push(...fileNames);",
        "after": "if (Array.isArray(existing)) existing.push(...fileNames);"
    },
    {
        "file": "transformers/generators.ts",
        "owner": "writeOperation",
        "before": "return writeBreak(args[0] as Label, location);",
        "after": "const breakLabel = args[0];\n                Debug.assert(typeof breakLabel === \"number\");\n                return writeBreak(breakLabel, location);"
    },
    {
        "file": "transformers/generators.ts",
        "owner": "writeOperation",
        "before": "return writeBreakWhenTrue(args[0] as Label, args[1] as Expression, location);",
        "after": "const breakWhenTrueLabel = args[0];\n                Debug.assert(typeof breakWhenTrueLabel === \"number\");\n                return writeBreakWhenTrue(breakWhenTrueLabel, args[1] as Expression, location);"
    },
    {
        "file": "transformers/generators.ts",
        "owner": "writeOperation",
        "before": "return writeBreakWhenFalse(args[0] as Label, args[1] as Expression, location);",
        "after": "const breakWhenFalseLabel = args[0];\n                Debug.assert(typeof breakWhenFalseLabel === \"number\");\n                return writeBreakWhenFalse(breakWhenFalseLabel, args[1] as Expression, location);"
    }
];
const hashes = {
    "builder.ts": {
        "before": "f1624d3e9d6106738e9ab8e92a21e88a405b84844f5dbdbba57e6ff885f4dedc",
        "after": "ff1909b2814808697b6a8d41488c6e8fab84959f630650e1622e973d84f52e77"
    },
    "checker.ts": {
        "before": "28d07f25c5c23f491f7ab63acf5767e583fc520fa5b667e9c3ba7ea49f93b17a",
        "after": "f54b8de63093e63c4d8b370657f18fc6b9f5aba49d58377bf626e1e048ab5977"
    },
    "commandLineParser.ts": {
        "before": "afd8af88ec70e2065792c6599f0f721ab4736ea72bcf7b611b4f64aa3c17e810",
        "after": "b551febe5fe1cefc1e258a401b7111ddea30a97e0a4d6159cd16fdbaed29887b"
    },
    "program.ts": {
        "before": "4197d5ddd1decfeb63ce93fc03761c17f0b723ce017436af3c639f82c7f12c0b",
        "after": "a967d016e05e1262cdafedc99c4bf5a75c8eeeb40c9f8882e75d567dbd9fe895"
    },
    "sys.ts": {
        "before": "dfbbe67d62a12946a181a32913883b51e51e1636a4306c2d25d9abfe977aae22",
        "after": "83ab447c4d131174c0f14e86176650b7361015a2273ed6bb95e3e1aeb55e14b6"
    },
    "transformers/generators.ts": {
        "before": "1e28dec8dc41d832328af3211e7e51aeea9266468922887a0faceb0bd1b28754",
        "after": "214e614d46a0a6303ed26758f8fbad171a6d05af298f146a8f06a89f0e2aeb9a"
    }
};

function digest(text) { return crypto.createHash('sha256').update(text).digest('hex'); }
function nodes(root, predicate) {
    const found = [];
    (function visit(n) { if (predicate(n)) found.push(n); ts.forEachChild(n, visit); })(root);
    return found;
}
function locate(file, raw) {
    const text = raw.replace(/\r\n/g, '\n');
    const hash = digest(text), expected = hashes[file];
    assert(hash === expected.before || hash === expected.after, `${file}: changed or partly adapted file`);
    const adapted = hash === expected.after;
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    const edits = [];
    for (const site of sites.filter(s => s.file === file)) {
        const owners = nodes(source, n => ts.isFunctionDeclaration(n) && n.body && n.name?.text === site.owner);
        assert.equal(owners.length, 1, `${file}: ${site.owner} owner count drift`);
        const spelling = adapted ? site.after : site.before;
        const start = text.indexOf(spelling, owners[0].getStart(source));
        assert(start >= owners[0].getStart(source) && start + spelling.length <= owners[0].end, `${file}: statement left owner`);
        if (!adapted) {
            const statements = nodes(owners[0], n => ts.isStatement(n) && n.getText(source) === site.before);
            assert.equal(statements.length, 1, `${file}: exact statement count drift`);
            assert.equal(statements[0].getStart(source), start);
            edits.push({start, end: start + spelling.length, after: site.after});
        }
    }
    return {text, adapted, edits};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = Object.keys(hashes).map(file => {
        const name = path.join(tree, 'src/compiler', file), raw = fs.readFileSync(name, 'utf8');
        const state = locate(file, raw);
        let result = state.text;
        for (const edit of state.edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.after + result.slice(edit.end);
        assert(locate(file, result).adapted, `${file}: invalid planned result`);
        if (raw.includes('\r\n')) result = result.replace(/\n/g, '\r\n');
        return {name, raw, result, ...state};
    });
    assert(plans.every(p => p.adapted === plans[0].adapted), 'partly adapted tree');
    for (const plan of plans) assert.equal(fs.readFileSync(plan.name, 'utf8'), plan.raw, 'source changed while planning');
    if (!plans[0].adapted) for (const plan of plans) fs.writeFileSync(plan.name, plan.result);
    return {files: plans[0].adapted ? 0 : plans.length, edits: plans.reduce((n, p) => n + p.edits.length, 0), javascriptChanging: true, behaviorChanging: false};
}
module.exports = {adapt, sites, hashes};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
