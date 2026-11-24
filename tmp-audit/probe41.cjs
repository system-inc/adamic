"use strict";

// Correctness's probes of family 41's reflection check, run on a tree stage3/apply.sh wrote. Each probe adds lines to
// types.ts in a fixture of the -before ref's src/, none naming commentPragmas, runs the head's adapter and prints
// whether it refused or accepted. Then a census over the before snapshot's src/: every use of a namespace import of a
// _namespaces/ts.js or types.js barrel that isn't a property access by name or a type position.
const { execFileSync, spawnSync } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const ts = require("typescript");

const tree = path.resolve(process.argv[2]);
const adapter = path.resolve(process.argv[3] || path.join(__dirname, "../stage3/adapt/99-seats-build22-41-pragma-arg-regexes/adapt.cjs"));
const ref = "refs/stage3/99-seats-build22-41-pragma-arg-regexes-before";
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "probe41-"));
const base = path.join(tmp, "base");
fs.mkdirSync(base);
execFileSync("git", ["-C", tree, "archive", "--format=tar", `--output=${tmp}/src.tar`, ref, "src"]);
execFileSync("tar", ["-xf", `${tmp}/src.tar`, "-C", base]);
const typesFile = "src/compiler/types.ts";
const typesText = fs.readFileSync(path.join(base, typesFile), "utf8");
const newline = typesText.includes("\r\n") ? "\r\n" : "\n";
const anchor = "export type IntroducesNewScopeNode = SignatureDeclaration | JSDocSignature | MappedTypeNode;";

const factoryArgs = '{ args: [...(entry as { reference: { args: object[] } }).reference.args, { name: "factory", optional: true }] }';
const probes = [
    ["R1 an alias of the namespace, then a computed key", [
        'import * as ts from "./_namespaces/ts.js";',
        "const namespaceAlias = ts;",
        "export function writeThroughAlias(key: keyof typeof ts) { Object.assign(namespaceAlias[key] as object, { reference: undefined }); }"]],
    ["R2 Object.values over the namespace, appending an eighth XML argument", [
        'import * as ts from "./_namespaces/ts.js";',
        'export function appendThroughValues() { for (const entry of Object.values(ts)) if (entry && typeof entry === "object" && "reference" in entry && "jsx" in entry) Object.assign((entry as { reference: object }).reference, ' + factoryArgs + "); }"]],
    ["R3 the namespace passed to a function that reflects on it", [
        'import * as ts from "./_namespaces/ts.js";',
        "function reflectOn(target: object, key: string) { Object.assign((target as Record<string, object>)[key], { reference: undefined }); }",
        "export function writeThroughCall(key: string) { reflectOn(ts, key); }"]],
    ["R4 Reflect.get with a computed key", [
        'import * as ts from "./_namespaces/ts.js";',
        "export function writeThroughReflect(key: string) { Object.assign(Reflect.get(ts, key) as object, { reference: undefined }); }"]],
    ["R5 a spread copy's entries (reads the same objects)", [
        'import * as ts from "./_namespaces/ts.js";',
        "export function writeThroughSpread(key: string) { Object.assign(({ ...ts } as Record<string, object>)[key], { reference: undefined }); }"]],
    ["R6 Object.getOwnPropertyNames then a property read by computed key on the alias", [
        'import * as ts from "./_namespaces/ts.js";',
        "const viaNames = ts as unknown as Record<string, object>;",
        "export function writeThroughNames() { for (const key of Object.getOwnPropertyNames(viaNames)) if (key.startsWith(\"comment\")) Object.assign(viaNames[key]!, { reference: undefined }); }"]],
    ["S1 the shipped reflection shape (control, expect refusal)", [
        'import * as ts from "./_namespaces/ts.js";',
        "export function writeNamespaceEntry(key: keyof typeof ts) { Object.assign(ts[key] as object, { reference: undefined }); }"]],
    ["S2 for...in over the namespace (expect refusal, no shipped fixture reaches this arm)", [
        'import * as ts from "./_namespaces/ts.js";',
        "export function writeThroughForIn() { for (const key in ts) if (key.startsWith(\"comment\")) Object.assign((ts as Record<string, object>)[key]!, { reference: undefined }); }"]],
];

const digest = file => createHash("sha256").update(fs.readFileSync(file)).digest("hex").slice(0, 12);
try {
    for (const [label, lines] of probes) {
        const root = path.join(tmp, "probe");
        fs.rmSync(root, { recursive: true, force: true });
        fs.cpSync(base, root, { recursive: true });
        const changed = typesText.replace(anchor + newline, [anchor, "", ...lines, ""].join(newline));
        if (changed === typesText) throw new Error("anchor not found");
        if (changed.includes("commentPragmas")) {
            const count = (changed.match(/commentPragmas/g) || []).length - (typesText.match(/commentPragmas/g) || []).length;
            if (count !== 0) throw new Error(`${label} names commentPragmas`);
        }
        const syntax = ts.createSourceFile(typesFile, changed, ts.ScriptTarget.Latest, true);
        fs.writeFileSync(path.join(root, typesFile), changed);
        const files = [typesFile, "src/compiler/parser.ts"].map(file => path.join(root, file));
        const before = files.map(digest);
        const result = spawnSync(process.execPath, [adapter, root], { encoding: "utf8" });
        const after = files.map(digest);
        const why = result.status === 0 ? result.stdout.trim()
            : (result.stderr.split("\n").find(line => line.startsWith("AssertionError")) || result.stderr.trim().split("\n")[0]);
        console.log(`PROBE ${label}: ${result.status === 0 ? "ACCEPTED" : "refused"} exit ${result.status}, parse diagnostics ${syntax.parseDiagnostics.length}, parser ${before[1]} -> ${after[1]}; ${why}`);
    }

    // The census: how a namespace-use allowlist would price on today's source.
    const files = execFileSync("git", ["-C", tree, "ls-tree", "-r", "--name-only", ref, "src"], { encoding: "utf8" }).trim().split("\n");
    const byExtension = {};
    for (const file of files) { const extension = path.extname(file) || "(none)"; byExtension[extension] = (byExtension[extension] || 0) + 1; }
    console.log("CENSUS src files by extension at the before ref: " + JSON.stringify(byExtension));
    let namespaceFiles = 0, propertyUses = 0, typeUses = 0;
    const other = [];
    for (const file of files.filter(file => file.endsWith(".ts"))) {
        const text = fs.readFileSync(path.join(base, file), "utf8");
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        const names = new Set();
        for (const statement of source.statements) {
            if (!ts.isImportDeclaration(statement) || !/(?:^|\/)(?:_namespaces\/ts|types)\.js$/.test(statement.moduleSpecifier.text)) continue;
            const binding = statement.importClause?.namedBindings;
            if (binding && ts.isNamespaceImport(binding)) names.add(binding.name.text);
        }
        if (names.size === 0) continue;
        namespaceFiles++;
        (function visit(node) {
            if (ts.isIdentifier(node) && names.has(node.text) && !ts.isNamespaceImport(node.parent)) {
                const parent = node.parent;
                if (ts.isPropertyAccessExpression(parent) && parent.expression === node) propertyUses++;
                else if (ts.isQualifiedName(parent) && parent.left === node) typeUses++;
                else if (ts.isPropertyAccessExpression(parent) && parent.name === node) { /* another object's member */ }
                else if (ts.isPropertyAssignment(parent) && parent.name === node) { /* a key */ }
                else other.push(`${file}:${source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1} ${ts.SyntaxKind[parent.kind]}: ${parent.getText(source).slice(0, 90).replace(/\s+/g, " ")}`);
            }
            ts.forEachChild(node, visit);
        })(source);
    }
    console.log(`CENSUS namespace imports: ${namespaceFiles} files, ${propertyUses} property accesses by name, ${typeUses} qualified type names, ${other.length} other uses`);
    for (const line of other.slice(0, 60)) console.log("CENSUS other " + line);
} finally {
    fs.rmSync(tmp, { recursive: true, force: true });
}
