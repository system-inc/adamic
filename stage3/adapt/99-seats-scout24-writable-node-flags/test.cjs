"use strict";

const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

assert.equal(process.argv.length, 3, "usage: test.cjs APPLIED_TREE");
const applied = path.resolve(process.argv[2]);
const adapter = path.join(__dirname, "adapt.cjs");
const ref = "refs/stage3/99-seats-scout24-writable-node-flags";
const sixth = "src/compiler/zzFlagsForm.ts";
const edited = "src/compiler/utilities.ts";

function run(command, args, options) {
    const result = spawnSync(command, args, { encoding: "utf8", timeout: 60000, ...options });
    if (result.error) throw result.error;
    if (result.status === null) throw new Error(`${command} stopped by ${result.signal}`);
    return result;
}

function extract(root, side) {
    const dir = path.join(root, side);
    const archive = path.join(root, `${side}.tar`);
    fs.mkdirSync(dir);
    const exported = run("git", ["-C", applied, "archive", "--output=" + archive, `${ref}-${side}`, "src"]);
    assert.equal(exported.status, 0, `git archive ${side}`);
    const unpacked = run("tar", ["-xf", archive, "-C", dir]);
    assert.equal(unpacked.status, 0, `tar ${side}`);
    fs.unlinkSync(archive);
    return dir;
}

function hashes(tree) {
    const entries = fs.readdirSync(path.join(tree, "src"), { recursive: true, withFileTypes: true });
    const files = entries.filter(entry => entry.isFile());
    return new Map(files.map(entry => {
        const file = path.join(entry.parentPath, entry.name);
        const digest = crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex");
        return [path.relative(tree, file), digest];
    }));
}

function equalHashes(actual, expected) {
    for (const [file, digest] of expected) {
        if (actual.get(file) !== digest) throw new Error(`changed or missing file ${file}`);
    }
    for (const file of actual.keys()) {
        if (!expected.has(file)) throw new Error(`unexpected file ${file}`);
    }
}

function checkControl(result, declarations, casts) {
    assert.equal(result.status, 0, "adapter exit");
    const { declarationEdits, castEdits } = JSON.parse(result.stdout);
    assert.equal(declarationEdits, declarations, "declaration edits");
    assert.equal(castEdits, casts, "cast edits");
}

function appendFixture(tree, file, expression, setup, qualified, offset) {
    const target = path.join(tree, file);
    let prefix = "";
    if (file === edited) {
        const existing = fs.readFileSync(target, "utf8");
        const imports = existing.match(/import \{[\s\S]*?\} from "\.\/_namespaces\/ts\.js";/);
        assert(imports && /^    type Node,\r?$/m.test(imports[0]), "utilities.ts must import Node");
        assert(/^export type Mutable<T extends object> = \{ -readonly \[K in keyof T\]: T\[K\]; \};\r?$/m.test(existing),
            "utilities.ts must declare Mutable");
        prefix = existing + "\n";
    } else {
        prefix = 'import { Mutable, Node } from "./_namespaces/ts.js";\n';
    }
    if (qualified) prefix += 'import * as Scope from "./_namespaces/ts.js";\n';
    const lines = [
        "export function zzFlagsForm(node: Node): void {",
        ...setup,
        `    ${expression} |= 1;`,
        "}",
        "",
    ];
    fs.writeFileSync(target, prefix + lines.join("\n"));
    const line = prefix.split("\n").length + 1 + setup.length;
    return `${file}:${line}:${5 + offset}`;
}

const unexpected = "unexpected Node flags write";
const unresolved = "unexpected unresolved flags write";
const keyed = "unexpected keyed flags write";
const forms = [
    ["qualified", "(node as Scope.Mutable<typeof node>).flags", []],
    ["angle-bracket", "(<Mutable<typeof node>>node).flags", []],
    ["extra-parentheses", "((node as Mutable<typeof node>)).flags", []],
    ["alias", "view.flags", ["    const view = node as Mutable<typeof node>;"]],
];
const probes = [
    ["type alias", "(node as M).flags", ["    type M = Mutable<Node>;"]],
    ["conditional receiver", "(pick ? (node as Mutable<Node>) : node).flags", ["    const pick = node.pos > 0;"]],
    ["call receiver", "id(node as Mutable<Node>).flags", ["    function id<T>(value: T): T { return value; }"]],
    ["property receiver", "holder.view.flags", ["    const holder = { view: node as Mutable<Node> };"]],
    ["alias of alias", "copy.flags", ["    const view = node as Mutable<Node>;", "    const copy = view;"]],
    ["assigned alias", "view.flags", ["    let view;", "    view = node as Mutable<Node>;"]],
    ["non-null initializer", "view.flags", ["    const view = (node as Mutable<Node>)!;"]],
];
const edges = [
    ["template key", '(node as Mutable<Node>)[`flags`]', unexpected, []],
    ["any receiver", "(node as any).flags", unresolved, []],
    ["unresolved name", "missingFlagsHolder.flags", unresolved, []],
    ["literal-typed key", "(node as Mutable<Node>)[key]", keyed, ['    const key: "flags" = "flags";']],
    ["string key", "(node as Mutable<Node>)[key]", keyed, ['    const key: string = "flags";']],
];

const factory = "src/compiler/factory/nodeFactory.ts";
const optionalChain = "node.flags |= NodeFlags.OptionalChain;";
const cloneFlags = "clone.flags |= node.flags & ~NodeFlags.Synthesized;";
function editPinned(tree, file, owner, statement, edit, kind = unexpected) {
    const target = path.join(tree, file);
    const text = fs.readFileSync(target, "utf8");
    const header = owner === undefined ? statement : `function ${owner}(`;
    const start = text.indexOf(header);
    assert(start !== -1 && text.indexOf(header, start + header.length) === -1, `expected one ${header}`);
    const at = owner === undefined ? start : text.indexOf(statement, start + header.length);
    assert(at !== -1, `missing pinned statement in ${owner}`);
    const lineStart = text.lastIndexOf("\n", at) + 1;
    const newline = text.indexOf("\n", at);
    const lineEnd = newline + 1;
    assert(newline !== -1, `missing newline in ${owner}`);
    const [modified, writeAt] = edit(text, at, lineStart, lineEnd);
    fs.writeFileSync(target, modified);
    if (writeAt === undefined) return `missing pinned Node flags write in ${file}`;
    const line = modified.slice(0, writeAt).split("\n").length;
    const column = writeAt - modified.lastIndexOf("\n", writeAt);
    return `${kind} at ${file}:${line}:${column}`;
}

const root = fs.mkdtempSync(path.join(os.tmpdir(), "adamic-flags-forms-"));
let failed = 0;
try {
    const snapshots = { before: extract(root, "before"), after: extract(root, "after") };
    const afterHashes = hashes(snapshots.after);
    function row(name, side, check) {
        const scratch = fs.mkdtempSync(path.join(root, "row-"));
        try {
            fs.cpSync(path.join(snapshots[side], "src"), path.join(scratch, "src"), { recursive: true });
            check(scratch, () => run(process.execPath, [adapter, scratch]));
            console.log(`${name} ok`);
        } catch (error) {
            failed++;
            console.log(`${name} failed: ${error.message.split("\n")[0]}`);
        } finally {
            fs.rmSync(scratch, { recursive: true, force: true });
        }
    }
    row("control first apply", "before", (tree, adapt) => {
        checkControl(adapt(), 1, 20);
        equalHashes(hashes(tree), afterHashes);
    });
    row("control replay", "after", (tree, adapt) => {
        checkControl(adapt(), 0, 0);
        equalHashes(hashes(tree), afterHashes);
    });
    function checkRefusal(tree, adapt, expected) {
        const before = hashes(tree);
        const result = adapt();
        equalHashes(hashes(tree), before);
        assert.notEqual(result.status, 0, "expected refusal");
        const message = result.stderr.split("\n").find(line => line.startsWith("AssertionError"));
        const line = `AssertionError [ERR_ASSERTION]: ${expected}`;
        if (message !== line) throw new Error(`expected ${line}; got ${message ?? "no assertion"}`);
    }
    function refusal(name, side, file, expression, kind, setup, offset = 0) {
        row(`${name} ${side} ${file}`, side, (tree, adapt) => {
            const location = appendFixture(tree, file, expression, setup, name === "qualified", offset);
            checkRefusal(tree, adapt, `${kind} at ${location}`);
        });
    }
    for (const side of ["before", "after"]) {
        refusal("direct", side, sixth, "(node as Mutable<typeof node>).flags", unexpected, []);
        for (const [name, expression, setup] of forms) {
            for (const file of [sixth, edited]) refusal(name, side, file, expression, unexpected, setup);
        }
    }
    for (const side of ["before", "after"]) {
        function pinned(name, file, owner, statement, edit) {
            row(`${name} ${side}`, side, (tree, adapt) => {
                checkRefusal(tree, adapt, editPinned(tree, file, owner, statement, edit));
            });
        }
        pinned("alias initializer swapped", factory, "createPropertyAccessChain", optionalChain, (text, at, lineStart) => {
            const start = text.indexOf("const node = ", text.indexOf("function createPropertyAccessChain("));
            const end = text.indexOf(";", start) + 1;
            assert(start >= 0 && end > start && end < lineStart, "missing alias initializer");
            const replacement = "const node = expression as unknown as Mutable<PropertyAccessChain>;";
            return [text.slice(0, start) + replacement + text.slice(end), at + replacement.length - (end - start)];
        });
        pinned("final alias write removed", factory, "cloneGeneratedPrivateIdentifier", cloneFlags, (text, _, start, end) =>
            [text.slice(0, start) + text.slice(end)]);
        pinned("ninth alias write", factory, "cloneGeneratedPrivateIdentifier", cloneFlags, (text, at, start, end) =>
            [text.slice(0, end) + text.slice(start, end) + text.slice(end), end + at - start]);
        pinned("pinned alias write changed", factory, "createPropertyAccessChain", optionalChain, (text, at) =>
            [text.slice(0, at) + "node.flags |= NodeFlags.Synthesized;" + text.slice(at + optionalChain.length), at]);
        for (const [name, expression, setup] of probes) refusal(name, side, sixth, expression, unexpected, setup);
        pinned("this write changed", edited, "Token", "this.flags = NodeFlags.None;", (text, at) =>
            [text.slice(0, at) + "this.flags = NodeFlags.Synthesized;" + text.slice(at + "this.flags = NodeFlags.None;".length), at]);
        pinned("direct write in another function", factory, "createPropertyAccessChain", optionalChain, (text, at, start, end) => {
            const receiver = side === "before" ? "(node as Mutable<Node>)" : "node";
            const indent = text.slice(start, at);
            const eol = text.slice(at + optionalChain.length, end);
            return [text.slice(0, end) + indent + receiver + ".flags |= NodeFlags.Synthesized;" + eol +
                text.slice(end), end + indent.length];
        });
        for (const [name, expression, kind, setup] of edges) refusal(name, side, sixth, expression, kind, setup);
    }
    for (const side of ["before", "after"]) {
        row(`generic key ${side} ${sixth}`, side, (tree, adapt) => {
            fs.writeFileSync(path.join(tree, sixth),
                'import type { Node, Mutable } from "./_namespaces/ts.js";\n' +
                'function probe<K extends keyof Node>(node: Mutable<Node>, key: K, value: Mutable<Node>[K]) { node[key] = value; }\n');
            checkRefusal(tree, adapt, `${keyed} at ${sixth}:2:94`);
        });

        const cloneKey = "clone[key] = node[key];";
        function keyedPin(name, edit) {
            row(`${name} ${side}`, side, (tree, adapt) => {
                checkRefusal(tree, adapt, editPinned(tree, factory, undefined, cloneKey, edit, keyed));
            });
        }
        keyedPin("keyed write duplicated", (text, at, start, end) =>
            [text.slice(0, end) + text.slice(start, end) + text.slice(end), end + at - start]);
        keyedPin("keyed write removed", (text, _, start, end) =>
            [text.slice(0, start) + text.slice(end)]);
        keyedPin("keyed write moved into an arrow", (text, at) => {
            const prefix = "const copyKey = () => { ";
            return [text.slice(0, at) + prefix + cloneKey + " };" + text.slice(at + cloneKey.length), at + prefix.length];
        });

        for (const [name, expression] of [
            ["as target", "(node.flags as number)"],
            ["angle target", "(<any>node.flags)"],
            ["satisfies target", "(node.flags satisfies number)"],
            ["non-null target", "node.flags!"],
            ["parenthesized target", "(node.flags)"],
        ]) {
            refusal(name, side, sixth, expression, unexpected, [], expression.indexOf("node.flags"));
        }

        for (const [name, prefix, suffix] of [
            ["an arrow", "const deferFlags = () => { ", " };"],
            ["a function expression", "const deferFlags = function () { ", " };"],
            ["a nested named function", "function deferFlags() { ", " }"],
            ["a class static block", "class DeferFlags { static { ", " } }"],
            ["a class field initializer", "class DeferFlags { deferred = ", " }"],
        ]) {
            row(`pinned write moved into ${name} ${side}`, side, (tree, adapt) => {
                checkRefusal(tree, adapt, editPinned(tree, factory, "createPropertyAccessChain", optionalChain,
                    (text, at) => [text.slice(0, at) + prefix + optionalChain + suffix +
                        text.slice(at + optionalChain.length), at + prefix.length]));
            });
        }

        refusal("template-literal key", side, sixth, "(node as Mutable<Node>)[key]", keyed,
            ['    const key = "flags" as `fl${string}`;']);
        refusal("string-mapping key", side, sixth, "(node as Mutable<Node>)[key]", keyed,
            ['    const key = "flags" as Lowercase<string>;']);
    }

} finally {
    fs.rmSync(root, { recursive: true, force: true });
}
if (failed) process.exitCode = 1;
