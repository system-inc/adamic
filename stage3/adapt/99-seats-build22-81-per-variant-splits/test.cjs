"use strict";

const assert = require("node:assert/strict");
const { execFileSync, spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const vm = require("node:vm");

const packages = path.resolve(__dirname, "../../api/node_modules");
const ts = require(path.join(packages, "typescript"));
assert.equal(ts.version, "6.0.3");
assert(process.argv.length === 3 || process.argv.length === 4, "usage: node test.cjs <applied tree> [adapter]");
const applied = path.resolve(process.argv[2]);
const adapter = path.resolve(process.argv[3] ?? path.join(__dirname, "adapt.cjs"));
const snapshot = "refs/stage3/99-seats-build22-81-per-variant-splits";
const builder = "src/compiler/builder.ts";
const utilities = "src/compiler/transformers/utilities.ts";
const files = [builder, utilities];

function readSnapshot(ref, file) {
    return execFileSync("git", ["-C", applied, "show", snapshot + "-" + ref + ":" + file]);
}

function apply(tree) {
    const result = spawnSync(process.execPath, [adapter, tree], {
        env: { ...process.env, NODE_PATH: packages },
        encoding: "utf8",
        timeout: 30_000,
    });
    if (result.error) throw result.error;
    assert.equal(result.status, 0, result.stderr || result.stdout || result.signal);
}

function digest(bytes) {
    return createHash("sha256").update(bytes).digest("hex");
}

function load(text, file, names, overrides = {}) {
    const source = ts.createSourceFile(file, text.toString("utf8"), ts.ScriptTarget.ES2020, true);
    assert.deepEqual(source.parseDiagnostics, []);
    const declarations = names.map(name => {
        const matches = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
        assert.equal(matches.length, 1, name);
        return matches[0].getText(source).replace(/^export /, "");
    });
    const code = ts.transpileModule([...declarations, names[0]].join("\n"), {
        compilerOptions: { target: ts.ScriptTarget.ES2020 },
        reportDiagnostics: true,
    });
    assert.deepEqual(code.diagnostics, []);
    return vm.runInNewContext(code.outputText, { ...ts, ...overrides });
}

function method(text, bundle, empty) {
    const visits = [];
    const reads = [];
    let callbackIndex;
    let traversals = 0;
    const fileInfos = empty ? [] : ["v1", , {
        get version() {
            assert.equal(callbackIndex, 2);
            reads.push(["version", callbackIndex]);
            return "v3";
        },
    }, "v4"];
    const fileNames = ["a.ts", "hole.ts", "b.ts", "c.ts"];
    for (const index of [0, 2, 3]) {
        const name = fileNames[index];
        Object.defineProperty(fileNames, index, {
            get() {
                assert.equal(callbackIndex, index);
                reads.push(["path", index]);
                return name;
            },
        });
    }
    const program = {
        options: bundle ? { outFile: "bundle.js" } : {},
        fileNames,
        fileInfos,
        root: [1, [3, 4]],
        resolvedRoot: [[3, 2]],
    };
    Object.defineProperty(fileInfos, "forEach", {
        value(callback) {
            assert.equal(this, fileInfos, "forEach must use the original file-info array");
            traversals++;
            Array.prototype.forEach.call(this, (value, index, receiver) => {
                visits.push(index);
                callbackIndex = index;
                callback(value, index, receiver);
                callbackIndex = undefined;
            });
        },
    });
    const map = load(text, builder, ["getBuildInfoFileVersionMap", "isIncrementalBundleEmitBuildInfo"], {
        isString(value) {
            const stringVersion = ts.isString(value);
            // A string has no getter. Its type check selects the version value.
            if (stringVersion) reads.push(["version", callbackIndex]);
            return stringVersion;
        },
    });
    const result = map(program, "/project/build.tsbuildinfo", {
        getCurrentDirectory: () => "/project",
        useCaseSensitiveFileNames: () => true,
    });
    assert.equal(program.fileInfos, fileInfos, "program.fileInfos must still hold the original file-info array");
    // Create the tuples here because the returned maps belong to the VM realm.
    const entries = map => Array.from(map, ([key, value]) => [key, value]);
    return { traversals, visits, reads, fileInfos: entries(result.fileInfos), roots: entries(result.roots) };
}

function checkMethods(before, after) {
    const roots = [
        ["/project/a.ts", undefined],
        ["/project/hole.ts", "/project/b.ts"],
        ["/project/c.ts", undefined],
    ];
    const fileInfos = [
        ["/project/a.ts", "v1"],
        ["/project/b.ts", "v3"],
        ["/project/c.ts", "v4"],
    ];
    for (const bundle of [true, false]) {
        const variant = bundle ? "bundle" : "multifile";
        for (const empty of [false, true]) {
            const description = empty ? variant + " empty" : variant + " sparse";
            const original = method(before, bundle, empty);
            const adapted = method(after, bundle, empty);
            const indices = empty ? [] : [0, 2, 3];
            assert.equal(original.traversals, 1, description);
            assert.deepEqual(original.visits, indices, description);
            assert.deepEqual(original.reads, indices.flatMap(index => [["path", index], ["version", index]]), description);
            assert.deepEqual(original.fileInfos, empty ? [] : fileInfos, description);
            assert.deepEqual(original.roots, empty ? [] : roots, description);
            assert.deepEqual(adapted.roots, original.roots, description + " roots comparison must use the forEach index");
            assert.deepEqual(adapted, original, description + " file versions and callback observations");
            console.log("PASS: " + description + " method comparison, indices [" + adapted.visits + "], " + adapted.roots.length + " roots");
        }
    }
}

function iteration(text) {
    const collect = load(text, utilities, ["collectExportedVariableInfo", "multiMapSparseArrayAdd"]);
    const file = ts.createSourceFile("probe.ts",
        "const { first, split: [, inside, ,], second: {last}, dup: first } = input; " +
        "const [, tail, , {end},] = input;",
        ts.ScriptTarget.ES2020, true);
    assert.deepEqual(file.parseDiagnostics, []);
    const declarations = file.statements.flatMap(statement => Array.from(statement.declarationList.declarations));
    const arrays = [];
    const visits = [];
    const metadata = elements => [elements.pos, elements.end, elements.hasTrailingComma, elements.transformFlags];

    function instrument(node) {
        if (ts.isBindingPattern(node)) {
            const elements = node.elements;
            arrays.push({ node, elements, before: metadata(elements) });
        }
        if (ts.isIdentifier(node) && node.text === "inside") ts.setEmitFlags(node, ts.EmitFlags.LocalName);
        ts.forEachChild(node, instrument);
    }

    const unique = new Map();
    const bindings = [];
    let names;
    for (const declaration of declarations) instrument(declaration);
    for (const [index, { node, elements }] of arrays.entries()) {
        const iterator = elements[Symbol.iterator];
        Object.defineProperty(elements, Symbol.iterator, {
            value() {
                assert.equal(this, elements);
                assert.equal(this, node.elements);
                visits.push(index);
                return iterator.call(this);
            },
        });
    }
    for (const declaration of declarations) names = collect(declaration, unique, names, bindings);
    for (const { node, elements, before } of arrays) {
        assert.equal(node.elements, elements);
        assert.deepEqual(metadata(elements), before);
    }
    return {
        visits,
        metadata: arrays.map(({ before }) => before),
        names: Array.from(names, identifier => identifier.text),
        unique: Array.from(unique.keys()),
        bindings: Object.values(bindings).map(list => Array.from(list, identifier => identifier.text)),
    };
}

function checkIteration(before, after) {
    const original = iteration(before);
    assert.deepEqual(original.visits, [0, 1, 2, 3, 4]);
    assert.deepEqual(original.names, ["first", "inside", "last", "tail", "end"]);
    assert.deepEqual(original.unique, original.names);
    assert.deepEqual(original.bindings, [["inside"]]);
    let adapted;
    assert.doesNotThrow(() => { adapted = iteration(after); }, "iteration comparison: adapted binding walk must handle omitted elements");
    assert.deepEqual(adapted, original, "iteration comparison: names, bindings and original NodeArray walks");
    console.log("PASS: iteration comparison, five NodeArray walks with metadata, " + adapted.names.join(", "));
}

const tree = fs.mkdtempSync(path.join(os.tmpdir(), "per-variant-splits-"));
try {
    const original = files.map(file => {
        const bytes = readSnapshot("before", file);
        const target = path.join(tree, file);
        fs.mkdirSync(path.dirname(target), { recursive: true });
        fs.writeFileSync(target, bytes);
        return bytes;
    });
    apply(tree);
    const adapted = files.map((file, index) => {
        const bytes = fs.readFileSync(path.join(tree, file));
        assert(!bytes.equals(original[index]), file + " did not change on the first apply");
        return bytes;
    });
    const hashes = adapted.map(digest);
    apply(tree);
    for (const [index, file] of files.entries()) {
        const repeated = fs.readFileSync(path.join(tree, file));
        assert.equal(digest(repeated), hashes[index], file + " second-apply hash");
        assert(repeated.equals(adapted[index]), file + " second-apply bytes");
        console.log("PASS: " + file + " changed on first apply and held SHA-256 " + hashes[index] + " on repeat");
    }
    checkMethods(original[0], adapted[0]);
    checkIteration(original[1], adapted[1]);
    for (const [index, file] of files.entries()) {
        assert(adapted[index].equals(readSnapshot("after", file)), file + " must match the pipeline's after snapshot");
        console.log("PASS: " + file + " matches the pipeline's after snapshot");
    }
}
finally {
    fs.rmSync(tree, { recursive: true, force: true });
}
