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
const owners = [
    {
        "file": "builder.ts",
        "owner": "createBuilderProgramUsingIncrementalBuildInfo",
        "ownerBeforeSha256": "965bed9f73584c873456ca4b443b690a3e02aab1fc41df417ef2830433ee4985",
        "ownerAfterSha256": "758b6b93ee05d9c81a7bf3d2e9cbb022414cf15ec75866589a7f6973b8f32869",
        "ownerPath": [
            "createBuilderProgramUsingIncrementalBuildInfo"
        ]
    },
    {
        "file": "checker.ts",
        "owner": "compareTypes",
        "ownerBeforeSha256": "057ec655125a893e098601a737a46d6e677cbc591e57fc2f20f2600a7db57d9a",
        "ownerAfterSha256": "1269a511cfeb932615782b511faa0b856befd02fc21c88eef1388955e08db30e",
        "ownerPath": [
            "createTypeChecker",
            "compareTypes"
        ]
    },
    {
        "file": "commandLineParser.ts",
        "owner": "serializeOptionBaseObject",
        "ownerBeforeSha256": "a5b45fea04f9349ca3a04d8a30144b4a4921a2f38e5effec238ae822910b8677",
        "ownerAfterSha256": "d48c510063b48bbf4a67df85f3ee19ee186848872fcc19ad705858567ce01bfe",
        "ownerPath": [
            "serializeOptionBaseObject"
        ]
    },
    {
        "file": "commandLineParser.ts",
        "owner": "convertToOptionValueWithAbsolutePaths",
        "ownerBeforeSha256": "ea142ef58b5efbd6659c5bf4d0697082ef2c863eefd5d0a73a6484af99a39644",
        "ownerAfterSha256": "536df146babb623f40e753c83e4b70fc9fdfe96a973c27f283a776aa3f6c75d4",
        "ownerPath": [
            "convertToOptionValueWithAbsolutePaths"
        ]
    },
    {
        "file": "commandLineParser.ts",
        "owner": "getConfigFileSpecs",
        "ownerBeforeSha256": "d443e80f38a74efddcd6756b986ac6e21e4f6eac65d257684a57f3a635d2e99a",
        "ownerAfterSha256": "880eb8e312456ad49ebc946b468a206dc8ecafc17c7a37e0739f58dd5527de42",
        "ownerPath": [
            "parseJsonConfigFileContentWorker",
            "getConfigFileSpecs"
        ]
    },
    {
        "file": "commandLineParser.ts",
        "owner": "convertJsonOption",
        "ownerBeforeSha256": "5be7e83e44e5589474278b557f0d7e1d04eb17d6a1326c211c00aa1c202bab71",
        "ownerAfterSha256": "3760308cedabccc1ff3a3263f7ccd6e9a923e75078b9351d38d6cbb7daef891b",
        "ownerPath": [
            "convertJsonOption"
        ]
    },
    {
        "file": "program.ts",
        "owner": "resolveTypeReferenceDirectiveNamesReusingOldState",
        "ownerBeforeSha256": "74f982730d1c212089bbea02a524b236d428c640ba3440996c4d3eb3bac6c984",
        "ownerAfterSha256": "ac6510e248df2ecf7b46397b9f236b6232ef38786fcf3ddaf2d9412af3880d53",
        "ownerPath": [
            "createProgram",
            "resolveTypeReferenceDirectiveNamesReusingOldState"
        ]
    },
    {
        "file": "sys.ts",
        "owner": "invokeCallbacks",
        "ownerBeforeSha256": "20e5713d49311137ab280d33cf0bdd579dd6f2395e7f836ac5eeb73879be5379",
        "ownerAfterSha256": "28b9432a1aaa26c07d9318e994296c6c04313225fad5131c6a38b12cf6c9d44f",
        "ownerPath": [
            "createDirectoryWatcherSupportingRecursive",
            "invokeCallbacks"
        ]
    },
    {
        "file": "transformers/generators.ts",
        "owner": "writeOperation",
        "ownerBeforeSha256": "dd271feb4e97a57dd06e17f970fa8a1e54aa037f3faef014fcf8cba0cdfb49f2",
        "ownerAfterSha256": "2ac8098c4b17dcd10bac2bcf41d66d0d5558bb8ba5fced758ee869a5d6c90e0e",
        "ownerPath": [
            "transformGenerators",
            "writeOperation"
        ]
    }
];

function digest(text) { return crypto.createHash('sha256').update(text).digest('hex'); }
function nodes(root, predicate) {
    const found = [];
    (function visit(n) { if (predicate(n)) found.push(n); ts.forEachChild(n, visit); })(root);
    return found;
}
function locate(file, raw) {
    const text = raw.replace(/\r\n/g, '\n');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    const edits = [], states = [];
    for (const pin of owners.filter(o => o.file === file)) {
        const found = nodes(source, n => ts.isFunctionDeclaration(n) && n.body && n.name?.text === pin.owner);
        assert.equal(found.length, 1, `${file}: ${pin.owner} owner count drift`);
        const owner = found[0], hash = digest(owner.getText(source));
        const ownerPath = [];
        for (let node = owner; node; node = node.parent) if (ts.isFunctionDeclaration(node) && node.name) ownerPath.unshift(node.name.text);
        assert.deepEqual(ownerPath, pin.ownerPath, `${file}: moved owner: ${pin.owner}`);
        assert([pin.ownerBeforeSha256, pin.ownerAfterSha256].includes(hash), `${file}: changed or partly adapted owner: ${pin.owner}`);
        const adapted = hash === pin.ownerAfterSha256;
        states.push(adapted);
        for (const site of sites.filter(s => s.file === file && s.owner === pin.owner)) {
            const spelling = adapted ? site.after : site.before;
            const start = text.indexOf(spelling, owner.getStart(source));
            assert(start >= owner.getStart(source) && start + spelling.length <= owner.end, `${file}: statement left owner`);
            assert(text.indexOf(spelling, start + 1) < 0 || text.indexOf(spelling, start + 1) >= owner.end, `${file}: exact statement count drift`);
            if (!adapted) {
                const statements = nodes(owner, n => ts.isStatement(n) && n.getText(source) === site.before);
                assert.equal(statements.length, 1, `${file}: exact statement count drift`);
                assert.equal(statements[0].getStart(source), start);
                edits.push({start, end: start + spelling.length, after: site.after});
            }
        }
    }
    assert(states.length && states.every(s => s === states[0]), 'partly adapted file');
    return {text, adapted: states[0], edits};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = [...new Set(owners.map(o => o.file))].map(file => {
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
module.exports = {adapt, sites, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
