"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const pins = require("./pins.json");
const { twins, watchKeys } = require("./twins.cjs");

const normalize = text => text.replaceAll("\r\n", "\n");
const twinName = name => name.replace(/^create/, "createDefault");

function assertInternal(node) {
    assert(ts.getJSDocTags(node).some(tag => tag.tagName.text === "internal"),
        `${node.name.text} must carry @internal`);
}

function adapt(tree) {
    const cli = "executeCommandLine.ts";
    const directory = path.join(tree, "src/compiler");
    const names = [...Object.keys(pins.owners), cli];
    const files = Object.fromEntries(names.map(name => [name, fs.readFileSync(path.join(directory, name), "utf8")]));
    function parse(name, text) {
        const file = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
        assert.equal(file.parseDiagnostics.length, 0, name);
        return file;
    }
    function one(nodes, name) {
        assert.equal(nodes.length, 1, name);
        return nodes[0];
    }
    function declaration(file, name) {
        return one(file.statements.filter(n => n.name?.text === name && (name !== "createWatchHost" || n.body)), name);
    }
    function imports(file) {
        const statement = one(file.statements.filter(n => ts.isImportDeclaration(n)
            && n.moduleSpecifier.text === "./_namespaces/ts.js"), file.fileName);
        assert(ts.isNamedImports(statement.importClause.namedBindings));
        return statement.importClause.namedBindings;
    }
    function scan(node, predicate) {
        const found = [];
        function visit(child) {
            if (predicate(child)) found.push(child);
            ts.forEachChild(child, visit);
        }
        visit(node);
        return found;
    }
    function ownerText(node, originalWatchResult = false) {
        const start = node.getStart();
        const text = node.getText();
        if (!originalWatchResult) return text;
        return text.slice(0, node.type.getStart() - start) + "WatchHost" + text.slice(node.type.end - start);
    }
    function oldOwners(parsed, adapted) {
        for (const [file, owners] of Object.entries(pins.owners)) {
            for (const [name, expected] of Object.entries(owners)) {
                const pinned = adapted && name === "createWatchHost" ? pins.watchImplementationAfter : expected;
                assert.equal(normalize(ownerText(declaration(parsed[file], name))), pinned, name);
            }
        }
    }
    function parseFiles(input) {
        return Object.fromEntries(names.map(name => [name, parse(name, input[name])]));
    }
    const parsed = parseFiles(files);
    const edits = Object.fromEntries(names.map(name => [name, []]));
    function edit(file, start, end, value) {
        const newline = files[file].includes("\r\n") ? "\r\n" : "\n";
        edits[file].push({ start, end, value: value.replaceAll("\n", newline) });
    }
    const states = [];
    for (const [name, expected] of Object.entries(pins.calls)) {
        const replacement = twinName(name);
        const call = one(scan(parsed[cli], n => ts.isCallExpression(n) && ts.isIdentifier(n.expression)
            && [name, replacement].includes(n.expression.text)), name);
        const spec = one(imports(parsed[cli]).elements.filter(n => [name, replacement].includes(n.name.text)), name);
        const moved = call.expression.text === replacement;
        assert.equal(spec.getText(), moved ? replacement : name, name);
        assert.equal(normalize(call.getText()), moved ? redirected(expected, name) : expected, name);
        states.push(moved);
        if (!moved) {
            edit(cli, call.getStart(), call.end, redirected(expected, name));
            edit(cli, spec.getStart(), spec.end, replacement);
        }
    }
    assert(states.every(state => state === states[0]), "partial CLI adaptation");
    const moved = states[0];
    oldOwners(parsed, moved);
    function watchDeclarations(file, adapted) {
        const implementation = declaration(file, "createWatchHost");
        const nodes = file.statements.filter(n => n.name?.text === "OwnWatchHost"
            || (n.name?.text === "createWatchHost" && !n.body));
        assert.deepEqual(nodes.map(n => normalize(n.getText())), adapted ? pins.watchDeclarations : []);
        for (const node of nodes) assertInternal(node);
        if (adapted) {
            const offset = file.statements.indexOf(implementation) - nodes.length;
            assert(nodes.every((node, i) => file.statements[offset + i] === node), "watch declaration order");
        }
        return implementation;
    }
    const implementation = watchDeclarations(parsed["watch.ts"], moved);
    if (!moved) {
        edit("watch.ts", implementation.type.getStart(), implementation.type.end, "OwnWatchHost");
        const start = implementation.getFullStart();
        edit("watch.ts", start, start, "\n\n" + pins.watchDeclarations
            .map(text => "/** @internal */\n" + text).join("\n\n"));
    }
    for (const [file, additions] of Object.entries(twins)) {
        const nodes = additions.map(text => {
            const parsed = parse(file, text);
            const node = one(parsed.statements, file);
            assert(ts.isFunctionDeclaration(node));
            assertInternal(node);
            assert(!node.typeParameters);
            assert.equal(scan(node.body, n => ts.isAsExpression(n) || ts.isTypeAssertionExpression(n)).length, 0,
                `${node.name.text} must not contain an assertion`);
            return node;
        });
        for (const [i, node] of nodes.entries()) {
            const existing = parsed[file].statements.filter(n => n.name?.text === node.name.text);
            if (moved) assert.equal(normalize(one(existing, node.name.text).getFullText()).trim(), additions[i]);
            else assert.equal(existing.length, 0, node.name.text);
        }
        if (!moved) {
            const anchor = declaration(parsed[file], file === "watch.ts"
                ? "createWatchCompilerHostOfFilesAndCompilerOptions" : "createSolutionBuilderWithWatchHost");
            edit(file, anchor.end, anchor.end, "\n\n" + additions.join("\n\n"));
        }
    }
    const extraImports = ["createEmitAndSemanticDiagnosticsBuilderProgram", "createWatchStatusReporter", "getDefaultLibFileName", "memoize", "normalizePath"];
    const items = imports(parsed["tsbuildPublic.ts"]);
    for (const name of extraImports) {
        const matching = items.elements.filter(n => n.name.text === name);
        assert.equal(matching.length, Number(moved), name);
        if (moved) assert.equal(matching[0].getText(), name, `value import drift: ${name}`);
        assert(!items.parent.isTypeOnly, "expected value import clause");
    }
    if (!moved) edit("tsbuildPublic.ts", items.getStart() + 1, items.getStart() + 1,
        "\n" + extraImports.map(name => `    ${name},`).join("\n"));
    const output = {};
    for (const name of names) {
        let text = files[name];
        let limit = text.length;
        for (const { start, end, value } of edits[name].sort((a, b) => b.start - a.start)) {
            assert(end <= limit, name);
            text = text.slice(0, start) + value + text.slice(end);
            limit = start;
        }
        output[name] = text;
    }
    const planned = parseFiles(output);
    oldOwners(planned, true);
    watchDeclarations(planned["watch.ts"], true);
    for (const [file, owners] of Object.entries(pins.owners)) {
        for (const name of Object.keys(owners)) {
            const watch = name === "createWatchHost";
            assert.equal(ownerText(declaration(planned[file], name), watch),
                ownerText(declaration(parsed[file], name), watch), name);
        }
    }
    const signatures = assertTwinReferences(tree, output, Object.keys(pins.calls).map(twinName));
    for (const name of names) {
        if (files[name] !== output[name]) fs.writeFileSync(path.join(directory, name), output[name]);
    }
    const readback = Object.fromEntries(names.map(name => [name, fs.readFileSync(path.join(directory, name), "utf8")]));
    assert.deepEqual(readback, output);
    const checked = parseFiles(readback);
    oldOwners(checked, true);
    watchDeclarations(checked["watch.ts"], true);
    return signatures;
}

// Pin tsc's own namespace reads by file, enclosing function, and exact text.
// The internal overload is a type contract. createWatchHost sets all six members
// on every call, whatever the reporter. These sites cannot change its behavior
// and do not hand the narrower type to a caller the walk did not count.
const namespaceValuePins = {
    "compiler/debug.ts": [
        { function: "formatSyntaxKind", statement: "return formatEnum(kind, (ts as Readonly<Record<\"SyntaxKind\", Readonly<Record<string, string | number>>>>).SyntaxKind, /*isFlags*/ false);", expression: "ts" },
        { function: "formatSnippetKind", statement: "return formatEnum(kind, (ts as Readonly<Record<\"SnippetKind\", Readonly<Record<string, string | number>>>>).SnippetKind, /*isFlags*/ false);", expression: "ts" },
        { function: "formatScriptKind", statement: "return formatEnum(kind, (ts as Readonly<Record<\"ScriptKind\", Readonly<Record<string, string | number>>>>).ScriptKind, /*isFlags*/ false);", expression: "ts" },
        { function: "formatNodeFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"NodeFlags\", Readonly<Record<string, string | number>>>>).NodeFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatNodeCheckFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"NodeCheckFlags\", Readonly<Record<string, string | number>>>>).NodeCheckFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatModifierFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"ModifierFlags\", Readonly<Record<string, string | number>>>>).ModifierFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatTransformFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"TransformFlags\", Readonly<Record<string, string | number>>>>).TransformFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatEmitFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"EmitFlags\", Readonly<Record<string, string | number>>>>).EmitFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatSymbolFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"SymbolFlags\", Readonly<Record<string, string | number>>>>).SymbolFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatTypeFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"TypeFlags\", Readonly<Record<string, string | number>>>>).TypeFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatSignatureFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"SignatureFlags\", Readonly<Record<string, string | number>>>>).SignatureFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatObjectFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"ObjectFlags\", Readonly<Record<string, string | number>>>>).ObjectFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatFlowFlags", statement: "return formatEnum(flags, (ts as Readonly<Record<\"FlowFlags\", Readonly<Record<string, string | number>>>>).FlowFlags, /*isFlags*/ true);", expression: "ts" },
        { function: "formatRelationComparisonResult", statement: "return formatEnum(result, (ts as Readonly<Record<\"RelationComparisonResult\", Readonly<Record<string, string | number>>>>).RelationComparisonResult, /*isFlags*/ true);", expression: "ts" },
        { function: "formatCheckMode", statement: "return formatEnum(mode, (ts as Readonly<Record<\"CheckMode\", Readonly<Record<string, string | number>>>>).CheckMode, /*isFlags*/ true);", expression: "ts" },
        { function: "formatSignatureCheckMode", statement: "return formatEnum(mode, (ts as Readonly<Record<\"SignatureCheckMode\", Readonly<Record<string, string | number>>>>).SignatureCheckMode, /*isFlags*/ true);", expression: "ts" },
        { function: "formatTypeFacts", statement: "return formatEnum(facts, (ts as Readonly<Record<\"TypeFacts\", Readonly<Record<string, string | number>>>>).TypeFacts, /*isFlags*/ true);", expression: "ts" },
    ],
    "compiler/emitter.ts": [
        { function: "makeName", statement: "return Debug.fail(`Unsupported GeneratedIdentifierKind: ${Debug.formatEnum(autoGenerate.flags & GeneratedIdentifierFlags.KindMask, (ts as Readonly<Record<\"GeneratedIdentifierFlags\", Readonly<Record<string, string | number>>>>).GeneratedIdentifierFlags, /*isFlags*/ true)}.`);", expression: "ts" },
    ],
    "server/project.ts": [
        { function: "Project.updateGraphWorker", statement: "this.writeLog(`Finishing updateGraphWorker: Project: ${this.getProjectName()} projectStateVersion: ${this.projectStateVersion} projectProgramVersion: ${this.projectProgramVersion} structureChanged: ${hasNewProgram}${this.program ? ` structureIsReused:: ${(ts as any).StructureIsReused[this.program.structureIsReused]}` : \"\"} Elapsed: ${elapsed}ms`);", expression: "ts" },
        { function: "Project.enableProxy", statement: "const pluginModule = pluginModuleFactory({ typescript: ts });", expression: "ts" },
    ],
    "harness/fourslashImpl.ts": [
        { function: "runCode", statement: "f(ts, test, goTo, config, verify, edit, debug, format, cancellation, FourSlashInterface.classification, FourSlashInterface.Completion, verifyOperationIsCancelled, ignoreInterpolations);", expression: "ts" },
    ],
    "testRunner/unittests/config/matchFiles.ts": [
        { function: "baselineMatches", statement: "if (parsed.wildcardDirectories) ts.getOwnKeys(parsed.wildcardDirectories).forEach(dir => wildcardDirectories![dir] = `WatchDirectoryFlags.${(ts as any).WatchDirectoryFlags[parsed.wildcardDirectories![dir]]}`);", expression: "ts" },
    ],
    "testRunner/unittests/config/tsconfigParsing.ts": [
        { function: "baselineWildcards", statement: "ts.getOwnKeys(parsed.wildcardDirectories!).forEach(dir => baseline.push(`${dir}: WatchDirectoryFlags.${(ts as any).WatchDirectoryFlags[parsed.wildcardDirectories![dir]]}`));", expression: "ts" },
    ],
    "testRunner/unittests/evaluation/esDecorators.ts": [
        { function: null, statement: "const targetName = ts.Debug.formatEnum(target, (ts as any).ScriptTarget);", expression: "ts" },
    ],
    "testRunner/unittests/evaluation/esDecoratorsMetadata.ts": [
        { function: null, statement: "const targetName = ts.Debug.formatEnum(target, (ts as any).ScriptTarget);", expression: "ts" },
    ],
    "testRunner/unittests/moduleResolution.ts": [
        { function: null, statement: "baselines.push(`Program Reused:: ${(ts as any).StructureIsReused[program2.structureIsReused]}`);", expression: "ts" },
    ],
    "testRunner/unittests/helpers/baseline.ts": [
        { function: "baselineProgram", statement: "baseline.push(`Program structureReused: ${(ts as any).StructureIsReused[program.structureIsReused]}`);", expression: "ts" },
    ],
    "testRunner/unittests/reuseProgramStructure.ts": [
        { function: "baselineDiagnostics", statement: "baselines.push(`Program ${program.version} Reused:: ${(ts as any).StructureIsReused[program.structureIsReused]}`);", expression: "ts" },
        { function: "baselineProgram", statement: "baselines.push(`Program ${program.version} Reused:: ${(ts as any).StructureIsReused[program.structureIsReused]}`);", expression: "ts" },
    ],
    "testRunner/unittests/skipJSDocParsing.ts": [
        { function: "diffSourceFiles", statement: "const kindName = ts.Debug.formatEnum(jsDocParsingMode, (ts as any).JSDocParsingMode);", expression: "ts" },
    ],
    "testRunner/runner.ts": [
        { function: "startTestEnvironment", statement: "(globalThis as any).ts = ts;", expression: "(globalThis as any).ts = ts" },
        { function: "startTestEnvironment", statement: "(globalThis as any).ts = ts;", expression: "ts" },
    ],
};
assert.equal(Object.keys(namespaceValuePins).length, 13);
assert.equal(Object.values(namespaceValuePins).flat().length, 32);

// The four additional TypeScript sites admit only the named refusal and export surface.
const moduleSurfacePins = {
    "server/session.ts": [{
        pin: "session-command-names",
        admits: "module",
        module: "server/protocol.ts",
        function: null,
        statement: "export const CommandNames: any = (protocol as any).CommandTypes;",
        expression: "protocol",
        surface: {
            "server/protocol.ts": "21e97a37391921ae2e262c35fd2b3e83c0ea19ca4c8dbcaf10d9e9e05df6f319",
        },
    }],
    "testRunner/unittests/services/textChanges.ts": [{
        pin: "text-changes-delete-node",
        admits: "module",
        module: "services/_namespaces/ts.textChanges.ts",
        function: null,
        statement: "const { deleteNode } = ts.textChanges;",
        expression: "ts.textChanges",
        surface: {
            "services/_namespaces/ts.textChanges.ts": "b41b21d00d07944b5614d2de28de4b1b040031f91fa1bac4dc93effbc37beed6",
            "services/textChanges.ts": "30e7fb5ccf052df3d7346df65e5d06f2dd4c4e8418e26b826572a0082139172a",
        },
    }],
    "testRunner/unittests/tsserver/session.ts": [{
        pin: "tsserver-command-names",
        admits: "module",
        module: "server/_namespaces/ts.server.protocol.ts",
        function: null,
        statement: "const allCommandNames: ts.server.protocol.CommandTypes[] = Object.values((ts.server.protocol as any).CommandTypes);",
        expression: "ts.server.protocol",
        surface: {
            "server/_namespaces/ts.server.protocol.ts": "62408d55cf60c2b7913ddbdc6c41c68fac5b3193ebe22b70f35460ddd82b73dc",
            "server/protocol.ts": "21e97a37391921ae2e262c35fd2b3e83c0ea19ca4c8dbcaf10d9e9e05df6f319",
        },
    }],
    "testRunner/runner.ts": [{
        pin: "runner-import-tests",
        admits: "import",
        module: "testRunner/tests.ts",
        function: "importTests",
        statement: "return import(\"./tests.js\");",
        expression: "import(\"./tests.js\")",
        surface: {
            "testRunner/tests.ts": "072335697fb9d3538f10aa8ce927eac23d7247febb28c3dc71ddd6974c331ae2",
            "testRunner/unittests/asserts.ts": "6257260ed1db919e892f24caae17b4e36d29a0b9bf3ff16acaceba3c3fb52660",
            "testRunner/unittests/base64.ts": "9f0562b9765dfeacdba1ce4ada165fd2f69bad86a674805c6ccd45b43fe03fbd",
            "testRunner/unittests/builder.ts": "8a50d73d5f04a2f52675277cd8aa910568c3c7c330a5cb2b7e270398fad24633",
            "testRunner/unittests/canWatch.ts": "e82b781976a8248d84f83805aec5ef7d283d29080158e787996e00f9beaa5baa",
            "testRunner/unittests/comments.ts": "1b966a4d4e40e5d6d0a3eb94fc076cbf85f1d209f61f5f079af6456b7b9855e4",
            "testRunner/unittests/compilerCore.ts": "60c222e81525b9ca6e87d1653688383be1799216d602e05b4ed579aa5c87d23a",
            "testRunner/unittests/config/commandLineParsing.ts": "ec153cca62f6783666fa7dddf0d0440f9e6c41aef882e9812c5a644c19d5df94",
            "testRunner/unittests/config/configurationExtension.ts": "6c0c211b14802cf6a73c3487fa5f9351ae71e43e10fbc0e719880e1f9266e8b4",
            "testRunner/unittests/config/convertCompilerOptionsFromJson.ts": "c0d22d56025eae9caad9c5547e011367d8f3f26dcb3de2045bd1ce59185ab6c2",
            "testRunner/unittests/config/convertTypeAcquisitionFromJson.ts": "be682546c4788b0f3c9d9597957c769c081f51db94ace977f011ac21ba5a325e",
            "testRunner/unittests/config/initializeTSConfig.ts": "b2f8c54b87d6c8c7145cb80bb8957b7830f3a2662f9c014bed07a96a54db1765",
            "testRunner/unittests/config/matchFiles.ts": "9a60afd00d79310e50401945db31662500e00c0bf12b21b362030930f33b5dc5",
            "testRunner/unittests/config/showConfig.ts": "4fae85c438aa2c22a6bcd7f8cc9103f4f7516ce7944221851a0979f7147658f5",
            "testRunner/unittests/config/tsconfigParsing.ts": "bddfa6417d8a0cc547e19ba234635bb76b8de4dd5389553dfb441bcadc3807f8",
            "testRunner/unittests/config/tsconfigParsingWatchOptions.ts": "eadef6944e857a420ebcb30acd1884d6c6787e554d6255f7cab99d7acb2f3e51",
            "testRunner/unittests/convertToBase64.ts": "f666a75a2bc148eb7635925176e7338116c604956ed025954b65e50c28cb9c69",
            "testRunner/unittests/customTransforms.ts": "e39289427cb0cca5211dae32049d19dd4cc2685031bd64bda51902e9eee9c684",
            "testRunner/unittests/debugDeprecation.ts": "22add32f0a9f5249aaf96212bb53e5f5f166a438f4a7c204b4247531444ecb31",
            "testRunner/unittests/diagnosticCollection.ts": "8943cd5ce92034d2d683ff6dd0b9bbe0f09980caf7cdd48876034919ddb81160",
            "testRunner/unittests/evaluation/arraySpread.ts": "8a5d40c9388f5a9f8ded3270c9d608247739ae1f96f3b16575bb98f955217a22",
            "testRunner/unittests/evaluation/asyncArrow.ts": "e69029604b8939b9ff71ca9371d5f9c69160439ebe2d5f163e711b8888a24cc5",
            "testRunner/unittests/evaluation/asyncGenerator.ts": "154394ce562cffcf1da313f7a90fa4aba1bc1f28c9e256fe074b1eaa5f4304da",
            "testRunner/unittests/evaluation/autoAccessors.ts": "aeb35602ce313addd2e4d596fc399ae0a54112a7178ef5487171c64080fe4846",
            "testRunner/unittests/evaluation/awaiter.ts": "298318f0146495a1a973043a2605911e4a83a5fef69006476edb56ead4090498",
            "testRunner/unittests/evaluation/awaitUsingDeclarations.ts": "5cf55787890a6ebf702a09f8439a2bb4f11e1f3b57cf23ea6a268405339a71cd",
            "testRunner/unittests/evaluation/constEnum.ts": "7191650a4341418f9bd4f009e4d7d76ecd582e0d3613181467edc946d5496131",
            "testRunner/unittests/evaluation/destructuring.ts": "46836eec070082fb890a3f473c5c662a52312f208d341a933bd840922cf17b48",
            "testRunner/unittests/evaluation/esDecorators.ts": "8149ab4ab89fadcce74aa8d5f6e52d8ab788acdbde5129e7cba33d6d59b30fb3",
            "testRunner/unittests/evaluation/esDecoratorsMetadata.ts": "e890f90e5592988cce31479bdbf762b11db07dba36280d0b069e85027d7ac176",
            "testRunner/unittests/evaluation/externalModules.ts": "efd4ea3b68507153080b9fb694f7b913d062c0848cf52cc7576d421d80b0683a",
            "testRunner/unittests/evaluation/forAwaitOf.ts": "6a4fc344e7a84cb615ee04ed3051c92e736fab742adbed8f4721d2e8bd44f9e9",
            "testRunner/unittests/evaluation/forOf.ts": "100259914dd54a6697184b6455b26ce974a9594944366d8a548e6548ec58038f",
            "testRunner/unittests/evaluation/generator.ts": "ad2a3dddc394b0239d62930cf9d0451dcc9f6a90fa437b4d9c58031ca4f43f65",
            "testRunner/unittests/evaluation/objectRest.ts": "dbab262d609ff01ec352e624c59e8d0c77fb4ccc307f05a80ab6d3a898468c4d",
            "testRunner/unittests/evaluation/optionalCall.ts": "05bc4dfb2582549972013cea904f684fcf5d0330c8221f5f0a7422b5c189c2bb",
            "testRunner/unittests/evaluation/superInStaticInitializer.ts": "f93533a08e6296a063e9e532358adc93fb538151aae3e5857d70191c23afd65d",
            "testRunner/unittests/evaluation/templateLiteral.ts": "84ebe8e9ea0cf0e73dc426cad280db0e26d2a735263df1836f40551a35633ba4",
            "testRunner/unittests/evaluation/updateExpressionInModule.ts": "ded4acbc9f488ff71d3c18a7139054e09858cb01e07966358d6e191995da6c5b",
            "testRunner/unittests/evaluation/usingDeclarations.ts": "d5b767d54691d7854397b2f48c61606bb5de074f846638538f514b13519d3f86",
            "testRunner/unittests/factory.ts": "724c71acfc1536a8d80e6d3d9b8044722ed62f0d08105fdd58fc61fdaca7f500",
            "testRunner/unittests/incrementalParser.ts": "bdc31823e8bab7afed98b5287c4b581a27157b879a36fdea94794ada4b5577ac",
            "testRunner/unittests/jsDocParsing.ts": "f281cb158bc876080c998b3619e27edd412bceb0f291806f351f10c4c9ed3e48",
            "testRunner/unittests/jsonParserRecovery.ts": "e34713111cdefc939337b5df3e6bab6d4f296c3607cf39c26fa29f692bf1a45f",
            "testRunner/unittests/moduleResolution.ts": "50efbed6992432c321818e90acc2b1f451a16bb8436722b964cdaa6203e86cd5",
            "testRunner/unittests/parsePseudoBigInt.ts": "00a7dd64f1649abc470ae78b7360590ea983b91416b8c6de32cf5ee3f48ef309",
            "testRunner/unittests/paths.ts": "fe3ee69f8641346f590e2abc83587bfccdcfaf0199e75ba7da9fdbd7830a6978",
            "testRunner/unittests/printer.ts": "c8ef237db7e8f2908c862010a442bf51e875a2261e81ec41ef25f17913b0d0f0",
            "testRunner/unittests/programApi.ts": "99dda8f8d04bb8aec1cfdf675b669cdade00766056f853808cc5d41c6e1c5617",
            "testRunner/unittests/publicApi.ts": "27623d0bc3376ebdcd7b09cc1a9a020eece608287d7138eccc24339c1fe33166",
            "testRunner/unittests/regExpScannerRecovery.ts": "a527744f3828ecf31f04d5df91fc443025821e940c8d2eff87401f1b5c790141",
            "testRunner/unittests/reuseProgramStructure.ts": "dd3b9c0e2a83e9e320aba1ec7180c8ca3a2bb2716d2c1d12df123ef8a74166ab",
            "testRunner/unittests/semver.ts": "dff279b017ac0116f6fefe3d6bfb347f8648dab65210a00051ba7371bf827a24",
            "testRunner/unittests/services/cancellableLanguageServiceOperations.ts": "99eff01081bba724d8644c711c7aa79f60f3d71235c514539074d3095b8f64e8",
            "testRunner/unittests/services/colorization.ts": "64ef59523213d2619b249dcdbac56b4141c7d4e0752861422b6ce4a6de033d2e",
            "testRunner/unittests/services/convertToAsyncFunction.ts": "689fd977168610e15d38f78072e76bec4f2d5d202c5ecc57682e254163240ef8",
            "testRunner/unittests/services/documentRegistry.ts": "0fcc606d7e6d5f10bb8f11f5c2edffb4d3a5dc595926c79baa8e56f3ebab38de",
            "testRunner/unittests/services/extract/constants.ts": "fca5a795b5dc4ec66fee30a6c24d08524827ae0ab5702392aa3d2139326749e9",
            "testRunner/unittests/services/extract/functions.ts": "adab7479d0154074fde394a2ee62b77f6a1420c148bb7ed450a2212c44dcfde3",
            "testRunner/unittests/services/extract/ranges.ts": "df8fff151b8ce8554de63af6c6f9821482a7ba50250cc50f9825aeecd650dfd5",
            "testRunner/unittests/services/extract/symbolWalker.ts": "150dc39684a1da542be04d878273a9b28457179a28adb5e5dbe561d752786c4c",
            "testRunner/unittests/services/hostNewLineSupport.ts": "a9834f974e741032eaae43f6ab3cf65ed223088998568bdec8bc92d9788b6713",
            "testRunner/unittests/services/languageService.ts": "b0d578da1cb5091d8da80f336ed30e87929941a8118ba2eae564c50946c0e877",
            "testRunner/unittests/services/organizeImports.ts": "f9c1a5e89c13da46bd5bc5909320fcbbf65b6b0788aab6bea9db80dbe8d8428d",
            "testRunner/unittests/services/patternMatcher.ts": "f4c44045f7e7ed0f4efb00d5a902a20878b3f4bcd53653779200f9f41c17d00a",
            "testRunner/unittests/services/preProcessFile.ts": "0da6d4a82faabf104c98cc262f7fbbd8809710c057cc2b9ab77e55a2837844b0",
            "testRunner/unittests/services/textChanges.ts": "56d3ac414360e094ee2c4afc35a7afbdb5ce4870f180bf81581409c03e761ece",
            "testRunner/unittests/services/transpile.ts": "155a2d97aa8d8a25ce490b8cade10e8945366c2e1300f763eaf3beb5d2b8e8a4",
            "testRunner/unittests/services/utilities.ts": "3647c449d13c7c5b16ff29fafcb98bb0b9c49f389c3d4f8885624c6420ddaa0d",
            "testRunner/unittests/skipJSDocParsing.ts": "93e4fc230e4ba7384a43b1865b7eefe0db8b18bc54f741fddc263c36f119803a",
            "testRunner/unittests/sys/symlinkWatching.ts": "d53ff404f27e97747507205c201cc8e2de24fafef902e55bc7e5e98832191cb0",
            "testRunner/unittests/transform.ts": "35a894f2de63e8050e95d7cbf3730b85313eb685867e25f595631fee248aa877",
            "testRunner/unittests/tsbuild/amdModulesWithOut.ts": "2106fd6826bbeb86fe89e8389996688a26f3e7223e05a8e5330f8f16efa9eb6c",
            "testRunner/unittests/tsbuild/clean.ts": "a88593dfdc8b9a809f64e30d66786c26b140bb444318b5de2adc96fba036bc36",
            "testRunner/unittests/tsbuild/commandLine.ts": "d08aa5c1e16f783af42938d678d62e645b5ea0040700f45de8ca7fb5824c3d78",
            "testRunner/unittests/tsbuild/configFileErrors.ts": "e62aea63dad7a47ef0e8efcd6cd979fe393f90b628d99a226d3566392a4c6297",
            "testRunner/unittests/tsbuild/configFileExtends.ts": "b04bd2691fab4a7672a4a534d80ff56da98765dc465cb57d6eb3017f34b9394e",
            "testRunner/unittests/tsbuild/containerOnlyReferenced.ts": "f13f3a6062c92ce06466dbed70b7d52c7658dc0098a8529481f2ec234435a7e7",
            "testRunner/unittests/tsbuild/declarationEmit.ts": "31fc82a0a3fae83e67b0a6b816f57b8ab0a63816568706b30b6a5a1c9ad7fd39",
            "testRunner/unittests/tsbuild/demo.ts": "73c7559a4af88634f70f9064fcf690e9252a7c42674c8d17fdd754407750cec5",
            "testRunner/unittests/tsbuild/emitDeclarationOnly.ts": "b8f8255b6774ac85e4d253fd470a25cf5df1ed21856893be7fb74309928ae752",
            "testRunner/unittests/tsbuild/emptyFiles.ts": "2d74dbd81659e063624e1ce12f4bc8d26e374597f0ff74c6b1a3bf8bd6a52903",
            "testRunner/unittests/tsbuild/exitCodeOnBogusFile.ts": "7c7e233248b07110fedc204874d78b563f272bda4d52385eba0febc0468204ed",
            "testRunner/unittests/tsbuild/extends.ts": "79250f37226a6b4b5e4a5cfcecf58e20b0237133fdc72e963e65f1cd07a42330",
            "testRunner/unittests/tsbuild/fileDelete.ts": "985de042860d9b88a019572b0723ce9d72adb6ba8f2a5b2f8262b271c1ccf5c2",
            "testRunner/unittests/tsbuild/graphOrdering.ts": "341925b89ab4cc7ee4313ac181b1805245e4fdf3838e8017fbd551de5d82f498",
            "testRunner/unittests/tsbuild/inferredTypeFromTransitiveModule.ts": "0a3cc7296ba6f49b4033521c36bee4d40dca4fde9934d435e49592bfbde0d461",
            "testRunner/unittests/tsbuild/javascriptProjectEmit.ts": "cba049e261f977181a770a303a8611dbc4f42106943fc77ba0708346261cafbc",
            "testRunner/unittests/tsbuild/lateBoundSymbol.ts": "3432810b7fb195016ec9ecb757b0837c623c6a7217ac6612afc4029983bca362",
            "testRunner/unittests/tsbuild/libraryResolution.ts": "59f0a88b7b1d62368dfb638c536debac9785a3d6a8625c14a075b1f224a321b3",
            "testRunner/unittests/tsbuild/moduleResolution.ts": "f7ca60ab3c5b1afd1e8cd10b7833d0985acddaaac27af4cd23cc77e1d377aacc",
            "testRunner/unittests/tsbuild/moduleSpecifiers.ts": "c49b69486476168043103f84092fdc3f141e1a93877cca132be5e8bb137a1272",
            "testRunner/unittests/tsbuild/noCheck.ts": "88b10fc50d90d78cef8bc7f03f96349a54c8e6189f52dbaeb53c73d9b02e4290",
            "testRunner/unittests/tsbuild/noEmit.ts": "45489d43fc9be628a37855fc3edcb574669f7a5186487a90a65a4d43a6a24919",
            "testRunner/unittests/tsbuild/noEmitOnError.ts": "100df17c8ab3d5dbe9b3b47f45f93f63b5ea96ea6ba7baba8f623a6063ec0809",
            "testRunner/unittests/tsbuild/outFile.ts": "d655301111d20372ec79a98584d7ce2b21151e98bec16c0feb783fc0aa1ab298",
            "testRunner/unittests/tsbuild/outputPaths.ts": "dcd38c8e6558a4f2204587bb0db726bfda891854ea09cbcc6fceda8ee63463d4",
            "testRunner/unittests/tsbuild/publicApi.ts": "7c80e26d4c9bde89810c545aee22209ce3fb67f4e596012c82e7ac9353f91334",
            "testRunner/unittests/tsbuild/referencesWithRootDirInParent.ts": "6525a79cac1d1b351be62ab47f9b103d474ebe920b69dffc3dfb6b00b047d741",
            "testRunner/unittests/tsbuild/resolveJsonModule.ts": "bba85f290991b8900b0e46a669d59d0582a162618cb19c2fe7dff8551cc91bee",
            "testRunner/unittests/tsbuild/roots.ts": "e9d5650fb5d58d00cd1334dfc497c49c2c294248d796bed2c435dec00e323c3d",
            "testRunner/unittests/tsbuild/sample.ts": "d0b0f5fdde456e217ecdc342a0e04e4d77a459e47c64f3a7f0ecb12aac77232b",
            "testRunner/unittests/tsbuild/transitiveReferences.ts": "caacb49af0be495674cc084be3f236454c569ad6aea3e5ac59ef964b02cb8913",
            "testRunner/unittests/tsbuildWatch/configFileErrors.ts": "63142d5532465cae717ea8b038f1a00899c56fe413048438e518620b8c276e1f",
            "testRunner/unittests/tsbuildWatch/demo.ts": "fda7f9701feea6ac2395296b5c740341d87f428db9d153d258b38ef66f6beb6b",
            "testRunner/unittests/tsbuildWatch/extends.ts": "83acf6085faee249d07479f6d0f2a39b62440e4a4ba4a7ba860afce707cc0bcd",
            "testRunner/unittests/tsbuildWatch/libraryResolution.ts": "42ac24784552148329f613216ed2c8b4304cd64b35f80ce22b089f5f2a161dc5",
            "testRunner/unittests/tsbuildWatch/moduleResolution.ts": "e997e89c6aaaae3958f8831fad2df318f688ade89579f30ca48aa27478949a12",
            "testRunner/unittests/tsbuildWatch/noEmit.ts": "7642cf63d5cfc96390c1ca7ac80f595363ab67eb0f4b11cc23c5d10483d3c0db",
            "testRunner/unittests/tsbuildWatch/noEmitOnError.ts": "9e71a34adfe2b47934d69b8b48a171166ecd02dedda3052ae41ce15a62edd1ee",
            "testRunner/unittests/tsbuildWatch/programUpdates.ts": "58cbb6097b374998218dbfe178e7f26ebb557dc263099b9440209bbaf774dd71",
            "testRunner/unittests/tsbuildWatch/projectsBuilding.ts": "ce705462ec7c4ffe3fd72fbe3ea6cd77143a1ed4ea9f18142e68d87fcef47130",
            "testRunner/unittests/tsbuildWatch/publicApi.ts": "9e6b39de927bbb092aafc5934d1ec2d863814e7e157e0c8fb21eb126f6128e27",
            "testRunner/unittests/tsbuildWatch/reexport.ts": "c9ac883bf489cd242c6a8ca134f59e6720327585114d3ef5326cac441732db67",
            "testRunner/unittests/tsbuildWatch/roots.ts": "023649e8ca1d9dd1e69d6691ecec1e9c71a5b670e0f05ae1b00d6506717028c5",
            "testRunner/unittests/tsbuildWatch/watchEnvironment.ts": "ac3c6f567ff6981335b0ad7e7f3e85d616a2967cd2b6a223500e588268849f75",
            "testRunner/unittests/tsc/cancellationToken.ts": "bba4eba1b8947148ec78b7fadc5164714ad5521cc7a3ac0f9d2ef753a877af23",
            "testRunner/unittests/tsc/commandLine.ts": "be84144f1090d09a8c77b78d95666092da268884473657c08a1dfccefa14d113",
            "testRunner/unittests/tsc/composite.ts": "ce0908aafb9c532efa79c23ccf2acadaa0e9e8331886d85a5c445eee8218c900",
            "testRunner/unittests/tsc/declarationEmit.ts": "2950af3f5f9c350b8df7253594cc5edb98808b751ee43f12fcd3691dd7947647",
            "testRunner/unittests/tsc/extends.ts": "bb904e32358751c4cde66c010be4d904f5880e88aef574d048984de112a86b3b",
            "testRunner/unittests/tsc/forceConsistentCasingInFileNames.ts": "26376ffa9ee7a71da95fdb219796eb902269dbeacaf8fd5bafc3f7c71b3f64f4",
            "testRunner/unittests/tsc/ignoreConfig.ts": "c120d45202a2de01bce9c433dff703f6f924c293d3d5a84df3ced21aeba4baf7",
            "testRunner/unittests/tsc/incremental.ts": "8e92b6e69ba6d76c1fd86f9ef5a6892d9c22c3c11c51c9b437a32f0ede023f3e",
            "testRunner/unittests/tsc/libraryResolution.ts": "0b4bc2df839597cfb37fc13a3855dd46b5b586f92ae2ec5e52d3ef6f1cd17b0e",
            "testRunner/unittests/tsc/listFilesOnly.ts": "9294409a69cdee7956a5592d6714539ff2ef60aaf3bb48e73221507e9f31116e",
            "testRunner/unittests/tsc/moduleResolution.ts": "643979a61323450b4215f1e670401981fc48667a8a9adcecbee9a0712bdbd2fc",
            "testRunner/unittests/tsc/noCheck.ts": "bf2a8be638df80bb3340250fc91846a97b92cc43822458dc2fbcc8ce2bfa4546",
            "testRunner/unittests/tsc/noEmit.ts": "2f350d5b0cf5a04f59e9d718d9e32eae07ba5e3dde07e7acf9c13b2d28e7b2ea",
            "testRunner/unittests/tsc/noEmitOnError.ts": "d8373199f59a9bacd9d4f850357893d72c41bf26c0866ab4920b0643b8ff8307",
            "testRunner/unittests/tsc/projectReferences.ts": "eb8683a0c2cd8b1eb364dfce528bd31e6fbbd9b695c39da20ec278c2aec992fd",
            "testRunner/unittests/tsc/projectReferencesConfig.ts": "378e4d1ac7293d376ff246153eed16e273e98a885572886c41317cb20c988f36",
            "testRunner/unittests/tsc/redirect.ts": "c96a90c17cdb730ba2e1518117b68ded888e36107135a44665f6f0abec92d979",
            "testRunner/unittests/tscWatch/consoleClearing.ts": "9beb8bd414135515197c920aad77d73a263e2afcb612b49ccade0a4d0a8bdd92",
            "testRunner/unittests/tscWatch/emit.ts": "6e4e7ecfe4576672105bfc2d285e15f0454de03386db2cb09ce12091b63b19a2",
            "testRunner/unittests/tscWatch/emitAndErrorUpdates.ts": "a2f365ff749c2dc7729b73efa19c85ae9470b657f8ccd7ff4c1a17f8f206ee54",
            "testRunner/unittests/tscWatch/extends.ts": "4e5554537b1797dec4a8dacd8995670ca736c95907ae57c9da877849434062a3",
            "testRunner/unittests/tscWatch/forceConsistentCasingInFileNames.ts": "da86c0f7dc53623e9fd61c685cdf1bf632b1522d12d78daa4b76b743f056abf4",
            "testRunner/unittests/tscWatch/incremental.ts": "e2510ee8cd4fcfe442559fcfb9e3e85cd8949ecf4fd6745bc2203a03d49b25a3",
            "testRunner/unittests/tscWatch/libraryResolution.ts": "ef869e87d9fe0510443a6d34578fb481168b2617822188f3c1184b88b5fb52b7",
            "testRunner/unittests/tscWatch/listFilesOnly.ts": "1443b377dae86ff891c005d8836ee911f3fb77dd3c8155f36d09b43b4aff2cb5",
            "testRunner/unittests/tscWatch/moduleResolution.ts": "cba6dcd22d75c0eb224d8204132a1a9a14880ea3fdd5ceddfec9240f2e10e842",
            "testRunner/unittests/tscWatch/nodeNextWatch.ts": "41218a634a6b04d28f87e59794a894a1b78ca07013b88312b8d691c2f73d3b4b",
            "testRunner/unittests/tscWatch/noEmit.ts": "2f3d15afeee45261ed234468863d893380967ee7349031f52006897e26c8c66e",
            "testRunner/unittests/tscWatch/noEmitOnError.ts": "2604780f92bc98afb3ed03db97d91971f03d4a6422fdc057f0663b5555172a7b",
            "testRunner/unittests/tscWatch/programUpdates.ts": "ede36c72e0a7daca7d81e5f1450a5621a5bf294a9b81e565df59aa30c54a8d19",
            "testRunner/unittests/tscWatch/projectsWithReferences.ts": "916ef4e10e653a8c447ea0ebc2b723a2499486c859a001bbee1b8836bf5abd22",
            "testRunner/unittests/tscWatch/resolutionCache.ts": "8fe840938da5ed33a4a4861e6f5907df18dd12d0cc91f286acc3760950efef0c",
            "testRunner/unittests/tscWatch/resolveJsonModuleWithIncremental.ts": "e19ce2655b84b8dc0934e715515822c10a916ac005edd4d252fe60fb6fa705e8",
            "testRunner/unittests/tscWatch/sourceOfProjectReferenceRedirect.ts": "7115e7ee7e48cce5b5221d7936dc54c1696f23e2ee61713310e23db3a1f4b15a",
            "testRunner/unittests/tscWatch/symlinks.ts": "0ada98a6b82db0ef879a8b9942af05d0156bd0f8c096f534cb9f7a59cec3d044",
            "testRunner/unittests/tscWatch/watchApi.ts": "30858c853fed450321aecbb9da449432a093cf2d742c634575ed11913acf345e",
            "testRunner/unittests/tscWatch/watchEnvironment.ts": "68149a9772df646cb31b74c147885ed2b9619a7de56b14e4fbdc9538910a0fd7",
            "testRunner/unittests/tsserver/applyChangesToOpenFiles.ts": "06b82463b709614502e69e90f9bf3043109cc764bbfc1f8f7821ebc5940f499c",
            "testRunner/unittests/tsserver/autoImportProvider.ts": "870956eacfb87979b7414edc7569532b8027bf87cc1eaf7c44f93f46392a6284",
            "testRunner/unittests/tsserver/auxiliaryProject.ts": "bfffe2677ccd668455410a9cab577077b7afebe178b38662cd68f179c770a73e",
            "testRunner/unittests/tsserver/cachingFileSystemInformation.ts": "edf6bd787dda934fb1f3bd6b71b309d724904562b32ffa84242d1d3673f1f190",
            "testRunner/unittests/tsserver/cancellationToken.ts": "2e7b2054065454cc7e3f544737518677b075d4a3d20c8c8d80d7d478208dcd77",
            "testRunner/unittests/tsserver/codeFix.ts": "40d8da54a78f3a5ec84381aa6df66060b4348ed10362d450318f84c64d1fb7b8",
            "testRunner/unittests/tsserver/compileOnSave.ts": "2843221822a1e001934f6ac99b97cf064922eb5773910878d1eda6fe4e523061",
            "testRunner/unittests/tsserver/completions.ts": "f2eb0de3be11809f60d20bcb393e02296d4decf739f2f2ff888be48f671d4219",
            "testRunner/unittests/tsserver/completionsIncomplete.ts": "a804267c290210789ab6dee5d8ed80a6d987246aa72c012c896c171b8fc91aba",
            "testRunner/unittests/tsserver/configFileSearch.ts": "b68bbde57cdaca6b872ce3b0557389f10999f566fa1e5813222fee1f736830b3",
            "testRunner/unittests/tsserver/configuredProjects.ts": "df6d2656a3a65f71e045819607be00f78bdcda898214eed65220b75ccba4690f",
            "testRunner/unittests/tsserver/declarationFileMaps.ts": "33f99a5b188e3b741cdc3fb158be91d9e4793b59eae4e048e214453f4b0204be",
            "testRunner/unittests/tsserver/documentRegistry.ts": "367d2c796ca6481685be266c93349ba69b3212dd10da58fd83adeb997c094c9f",
            "testRunner/unittests/tsserver/duplicatePackages.ts": "cf3e41214f0e29174e20c5b437cb2fc17c0a314871383e89032c90b5693ff327",
            "testRunner/unittests/tsserver/dynamicFiles.ts": "cc5835194f3641baeaa79610d4a750450ca4b8351bd3a7185c3f0a2dc114b27b",
            "testRunner/unittests/tsserver/events/largeFileReferenced.ts": "237447d1ca094a88f0bf812fab0785baa235599924bc714a6de4ac991eab77bd",
            "testRunner/unittests/tsserver/events/projectLanguageServiceState.ts": "d3403fc6d267f80be4292bf6bd98b564385e5f806886c96a099c22293505ba0c",
            "testRunner/unittests/tsserver/events/projectLoading.ts": "967bedfd095293d178abdd8c959d02ffe071fd12276696536f9bb847f2aa121d",
            "testRunner/unittests/tsserver/events/projectUpdatedInBackground.ts": "afe006db9749190e739163225e73cd66fdace8db3f285dfb2803a5c5fa1235d6",
            "testRunner/unittests/tsserver/events/watchEvents.ts": "d96874b4276f4482208413350de642649b56e70627256ce5de72eff955f385af",
            "testRunner/unittests/tsserver/exportMapCache.ts": "4ee40649cd854277b911bcec3145963827860aef175087b5503a740d945cf13f",
            "testRunner/unittests/tsserver/extends.ts": "dffdadeec18eac2feff05f3ff66e273da5ab5d1c7944abba6c7956e7ba9a3832",
            "testRunner/unittests/tsserver/externalProjects.ts": "3cd2e2fe648f0e93e398f51c1097798fbca2f1beb17c1bb662a8e70fc39598e0",
            "testRunner/unittests/tsserver/findAllReferences.ts": "ce0171a67501ea51f6b5d96a54ad0ab87dea84602c471da10336ae0b2c43674a",
            "testRunner/unittests/tsserver/forceConsistentCasingInFileNames.ts": "b194b571b53209659aaa5391632e24291eb8c236131743aeb662d2b3f309e3f5",
            "testRunner/unittests/tsserver/formatSettings.ts": "c1ef84c8f3e914484f5d26ed08b25d2a0bad65e34a8edb395473381059219e9e",
            "testRunner/unittests/tsserver/getApplicableRefactors.ts": "861c87d80c1dbc0e6ae9b7e9574cc84787218b9bf50e8ecd21cda47f0b8aebf2",
            "testRunner/unittests/tsserver/getEditsForFileRename.ts": "00924ef4c97b6a97314187124e3e366c48462ca6a7b5fc40fd84914665071bcf",
            "testRunner/unittests/tsserver/getExportReferences.ts": "1af5a2df9f53b522c36760a31cac01cb4fcea7d5d6b5edf65a934eab04e9acbf",
            "testRunner/unittests/tsserver/getFileReferences.ts": "2fd187da4fc7c230f7a4f762bb4dacccf31d55b243337b581dbc4c612c75576c",
            "testRunner/unittests/tsserver/getMoveToRefactoringFileSuggestions.ts": "11b946013c4500faf7622ab19c7c645899c051f200464da14884370a346a16ef",
            "testRunner/unittests/tsserver/goToDefinition.ts": "d4f378a61ebb62f6666d71bebbaf2facbde4b19ed1b67d456b44fec26e406ecb",
            "testRunner/unittests/tsserver/importHelpers.ts": "bb18611917845461dc94445be4e9d9ab46779c456fbd90015acbf4161ad33f29",
            "testRunner/unittests/tsserver/inconsistentErrorInEditor.ts": "64ac0e01283f0474cb684eee415d481e942f23ed79e3046fb1d4ee173e33bad0",
            "testRunner/unittests/tsserver/inferredProjects.ts": "2343159b93c9b97c515368b2876427c31ca58150f2f50e90dc2c55e5af17a1bd",
            "testRunner/unittests/tsserver/inlayHints.ts": "b501e117c41bd79bf382bea37036104e02109ceb74ea884d15fe5df0f23228da",
            "testRunner/unittests/tsserver/jsdocTag.ts": "27dd36b01e3e7979bc2d5e1a168699a5ed01d36ebeab15e915cf37b814e7c2d6",
            "testRunner/unittests/tsserver/languageService.ts": "afa3edc91849e32094d6ec6bda56127833357c3578bee83ac27942f5ddbea761",
            "testRunner/unittests/tsserver/libraryResolution.ts": "ea55dc8d4b587f0928afe7d55a01df4590fbe2a2836d8c0f7d65fb38aa9c62d6",
            "testRunner/unittests/tsserver/maxNodeModuleJsDepth.ts": "c4940d878b1f9dfe91191597cde70ccd98c7ffeca01adf276c974dd1855fc559",
            "testRunner/unittests/tsserver/metadataInResponse.ts": "e41e1cabad8d904f74560c395e6274a9e6c890514c643f4439e39174485087d5",
            "testRunner/unittests/tsserver/moduleResolution.ts": "b1acdeddb6d514a709e9db7e6d9d0501512165f014c664e3b73c71287b3658ef",
            "testRunner/unittests/tsserver/moduleSpecifierCache.ts": "ef10ba25a961a4a5d2d3b7b7f7c8eddb99d624f527385974d0c68712954b4c2f",
            "testRunner/unittests/tsserver/navTo.ts": "4dda577a7798af88fc7a5935f3421564431015d4ff1a12e26fed7ba4aab489cf",
            "testRunner/unittests/tsserver/occurences.ts": "084796507f01b1264c99214fa701c8f29b5520e79ec8027132c1ec204151d227",
            "testRunner/unittests/tsserver/openFile.ts": "9770f2c83fecb41ac7c8f444f44fc7cb649eee3c0662e4440c716b0410abf6a3",
            "testRunner/unittests/tsserver/packageJsonInfo.ts": "ceaa2c8fdfb8881ce84462cec584a52a095cbfc77fd5e909e1c3b579e3db4d30",
            "testRunner/unittests/tsserver/partialSemanticServer.ts": "06a8080d19fdffff07a9ba76a4eb040ec356316804c31dfaa90c49f171da5799",
            "testRunner/unittests/tsserver/pasteEdits.ts": "85c05623310dc69355e79d9de488ca10df25a9b50768f80035f6f94c7a5c6aa9",
            "testRunner/unittests/tsserver/plugins.ts": "134abece7bbd4b22b1c99d989bde5c0e7f54499a45bbc95b73b48167841f09c8",
            "testRunner/unittests/tsserver/pluginsAsync.ts": "c5914b1ae0e85053f1c9ecb80700a57c46a850bd40754e431d10f5968bfbf12b",
            "testRunner/unittests/tsserver/projectErrors.ts": "7abd961b79b56011e2842213ee723370fec2a214401f4b326843c0d7693ae22e",
            "testRunner/unittests/tsserver/projectReferenceCompileOnSave.ts": "d89f60121feb29bfd3b1ce5affa06c1f7751d3dd14b3504774e29a530b71df93",
            "testRunner/unittests/tsserver/projectReferenceErrors.ts": "7aff9d8cb226cd8ab1281cc490d557b446183adbdce8e8799da7eb5228bb4ba2",
            "testRunner/unittests/tsserver/projectReferences.ts": "16f062db2aa2f231905f580cd425c78b8a07281dcdf19650ef5985af95813017",
            "testRunner/unittests/tsserver/projectReferencesSourcemap.ts": "83e50dd3d1dae9929f4d9109d40335596985f876cfaff2717b1f2e3bab639f86",
            "testRunner/unittests/tsserver/projectRootFiles.ts": "9e711673411a803823839df5edc160196c576521ffd764b12d3ea0467c85986c",
            "testRunner/unittests/tsserver/projects.ts": "b6577f8514ad720cc872593ed47feba0911aa1904e2be2670af6623030dc6ab5",
            "testRunner/unittests/tsserver/projectsWithReferences.ts": "c5c2109705c329d9b2bdac60fcf7bba23d770d82c2e0e9f62be4545085638d19",
            "testRunner/unittests/tsserver/refactors.ts": "e4eba44568f756f9ab80c028c15afcfe52891c2f03204f6312f2a4d338d2d98c",
            "testRunner/unittests/tsserver/regionDiagnostics.ts": "52be7e01e6cc25ebe3c60126d11987b4121c56b372a9222aa86c33f0db962231",
            "testRunner/unittests/tsserver/reload.ts": "842a78be696b92a529a4440721680327adeb317bf337733c812e6a9957231874",
            "testRunner/unittests/tsserver/reloadProjects.ts": "b0eec22fab7094ee5f35f59aa0c5a7e64bea38c05b3f19e303fe1b0bad08b3db",
            "testRunner/unittests/tsserver/rename.ts": "64418045f1d21a76ac8e309dc2a79cc54cd268812db3aa07decbad28c5836f2a",
            "testRunner/unittests/tsserver/resolutionCache.ts": "2baf9544120402e14c7323e77f580d82344c9a548e5a9a6d3497ac00620b6446",
            "testRunner/unittests/tsserver/session.ts": "07e4c722ef9f26c3ec61a0afa3d8d921832e6e0065103de224a5dbbad4cae226",
            "testRunner/unittests/tsserver/skipLibCheck.ts": "fbd2d88631d534f67ee8ce8b4bfdc731d6f3761988621b8d8f3c89c0a09754e5",
            "testRunner/unittests/tsserver/smartSelection.ts": "11ef75277a38de0d26dbc619041a1a2120ef05b2fe98980753d66845034ab744",
            "testRunner/unittests/tsserver/symlinkCache.ts": "9deb619ff8aec8f569ac073470cf8e482812f90c6ef976e8f4346d9bcceda1b1",
            "testRunner/unittests/tsserver/symLinks.ts": "bd66270ddaa3bca52abd39ee2ed941ea4cf7abe8356bd741b79413ef8777c08c",
            "testRunner/unittests/tsserver/syntacticServer.ts": "f2f29a8b91e8a219717ab2b39dea686af7ac3e329aaeed8a1846d6eaba4d628b",
            "testRunner/unittests/tsserver/syntaxOperations.ts": "587ffa5c3832e32f0ced74ff14ac02978981a103a36ba70071b731e685fbf7a3",
            "testRunner/unittests/tsserver/telemetry.ts": "9bc21ba22b909a7da41dbfe74821f3a9c14e4bafddeb44df631033ed6e7fa9df",
            "testRunner/unittests/tsserver/textStorage.ts": "961484f7286f9a6367061ed1fb829da181801fe48b98669617390586bdbb8300",
            "testRunner/unittests/tsserver/typeAquisition.ts": "2b371e77cc5b6d1ad82480428832b27f95cb3e02086af320892dcbb0f09d7229",
            "testRunner/unittests/tsserver/typeOnlyImportChains.ts": "35158d1d10aaafb50f50c74a61eeee3153046e4a2cb6e28fb3d22338b14f7081",
            "testRunner/unittests/tsserver/typeReferenceDirectives.ts": "c5c8cf0765d6bd3f7004d653e2ce2f071d2a32cdad23a97fab36b462f956d492",
            "testRunner/unittests/tsserver/typingsInstaller.ts": "0d04e875d2ec26a4a9f9fea551b68517f9fae0337eda34083d9c8b60b986304e",
            "testRunner/unittests/tsserver/versionCache.ts": "05ff45240bb44133774bce53e80a9498d92d19259fe09951a951a1ac06633bd6",
            "testRunner/unittests/tsserver/watchEnvironment.ts": "a6320f2db56e2fe7788570809d35ce5abc3f8dcf70306404419033970f27e2fc",
            "testRunner/unittests/typeParameterIsPossiblyReferenced.ts": "16ebd4bfc49f2bf6b6ca966295917e13fb5aa511373f7540add90462fa0d50f9",
        },
    }],
};
assert.equal(Object.keys(moduleSurfacePins).length, 4);
assert.equal(Object.values(moduleSurfacePins).flat().length, 4);

function assertTwinReferences(tree, output, names) {
    const src = path.resolve(tree, "src");
    const directory = path.join(src, "compiler");
    const roots = ts.sys.readDirectory(src, [".ts", ".tsx", ".mts", ".cts"]);
    // Without strictNullChecks the stock optional reporter selects the required overload.
    const options = { strict: true, exactOptionalPropertyTypes: true, target: ts.ScriptTarget.Latest, module: ts.ModuleKind.NodeNext, types: [] };
    const host = ts.createCompilerHost(options, true);
    const read = host.readFile.bind(host);
    host.readFile = file => path.dirname(file) === directory && Object.hasOwn(output, path.basename(file))
        ? output[path.basename(file)] : read(file);
    const program = ts.createProgram(roots, options, host);
    const checker = program.getTypeChecker();
    const symbols = new Map();
    const permitted = new Set();
    const calls = new Set();
    for (const file of ["watch.ts", "tsbuildPublic.ts"]) {
        const source = program.getSourceFile(path.join(directory, file));
        for (const declaration of source.statements) {
            if (ts.isFunctionDeclaration(declaration) && names.includes(declaration.name?.text)) {
                const symbol = checker.getSymbolAtLocation(declaration.name);
                symbols.set(symbol, declaration.name.text);
                permitted.add(declaration.name);
            }
        }
    }
    assert.equal(symbols.size, names.length);
    const cli = program.getSourceFile(path.join(directory, "executeCommandLine.ts"));
    function unalias(symbol) {
        return symbol?.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(symbol) : symbol;
    }
    const watch = program.getSourceFile(path.join(directory, "watch.ts"));
    const watchFunctions = watch.statements.filter(n => n.name?.text === "createWatchHost");
    const [required, general, implementation] = watchFunctions;
    const watchSymbol = unalias(checker.getSymbolAtLocation(implementation.name));
    const watchNames = new Set(watchFunctions.map(declaration => declaration.name));
    const overloads = new Map([[required, "required"], [general, "general"]]);
    const countedCalls = new Set();
    const watchCalls = [];
    let escapedWatch;
    function location(node) {
        const source = node.getSourceFile();
        const { line, character } = source.getLineAndCharacterOfPosition(node.getStart());
        return { file: path.relative(src, source.fileName), line: line + 1, column: character + 1 };
    }
    function site(node) {
        const { file, line, column } = location(node);
        return `src/${file}:${line}:${column}`;
    }
    function recordCall(node) {
        const signature = checker.getResolvedSignature(node);
        const overload = overloads.get(signature?.declaration);
        if (!overload) return;
        let owner = node.parent;
        while (owner && !ts.isFunctionDeclaration(owner)) owner = owner.parent;
        countedCalls.add(node);
        watchCalls.push({ owner: owner?.name?.text ?? "<unknown>", overload, ...location(node),
            signature: checker.signatureToString(signature) });
    }
    function inspectReference(node, target) {
        const parent = node.parent;
        const specifier = (ts.isImportSpecifier(parent) || ts.isExportSpecifier(parent))
            && (parent.name === node || parent.propertyName === node);
        if (target === watchSymbol) {
            if (watchNames.has(node) || specifier) return;
            let callee = ts.isPropertyAccessExpression(parent) && parent.name === node ? parent : node;
            while (ts.isParenthesizedExpression(callee.parent)) callee = callee.parent;
            if (countedCalls.has(callee.parent) && callee.parent.expression === callee) return;
            escapedWatch ??= node;
        }
        const name = symbols.get(target);
        if (name && !permitted.has(node)) {
            const direct = node.getSourceFile() === cli && ts.isIdentifier(node);
            const imported = direct && ts.isImportSpecifier(parent) && parent.name === node
                && node.text === name;
            const call = direct && ts.isCallExpression(parent) && parent.expression === node
                && node.text === name && !calls.has(name);
            assert(imported || call, `unreviewed twin reference: ${site(node)}: ${name}`);
            if (call) calls.add(name);
        }
    }
    function inspectElement(node) {
        const receiver = checker.getTypeAtLocation(node.expression);
        const keyType = checker.getTypeAtLocation(node.argumentExpression);
        const keys = keyType.isUnion() ? keyType.types : [keyType];
        if (keys.every(key => key.isStringLiteral())) {
            for (const key of keys) {
                const target = unalias(checker.getPropertyOfType(receiver, key.value));
                inspectReference(node, target);
            }
            return;
        }
        assert(!receiver.getProperties().some(property => {
            const target = unalias(property);
            return target === watchSymbol || symbols.has(target);
        }), `unresolved bracket key on guarded receiver: ${site(node)}`);
    }
    const usedNamespacePins = new Set();
    const namespaceNamesByType = new Map();
    function heldNames(type) {
        if (!namespaceNamesByType.has(type)) {
            const names = new Set();
            for (const member of type.isUnion() ? type.types : [type]) {
                for (const property of member.getProperties()) {
                    const target = unalias(property);
                    const name = target === watchSymbol ? "createWatchHost" : symbols.get(target);
                    if (name) names.add(name);
                }
            }
            namespaceNamesByType.set(type, [...names].sort());
        }
        return namespaceNamesByType.get(type);
    }
    function isNonValueName(node) {
        const parent = node.parent;
        return ts.isDeclarationName(node)
            || ts.isPartOfTypeQuery(node)
            || ts.isPartOfTypeNode(node)
            || ts.isQualifiedName(parent)
            || ts.isImportSpecifier(parent)
            || ts.isExportSpecifier(parent);
    }
    function pinFor(node) {
        let owner = null;
        for (let ancestor = node.parent; ancestor; ancestor = ancestor.parent) {
            if (ts.isFunctionDeclaration(ancestor)) {
                owner = ancestor.name?.getText() ?? null;
                break;
            }
            if (ts.isMethodDeclaration(ancestor) || ts.isGetAccessorDeclaration(ancestor)
                || ts.isSetAccessorDeclaration(ancestor) || ts.isConstructorDeclaration(ancestor)) {
                const member = ts.isConstructorDeclaration(ancestor) ? "constructor" : ancestor.name.getText();
                const container = ancestor.parent;
                owner = ts.isClassDeclaration(container) && container.name
                    ? container.name.getText() + "." + member : member;
                break;
            }
        }
        for (let statement = node; statement.parent; statement = statement.parent) {
            const parent = statement.parent;
            if (ts.isBlock(parent) || ts.isModuleBlock(parent) || ts.isSourceFile(parent)
                || ts.isCaseClause(parent) || ts.isDefaultClause(parent)) {
                return { function: owner, statement: statement.getText(), expression: node.getText() };
            }
        }
        return undefined;
    }
    function sourceModule(symbol) {
        return symbol?.declarations?.find(ts.isSourceFile);
    }
    function surfaceModules(file) {
        const source = program.getSourceFile(path.join(src, file));
        if (!source) return undefined;
        const modules = new Set([file]);
        for (const statement of source.statements) {
            if (!ts.isExportDeclaration(statement) || !statement.moduleSpecifier
                || (statement.exportClause && !ts.isNamespaceExport(statement.exportClause))) continue;
            const reached = sourceModule(unalias(checker.getSymbolAtLocation(statement.moduleSpecifier)));
            if (!reached?.fileName.startsWith(src + path.sep)) return undefined;
            modules.add(path.relative(src, reached.fileName).split(path.sep).join("/"));
        }
        return [...modules];
    }
    function countedStatements(source) {
        const extra = new Set();
        for (const statement of source.statements) {
            if (!ts.isExportDeclaration(statement) || statement.moduleSpecifier
                || !statement.exportClause || !ts.isNamedExports(statement.exportClause)) continue;
            for (const item of statement.exportClause.elements) {
                const local = checker.getExportSpecifierLocalTargetSymbol(item);
                for (const declaration of local?.declarations ?? []) {
                    let top = declaration;
                    while (top.parent && top.parent !== source && !ts.isSourceFile(top.parent)) top = top.parent;
                    if (top.parent === source) extra.add(top);
                }
            }
        }
        return source.statements.filter(statement => ts.isExportDeclaration(statement)
            || ts.isExportAssignment(statement) || ts.isImportDeclaration(statement)
            || ts.isImportEqualsDeclaration(statement) || extra.has(statement)
            || statement.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword))
            .map(statement => statement.getText(source));
    }
    function surfaceHash(file, statements) {
        return require("node:crypto").createHash("sha256").update([file, ...statements, ""].join("\n")).digest("hex");
    }
    function valueSymbols(node, shorthand) {
        if (shorthand) return [unalias(checker.getShorthandAssignmentValueSymbol(node.parent))];
        if (ts.isElementAccessExpression(node)) {
            const key = checker.getTypeAtLocation(node.argumentExpression);
            const keys = key.isUnion() ? key.types : [key];
            if (!keys.every(part => part.isStringLiteral())) return [];
            const receiver = checker.getTypeAtLocation(node.expression);
            return keys.map(part => unalias(checker.getPropertyOfType(receiver, part.value)));
        }
        if (ts.isPropertyAccessExpression(node)) return [unalias(checker.getSymbolAtLocation(node.name))];
        return ts.isIdentifier(node) ? [unalias(checker.getSymbolAtLocation(node))] : [];
    }
    function checkSurface(node, pin) {
        const recorded = Object.keys(pin.surface);
        const actual = surfaceModules(pin.module);
        const drift = file => `module surface drift: ${site(node)}: ${pin.pin}: ${file}`;
        assert(actual && actual.length === recorded.length && actual.every((file, i) => file === recorded[i]),
            drift(pin.module));
        for (const file of recorded) {
            const source = program.getSourceFile(path.join(src, file));
            assert(source && surfaceHash(file, countedStatements(source)) === pin.surface[file], drift(file));
        }
    }
    function checkNamespaceRefusal(node, held, module, dynamic, importEquals) {
        if (!held.length && !module && !dynamic && !importEquals) return;
        const file = location(node).file;
        const position = pinFor(node);
        const unusedMatch = candidate => position && !usedNamespacePins.has(candidate)
            && candidate.function === position.function
            && candidate.statement === position.statement && candidate.expression === position.expression;
        const original = namespaceValuePins[file]?.find(unusedMatch);
        if (original) {
            usedNamespacePins.add(original);
            return;
        }
        assert(!held.length, `unreviewed namespace value: ${site(node)}: ${held.join(", ")}`);
        const rule = dynamic ? "import" : "module";
        const target = dynamic || module || importEquals;
        const pin = !importEquals && !(dynamic && module) && moduleSurfacePins[file]?.find(candidate =>
            unusedMatch(candidate) && candidate.admits === rule && candidate.module === target);
        assert(pin, dynamic ? `unreviewed dynamic import: ${site(node)}: ${target}`
            : `unreviewed namespace value: ${site(node)}: module ${target}`);
        checkSurface(node, pin);
        usedNamespacePins.add(pin);
    }
    function dynamicImportTarget(node) {
        const specifier = node.arguments[0];
        const target = specifier && ts.isStringLiteralLike(specifier)
            ? unalias(checker.getSymbolAtLocation(specifier)) : undefined;
        const declaration = sourceModule(target) ?? target?.declarations?.find(decl =>
            ts.isModuleDeclaration(decl) && ts.isStringLiteral(decl.name));
        if (!declaration) return "unresolved";
        const file = declaration.getSourceFile().fileName;
        return file.startsWith(src + path.sep) ? path.relative(src, file) : undefined;
    }
    function inspectImportEquals(node) {
        if (!ts.isImportEqualsDeclaration(node) || ts.isExternalModuleReference(node.moduleReference)) return;
        const reference = node.moduleReference;
        const file = sourceModule(unalias(checker.getSymbolAtLocation(reference)));
        if (file) checkNamespaceRefusal(reference, [], undefined, undefined, path.relative(src, file.fileName));
    }
    function inspectNamespaceValue(node) {
        const dynamic = ts.isImportCall(node) ? dynamicImportTarget(node) : undefined;
        const parent = node.parent;
        if (!parent || ts.isParenthesizedExpression(node)) return;
        const shorthand = ts.isShorthandPropertyAssignment(parent) && parent.name === node;
        if (ts.isIdentifier(node) && !shorthand && isNonValueName(node)) return;
        if (!(shorthand || ts.isExpressionNode(node) || (ts.isExportAssignment(parent) && parent.expression === node))
            || (ts.isPropertyAccessExpression(parent) && parent.name === node)) return;
        let expression = node;
        while (ts.isParenthesizedExpression(expression.parent)) expression = expression.parent;
        const access = expression.parent;
        const isReceiver = (ts.isPropertyAccessExpression(access) || ts.isElementAccessExpression(access))
            && access.expression === expression;
        const names = isReceiver ? [] : heldNames(checker.getTypeAtLocation(node));
        const moduleFiles = isReceiver ? [] : [...new Set(valueSymbols(node, shorthand).map(sourceModule)
            .filter(Boolean).map(file => path.relative(src, file.fileName)))].sort();
        checkNamespaceRefusal(node, names, moduleFiles.join(", "), dynamic);
    }

    function inspect(node) {
        if (ts.isCallExpression(node)) recordCall(node);
        if (ts.isIdentifier(node)) {
            const parent = node.parent;
            let symbol;
            if ((ts.isPropertyAssignment(parent) || ts.isShorthandPropertyAssignment(parent)) && parent.name === node
                && ts.isObjectLiteralExpression(parent.parent)
                && ts.isArrayLiteralOrObjectLiteralDestructuringPattern(parent.parent)) {
                symbol = checker.getPropertySymbolOfDestructuringAssignment(node);
            } else if (ts.isShorthandPropertyAssignment(parent) && parent.name === node) {
                symbol = checker.getShorthandAssignmentValueSymbol(parent);
            } else if (ts.isBindingElement(parent) && parent.name === node && !parent.propertyName
                && !parent.dotDotDotToken && ts.isObjectBindingPattern(parent.parent)) {
                symbol = checker.getPropertyOfType(checker.getTypeAtLocation(parent.parent), node.text);
            } else {
                symbol = checker.getSymbolAtLocation(node);
            }
            inspectReference(node, unalias(symbol));
        }
        if (ts.isElementAccessExpression(node)) inspectElement(node);
        inspectImportEquals(node);
        inspectNamespaceValue(node);
        ts.forEachChild(node, inspect);
    }
    for (const source of program.getSourceFiles()) if (source.fileName.startsWith(src + path.sep)) inspect(source);
    assert.equal(calls.size, names.length, "four resolved CLI calls");
    if (escapedWatch) {
        assert.fail(`unreviewed createWatchHost value reference: ${site(escapedWatch)} (${watchCalls.length} resolved overload calls)`);
    }
    assert.deepEqual(watchCalls.map(({ owner, overload }) => `${overload} ${owner}`).sort(), [
        "general createSolutionBuilderWithWatchHost",
        "general createWatchCompilerHost",
        "required createDefaultSolutionBuilderWithWatchHost",
        "required createDefaultWatchCompilerHostOfConfigFile",
        "required createDefaultWatchCompilerHostOfFilesAndCompilerOptions",
    ], "createWatchHost overload call drift");
    const own = checker.getTypeAtLocation(watch.statements.find(n => n.name?.text === "OwnWatchHost"));
    const published = checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(general));
    const literal = implementation.body.statements.find(ts.isReturnStatement).expression;
    const members = checker.getPropertiesOfType(own);
    assert.deepEqual(members.map(member => member.name), watchKeys);
    assert.deepEqual(literal.properties.map(property => property.name.text), watchKeys);
    for (const [i, member] of members.entries()) {
        assert.equal(Boolean(member.flags & ts.SymbolFlags.Optional), false, member.name);
        const property = literal.properties[i];
        const value = ts.isShorthandPropertyAssignment(property)
            ? checker.getShorthandAssignmentValueSymbol(property) : undefined;
        const written = value ? checker.getTypeOfSymbolAtLocation(value, property)
            : checker.getTypeAtLocation(property.initializer);
        const internal = checker.getTypeOfSymbolAtLocation(member, member.valueDeclaration);
        const external = checker.getTypeOfPropertyOfType(published, member.name);
        assert.equal(checker.isTypeAssignableTo(written, internal), true, member.name);
        assert.equal(checker.isTypeAssignableTo(internal, external), true, member.name);
    }
    return watchCalls;
}

function redirected(text, name) {
    const replacement = text.replace(name, twinName(name));
    return name.startsWith("createSolution") ? replacement.replace(/\n +\/\*createProgram\*\/ undefined,/, "") : replacement;
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(ts.version, "6.0.3");
    assert.equal(process.argv.length, 3);
    adapt(process.argv[2]);
}
