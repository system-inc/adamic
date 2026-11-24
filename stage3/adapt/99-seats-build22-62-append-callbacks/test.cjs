"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");

const [tree] = process.argv.slice(2);
assert(process.argv.length === 3 && ts.version === "6.0.3", "usage: node test.cjs <applied tree>");
const adapter = path.join(__dirname, "adapt.cjs");
const refs = "refs/stage3/99-seats-build22-62-append-callbacks";
const cli = "src/compiler/commandLineParser.ts";
const parser = "src/compiler/parser.ts";
const files = [cli, parser];
const capture = Symbol.for("adamic.family62.parseDiagnostics");
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-append-output-"));
const beforeTree = path.join(scratch, "before");
const afterTree = path.join(scratch, "after");
const staged = file => path.join(afterTree, file);

function git(...args) {
    return execFileSync("git", ["-C", tree, ...args]);
}

function apply() {
    return JSON.parse(execFileSync(process.execPath, [adapter, afterTree], { encoding: "utf8" }));
}

function emit(label, sourceTree) {
    const outDir = path.join(scratch, label);
    const config = ts.getParsedCommandLineOfConfigFile(path.join(sourceTree, "src/tsconfig-base.json"), {}, {
        ...ts.sys,
        onUnRecoverableConfigFileDiagnostic(diagnostic) { throw new Error(String(diagnostic.messageText)); },
    });
    assert.ok(config);
    const options = {
        ...config.options, outDir, rootDir: path.join(sourceTree, "src"),
        composite: false, incremental: false, tsBuildInfoFile: undefined,
        declaration: false, declarationMap: false, emitDeclarationOnly: false,
        isolatedDeclarations: false, sourceMap: false, types: ["node"],
        typeRoots: [path.resolve(path.dirname(require.resolve("@types/node/package.json")), "..")],
    };
    const host = ts.createCompilerHost(options);
    host.getCurrentDirectory = () => sourceTree;
    const program = ts.createProgram([path.join(sourceTree, "src/typescript/typescript.ts")], options, host);
    const diagnostics = ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length, 0, ts.formatDiagnosticsWithColorAndContext(diagnostics.slice(0, 10), host));
    let observed = false;
    const result = program.emit(undefined, (file, text, ...args) => {
        if (file === path.join(outDir, "compiler/parser.js")) {
            // Observe the real array at construction, before either JSON path starts conversion.
            const anchor = /sourceFile\.parseDiagnostics = [^\n]+;/g;
            assert.equal([...text.matchAll(anchor)].length, 2);
            text = text.replace(anchor, assignment =>
                assignment + '\n        globalThis[Symbol.for("adamic.family62.parseDiagnostics")]?.(sourceFile);');
            observed = true;
        }
        host.writeFile(file, text, ...args);
    });
    assert.equal(result.emitSkipped, false);
    assert.equal(result.diagnostics.length, 0);
    assert.ok(observed, "expected the emitted parser instrument");
    return require(path.join(outDir, "typescript/typescript.js"));
}

function observe(run) {
    const parsed = [];
    globalThis[capture] = file => parsed.push({ file, diagnostics: file.parseDiagnostics });
    try {
        const output = run();
        assert.equal(parsed.length, 1, "expected one parse of the input");
        const { file, diagnostics } = parsed[0];
        assert.equal(file.parseDiagnostics, diagnostics, "source file must retain its original diagnostic array");
        return { output, file, diagnostics };
    } finally {
        delete globalThis[capture];
    }
}

const malformed = "{bare: 'v', 'quoted': 'second', short, optional?: 1, bad: foo}";
const fileName = "/tsconfig.json";
const codesAndSpans = [
    [1327, 1, 4], [1327, 7, 3], [1327, 12, 8], [1327, 22, 8], [1136, 32, 5],
    [8009, 47, 1], [1327, 39, 8], [1327, 53, 3], [1328, 58, 3],
];

function behavior(ts) {
    const diag = items => items.map(d => [
        d.code, d.start, d.length, ts.flattenDiagnosticMessageText(d.messageText, "\n"),
    ]);
    const source = ts.parseJsonText(fileName, malformed);
    const seed = { code: 1, start: 0, length: 0, messageText: "existing diagnostic" };
    const errors = [seed];
    const value = ts.convertToObject(source, errors);
    assert.ok(errors.some(d => d.code === 1327 && d.start === 12 && d.length === 8),
        "missing single-quoted-name diagnostic at 12:8");
    assert.equal(errors[0], seed, "append order must keep the caller array's existing diagnostic first");
    assert.deepEqual(diag(errors.slice(1)).map(d => d.slice(0, 3)), codesAndSpans,
        "conversion diagnostic order and source spans");
    assert.deepEqual(value, { bare: "v", quoted: "second", optional: 1, bad: undefined });

    const textCases = [
        [malformed, value, codesAndSpans],
        ["[{'recovered': 1}]", { recovered: 1 }, [[5092, 0, 18], [1327, 2, 11]]],
        ["17", {}, [[5092, 0, 2]]],
    ].map(([text, expected, spans]) => {
        const { output, diagnostics } = observe(() => ts.parseConfigFileTextToJson(fileName, text));
        assert.equal(output.error, diagnostics[0], "error must come from the original array after conversion");
        assert.deepEqual(output.config, expected);
        assert.deepEqual(diag(diagnostics).map(d => d.slice(0, 3)), spans, text);
        return { config: output.config, diagnostics: diag(diagnostics), error: diag([output.error]) };
    });
    const parserCase = observe(() => ts.createSourceFile(fileName, malformed,
        ts.ScriptTarget.ES2020, true, ts.ScriptKind.JSON));
    assert.equal(parserCase.output, parserCase.file);
    assert.deepEqual(diag(parserCase.diagnostics).map(d => d.slice(0, 3)), codesAndSpans);

    const configCases = ["17", "foo"].map(target => {
        const config = ts.parseJsonSourceFileConfigFileContent(
            ts.parseJsonText(fileName, `{"compilerOptions":{"target":${target}}}`),
            { useCaseSensitiveFileNames: true, readDirectory: () => ["/input.ts"],
                fileExists: () => false, readFile: () => undefined }, "/");
        assert.deepEqual(diag(config.errors).map(d => d.slice(0, 3)), [[5024, 29, target.length]]);
        assert.deepEqual(config.raw, { compilerOptions: { target: target === "17" ? 17 : undefined } });
        return { raw: config.raw, options: config.options, diagnostics: diag(config.errors) };
    });
    return { value, errors: diag(errors), textCases, parser: diag(parserCase.diagnostics), configCases };
}

try {
    const archive = path.join(scratch, "before.tar");
    fs.mkdirSync(beforeTree);
    git("archive", "--format=tar", "-o", archive, refs + "-before");
    execFileSync("tar", ["-xf", archive, "-C", beforeTree]);
    fs.unlinkSync(archive);
    const before = behavior(emit("before-emit", beforeTree));
    fs.cpSync(beforeTree, afterTree, { recursive: true });
    assert.deepEqual(apply(), { edits: 15 });
    const after = files.map(file => fs.readFileSync(staged(file)));
    assert.deepEqual(behavior(emit("after-emit", afterTree)), before);
    const changed = git("diff", "--name-only", refs + "-before", refs + "-after").toString().trim().split("\n");
    assert.deepEqual(changed, files);
    for (const [index, file] of files.entries()) {
        assert.deepEqual(after[index], git("show", refs + "-after:" + file), file + " differs from after snapshot");
    }
    assert.deepEqual(apply(), { edits: 0 });
    assert.deepEqual(files.map(file => fs.readFileSync(staged(file))), after);
    console.log("PASS: 15 edits match the after snapshot, then 0 with identical bytes; seven append sites, diagnostics, values, and caller arrays");

    for (const file of files) fs.copyFileSync(path.join(beforeTree, file), staged(file));
    const text = fs.readFileSync(path.join(beforeTree, cli), "utf8");
    const source = ts.createSourceFile(cli, text, ts.ScriptTarget.Latest, true);
    const worker = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "convertToJson");
    assert.ok(worker?.body);
    const at = worker.body.end - 1;
    const extra = "    errors.push(createDiagnosticForNodeInSourceFile(sourceFile, sourceFile, Diagnostics.Property_assignment_expected));\r\n";
    fs.writeFileSync(staged(cli), text.slice(0, at) + extra + text.slice(at));
    const drift = files.map(file => fs.readFileSync(staged(file)));
    const result = spawnSync(process.execPath, [adapter, afterTree], { encoding: "utf8" });
    assert.ifError(result.error);
    assert.notEqual(result.status, 0, `reference guard accepted an extra errors.push: ${result.stdout.trim()}`);
    assert.match(result.stderr, /convertToJson: expected 6 rewritten errors references/);
    assert.deepEqual(files.map(file => fs.readFileSync(staged(file))), drift, "drift rejection must write neither file");
    console.log("PASS: extra errors.push rejected and both drift input files unchanged");

    for (const file of files) fs.copyFileSync(path.join(beforeTree, file), staged(file));
    const nested = worker.body.statements.find(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "convertObjectLiteralExpressionToJson");
    assert.ok(nested?.body?.statements.length);
    const start = nested.body.statements[0].getStart(source);
    const indent = text.slice(text.lastIndexOf("\n", start) + 1, start);
    const newline = text.includes("\r\n") ? "\r\n" : "\n";
    const shadow = [
        "const appendDiagnostic = (_diagnostic: DiagnosticWithLocation) => 0;",
        "void appendDiagnostic;",
        "",
    ].join(newline + indent);
    fs.writeFileSync(staged(cli), text.slice(0, start) + shadow + text.slice(start));
    const shadowed = files.map(file => fs.readFileSync(staged(file)));
    const refusal = spawnSync(process.execPath, [adapter, afterTree], { encoding: "utf8" });
    assert.ifError(refusal.error);
    assert.notEqual(refusal.status, 0, `appendDiagnostic guard accepted shadowing: ${refusal.stdout.trim()}`);
    assert.match(refusal.stderr, /convertToJson: appendDiagnostic guard expected 0 pinned identifiers, found 2/);
    assert.deepEqual(files.map(file => fs.readFileSync(staged(file))), shadowed, "shadow rejection must write neither file");
    console.log("PASS: nested appendDiagnostic shadow rejected and both input files unchanged");
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
