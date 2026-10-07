"use strict";
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const [pristine, tree, ledger] = process.argv.slice(2);
assert(pristine && tree && ledger, "usage: api-mutants pristine tree ledger");
const check = path.join(__dirname, "check-api-baselines.cjs");
const mutants = [];
function killed(name, file, mutate, message) {
    const original = fs.readFileSync(file);
    try {
        const changed = mutate(original.toString("utf8"));
        assert.notEqual(changed, original.toString("utf8"));
        fs.writeFileSync(file, changed);
        const result = spawnSync(process.execPath, [check, pristine, tree, ledger], { encoding: "utf8" });
        assert.equal(result.status, 1, name + ": not caught");
        assert.match(result.stderr, message);
        mutants.push({ name, exit: result.status, caught: result.stderr.match(/(?:Error|AssertionError[^:]*):[^\n]*/)?.[0] || message.source });
    } finally { fs.writeFileSync(file, original); }
    assert.deepEqual(fs.readFileSync(file), original, name + ": restore bytes");
}
const api = path.join(tree, "built/local/typescript.d.ts");
killed("public brand token restored to any", api, text => text.replace("__pathBrand: undefined;", "__pathBrand: any;"), /API snapshot differs from parsed owner edits/);
killed("ErrorCallback payload changed to boolean", api, text => text.replace("arg0?: string | number", "arg0?: boolean"), /API snapshot differs from parsed owner edits/);
killed("unlisted public API reference declaration", path.join(tree, "tests/baselines/reference/api/typescript.d.ts"), text => text + "\ntype Unreviewed = string;\n", /API reference was changed without this proof/);
const root = path.join(tree, "tests/baselines/reference");
function firstOther(directory) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true }).sort((a,b) => a.name.localeCompare(b.name))) {
        const file = path.join(directory, entry.name);
        if (entry.isDirectory()) { const found = firstOther(file); if (found) return found; }
        else if (entry.isFile() && path.relative(root, file) !== "api/typescript.d.ts") return file;
    }
}
killed("unrelated reference baseline byte", firstOther(root), text => text + "\n", /unsanctioned reference baseline edit/);
const rerun = spawnSync(process.execPath, [check, pristine, tree, ledger], { encoding: "utf8" });
assert.equal(rerun.status, 0, rerun.stderr);
console.log(JSON.stringify({ status: "pass", mutants, restored_check: "pass" }, null, 2));
