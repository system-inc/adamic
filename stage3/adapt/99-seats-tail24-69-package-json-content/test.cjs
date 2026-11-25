"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const { createRequire } = require("node:module");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");
const { findUnit } = require("./adapt.cjs");

const snapshot = "refs/stage3/99-seats-tail24-69-package-json-content";
const resolver = "src/compiler/moduleNameResolver.ts";
const specific = "loadModuleFromSpecificNodeModulesDirectory";
const exportsOrImports = "loadModuleFromExportsOrImports";
const adapter = path.join(__dirname, "adapt.cjs");
const read = (tree, side) => execFileSync("git", ["-C", tree, "show", `${snapshot}-${side}:${resolver}`], { encoding: "utf8" });
const parse = (file, text) => ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const functionText = (text, name) => {
    const source = parse(resolver, text);
    return findUnit(source, name).getText(source);
};

function structure(tmp, before, after) {
    const target = path.join(tmp, resolver);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    function apply(text, success, label) {
        fs.writeFileSync(target, text);
        fs.utimesSync(target, 1, 1);
        const stamp = fs.statSync(target).mtimeMs;
        const run = spawnSync(process.execPath, [adapter, tmp], { encoding: "utf8" });
        if (run.error) throw run.error;
        assert.equal(run.status === 0, success, run.stderr);
        assert.equal(fs.readFileSync(target, "utf8"), success ? after : text, label);
        if (!success || text === after) assert.equal(fs.statSync(target).mtimeMs, stamp, label);
        console.log(`PASS ${label}`);
    }

    apply(before, true, "snapshot bytes");
    apply(after, true, "idempotence with no write");
    const call = /hasProperty\(\(rootPackageInfo = getPackageJsonInfo\(packageDirectory, !nodeModulesDirectoryExists, state\)\)\?\.contents\.packageJsonContent \?\? emptyArray, "exports"\)/;
    const match = before.match(call);
    assert(match);
    apply(before.replace(call, "false"), false, "missing call");
    apply(before.replace(call, match[0].replace("?? emptyArray", "?? []")), false, "altered call");
    apply(before.replace(call, `(${match[0]} || ${match[0]})`), false, "duplicated call");
    assert.equal(functionText(before, exportsOrImports), functionText(after, exportsOrImports));
    console.log("PASS unit 1 source unchanged");
}

function types(tree, tmp) {
    const root = path.join(tmp, "after");
    fs.mkdirSync(root);
    const archive = path.join(tmp, "after.tar");
    execFileSync("git", ["-C", tree, "archive", "--format=tar", `--output=${archive}`, `${snapshot}-after`, "src"]);
    execFileSync("tar", ["-xf", archive, "-C", root]);
    const typeRoots = [path.resolve(path.dirname(require.resolve("typescript/package.json")), "../@types")];
    const options = ts.convertCompilerOptionsFromJson({
        strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
        erasableSyntaxOnly: false, verbatimModuleSyntax: true, allowImportingTsExtensions: true,
        noEmit: true, module: "ESNext", moduleDetection: "force", moduleResolution: "bundler",
        target: "ES2024", lib: ["es2024"], types: ["node"], typeRoots,
    }, root);
    assert.deepEqual(options.errors, []);
    const program = ts.createProgram([path.join(root, "src/tsc/tsc.ts")], options.options);
    const diagnostics = ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length, 0, ts.formatDiagnosticsWithColorAndContext(diagnostics, {
        getCanonicalFileName: name => name, getCurrentDirectory: () => root, getNewLine: () => "\n",
    }));
    console.log("PASS snapshot checker clean");
    const checker = program.getTypeChecker();
    const file = program.getSourceFile(path.join(root, resolver));
    for (const [name, wanted] of [[specific, "PackageJsonPathFields"], [exportsOrImports, "object"]]) {
        const matches = [];
        function visit(node) {
            if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === "hasProperty") matches.push(node);
            ts.forEachChild(node, visit);
        }
        visit(findUnit(file, name));
        assert.equal(matches.length, 1);
        const args = checker.getTypeArgumentsForResolvedSignature(checker.getResolvedSignature(matches[0]));
        assert.equal(args.length, 1);
        assert.equal(checker.typeToString(args[0]), wanted);
        console.log(`PASS ${name} T=${wanted}`);
    }
}

function splice(text) {
    const bundle = require.resolve("typescript");
    const bytes = fs.readFileSync(bundle, "utf8");
    const source = parse(bundle, bytes);
    const matches = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === specific) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, 1);
    const node = matches[0];
    const replacement = ts.transpileModule(text, { compilerOptions: { target: ts.ScriptTarget.ES2024 } }).outputText;
    const module = { exports: {} };
    vm.compileFunction(bytes.slice(0, node.getStart(source)) + replacement + bytes.slice(node.end),
        ["exports", "require", "module", "__filename", "__dirname"], { filename: bundle })(
        module.exports, createRequire(bundle), module, bundle, path.dirname(bundle));
    return module.exports;
}

const rootExports = '{"name":"pkg","exports":{"./sub":{"types":"./sub/index.d.ts"}}}';
const cases = [
    ["undefined content", undefined, "pkg/sub"],
    ["content without exports", '{"name":"pkg"}', "pkg/sub"],
    ["content with exports", rootExports, "pkg/sub"],
    ["null package JSON", "null", "pkg/sub"],
    ["unparseable package JSON", "{ invalid", "pkg/sub"],
    ["no subpath", rootExports, "pkg"],
    ["node10 control", rootExports, "pkg/sub", "Node10"],
];

function observe(api, row, resolution, cached) {
    const [, rootText, moduleName, override] = row;
    const files = new Map([
        ["/app/index.ts", `import "${moduleName}";`],
        ["/app/node_modules/pkg/index.d.ts", "export {};"],
        ["/app/node_modules/pkg/sub/package.json", '{"name":"pkg-sub","version":"1.0.0","types":"./index.d.ts"}'],
        ["/app/node_modules/pkg/sub/index.d.ts", "export {};"],
    ]);
    if (rootText !== undefined) files.set("/app/node_modules/pkg/package.json", rootText);
    const calls = [];
    const traces = [];
    function tracked(name, operation) {
        return (...args) => {
            calls.push([name, ...args]);
            return operation(...args);
        };
    }
    const host = {
        fileExists: tracked("fileExists", file => files.has(file)),
        readFile: tracked("readFile", file => files.get(file)),
        directoryExists: tracked("directoryExists", directory => [...files.keys()].some(file => file.startsWith(directory + "/"))),
        realpath: tracked("realpath", file => file),
        getCurrentDirectory: tracked("getCurrentDirectory", () => "/app"),
        useCaseSensitiveFileNames: tracked("useCaseSensitiveFileNames", () => true),
        trace: line => traces.push(line),
    };
    const options = { traceResolution: true, moduleResolution: api.ModuleResolutionKind[override ?? resolution] };
    const cache = cached ? api.createModuleResolutionCache("/app", file => file, options) : undefined;
    let result, error;
    try {
        result = api.resolveModuleName(moduleName, "/app/index.ts", options, host, cache, undefined, override === "Node10" ? undefined : api.ModuleKind.ESNext);
    } catch (caught) {
        error = { name: caught.name, message: caught.message };
    }
    // Normalize only object realm identity, not key order in either ordered log.
    return structuredClone({ result, error, calls, traces });
}

function runtime(before, after) {
    const control = splice(functionText(before, specific));
    const adapted = splice(functionText(after, specific));
    for (const row of cases) for (const mode of ["NodeNext", "Bundler"]) for (const cached of [false, true]) {
        const label = `${row[0]}, ${mode}, cache=${cached}`;
        const stock = observe(ts, row, mode, cached);
        const beforeResult = observe(control, row, mode, cached);
        const afterResult = observe(adapted, row, mode, cached);
        try {
            assert.deepEqual(beforeResult, stock, label + " control");
            assert.deepEqual(afterResult, stock, label + " adapted");
        } catch (error) {
            console.error(JSON.stringify({ label, stock, beforeResult, afterResult }, null, 2));
            throw error;
        }
        assert.equal(stock.error, undefined, label);
        console.log(`PASS ${label}`);
    }
}

if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node test.cjs APPLIED_TREE");
    assert.equal(ts.version, "6.0.3");
    const tree = path.resolve(process.argv[2]);
    const before = read(tree, "before");
    const after = read(tree, "after");
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "resolver-presence-test-"));
    try {
        structure(tmp, before, after);
        types(tree, tmp);
        runtime(before, after);
    } finally {
        fs.rmSync(tmp, { recursive: true });
    }
}

module.exports = { read, functionText, splice, observe, cases };
