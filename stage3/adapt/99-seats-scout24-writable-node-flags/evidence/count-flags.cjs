"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");

const [repository, tree, output] = process.argv.slice(2).map(p => path.resolve(p));
const packages = path.join(repository, "stage3/api/node_modules");
const ts = require(path.join(packages, "typescript"));
assert.equal(ts.version, "6.0.3");
const { compiler } = require(path.join(repository, "stage3/adapt/47-host-errors/adapt.cjs"));
const src = path.join(tree, "src");
const roots = ts.sys.readDirectory(src, [".ts"]).sort();
const options = {
    ...compiler(tree, ts).getCompilerOptions(),
    typeRoots: [path.join(packages, "@types")],
    skipLibCheck: true,
    noEmit: true,
};
const program = ts.createProgram(roots, options);
const checker = program.getTypeChecker();
const compilerTypes = program.getSourceFile(path.join(src, "compiler/types.ts"));
const [nodeDeclaration] = compilerTypes.statements.filter(node =>
    ts.isInterfaceDeclaration(node) && node.name.text === "Node");
assert(nodeDeclaration);
const nodeType = checker.getDeclaredTypeOfSymbol(checker.getSymbolAtLocation(nodeDeclaration.name));
const flags = checker.getPropertyOfType(nodeType, "flags");
assert.equal(flags.declarations.length, 1);
const baseFlags = checker.getTypeOfSymbolAtLocation(flags, flags.valueDeclaration ?? flags.declarations[0]);
assert.equal(checker.typeToString(baseFlags), "NodeFlags");
const rows = new Map();
let resolvedTypes = 0;

function clean(text) {
    return text.replace(/[\r\n\t]+/g, " ").trim();
}

function record(at, owner, propertyType, kind) {
    const file = at.getSourceFile();
    const point = file.getLineAndCharacterOfPosition(at.getStart(file));
    const assignable = checker.isTypeAssignableTo(owner, nodeType);
    const narrower = !checker.isTypeAssignableTo(baseFlags, propertyType) &&
        checker.isTypeAssignableTo(propertyType, baseFlags);
    const identity = [path.relative(tree, file.fileName), point.line + 1, point.character + 1];
    const row = [...identity, clean(checker.typeToString(owner, at)), assignable,
        clean(checker.typeToString(propertyType, at)), narrower, kind];
    rows.set(row.join("\t"), { row, at, assignable, narrower });
}

function inspectType(node, type) {
    if (type.flags & ts.TypeFlags.TypeParameter) {
        type = checker.getBaseConstraintOfType(type) ?? type;
    }
    const member = checker.getPropertyOfType(type, "flags");
    if (!member) return;
    resolvedTypes++;
    const propertyType = checker.getTypeOfSymbolAtLocation(member, node);
    // The declaration pass already records ordinary inherited properties.
    // This pass records refinements and synthetic members on Node views,
    // including intersections, instantiated mapped types and constraints.
    if (checker.isTypeAssignableTo(type, nodeType) &&
        (!member.declarations?.length || !checker.isTypeAssignableTo(baseFlags, propertyType))) {
        const declaration = member.declarations?.find(d => d.getSourceFile() === node.getSourceFile());
        if (declaration && checker.getTypeAtLocation(declaration.parent) === type) return;
        record(node, type, propertyType, "resolved");
    }
}

for (const root of roots) {
    const source = program.getSourceFile(root);
    assert(source, root);
    function visit(node) {
        if (node.name && (ts.isPropertySignature(node) || ts.isPropertyDeclaration(node) ||
            ts.isPropertyAssignment(node) || ts.isShorthandPropertyAssignment(node) ||
            ts.isMethodSignature(node) || ts.isMethodDeclaration(node) ||
            ts.isGetAccessorDeclaration(node) || ts.isSetAccessorDeclaration(node) ||
            ts.isParameter(node) || ts.isEnumMember(node))) {
            const member = checker.getSymbolAtLocation(node.name);
            if (member?.name === "flags" && !ts.isParameter(node)) {
                record(node.name, checker.getTypeAtLocation(node.parent),
                    checker.getTypeOfSymbolAtLocation(member, node), "declaration");
            }
            if (member?.name === "flags" && ts.isParameter(node) && ts.isParameterPropertyDeclaration(node, node.parent)) {
                record(node.name, checker.getTypeAtLocation(node.parent.parent),
                    checker.getTypeOfSymbolAtLocation(member, node), "declaration");
            }
        }
        if (ts.isTypeNode(node) || ts.isTypeParameterDeclaration(node) ||
            ts.isInterfaceDeclaration(node) || ts.isClassLike(node) ||
            ts.isObjectLiteralExpression(node) || ts.isCallExpression(node) ||
            ts.isNewExpression(node) || ts.isVariableDeclaration(node) ||
            ts.isTypeAliasDeclaration(node)) {
            inspectType(node, checker.getTypeAtLocation(node));
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
}

const records = [...rows.values()].sort((a, b) => a.row[0].localeCompare(b.row[0], "en") ||
    a.row[1] - b.row[1] || a.row[2] - b.row[2]);
const narrow = records.filter(record => record.assignable && record.narrower);
const sites = new Map(narrow.map(record => [record.row.slice(0, 3).join(":"), record]));
fs.writeFileSync(path.join(output, "flags-declarations.tsv"),
    "file\tline\tcolumn\tdeclaring type\tassignable to Node\tdeclared flags type\tnarrower than NodeFlags\tkind\n" +
    records.map(record => record.row.join("\t")).join("\n") + "\n");
const summary = [
    `Narrower Node flags sites: ${sites.size}`,
    `TypeScript ${ts.version}, ${roots.length} .ts roots including ${roots.filter(file => file.endsWith(".d.ts")).length} .d.ts files`,
    `${records.length} declaration/refinement rows, ${resolvedTypes} flags-bearing type visits`,
];
for (const [where, { at }] of sites) {
    const file = at.getSourceFile();
    const { line } = file.getLineAndCharacterOfPosition(at.getStart(file));
    summary.push(`${where}: ${file.text.split(/\r?\n/)[line].trim()}`);
}
fs.writeFileSync(path.join(output, "flags-count.txt"), summary.join("\n") + "\n");
console.log(summary.join("\n"));
