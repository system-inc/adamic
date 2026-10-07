#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
const file = "transformers/module/system.ts";
const input = path.resolve(process.argv[2], "src/compiler", file);
const program = ts.createProgram([input], {noResolve:true, noLib:true, target:ts.ScriptTarget.Latest});
const source = program.getSourceFile(input);
const checker = program.getTypeChecker();
assert.equal(source.parseDiagnostics.length, 0);
const functions = new Map();
function collect(node) {
    if (ts.isFunctionDeclaration(node) && node.name) {
        assert(!functions.has(node.name.text), "duplicate local function name");
        functions.set(node.name.text, node);
    }
    ts.forEachChild(node, collect);
}
collect(source);
// Conservatively include references passed as callbacks and returned delegates,
// not merely direct calls. Nested declarations have their own graph vertices.
const graph = new Map();
for (const [name, fn] of functions) {
    const edges = new Set();
    function visit(node) {
        if (ts.isFunctionDeclaration(node)) return;
        if (ts.isIdentifier(node) && functions.has(node.text)) {
            const symbol = checker.getSymbolAtLocation(node);
            if (symbol?.declarations?.includes(functions.get(node.text))) edges.add(node.text);
        }
        ts.forEachChild(node, visit);
    }
    visit(fn.body);
    graph.set(name, edges);
}
function reachable(root) {
    const seen = new Set();
    function visit(name) {
        if (seen.has(name)) return;
        seen.add(name);
        for (const next of graph.get(name) || []) visit(next);
    }
    visit(root);
    return seen;
}
const transformation = reachable("transformSourceFile");
const emission = new Set([...reachable("onEmitNode"), ...reachable("onSubstituteNode")]);
const sites = require("./required-values.json").filter(site => site.file === file && site.kind === "Identifier");
assert.equal(sites.length, 25);
const conditional = new Set(["createExportExpression", "substituteMetaProperty"]);
const owners = [...new Set(sites.filter(site => !conditional.has(site.function)).map(site => site.function))];
for (const name of owners) {
    assert(transformation.has(name), "required reader outside transformation: " + name);
    assert(!emission.has(name), "required reader reachable in emission: " + name);
}
for (const name of transformation) assert(!graph.get(name).has("transformSourceFile"), "recursive source-file transformation");
const writes = new Map(["moduleInfo", "exportFunction", "contextObject"].map(name => [name, []]));
function writesVisit(node) {
    if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
        ts.isIdentifier(node.left) && writes.has(node.left.text)) {
        let owner = node.parent;
        while (owner && !ts.isFunctionDeclaration(owner)) owner = owner.parent;
        writes.get(node.left.text).push({node, owner:owner.name.text, right:node.right.getText(source)});
    }
    ts.forEachChild(node, writesVisit);
}
writesVisit(source);
const producers = {moduleInfo:"moduleInfoMap[id] = collectExternalModuleInfo(context, node)",
    exportFunction:'factory.createUniqueName("exports")', contextObject:'contextObjectMap[id] = factory.createUniqueName("context")'};
const maps = {moduleInfo:"moduleInfoMap", exportFunction:"exportFunctionsMap", contextObject:"contextObjectMap"};
let bodyCall;
function callsVisit(node) {
    if (ts.isCallExpression(node) && node.expression.getText(source) === "createSystemModuleBody") {assert(!bodyCall);bodyCall=node;}
    ts.forEachChild(node, callsVisit);
}
callsVisit(functions.get("transformSourceFile").body);assert(bodyCall);
for (const [name, list] of writes) {
    assert.deepEqual(list.map(write => [write.owner,write.right]), [
        ["transformSourceFile",producers[name]], ["transformSourceFile","undefined!"],
        ["onEmitNode",maps[name]+"[id]"], ["onEmitNode","undefined!"],
    ]);
    assert(list[0].node.end < bodyCall.pos && bodyCall.end < list[1].node.pos);
}
const exportCalls=[];
for (const name of emission) {
    const fn=functions.get(name);if(!fn)continue;
    function visit(node) {
        if(ts.isFunctionDeclaration(node))return;
        if(ts.isCallExpression(node)&&node.expression.getText(source)==="createExportExpression")exportCalls.push({name,node});
        ts.forEachChild(node,visit);
    }
    visit(fn.body);
}
assert.equal(exportCalls.length,1);
assert.equal(exportCalls[0].name,"substituteBinaryExpression");
let guard=exportCalls[0].node.parent;
while(guard&&!(ts.isIfStatement(guard)&&guard.expression.getText(source)==="exportedNames"))guard=guard.parent;
assert(guard,"export emission lost export-name guard");
let names;
function findNames(node){if(ts.isVariableDeclaration(node)&&node.name.getText(source)==="exportedNames")names=node;ts.forEachChild(node,findNames);}
findNames(functions.get("substituteBinaryExpression").body);
assert.equal(names.initializer.getText(source),"getExports(node.left)");
assert(names.end < guard.pos);
const meta=functions.get("substituteMetaProperty").body.statements[0];
assert(ts.isIfStatement(meta));assert.equal(meta.expression.getText(source),"isImportMeta(node)");
assert.equal(meta.thenStatement.statements[0].expression.arguments[0].getText(source),"contextObject!");
const observations=[];
for(const moduleDetection of [ts.ModuleDetectionKind.Legacy,ts.ModuleDetectionKind.Auto,ts.ModuleDetectionKind.Force]) {
    const name="/probe.ts",text="console.log(import.meta);";
    const options={target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.System,moduleDetection,noLib:true};
    const host=ts.createCompilerHost(options);host.getSourceFile=(f,l)=>f===name?ts.createSourceFile(f,text,l,true):undefined;
    host.fileExists=f=>f===name;host.readFile=f=>f===name?text:undefined;
    const probe=ts.createProgram([name],options,host).getSourceFile(name);
    // SourceFile parsing under Program installs the compiler's own detection callback.
    assert(ts.isExternalModule(probe));
    observations.push({moduleDetection:ts.ModuleDetectionKind[moduleDetection],import_meta_is_external:true});
}
console.log(JSON.stringify({status:"pass",required_reads:sites.length,transformation_only_reads:23,
    transformation_only_reader_functions:owners,all_three_contexts_initialized_before_body:true,
    reset_follows_body:true,no_recursive_source_file_transform:true,
    conditional_export_emission_guard_verified:true,conditional_import_meta_guard_verified:true,
    import_meta_detection:observations},null,2));
