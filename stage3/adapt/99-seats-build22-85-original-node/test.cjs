"use strict";

const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { test } = require("node:test");
const vm = require("node:vm");
const packages = path.resolve(__dirname, "../../api/node_modules");
const ts = require(path.join(packages, "typescript"));

assert.equal(process.argv.length, 3, "usage: node test.cjs <pristine TypeScript v6.0.3 checkout>");
assert.equal(ts.version, "6.0.3");
const pristine = path.resolve(process.argv[2]);
const pin = require("../../source.json");
assert.equal(execFileSync("git", ["-C", pristine, "rev-parse", "HEAD"], { encoding: "utf8" }).trim(), pin.commit);
const adapters = [
    path.join(__dirname, "../99-seats-build22-85-change-extension"),
    __dirname,
];
const files = new Set();
for (const adapter of adapters) {
    const pins = require(path.join(adapter, "pins.json"));
    files.add(pins.owner);
    for (const file of pins.files) files.add(file.file);
}

function evaluate(tree) {
    const sources = {
        "core.ts": ["equateValues", "equateStringsCaseSensitive", "equateStringsCaseInsensitive", "startsWith"],
        "types.ts": ["CharacterCodes", "Extension"],
        "path.ts": ["isAnyDirectorySeparator", "hasTrailingDirectorySeparator", "removeTrailingDirectorySeparator",
            "tryGetExtensionFromPath", "getAnyExtensionFromPathWorker", "getAnyExtensionFromPath", "changeAnyExtension"],
        "utilities.ts": ["extensionsToRemove", "changeExtension", "changeExtensionString"],
        "utilitiesPublic.ts": ["getOriginalNode", "getOriginalNodeWithoutPredicate"],
        "checker.ts": ["nextNodeId", "getNodeId"],
        "transformers/utilities.ts": ["getOriginalNodeId"],
    };
    const declarations = [];
    for (const [file, names] of Object.entries(sources)) {
        const fileName = path.join(tree, "src/compiler", file);
        const source = ts.createSourceFile(fileName, fs.readFileSync(fileName, "utf8"), ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, file);
        for (const name of names) {
            const matches = source.statements.filter(node =>
                ((ts.isFunctionDeclaration(node) && node.body) || ts.isEnumDeclaration(node)) && node.name?.text === name
                || ts.isVariableStatement(node) && node.declarationList.declarations.some(binding => binding.name.getText(source) === name));
            assert.equal(matches.length, 1, file + ": " + name);
            declarations.push(matches[0].getText(source).replace(/^export /, ""));
        }
    }
    const result = ts.transpileModule(declarations.join("\n"), { compilerOptions: { target: ts.ScriptTarget.ES2020 } });
    return vm.runInNewContext(result.outputText + "\n({ changeExtension, changeExtensionString, getOriginalNode, "
        + "getOriginalNodeWithoutPredicate, getOriginalNodeId, getNodeId });");
}

const tree = fs.mkdtempSync(path.join(os.tmpdir(), "family85-witness-"));
let tsHelpers;
try {
    for (const file of files) {
        const destination = path.join(tree, file);
        fs.mkdirSync(path.dirname(destination), { recursive: true });
        fs.copyFileSync(path.join(pristine, file), destination);
    }
    for (const file of ["core.ts", "types.ts", "path.ts"]) {
        fs.copyFileSync(path.join(pristine, "src/compiler", file), path.join(tree, "src/compiler", file));
    }
    for (const adapter of adapters) {
        execFileSync(process.execPath, [path.join(adapter, "adapt.cjs"), tree], {
            env: { ...process.env, NODE_PATH: packages }, stdio: "inherit",
        });
    }
    tsHelpers = evaluate(tree);
} finally {
    fs.rmSync(tree, { recursive: true, force: true });
}

test("getOriginalNodeWithoutPredicate follows every link and preserves undefined", () => {
    const root = { id: 1, original: undefined };
    const middle = { id: 2, original: root };
    const leaf = { id: 3, original: middle };
    assert.strictEqual(tsHelpers.getOriginalNode(leaf), root);
    assert.strictEqual(tsHelpers.getOriginalNodeWithoutPredicate(root), root);
    assert.strictEqual(tsHelpers.getOriginalNodeWithoutPredicate(undefined), undefined);
    assert.strictEqual(tsHelpers.getOriginalNodeWithoutPredicate(leaf), root);
});

test("getOriginalNodeId uses the original node through a two-hop transform chain", () => {
    const root = { id: 1 };
    const middle = { id: 2, original: root };
    const leaf = { id: 3, original: middle };
    assert.equal(tsHelpers.getOriginalNodeId(leaf), tsHelpers.getNodeId(root));
});

test("changeExtensionString agrees with changeExtension on 18 representative path/extension pairs", () => {
    for (const input of ["file.ts", "file.d.ts", "file.mts", "file.TS", "file", "dir/.file.ts"]) {
        for (const extension of [".js", ".d.ts", ""]) {
            assert.strictEqual(tsHelpers.changeExtensionString(input, extension), tsHelpers.changeExtension(input, extension),
                JSON.stringify([input, extension]));
        }
    }
});
