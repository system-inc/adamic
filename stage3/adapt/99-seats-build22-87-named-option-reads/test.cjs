"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");
const ts = require("typescript");

const adapter = path.join(__dirname, "adapt.cjs");
const snapshot = "refs/stage3/99-seats-build22-87-named-option-reads-before";
const resolver = "src/compiler/moduleNameResolver.ts";
const helpers = ["tryLoadModuleUsingPathsIfEligible", "tryLoadModuleUsingBaseUrl"];
const sha256 = bytes => createHash("sha256").update(bytes).digest("hex");

function exercise(text, label) {
    const source = ts.createSourceFile(resolver, text, ts.ScriptTarget.Latest, true);
    const declarations = helpers.map(name => {
        const matches = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
        assert.equal(matches.length, 1, name);
        return matches[0].getText(source);
    });
    const { outputText, diagnostics } = ts.transpileModule(declarations.join("\n"), {
        compilerOptions: { target: ts.ScriptTarget.ES2020 },
        reportDiagnostics: true,
    });
    assert.deepEqual(diagnostics, []);

    const mappings = [];
    const traces = [];
    const context = vm.createContext({
        pathIsRelative: name => name.startsWith("."),
        getPathsBasePath: options => options.baseUrl ?? "/fallback",
        tryParsePatterns: paths => Object.keys(paths),
        tryLoadModuleUsingPaths: (_extensions, name, base, paths) => {
            mappings.push(paths);
            return paths[name]?.[0] ? base + "/" + paths[name][0] : undefined;
        },
        normalizePath: value => value,
        combinePaths: (base, suffix) => base + "/" + suffix,
        directoryProbablyExists: () => true,
        getDirectoryPath: value => value,
        trace: (_host, message) => traces.push(message),
        Diagnostics: {
            baseUrl_option_is_set_to_0_using_this_value_to_resolve_non_relative_module_name_1: "base",
            paths_option_is_specified_looking_for_a_pattern_to_match_module_name_0: "paths",
            Resolving_module_name_0_relative_to_base_url_1_2: "resolve",
        },
    });
    const [pathsHelper, baseHelper] = vm.runInContext(outputText + "\n[" + helpers.join(",") + "]", context);
    const runtimeTypeError = vm.runInContext("TypeError", context);
    const loader = (_extensions, candidate) => candidate;
    const results = [];
    function expect(helper, name, state, expected, description) {
        const result = helper(0, name, loader, state);
        assert.equal(result, expected, `${label}: ${description}`);
        results.push(result);
    }

    const paths = { pkg: ["mapped.ts"] };
    const state = { compilerOptions: { baseUrl: "/base", paths }, traceEnabled: true };
    expect(pathsHelper, "pkg", state, "/base/mapped.ts", "ordinary paths mapping survives");
    assert.equal(mappings[0], state.compilerOptions.paths, `${label}: ordinary paths object identity`);
    assert.deepEqual(traces, ["base", "paths"]);

    const fallback = { compilerOptions: { paths } };
    expect(pathsHelper, "pkg", fallback, "/fallback/mapped.ts", "paths use the fallback base");
    assert.equal(mappings[1], fallback.compilerOptions.paths, `${label}: fallback paths object identity`);
    const withoutPaths = { compilerOptions: { baseUrl: "/base" } };
    const withoutOptions = { compilerOptions: {} };
    expect(pathsHelper, "pkg", withoutPaths, undefined, "base URL without paths skips the paths loader");
    expect(pathsHelper, "pkg", withoutOptions, undefined, "absent options skip the paths loader");
    expect(pathsHelper, "./pkg", state, undefined, "relative names skip the paths loader");
    assert.equal(mappings.length, 2);
    assert.deepEqual(traces, ["base", "paths"]);

    expect(baseHelper, "pkg", withoutPaths, "/base/pkg", "ordinary base URL resolution");
    expect(baseHelper, "pkg", withoutOptions, undefined, "absent base URL skips the base loader");
    expect(baseHelper, "pkg", { compilerOptions: { baseUrl: "" } }, undefined, "empty base URL skips the base loader");

    const order = [];
    const tracedReceiver = {
        get compilerOptions() {
            order.push("receiver");
            return {
                get baseUrl() {
                    order.push("baseUrl");
                    return "/base";
                },
                get paths() {
                    order.push("paths");
                    return paths;
                },
            };
        },
    };
    // A relative name isolates the declaration from getPathsBasePath's later read.
    expect(pathsHelper, "./pkg", tracedReceiver, undefined, "paths helper reads in pattern order");
    assert.deepEqual(order.splice(0), ["receiver", "baseUrl", "paths"]);
    expect(baseHelper, "pkg", tracedReceiver, "/base/pkg", "base helper reads the receiver and base URL once");
    assert.deepEqual(order, ["receiver", "baseUrl"]);

    for (const helper of [pathsHelper, baseHelper]) {
        assert.throws(() => helper(0, "pkg", loader, { compilerOptions: undefined }), runtimeTypeError);
    }
    console.log(`${label}: resolution, paths identity, absent values, traces, and read order passed`);
    return results;
}

const tree = process.argv[2];
assert.ok(tree, "usage: node test.cjs <cold adapted tree>");
const input = execFileSync("git", ["-C", tree, "show", `${snapshot}:${resolver}`]);
console.log(`Snapshot SHA-256: ${sha256(input)}`);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-family87-"));
try {
    const file = path.join(scratch, resolver);
    fs.mkdirSync(path.dirname(file), { recursive: true });
    fs.writeFileSync(file, input);
    const apply = () => spawnSync(process.execPath, [adapter, scratch], { encoding: "utf8" });
    const applied = apply();
    assert.ifError(applied.error);
    assert.equal(applied.status, 0, applied.stderr);

    const output = fs.readFileSync(file);
    const originalResults = exercise(input.toString("utf8"), "snapshot");
    const adaptedResults = exercise(output.toString("utf8"), "adapted");
    assert.deepEqual(adaptedResults, originalResults);

    const repeated = apply();
    assert.ifError(repeated.error);
    assert.equal(typeof repeated.status, "number");
    assert.notEqual(repeated.status, 0);
    assert.match(repeated.stderr, /heldCompilerOptions must be fresh/);
    assert.equal(sha256(fs.readFileSync(file)), sha256(output));
    console.log(`Direct reapplication exited ${repeated.status} and preserved SHA-256 ${sha256(output)}`);
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
