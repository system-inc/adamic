#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const {spawnSync} = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
if (process.argv.length !== 5) throw new Error("usage: node system-type-mutants.cjs <checker> <source-only-tree> <new-output>");
const [checker, tree, output] = process.argv.slice(2).map(value => path.resolve(value));
assert(!fs.existsSync(output), "mutant output exists");
fs.mkdirSync(output, {recursive:true});
const file = "transformers/module/system.ts", target = path.join(tree, "src/compiler", file);
const original = fs.readFileSync(target, "utf8");
const source = ts.createSourceFile(file, original, ts.ScriptTarget.Latest, true);
const results = [];
for (const site of require("./type-edits.json").filter(site => site.file === file)) {
    assert.equal(site.file, file);
    let node;
    function visit(value) {
        if (site.kind === "VariableDeclaration" && ts.isVariableDeclaration(value) && value.name.getText(source) === site.name) {
            assert(!node); node = value;
        }
        if (site.kind === "CallExpression" && ts.isCallExpression(value) && value.expression.getText(source) === site.name &&
            JSON.stringify(value.arguments.map(arg => arg.getText(source))) === JSON.stringify(site.arguments)) {
            assert(!node); node = value;
        }
        ts.forEachChild(value, visit);
    }
    visit(source); assert(node);
    const type = site.kind === "VariableDeclaration" ? node.type : node.typeArguments[0];
    const replacement = site.kind === "VariableDeclaration" ? site.before : "Identifier | undefined";
    const text = original.slice(0, type.getStart(source)) + replacement + original.slice(type.end);
    const out = path.join(output, site.name);
    try {
        fs.writeFileSync(target, text);
        const contract = spawnSync(process.execPath, [path.join(__dirname,"adapt.cjs"),"--check",tree], {encoding:"utf8"});
        assert.equal(contract.status, 1);
        assert(contract.stderr.includes(site.kind === "VariableDeclaration" ? "optional owner missing" : "unexpected overload type"));
        const census = spawnSync("bash", [path.join(__dirname,"latent-file-census.sh"),checker,tree,out], {encoding:"utf8"});
        fs.writeFileSync(path.join(output,site.name+".log"), census.stdout+census.stderr);
        assert.equal(census.status, 0);
        const diagnostics = JSON.parse(fs.readFileSync(path.join(out,"all-diagnostics.json"))).diagnostics.filter(text=>text.includes("/src/compiler/"+file+":"));
        assert.equal(diagnostics.length, 1);
        assert(diagnostics[0].includes("error TS2322:"));
        assert(diagnostics[0].includes({moduleInfo:":1773:13:",exportFunction:":1774:13:",contextObject:":1776:13:"}[site.name]));
        results.push({site,contract_exit:contract.status,contract_reason:contract.stderr.trim(),census_diagnostics:diagnostics});
    } finally {fs.writeFileSync(target,original);}
}
// A required transformation helper must not become reachable during emission.
let emit;
function findEmit(node) {if(ts.isFunctionDeclaration(node)&&node.name?.text==="onEmitNode")emit=node;ts.forEachChild(node,findEmit);}
findEmit(source);assert(emit);
const at = emit.body.statements[0].getStart(source);
const graphTree = path.join(output,"graph"), graphFile = path.join(graphTree,"src/compiler",file);
fs.mkdirSync(path.dirname(graphFile),{recursive:true});
fs.writeFileSync(graphFile, original.slice(0,at)+"appendExportsOfDeclaration(undefined, new IdentifierNameMap(), node as Declaration);\r\n"+original.slice(at));
const graph = spawnSync(process.execPath,[path.join(__dirname,"verify-system-context.cjs"),graphTree],{encoding:"utf8"});
assert.equal(graph.status,1);
assert(graph.stderr.includes("required reader reachable in emission"));
results.push({mutation:"invoke required transform helper from emission root",check:"context protocol",exit:graph.status,reason:graph.stderr});
for (const [name, functionName, condition, expected] of [
    ["meta-guard", "substituteMetaProperty", "isImportMeta(node)", "isImportMeta(node)"],
    ["export-guard", "substituteBinaryExpression", "exportedNames", "export emission lost export-name guard"],
]) {
    let fn, guard;
    function findFn(node) {if(ts.isFunctionDeclaration(node)&&node.name?.text===functionName)fn=node;ts.forEachChild(node,findFn);}
    findFn(source);assert(fn);
    function findGuard(node) {if(ts.isIfStatement(node)&&node.expression.getText(source)===condition){assert(!guard);guard=node;}ts.forEachChild(node,findGuard);}
    findGuard(fn.body);assert(guard);
    const guardTree=path.join(output,name), guardFile=path.join(guardTree,"src/compiler",file);
    fs.mkdirSync(path.dirname(guardFile),{recursive:true});
    fs.writeFileSync(guardFile, original.slice(0,guard.expression.getStart(source))+"true"+original.slice(guard.expression.end));
    const check=spawnSync(process.execPath,[path.join(__dirname,"verify-system-context.cjs"),guardTree],{encoding:"utf8"});
    assert.equal(check.status,1);assert(check.stderr.includes(expected));
    results.push({mutation:name,check:"conditional emission protocol",exit:check.status,reason:check.stderr});
}
console.log(JSON.stringify(results,null,2));
