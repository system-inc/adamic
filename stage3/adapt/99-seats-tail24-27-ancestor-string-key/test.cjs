"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const {execFileSync} = require("node:child_process");
const ts = require("typescript");
const {adapt, locate, nodes, resolver, rewrite, services} = require("./adapt.cjs");

assert.equal(ts.version, "6.0.3");
const tree = path.resolve(process.argv[2]);
const id = "refs/stage3/99-seats-tail24-27-ancestor-string-key";
const compilerPath = "src/compiler/path.ts";
const utilities = "src/compiler/utilities.ts";
const read = (side, relative) => execFileSync("git", ["-C", tree, "show", id + "-" + side + ":" + relative], {encoding: "utf8"});
const parse = (relative, raw) => ts.createSourceFile(relative, raw, ts.ScriptTarget.Latest, true);
const functions = (source, name) => source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body);
function implementation(source, name) {
    const found = functions(source, name);
    assert.equal(found.length, 1, name);
    return found[0];
}

for (const relative of [resolver, services]) {
    const before = read("before", relative);
    const after = read("after", relative);
    assert.notEqual(after, before, relative);
    const source = parse(relative, after);
    const targets = locate(source, relative);
    assert.equal(rewrite(source, targets, "before"), before, "Only the pinned nodes may change");
    if (relative === resolver) {
        assert.deepEqual(targets.wrapper.typeParameters.map(node => node.name.text), ["T"]);
        assert.equal(nodes(targets.wrapper.body, ts.isAsExpression).length, 0);
    } else {
        const owner = implementation(source, "getIsExcluded");
        assert.equal(nodes(owner.body, ts.isAsExpression).length, 0);
        assert.equal(nodes(owner.body, node => ts.isCallExpression(node) && node.expression.getText(source) === "toPath").length, 1);
    }
}
for (const relative of [compilerPath, utilities, "src/compiler/core.ts"]) {
    assert.equal(read("before", relative), read("after", relative), "Helper source changed");
}
console.log("PASS snapshot structure and unchanged helper sources");

function withBeforeTree(run) {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "ancestor-pins-"));
    try {
        const archive = path.join(scratch, "source.tar");
        execFileSync("git", ["-C", tree, "archive", "-o", archive, id + "-before", "src"]);
        execFileSync("tar", ["-xf", archive, "-C", scratch]);
        fs.unlinkSync(archive);
        run(scratch);
    } finally {
        fs.rmSync(scratch, {recursive: true, force: true});
    }
}

function recordWrites(run) {
    const writes = [];
    const write = fs.writeFileSync;
    try {
        fs.writeFileSync = (...args) => { writes.push(args[0]); return write(...args); };
        run();
    } finally {
        fs.writeFileSync = write;
    }
    return writes;
}

function filesAt(scratch) {
    return [resolver, services].map(relative => {
        const file = path.join(scratch, relative);
        return {bytes: fs.readFileSync(file, "utf8"), modified: fs.statSync(file, {bigint: true}).mtimeNs};
    });
}

withBeforeTree(scratch => {
    assert.deepEqual(adapt(scratch), {files: 2, edits: 3});
    const first = filesAt(scratch);
    assert.deepEqual(first.map(item => item.bytes), [read("after", resolver), read("after", services)]);
    assert.deepEqual(recordWrites(() => assert.deepEqual(adapt(scratch), {files: 0, edits: 0})), []);
    assert.deepEqual(filesAt(scratch), first);
    console.log("PASS forward replay matches both after snapshots and a second run preserves bytes and mtimes");
});

function load(side, observed) {
    const files = new Map([resolver, services, compilerPath, utilities].map(relative => [relative, parse(relative, read(side, relative))]));
    const wanted = [
        ...["forEachAncestorDirectory", "getDirectoryPath", "ensureTrailingDirectorySeparator", "toPath"].map(name => [compilerPath, name]),
        [resolver, "pathContainsNodeModules"],
        [utilities, "hostUsesCaseSensitiveFileNames"],
        [utilities, "hostGetCanonicalFileName"],
        [resolver, "forEachAncestorDirectoryStoppingAtGlobalCache"],
        ...["forEachExternalModuleToImportFrom", "getIsExcludedPatterns", "forEachExternalModule", "getIsExcluded"].map(name => [services, name]),
    ];
    const code = wanted.map(([relative, name]) => {
        const source = files.get(relative);
        return implementation(source, name).getText(source).replace(/^export /, "");
    }).join("\n");
    const js = ts.transpileModule(code, {compilerOptions: {target: ts.ScriptTarget.ES2020}}).outputText;
    return vm.runInNewContext(js + `
        const snapshotToPath = toPath;
        toPath = (dir, base, canonicalize) => {
            const output = snapshotToPath(dir, base, canonicalize);
            observe(dir, output);
            return output;
        };
        ({getIsExcluded, forEachExternalModuleToImportFrom});
    `, {...ts, observe: (input, output) => observed.push([input, output])});
}

const programOptions = {noLib: true, types: []};
function makeProgram(base, sensitive, fileName, host = ts.createCompilerHost(programOptions)) {
    host.getCurrentDirectory = () => base;
    host.useCaseSensitiveFileNames = () => sensitive;
    host.getCanonicalFileName = ts.createGetCanonicalFileName(sensitive);
    host.getSourceFile = name => name === fileName
        ? ts.createSourceFile(name, "export const member = 1;", ts.ScriptTarget.Latest)
        : undefined;
    const result = ts.createProgram([fileName], programOptions, host);
    assert.equal(result.getSourceFiles().length, 1, fileName);
    return result;
}

function execute(side, sample, main, provider) {
    const paths = [];
    const api = runners[side];
    const lookups = [];
    const calls = [];
    const read = (method, value) => () => { calls.push(method); return value; };
    const host = Object.assign(sample.host ?? {}, {
        getCurrentDirectory: read("getCurrentDirectory", sample.base),
        useCaseSensitiveFileNames: read("useCaseSensitiveFileNames", sample.sensitive),
        getGlobalTypingsCacheLocation: read("getGlobalTypingsCacheLocation", sample.stop),
        getPackageJsonAutoImportProvider: read("getPackageJsonAutoImportProvider", provider),
        getSymlinkCache: read("getSymlinkCache", {
            getSymlinkedDirectoriesByRealpath: read("getSymlinkedDirectoriesByRealpath", {
                size: 1,
                get(key) {
                    lookups.push(key);
                    return key === sample.hit ? ["/excluded"] : undefined;
                },
            }),
        }),
    });
    observations.length = 0;
    let result;
    if (provider) {
        result = [];
        api.forEachExternalModuleToImportFrom(main, host,
            {autoImportFileExcludePatterns: ["/excluded/**"]}, true,
            (_, file, program) => result.push([program === main ? "main" : "provider", file.path]));
    } else {
        result = api.getIsExcluded([/^\/excluded\//], host)(main.getSourceFiles()[0]);
    }
    paths.push(...observations);
    return {result, lookups, calls, paths};
}

const observations = [];
const runners = {before: load("before", observations), after: load("after", observations)};
const cases = [
    ["ancestor hit and excluded after symlink replacement", "/Work/node_modules/Pkg/nested/index.ts", {hit: "/Work/node_modules/Pkg/"}],
    ["POSIX miss", "/Work/node_modules/Pkg/index.ts"],
    ["upper drive", "C:/Work/node_modules/Pkg/index.ts"],
    ["lower drive", "c:/Work/node_modules/Pkg/index.ts"],
    ["UNC", "//Server/Share/node_modules/Pkg/index.ts"],
    ["file URL", "file:///Work/node_modules/Pkg/index.ts"],
    ["dynamic", "^/untitled/ts-nul-authority/node_modules/Pkg/index.ts"],
    ["mixed case and Unicode", "/Work/node_modules/İıßÄ/index.ts"],
    ["nested node_modules", "/Work/node_modules/Pkg/node_modules/Inner/index.ts"],
    ["root file", "/root.ts"],
    ["global cache stop", "/Work/node_modules/Pkg/nested/index.ts", {stop: "/Work/node_modules/Pkg"}],
];

function verify(sample, main, provider) {
    const before = execute("before", sample, main, provider);
    const after = execute("after", sample, main, provider);
    const canonicalize = ts.createGetCanonicalFileName(sample.sensitive);
    const expectedKeys = before.lookups.map(key => {
        const directory = ts.getRootLength(key) === key.length ? key : ts.removeTrailingDirectorySeparator(key);
        const rebuilt = ts.toPath(directory, sample.base, canonicalize);
        return ts.ensureTrailingDirectorySeparator(rebuilt);
    });
    if (sample.changed) expectedKeys.push("/provider/node_modules/", "/provider/", "/");
    assert.deepEqual(after.lookups, expectedKeys, sample.name + " lookup keys");
    assert.deepEqual(after.calls.filter(call => call !== "getCurrentDirectory" && call !== "useCaseSensitiveFileNames"),
        before.calls.filter(call => call !== "getCurrentDirectory" && call !== "useCaseSensitiveFileNames"));
    const reads = {};
    for (const method of ["getCurrentDirectory", "useCaseSensitiveFileNames"]) {
        reads[method] = after.calls.filter(call => call === method).length - before.calls.filter(call => call === method).length;
        assert.equal(reads[method], after.lookups.length);
    }
    const differences = [];
    assert.equal(after.paths.length, after.lookups.length);
    for (const [input, output] of after.paths) {
        if (ts.getRootLength(input) > 0) {
            for (const base of [sample.base, "", "/Unrelated"]) {
                assert.equal(ts.toPath(input, base, canonicalize), output, sample.name + " base independence");
            }
        }
        if (input !== output) differences.push([input, output]);
    }
    if (sample.changed) {
        assert.deepEqual(differences, sample.changed);
        assert.deepEqual(before.result, [["main", main.getSourceFiles()[0].path]]);
        assert.deepEqual(after.result, [...before.result, ["provider", provider.getSourceFiles()[0].path]]);
    } else {
        assert.deepEqual(differences, [], sample.name);
        assert.deepEqual(after.lookups, before.lookups);
        assert.deepEqual(after.result, before.result);
        if (sample.expectedKeys) {
            assert.deepEqual(before.lookups, sample.expectedKeys);
            assert.equal(before.result, false);
        }
        if (sample.hit && !provider) assert.equal(before.result, true);
        if (sample.stop) assert.equal(before.lookups.at(-1), ts.ensureTrailingDirectorySeparator(sample.stop));
        if (sample.name.startsWith("root file")) assert.deepEqual(before.lookups, []);
        if (!provider && !sample.hit && !sample.stop && !sample.name.startsWith("root file")) assert(before.lookups.length > 0, sample.name);
    }
    if (sample.outside) console.log("OBSERVED", sample.name, JSON.stringify({before, after, differences}));
    console.log("PASS", sample.name, JSON.stringify(reads));
}

const cacheCalls = ["getSymlinkCache", "getSymlinkedDirectoriesByRealpath", "getGlobalTypingsCacheLocation"];
const lookupReads = ["getCurrentDirectory", "useCaseSensitiveFileNames"];

function observeEmptyBaseProvider(sensitive, main) {
    const fileName = "rel/node_modules/UPPER/index.ts";
    const provider = makeProgram("", sensitive, fileName);
    const canonicalize = ts.createGetCanonicalFileName(sensitive);
    const stored = provider.getSourceFiles()[0].path;
    const directory = canonicalize("rel/node_modules/UPPER");
    assert.equal(stored, directory + "/index.ts");
    const absolute = canonicalize("/Work") + "/" + directory;
    const sample = {base: "/Work", sensitive, hit: absolute + "/"};
    const before = execute("before", sample, main, provider);
    const after = execute("after", sample, main, provider);
    const mainPath = main.getSourceFiles()[0].path;
    const mainDirs = [canonicalize("/Work/node_modules/Main"), canonicalize("/Work/node_modules"), canonicalize("/Work"), "/"];
    const mainKeys = mainDirs.map(ts.ensureTrailingDirectorySeparator);
    assert.deepEqual(before, {
        result: [["main", mainPath], ["provider", stored]],
        lookups: [...mainKeys, directory + "/", "rel/node_modules/", "rel/", "/"],
        calls: ["useCaseSensitiveFileNames", ...cacheCalls, "getPackageJsonAutoImportProvider", ...cacheCalls],
        paths: [],
    });
    assert.deepEqual(after, {
        result: [["main", mainPath]],
        lookups: [...mainKeys, sample.hit],
        calls: ["useCaseSensitiveFileNames", ...cacheCalls, ...mainDirs.flatMap(() => lookupReads),
            "getPackageJsonAutoImportProvider", ...cacheCalls, ...lookupReads],
        paths: [...mainDirs.map(dir => [dir, dir]), [directory, absolute]],
    });
    console.log("OBSERVED empty-base provider", sensitive ? "sensitive" : "folding", JSON.stringify({stored, before, after}));
}

function observeLivePolicy() {
    const host = ts.createCompilerHost(programOptions);
    const stored = "/Work/node_modules/UPPER/index.ts";
    const main = makeProgram("/Work", true, stored, host);
    assert.equal(main.getSourceFiles()[0].path, stored);
    const sample = {host, base: "/Work", sensitive: false, hit: "/Work/node_modules/UPPER/"};
    const before = execute("before", sample, main);
    const after = execute("after", sample, main);
    assert.deepEqual(before, {result: true, lookups: [sample.hit], calls: cacheCalls, paths: []});
    assert.deepEqual(after, {
        result: false,
        lookups: ["/work/node_modules/upper/", "/work/node_modules/", "/work/", "/"],
        calls: [...cacheCalls, ...Array.from({length: 4}, () => lookupReads).flat()],
        paths: [["/Work/node_modules/UPPER", "/work/node_modules/upper"], ["/Work/node_modules", "/work/node_modules"],
            ["/Work", "/work"], ["/", "/"]],
    });
    console.log("OBSERVED live case policy", JSON.stringify({stored, before, after}));
}

for (const sensitive of [true, false]) {
    const base = "/Work";
    const canon = ts.createGetCanonicalFileName(sensitive);
    for (const [name, fileName, options] of cases) {
        verify({name: name + (sensitive ? " sensitive" : " folding"), base, sensitive,
            stop: options?.stop && canon(options.stop), hit: options?.hit && canon(options.hit)},
        makeProgram(base, sensitive, fileName));
    }
    const main = makeProgram(base, sensitive, "/Work/node_modules/Main/index.ts");
    const providerCase = (name, providerBase, providerSensitive, fileName, outside, hit, changed) =>
        verify({name: name + (sensitive ? " sensitive host" : " folding host"), base, sensitive, outside, hit, changed},
            main, makeProgram(providerBase, providerSensitive, fileName));
    verify({name: "empty-base main " + (sensitive ? "sensitive" : "folding"), base: "", sensitive,
        expectedKeys: [canon("rel/node_modules/Pkg/"), "rel/node_modules/", "rel/", "/"]},
        makeProgram("", sensitive, "rel/node_modules/Pkg/index.ts"));
    observeEmptyBaseProvider(sensitive, main);
    providerCase("shared provider", base, sensitive, "/Work/node_modules/Provider/index.ts", false);
    providerCase("unrelated provider base", "/Other", sensitive, "relative/node_modules/Provider/index.ts", true);
    if (sensitive) {
        providerCase("outside provider folds", "/provider", false, "/provider/node_modules/UPPER/index.ts", true);
    } else {
        providerCase("outside sensitive provider with uppercase", "/provider", true, "/provider/node_modules/UPPER/index.ts", true,
            "/provider/node_modules/UPPER/", [["/provider/node_modules/UPPER", "/provider/node_modules/upper"]]);
        providerCase("outside sensitive provider with Unicode", "/provider", true, "/provider/node_modules/Äİ/index.ts", true,
            "/provider/node_modules/Äİ/", [["/provider/node_modules/Äİ", "/provider/node_modules/äİ"]]);
    }
}

observeLivePolicy();

withBeforeTree(scratch => {
    for (const [relative, needle, replacement, diagnostic] of [
        [resolver, "Stops at global cache location", "Stops at the global cache location", /wrapper exact text drift/],
        [services, "    stripQuotes,", "    stripQuotes as unusedStripQuotes,", /imports exact text drift/],
        [services, "ensureTrailingDirectorySeparator(dirPath)", "ensureTrailingDirectorySeparator((dirPath))", /key exact text drift/],
        [resolver, "return getPackageJsonInfo(candidate, /*onlyRecordFailures*/ false, moduleResolutionState);", "return getPackageJsonInfo(candidate, /*onlyRecordFailures*/ true, moduleResolutionState);", /caller text drift/],
    ]) {
        const destination = path.join(scratch, relative);
        const original = fs.readFileSync(destination, "utf8");
        const drift = original.replace(needle, replacement);
        assert.notEqual(drift, original);
        fs.writeFileSync(destination, drift);
        const other = path.join(scratch, relative === resolver ? services : resolver);
        const otherBefore = fs.readFileSync(other);
        assert.deepEqual(recordWrites(() => assert.throws(() => adapt(scratch), diagnostic)), []);
        assert.equal(fs.readFileSync(destination, "utf8"), drift);
        assert.deepEqual(fs.readFileSync(other), otherBefore);
        fs.writeFileSync(destination, original);
        console.log("PASS refuses drift", diagnostic.source);
    }
    const alias = path.join(scratch, "src/compiler/reviewerAlias.ts");
    const source = `import { forEachAncestorDirectoryStoppingAtGlobalCache, ModuleResolutionHost } from "./_namespaces/ts.js";

const ancestor = forEachAncestorDirectoryStoppingAtGlobalCache;

export function extraCaller(host: ModuleResolutionHost, directory: string): string | undefined {
    return ancestor(host, directory, current => current);
}
`;
    const diagnostic = /Unpinned ancestor wrapper reference at src\/compiler\/reviewerAlias\.ts:3:18/;
    function rejectsAlias(side) {
        const before = filesAt(scratch);
        fs.writeFileSync(alias, source);
        assert.deepEqual(recordWrites(() => assert.throws(() => adapt(scratch), diagnostic)), []);
        assert.deepEqual(filesAt(scratch), before);
        assert.equal(fs.readFileSync(alias, "utf8"), source);
        fs.unlinkSync(alias);
        console.log("PASS refuses reviewer alias before writing", side, diagnostic.source);
    }
    rejectsAlias("before");
    adapt(scratch);
    rejectsAlias("replay");
});
