"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { adapt } = require("./adapt.cjs");
const { watchKeys } = require("./twins.cjs");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "pass a full stage3/apply.sh tree");
const tree = path.resolve(process.argv[2]);
const snapshot = "refs/stage3/99-seats-scout24-whole-built-hosts";
const git = (...args) => execFileSync("git", ["-C", tree, ...args]);
const files = ["watch.ts", "tsbuildPublic.ts", "executeCommandLine.ts", "core.ts"];
const before = Object.fromEntries(files.map(file => [file, git("show", `${snapshot}-before:src/compiler/${file}`)]));
const after = Object.fromEntries(files.map(file => [file, git("show", `${snapshot}-after:src/compiler/${file}`)]));
assert.deepEqual(git("diff", "--name-only", `${snapshot}-before`, `${snapshot}-after`).toString().trim().split("\n"), [
    "src/compiler/executeCommandLine.ts", "src/compiler/tsbuildPublic.ts", "src/compiler/watch.ts",
]);
function programFor(directory, roots, overrides = {}) {
    const sdk = path.dirname(path.dirname(require.resolve("typescript")));
    const config = ts.getParsedCommandLineOfConfigFile(path.join(directory, "src/tsconfig-base.json"), {
        composite: false, incremental: false, declarationMap: false, sourceMap: false,
        stripInternal: true, types: ["node"], typeRoots: [path.join(sdk, "../@types")],
        ...overrides,
    }, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: diagnostic => { throw Error(format(diagnostic)); } });
    assert.equal(config.errors.length, 0, config.errors.map(format).join("\n"));
    return ts.createProgram(roots, config.options);
}
function format(diagnostic) {
    const file = diagnostic.file;
    const line = file ? file.getLineAndCharacterOfPosition(diagnostic.start).line + 1 : 0;
    return `${file?.fileName ?? ""}:${line} TS${diagnostic.code} ${ts.flattenDiagnosticMessageText(diagnostic.messageText, " ")}`;
}
function declarations(program, directory) {
    return files.slice(0, 3).map(name => {
        const emitted = [];
        const source = program.getSourceFile(path.join(directory, "src/compiler", name));
        const result = program.emit(source, (file, text) => {
            if (file.endsWith(".d.ts")) emitted.push(text);
        }, undefined, true);
        assert.equal(result.emitSkipped, false, name);
        assert.equal(result.diagnostics.length, 0, result.diagnostics.map(format).join("\n"));
        assert.equal(emitted.length, 1, name);
        return emitted[0];
    });
}
function rootsFor(directory) {
    return ts.sys.readDirectory(path.join(directory, "src/compiler"), [".ts"]);
}
function watchImplementation(sources) {
    const file = ts.createSourceFile("watch.ts", sources["watch.ts"].toString("utf8"), ts.ScriptTarget.Latest, true);
    const implementations = file.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "createWatchHost" && node.body);
    assert.equal(implementations.length, 1);
    return implementations[0];
}
const beforeWatchImplementation = watchImplementation(before);
const afterWatchImplementation = watchImplementation(after);
assert.equal(afterWatchImplementation.type.getText(), "OwnWatchHost");
assert.equal(afterWatchImplementation.body.getText(), beforeWatchImplementation.body.getText());
console.log("PASS createWatchHost implementation declares OwnWatchHost with its stock body");
const changedLines = new Map();
let changedFile;
for (const line of git("diff", "--unified=0", `${snapshot}-before`, `${snapshot}-after`, "--", "src").toString().split("\n")) {
    if (line.startsWith("+++ b/")) {
        changedFile = new Set();
        changedLines.set(line.slice(6), changedFile);
    }
    const hunk = /^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@/.exec(line);
    if (!hunk) continue;
    const start = Number(hunk[1]);
    for (let n = 0; n < Number(hunk[2] ?? 1); n++) changedFile.add(start + n);
}

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "whole-hosts-"));
try {
    const beforeArchive = path.join(scratch, "before.tar");
    git("archive", "--format=tar", `--output=${beforeArchive}`, `${snapshot}-before`, "src");
    execFileSync("tar", ["-xf", beforeArchive, "-C", scratch]);
    const originalDeclarations = declarations(programFor(scratch, rootsFor(scratch)), scratch);
    let watchSignatures;
    for (const attempt of [1, 2]) {
        watchSignatures = adapt(scratch);
        for (const file of files) {
            assert(fs.readFileSync(path.join(scratch, "src/compiler", file)).equals(after[file]), file);
        }
        console.log(`PASS application ${attempt} matches the after snapshot`);
    }
    const afterTree = path.join(scratch, "after");
    fs.mkdirSync(afterTree);
    const afterArchive = path.join(scratch, "after.tar");
    git("archive", "--format=tar", `--output=${afterArchive}`, `${snapshot}-after`, "src");
    execFileSync("tar", ["-xf", afterArchive, "-C", afterTree]);
    const entry = [path.join(afterTree, "src/tsc/tsc.ts")];
    const directory = rootsFor(afterTree);
    for (const [name, roots] of [["entry", entry], ["directory", directory]]) {
        const program = programFor(afterTree, roots);
        const diagnostics = ts.getPreEmitDiagnostics(program);
        assert.equal(diagnostics.length, 0, diagnostics.map(format).join("\n"));
        console.log(`PASS stock TypeScript ${name} graph, ${roots.length} roots, 0 diagnostics`);
        if (name === "directory") {
            for (const [index, text] of declarations(program, afterTree).entries()) {
                assert(text === originalDeclarations[index], `stripped declaration changed: ${files[index]}`);
            }
            console.log("PASS three stripped declaration files byte-identical");
        }
        const exact = programFor(afterTree, roots, {
            strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
        });
        const written = ts.getPreEmitDiagnostics(exact).filter(diagnostic => {
            if (!diagnostic.file) return false;
            const line = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start).line + 1;
            return changedLines.get(path.relative(afterTree, diagnostic.file.fileName))?.has(line);
        });
        assert.equal(written.length, 0, written.map(format).join("\n"));
        console.log(`PASS exact-optional ${name} graph, 0 diagnostics on adapter lines`);
    }
    for (const { owner, overload, file, line, column, signature } of watchSignatures) {
        console.log(`src/${file}:${line}:${column} ${owner} ${overload} ${signature}`);
    }
    console.log("PASS three twins select the required overload and two stock copies select the general overload");
    console.log("PASS all six required OwnWatchHost members carry the values written by createWatchHost");
    const expected = [
        ["copyProperties", "src/compiler/watch.ts", "createWatchCompilerHost", "createWatchHost(system, reportWatchStatus)", "WatchHost"],
        ["copyProperties", "src/compiler/tsbuildPublic.ts", "createSolutionBuilderWithWatchHost", "watchHost", "WatchHost"],
    ];
    function scanEnumeration(directory) {
        const command = [path.join(__dirname, "enumeration.cjs"), directory];
        const result = execFileSync(process.execPath, command);
        const { roots, enumeration } = JSON.parse(result);
        return { roots, rows: enumeration.map(([form, site, owner, expression, type]) =>
            [form, site.split(":")[0], owner, expression, type]) };
    }
    function enumeration(directory = scratch) {
        return scanEnumeration(directory).rows;
    }
    function checkEnumeration(actual) {
        assert.deepEqual(actual, expected);
    }
    const cleanScan = scanEnumeration(scratch);
    checkEnumeration(cleanScan.rows);
    console.log(`enumeration.cjs exit=0, ${cleanScan.roots} roots, ${cleanScan.rows.length} construction copies`);
    console.log("PASS only the two preserved construction copies enumerate these host types");

    function withDrift(name, transform, check, directory = scratch) {
        return withSourceDrift(`compiler/${name}`, transform, check, directory);
    }
    function withSourceDrift(name, transform, check, directory) {
        const file = path.join(directory, "src", name);
        const previous = fs.existsSync(file) ? fs.readFileSync(file) : undefined;
        fs.writeFileSync(file, transform(previous?.toString() ?? ""));
        const checked = [...new Set([name, ...files.map(file => `compiler/${file}`)])].map(item => {
            const target = path.join(directory, "src", item);
            return { target, bytes: fs.readFileSync(target), mtime: fs.statSync(target, { bigint: true }).mtimeNs };
        });
        try {
            check();
            for (const { target, bytes, mtime } of checked) {
                assert(fs.readFileSync(target).equals(bytes), target);
                assert.equal(fs.statSync(target, { bigint: true }).mtimeNs, mtime, target);
            }
        } finally {
            if (previous) fs.writeFileSync(file, previous);
            else fs.unlinkSync(file);
        }
    }
    function rejected(reason, directory = scratch) {
        const command = [path.join(__dirname, "adapt.cjs"), directory];
        const run = spawnSync(process.execPath, command, { encoding: "utf8" });
        if (run.error) throw run.error;
        assert.equal(run.status, 1, run.stderr);
        assert.match(run.stderr, reason);
        return `node ${command.join(" ")} exit=${run.status}, ${run.stderr.match(reason)[0]}`;
    }
    withDrift("wholeReviewSecondCaller.ts", () => [
        'import * as watch from "./watch.js";',
        'import { System } from "./_namespaces/ts.js";',
        '',
        'export function secondCaller(system: System) {',
        '    return watch["createDefaultWatchCompilerHostOfConfigFile"]({ configFileName: "tsconfig.json", system });',
        '}',
        '',
    ].join("\r\n"), () => rejected(/unreviewed twin reference/));
    console.log("PASS reviewer second caller rejected without changing bytes or mtime");
    withDrift("tsbuildPublic.ts", text => {
        const changed = text.replace("    memoize,", "    identity as memoize,");
        assert.notEqual(changed, text);
        return changed;
    }, () => rejected(/value import drift: memoize/));
    console.log("PASS reviewer import drift rejected without changing bytes or mtime");
    withDrift("executeCommandLine.ts", text => {
        const site = "    updateWatchCompilationHost(system, cb, watchCompilerHost);\r\n    watchCompilerHost.configFileParsingResult";
        assert(text.includes(site));
        return text.replace(site, "    Reflect.ownKeys(watchCompilerHost);\r\n" + site);
    }, () => {
        const actual = enumeration();
        assert.equal(actual.length, expected.length + 1);
        assert(actual.some(([operation, file]) => operation === "Reflect.ownKeys" && file.endsWith("/executeCommandLine.ts")));
        assert.throws(() => checkEnumeration(actual), { code: "ERR_ASSERTION" });
    });
    console.log("PASS reviewer enumeration drift adds a third reader and fails the oracle without writing");

    function assertBeforeSnapshot(directory) {
        for (const file of files) {
            const contents = fs.readFileSync(path.join(directory, "src/compiler", file));
            assert(contents.equals(before[file]), `${file} must match the before snapshot`);
        }
        assert(!before["watch.ts"].equals(after["watch.ts"]), "watch.ts must differ between snapshots");
    }

    function reviewerDrifts(directory, phase) {
        let receipt;
        if (phase === "first apply") assertBeforeSnapshot(directory);
        withDrift("reviewerWholeCaller.ts", () => [
            'import * as watch from "./watch.js";',
            'import { System } from "./_namespaces/ts.js";',
            '',
            'const twinKey = "createDefaultWatchCompilerHostOfConfigFile";',
            'export function secondCaller(system: System) {',
            '    return watch[twinKey]({ configFileName: "tsconfig.json", system });',
            '}',
            '',
        ].join("\r\n"), () => {
            receipt = rejected(/unreviewed twin reference: src\/compiler\/reviewerWholeCaller\.ts:6:12: createDefaultWatchCompilerHostOfConfigFile/, directory);
        }, directory);
        console.log(`PASS literal-type bracket key on ${phase}: ${receipt}; bytes and mtimes unchanged`);

        if (phase === "first apply") assertBeforeSnapshot(directory);
        withDrift("reviewerWholeOverload.ts", () => [
            'import { createWatchHost } from "./watch.js";',
            'import { System, WatchStatusReporter } from "./_namespaces/ts.js";',
            'const makeWatch = createWatchHost;',
            'export function extraWatch(system: System, reporter: WatchStatusReporter) {',
            '    return makeWatch(system, reporter);',
            '}',
            '',
        ].join("\r\n"), () => {
            receipt = rejected(/unreviewed createWatchHost value reference: src\/compiler\/reviewerWholeOverload\.ts:3:19 \(6 resolved overload calls\)/, directory);
        }, directory);
        console.log(`PASS overload alias on ${phase}: ${receipt}; bytes and mtimes unchanged`);

        withDrift("reviewerWholeReader.ts", () => [
            'import { WatchCompilerHostOfConfigFile, EmitAndSemanticDiagnosticsBuilderProgram } from "./_namespaces/ts.js";',
            '',
            'export function descriptors(host: WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>) {',
            '    return Object.getOwnPropertyDescriptors(host);',
            '}',
            '',
        ].join("\r\n"), () => {
            const { roots, rows } = scanEnumeration(directory);
            assert.equal(rows.length, expected.length + 1);
            const reader = rows.find(([operation, file]) => operation === "Object.getOwnPropertyDescriptors"
                && file === "src/compiler/reviewerWholeReader.ts");
            assert(reader);
            assert.throws(() => checkEnumeration(rows), { code: "ERR_ASSERTION" });
            receipt = `node ${path.join(__dirname, "enumeration.cjs")} ${directory} exit=0, ${roots} roots, ${rows.length} readers, oracle ERR_ASSERTION for ${reader[0]} in ${reader[1]}`;
        }, directory);
        console.log(`PASS descriptor enumeration on ${phase}: ${receipt}; bytes and mtimes unchanged`);
    }

    function namespaceDrifts(directory, phase) {
        const cases = [
            ["reviewer150TwinDestructure.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed twin reference: src\/compiler\/reviewer150TwinDestructure\.ts:4:9: createDefaultWatchCompilerHostOfConfigFile$/m, [
                "import * as watch from \"./watch.js\";",
                "import { System, WatchCompilerHostOfConfigFile, EmitAndSemanticDiagnosticsBuilderProgram } from \"./_namespaces/ts.js\";",
                "",
                "const { createDefaultWatchCompilerHostOfConfigFile } = watch;",
                "export function destructuredTwin(system: System): WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> {",
                "    return createDefaultWatchCompilerHostOfConfigFile({ configFileName: \"tsconfig.json\", system });",
                "}",
                "",
            ]],
            ["reviewer150TwinPick.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer150TwinPick\.ts:4:80: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { System, WatchCompilerHostOfConfigFile, EmitAndSemanticDiagnosticsBuilderProgram } from \"./_namespaces/ts.js\";",
                "",
                "const view: Pick<typeof watch, \"createDefaultWatchCompilerHostOfConfigFile\"> = watch;",
                "export function pickedTwin(system: System): WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> {",
                "    return view.createDefaultWatchCompilerHostOfConfigFile({ configFileName: \"tsconfig.json\", system });",
                "}",
                "",
            ]],
            ["reviewer150Destructure.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer150Destructure\.ts:5:29: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "const { createWatchHost } = watch;",
                "const makeWatch: (system: System, reportWatchStatus: WatchStatusReporter) => OwnWatchHost = createWatchHost;",
                "export function destructuredWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return makeWatch(system, reporter);",
                "}",
                "",
            ]],
            ["reviewer150View.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer150View\.ts:5:106: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "const view: { createWatchHost(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost; } = watch;",
                "export function viewWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return view.createWatchHost(system, reporter);",
                "}",
                "",
            ]],
            ["reviewer150Member.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer150Member\.ts:8:100: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "function member<T, K extends keyof T>(owner: T, key: K): T[K] {",
                "    return owner[key];",
                "}",
                "const makeWatch: (system: System, reportWatchStatus: WatchStatusReporter) => OwnWatchHost = member(watch, \"createWatchHost\");",
                "export function memberWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return makeWatch(system, reporter);",
                "}",
                "",
            ]],
            ["reviewer150DestructureDirect.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer150DestructureDirect\.ts:5:29: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "const { createWatchHost } = watch;",
                "export function destructuredDirect(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return createWatchHost(system, reporter);",
                "}",
                "",
            ]],
            ["reviewer150Loose.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer150Loose\.ts:5:20: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "const loose: any = watch;",
                "const makeWatch: (system: System, reportWatchStatus: WatchStatusReporter) => OwnWatchHost = loose.createWatchHost;",
                "export function looseWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return makeWatch(system, reporter);",
                "}",
                "",
            ]],
            ["reviewer150GenericKey.ts", /^AssertionError \[ERR_ASSERTION\]: unresolved bracket key on guarded receiver: src\/compiler\/reviewer150GenericKey\.ts:4:12$/m, [
                "import * as watch from \"./watch.js\";",
                "",
                "export function exportOf<K extends keyof typeof watch>(key: K): (typeof watch)[K] {",
                "    return watch[key];",
                "}",
                "",
            ]],
        ];
        for (const [file, reason, source] of cases) {
            if (phase === "first apply") assertBeforeSnapshot(directory);
            let receipt;
            withDrift(file, () => source.join("\n"), () => {
                receipt = rejected(reason, directory);
            }, directory);
            console.log(`PASS ${file} on ${phase}: ${receipt}; bytes and mtimes unchanged`);
        }

        if (phase === "first apply") assertBeforeSnapshot(directory);
        let receipt;
        withDrift("debug.ts", text => {
            const newline = text.includes("\r\n") ? "\r\n" : "\n";
            const opening = /function formatSyntaxKind\([^\n]*\{\r?\n([ \t]+)(?=return )/;
            assert.match(text, opening);
            return text.replace(opening, (_, indent) => _.slice(0, -indent.length)
                + indent + "const namespaceAgain = ts;" + newline + indent);
        }, () => {
            receipt = rejected(/^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/debug\.ts:420:32: createDefaultSolutionBuilderHost, createDefaultSolutionBuilderWithWatchHost, createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, directory);
        }, directory);
        console.log(`PASS 33rd namespace value on ${phase}: ${receipt}; bytes and mtimes unchanged`);

        withDrift("reviewer150Presence.ts", () => [
            "import { WatchCompilerHostOfConfigFile, EmitAndSemanticDiagnosticsBuilderProgram } from \"./_namespaces/ts.js\";",
            "",
            "const keysName = \"keys\";",
            "export function presence(host: WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>): readonly unknown[] {",
            "    return [",
            "        Object.getOwnPropertyDescriptor(host, \"watchFile\"),",
            "        Reflect.has(host, \"watchFile\"),",
            "        Reflect.getOwnPropertyDescriptor(host, \"watchFile\"),",
            "        Object[keysName](host),",
            "    ];",
            "}",
            "",
        ].join("\n"), () => {
            const { roots, rows } = scanEnumeration(directory);
            const file = "src/compiler/reviewer150Presence.ts";
            const reads = rows.filter(([, source]) => source === file).map(([operation]) => operation);
            const presence = ["Object.getOwnPropertyDescriptor", "Reflect.has", "Reflect.getOwnPropertyDescriptor"];
            assert.deepEqual(reads.filter(operation => operation !== "Object.keys"), presence);
            assert.equal(rows.filter(([operation, source]) => operation !== "Object.keys" || source !== file).length,
                expected.length + presence.length);
            assert.throws(() => checkEnumeration(rows), { code: "ERR_ASSERTION" });
            receipt = `node ${path.join(__dirname, "enumeration.cjs")} ${directory} exit=0, ${roots} roots, ${rows.length} readers, oracle ERR_ASSERTION for ${presence.join(", ")} in ${file}; Object[keysName](host) found=${reads.includes("Object.keys")}`;
        }, directory);
        console.log(`PASS presence enumeration on ${phase}: ${receipt}; bytes and mtimes unchanged`);
    }

    function moduleDrifts(directory, phase) {
        const cases = [
            ["compiler/reviewer15Nested.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer15Nested\.ts:7:118: module compiler\/reviewer15Nested\.ts$/m, [
                "import * as watch from \"./watch.js\";",
                "import * as self from \"./reviewer15Nested.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "export { watch };",
                "const view: { watch: { createWatchHost(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost; }; } = self;",
                "export function nestedWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return view.watch.createWatchHost(system, reporter);",
                "}",
                "",
            ].join("\n")],
            ["compiler/reviewer15NestedAssign.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer15NestedAssign\.ts:7:47: module compiler\/reviewer15NestedAssign\.ts$/m, [
                "import * as self from \"./reviewer15NestedAssign.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "export * as nested from \"./watch.js\";",
                "let makeWatch: (system: System, reportWatchStatus: WatchStatusReporter) => OwnWatchHost;",
                "({ nested: { createWatchHost: makeWatch } } = self);",
                "export function assignedWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return makeWatch(system, reporter);",
                "}",
                "",
            ].join("\n")],
            ["compiler/reviewer15Import.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed dynamic import: src\/compiler\/reviewer15Import\.ts:4:118: compiler\/watch\.ts$/m, [
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "const loading: Promise<{ createWatchHost(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost; }> = import(\"./watch.js\");",
                "export async function importedWatch(system: System, reporter: WatchStatusReporter): Promise<OwnWatchHost> {",
                "    return (await loading).createWatchHost(system, reporter);",
                "}",
                "",
            ].join("\n")],
            ["compiler/reviewer15Shorthand.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/reviewer15Shorthand\.ts:5:120: createDefaultWatchCompilerHostOfConfigFile, createDefaultWatchCompilerHostOfFilesAndCompilerOptions, createWatchHost$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "const held: { watch: { createWatchHost(system: System, reportWatchStatus: WatchStatusReporter): OwnWatchHost; }; } = { watch };",
                "export function shorthandWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return held.watch.createWatchHost(system, reporter);",
                "}",
                "",
            ].join("\n")],
            ["compiler/seats16TwinAssign.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed twin reference: src\/compiler\/seats16TwinAssign\.ts:5:12: createDefaultWatchCompilerHostOfConfigFile$/m, [
                "import * as watch from \"./watch.js\";",
                "import { System, WatchCompilerHostOfConfigFile, EmitAndSemanticDiagnosticsBuilderProgram } from \"./_namespaces/ts.js\";",
                "",
                "let assignedTwin: (input: { configFileName: string; system: System; }) => WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram>;",
                "({ held: { createDefaultWatchCompilerHostOfConfigFile: assignedTwin } } = { held: watch });",
                "export function assignedTwinHost(system: System): WatchCompilerHostOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> {",
                "    return assignedTwin({ configFileName: \"tsconfig.json\", system });",
                "}",
                "",
            ].join("\n")],
            ["compiler/seats16NamespaceHold.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed namespace value: src\/compiler\/seats16NamespaceHold\.ts:6:27: module compiler\/watch\.ts$/m, [
                "import * as watch from \"./watch.js\";",
                "import { OwnWatchHost } from \"./watch.js\";",
                "import { System, WatchStatusReporter } from \"./_namespaces/ts.js\";",
                "",
                "namespace hold {",
                "    export import inner = watch;",
                "}",
                "export function heldNamespaceWatch(system: System, reporter: WatchStatusReporter): OwnWatchHost {",
                "    return hold.inner.createWatchHost(system, reporter);",
                "}",
                "",
            ].join("\n")],
            ["compiler/seats16ImportUnresolved.ts", /^AssertionError \[ERR_ASSERTION\]: unreviewed dynamic import: src\/compiler\/seats16ImportUnresolved\.ts:2:12: unresolved$/m, [
                "export function loadModule(specifier: string): Promise<unknown> {",
                "    return import(specifier);",
                "}",
                "",
            ].join("\n")],
        ];
        for (const [file, reason, source] of cases) {
            if (phase === "first apply") assertBeforeSnapshot(directory);
            let receipt;
            withSourceDrift(file, () => source, () => {
                receipt = rejected(reason, directory);
            }, directory);
            console.log(`PASS ${path.basename(file)} on ${phase}: ${receipt}; bytes and mtimes unchanged`);
        }
        const surfaceCases = [
            ["server/protocol.ts",
                "export type DiagnosticPerformanceData = { [Kind in DiagnosticEventKind]?: number; };",
                "export type DiagnosticPerformanceData = { [Key in DiagnosticEventKind]?: number; };",
                /^AssertionError \[ERR_ASSERTION\]: module surface drift: src\/server\/session\.ts:323:35: session\-command\-names: server\/protocol\.ts$/m],
            ["services/textChanges.ts",
                "export type ValidateNonFormattedText = (node: Node, text: string) => void;",
                "export type ValidateNonFormattedText = (node: Node, newText: string) => void;",
                /^AssertionError \[ERR_ASSERTION\]: module surface drift: src\/testRunner\/unittests\/services\/textChanges\.ts:131:28: text\-changes\-delete\-node: services\/textChanges\.ts$/m],
        ];
        for (const [file, old, replacement, reason] of surfaceCases) {
            if (phase === "first apply") assertBeforeSnapshot(directory);
            let receipt;
            withSourceDrift(file, text => {
                assert.equal(text.split(old).length, 2, file);
                const changed = text.replace(old, replacement);
                assert.notEqual(changed, text, file);
                return changed;
            }, () => {
                receipt = rejected(reason, directory);
            }, directory);
            console.log(`PASS module surface drift in ${file} on ${phase}: ${receipt}; bytes and mtimes unchanged`);
        }
        if (phase === "first apply") assertBeforeSnapshot(directory);
        const file = "services/textChanges.ts";
        const line = 'export * as reviewer17Watch from "../compiler/watch.js";';
        let receipt;
        withSourceDrift(file, text => {
            assert.equal(text.split(line).length, 1, file);
            const newline = text.includes("\r\n") ? "\r\n" : "\n";
            const separator = text.endsWith(newline) ? "" : newline;
            const changed = text + separator + line + newline;
            assert.notEqual(changed, text, file);
            return changed;
        }, () => {
            receipt = rejected(/^AssertionError \[ERR_ASSERTION\]: module surface drift: src\/testRunner\/unittests\/services\/textChanges\.ts:131:28: text\-changes\-delete\-node: services\/textChanges\.ts$/m, directory);
        }, directory);
        console.log(`PASS module surface namespace export appended to ${file} on ${phase}: ${receipt}; bytes and mtimes unchanged`);
    }

    const firstApply = path.join(scratch, "first-apply");
    fs.mkdirSync(firstApply);
    execFileSync("tar", ["-xf", beforeArchive, "-C", firstApply]);
    const beforeScan = scanEnumeration(firstApply);
    checkEnumeration(beforeScan.rows);
    console.log(`enumeration.cjs before snapshot exit=0, ${beforeScan.roots} roots, ${beforeScan.rows.length} construction copies`);
    reviewerDrifts(firstApply, "first apply");
    namespaceDrifts(firstApply, "first apply");
    moduleDrifts(firstApply, "first apply");
    fs.rmSync(firstApply, { recursive: true, force: true });
    reviewerDrifts(scratch, "replay");
    namespaceDrifts(scratch, "replay");
    moduleDrifts(scratch, "replay");

} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}

const pairs = [
    ["watch.ts", "createWatchCompilerHostOfConfigFile", "createDefaultWatchCompilerHostOfConfigFile"],
    ["watch.ts", "createWatchCompilerHostOfFilesAndCompilerOptions", "createDefaultWatchCompilerHostOfFilesAndCompilerOptions"],
    ["tsbuildPublic.ts", "createSolutionBuilderHost", "createDefaultSolutionBuilderHost"],
    ["tsbuildPublic.ts", "createSolutionBuilderWithWatchHost", "createDefaultSolutionBuilderWithWatchHost"],
];
function body(sources, file, name) {
    const ast = ts.createSourceFile(file, sources[file].toString(), ts.ScriptTarget.Latest, true);
    const candidates = ast.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
    assert.equal(candidates.length, 1, name);
    return candidates[0].getText(ast).replace(/^export /, "");
}
const shared = [
    ["watch.ts", "createWatchHost"], ["watch.ts", "reportUnrecoverableDiagnostic"],
    ["core.ts", "memoize"], ["core.ts", "copyProperties"],
];
function compile(sources, stock) {
    const helpers = stock ? [
        ...shared, ["watch.ts", "createProgramHost"], ["watch.ts", "createWatchCompilerHost"],
        ["tsbuildPublic.ts", "createSolutionBuilderHostBase"],
    ] : shared;
    const picked = [...helpers, ...pairs.map(([file, before, after]) => [file, stock ? before : after])];
    const result = ts.transpileModule(picked.map(([file, name]) => body(sources, file, name)).join("\n"), {
        compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS },
        reportDiagnostics: true,
    });
    assert.deepEqual(result.diagnostics, []);
    return new vm.Script(`"use strict";\n${result.outputText}\n({${picked.map(([, name]) => name).join(",")}})`);
}
const stock = compile(before, true);
const revised = compile(after, false);

function fixture(present, provided) {
    const log = [];
    let phase = "building";
    const record = (...args) => log.push([phase, ...args]);
    const optional = new Set([
        "realpath", "getEnvironmentVariable", "createHash", "now", "watchFile", "watchDirectory",
        "setTimeout", "clearTimeout", "getModifiedTime", "setModifiedTime", "deleteFile", "preferNonRecursiveWatch",
    ]);
    const values = { newLine: "\n", useCaseSensitiveFileNames: true, storeSignatureInfo: true, preferNonRecursiveWatch: false };
    const methods = new Map();
    let nowReads = 0;
    const system = new Proxy({}, {
        get(target, key) {
            record("get", key);
            if (!present && optional.has(key)) return undefined;
            if (Object.hasOwn(values, key)) return values[key];
            if (key === "now") {
                const value = ++nowReads;
                return function () { record("call", key, this === system, []); return value; };
            }
            if (!methods.has(key)) methods.set(key, function (...args) {
                record("call", key, this === system, args);
                if (key === "getCurrentDirectory") return "/working";
                if (key === "getExecutingFilePath") return "/lib/tsc.js";
                return key;
            });
            return methods.get(key);
        },
    });
    const reporter = name => (...args) => record("reporter", name, args);
    const reportDiagnostic = reporter("diagnostic");
    const reportSolutionBuilderStatus = reporter("solutionStatus");
    const reportWatchStatus = reporter("watchStatus");
    const reportErrorSummary = reporter("errorSummary");
    let factoryResult;
    const defaultFactory = (...args) => {
        record("factory", args);
        factoryResult = {};
        return factoryResult;
    };
    const globals = {
        sys: system, hasOwnProperty: Object.prototype.hasOwnProperty,
        createEmitAndSemanticDiagnosticsBuilderProgram: defaultFactory,
        normalizePath: value => value,
        getDirectoryPath: value => value.slice(0, value.lastIndexOf("/")),
        getDefaultLibFileName: () => "lib.d.ts",
        combinePaths: (...args) => args.join("/"),
        createDiagnosticReporter: () => { record("create reporter", "defaultDiagnostic"); return reporter("defaultDiagnostic"); },
        createBuilderStatusReporter: () => { record("create reporter", "defaultBuildStatus"); return reporter("defaultBuildStatus"); },
        createWatchStatusReporter: (...args) => {
            record("create reporter", "defaultWatchStatus", args.map(arg => arg === system ? "system" : arg));
            assert.equal(args.length, 1);
            assert.equal(args[0] === system, true);
            return reporter("defaultWatchStatus");
        },
        returnNoopFileWatcher: () => "no watcher",
        returnUndefined: () => undefined,
        noop: () => undefined,
        ExitStatus: { DiagnosticsPresent_OutputsSkipped: 1 },
        getNewLineCharacter: () => "\n",
        createCompilerDiagnostic: (message, count) => ({ message, count }),
        getWatchErrorSummaryDiagnosticMessage: () => "watch summary",
        emitFilesAndReportErrors: (builder, report, write, summary) => {
            report("diagnostic");
            write("emit");
            summary(2);
        },
    };
    const input = {
        configFileName: "tsconfig.json", optionsToExtend: {}, watchOptionsToExtend: {}, extraFileExtensions: [],
        rootFiles: ["root.ts"], options: {}, watchOptions: {}, projectReferences: [],
        system,
        reportDiagnostic: provided ? reportDiagnostic : undefined,
        reportWatchStatus: provided ? reportWatchStatus : undefined,
    };
    return {
        log, globals, input,
        factoryResult: () => factoryResult, reportDiagnostic, reportSolutionBuilderStatus, reportWatchStatus, reportErrorSummary,
        built() { phase = "built"; },
    };
}

function build(program, index, isStock, present, provided) {
    const f = fixture(present, provided);
    const builders = program.runInNewContext(f.globals);
    const name = pairs[index][isStock ? 1 : 2];
    const args = index < 2 ? [f.input] : [
        f.input.system, ...(isStock ? [undefined] : []),
        provided ? f.reportDiagnostic : undefined,
        provided ? f.reportSolutionBuilderStatus : undefined,
        provided ? index === 2 ? f.reportErrorSummary : f.reportWatchStatus : undefined,
    ];
    const host = builders[name](...args);
    assert(!f.log.some(([, op]) => op !== "get" && op !== "create reporter"), "a host member ran during construction");
    f.built();
    if (index !== 2) {
        const creations = f.log.filter(([, operation, reporter]) =>
            operation === "create reporter" && reporter === "defaultWatchStatus");
        assert.equal(creations.length, provided ? 0 : 1);
        if (!provided) assert.deepEqual(f.log[f.log.indexOf(creations[0]) + 1], ["building", "get", "watchFile"]);
    }
    const construction = f.log.slice();
    const result = host.createProgram(...argsByKey.createProgram);
    const calls = f.log.splice(construction.length);
    assert.equal(calls.length, 1);
    assert.deepEqual(calls[0].slice(0, 2), ["built", "factory"]);
    assert.equal(calls[0][2].length, 6);
    for (const [position, value] of argsByKey.createProgram.entries()) assert(Object.is(calls[0][2][position], value));
    assert(Object.is(result, f.factoryResult()));
    if (provided && Object.hasOwn(host, "onWatchStatusChange")) assert(Object.is(host.onWatchStatusChange, f.reportWatchStatus));
    if (provided && Object.hasOwn(host, "reportDiagnostic")) assert(Object.is(host.reportDiagnostic, f.reportDiagnostic));
    if (provided && Object.hasOwn(host, "reportSolutionBuilderStatus")) assert(Object.is(host.reportSolutionBuilderStatus, f.reportSolutionBuilderStatus));
    if (provided && Object.hasOwn(host, "reportErrorSummary")) assert(Object.is(host.reportErrorSummary, f.reportErrorSummary));
    return { host, ...f };
}

const argsByKey = {
    createProgram: [["root.ts"], undefined, { host: true }, { previous: true }, [{ code: 1 }], undefined],
    getDefaultLibFileName: [{}],
    afterProgramCreate: [{ getCompilerOptions: () => ({}) }],
    getDirectories: ["/dir"],
    readDirectory: ["/dir", [".ts"], [], [], 1],
    watchFile: ["/file", () => {}, 1, {}],
    watchDirectory: ["/dir", () => {}, false, {}],
    setTimeout: [() => {}, 1],
    setModifiedTime: ["/file", new Date(0)],
};
function invoke(key, f) {
    const args = argsByKey[key] ?? ["/file", "data", false];
    const value = f.host[key](...args);
    if (key === "getCurrentDirectory") assert.equal(f.host[key](...args), value);
    return value;
}

for (const program of [stock, revised]) {
    const f = fixture(false, true);
    const watcher = program.runInNewContext(f.globals).createWatchHost(f.input.system, f.reportWatchStatus);
    const enumerated = [];
    for (const key in watcher) enumerated.push(key);
    assert.deepEqual(enumerated, Object.keys(watcher));
    assert.deepEqual(enumerated, watchKeys);
}
console.log("PASS WatchHost literals enumerate exactly six own keys in order");

for (const [index, [, label]] of pairs.entries()) {
    for (const present of [false, true]) {
        for (const provided of [false, true]) {
            const a = build(stock, index, true, present, provided);
            const b = build(revised, index, false, present, provided);
            assert.deepEqual(b.log, a.log, "construction access order");
            if (index === 0 && !present) {
                for (const [side, { log }] of [["stock", a], ["twin", b]]) {
                    const at = log.findIndex(([, op, key]) => op === "get" && key === "watchFile");
                    console.log(`reporter ${side} supplied=${provided} ${JSON.stringify(log.slice(at - 2, at + 1))}`);
                }
            }
            const keys = Object.keys(a.host);
            assert.deepEqual(Object.keys(b.host), keys, "own host key order");
            for (const key of watchKeys) {
                assert.equal(Object.hasOwn(b.host, key), Object.hasOwn(a.host, key), key);
            }
            if (index !== 2) {
                assert.deepEqual(keys.filter(key => watchKeys.includes(key)), watchKeys);
                assert(Object.hasOwn(b.host, "preferNonRecursiveWatch"));
                assert.equal(b.host.preferNonRecursiveWatch, present ? false : undefined);
            }
            for (const key of keys) {
                assert.equal(typeof b.host[key], typeof a.host[key], key);
                if (typeof a.host[key] !== "function") {
                    assert.deepEqual(b.host[key], a.host[key], key);
                    continue;
                }
                const old = build(stock, index, true, present, provided);
                const next = build(revised, index, false, present, provided);
                assert.deepEqual(invoke(key, next), invoke(key, old), key);
                assert.deepEqual(next.log, old.log, key);
                if (key === "getCurrentDirectory") {
                    assert.equal(next.log.filter(([, op, name]) => op === "call" && name === key).length, 1);
                }
            }
            if (index < 2) {
                for (const f of [a, b]) {
                    f.host.onWatchStatusChange = function (...args) {
                        f.log.push(["replacement", this === f.host, args]);
                    };
                    f.host.afterProgramCreate(argsByKey.afterProgramCreate[0]);
                    assert.deepEqual(f.log.at(-1).slice(0, 2), ["replacement", true]);
                }
                assert.deepEqual(b.log, a.log);
            }
            console.log(`PASS ${label} optional members ${present} supplied reporters ${provided}`);
        }
    }
}
console.log("PASS reporter creation and first member read agree for every watch host");
console.log("PASS all four factories forward six arguments exactly once and return the builder result");
