const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const specializeFactory = require("./specialize-factory.cjs");
const {checkSource} = require("./check-postfix.cjs");
const [root, installed] = process.argv.slice(2);
const ts = require(installed);
assert.equal(ts.version, "6.0.3");
const types = path.join(root, "src/compiler/types.ts");
const factory = path.join(root, "src/compiler/factory/nodeFactory.ts");
const watcher = path.join(root, "src/compiler/watchPublic.ts");

const beforeTypes = fs.readFileSync(types, "utf8");
const file = ts.createSourceFile(types, beforeTypes, ts.ScriptTarget.Latest, true);
const edits = [];
for (const name of ["JSDocReturnTag", "JSDocThrowsTag"]) {
    const interfaces = file.statements.filter(n => ts.isInterfaceDeclaration(n) && n.name.text === name);
    assert.equal(interfaces.length, 1, name);
    const members = interfaces[0].members.filter(n => n.name?.getText(file) === "typeExpression");
    assert.equal(members.length, 1, name);
    const member = members[0];
    assert(member.questionToken && member.modifiers?.some(m => m.kind === ts.SyntaxKind.ReadonlyKeyword), name);
    const type = member.type;
    if (type.getText(file) === "JSDocTypeExpression | undefined") continue;
    assert.equal(type.getText(file), "JSDocTypeExpression", name);
    edits.push(type.end);
}
let result = beforeTypes;
for (const end of edits.sort((a,b) => b-a)) result = result.slice(0,end) + " | undefined" + result.slice(end);

const original = "program: getCurrentBuilderProgram() || rootFileNames,";
const revised = "program: (getCurrentBuilderProgram() as T | undefined) || rootFileNames,";
const watch = fs.readFileSync(watcher, "utf8");
assert.equal(watch.split(original).length + watch.split(revised).length,3);
const source = fs.readFileSync(factory, "utf8");
const specialized = source.includes("getJSDocUnaryTypeCreateFunctionForJSDocOptionalType")
    ? source : specializeFactory(ts, factory, source);
assert(specialized.includes("getJSDocTypeLikeTagCreateFunctionForJSDocThrowsTag"));
assert(specialized.includes("): Mutable<JSDocNullableType> {"));
const core = fs.readFileSync(path.join(root, "src/compiler/core.ts"), "utf8");
checkSource(ts, factory, specialized, core);

if (specialized !== source) fs.writeFileSync(factory, specialized);
if (watch.includes(original)) fs.writeFileSync(watcher,watch.replace(original,revised));
if (result !== beforeTypes) fs.writeFileSync(types,result);
