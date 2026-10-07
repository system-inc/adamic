"use strict";
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const [tree, output] = process.argv.slice(2);
assert(tree && output);
const sf = ts.createSourceFile("core.ts", fs.readFileSync(path.join(tree,"src/compiler/core.ts"),"utf8"), ts.ScriptTarget.Latest, true);
function body(name) {
    const nodes = sf.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === name && n.body);
    assert.equal(nodes.length,1);
    return ts.transpileModule(nodes[0].getText(sf),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS}}).outputText;
}
const originalJS = body("addRange"), offsets = body("toOffset");
assert.equal(originalJS.split("to.push(from[i]);").length,2);
const checkedJS = originalJS.replace("to.push(from[i]);","to.push(requirePresent(from[i]));");
function make(text) { return new Function("exports", "requirePresent", offsets + text + "; return exports.addRange;")({}, value => {
    if(value === undefined) throw new Error("parser addRange second read missing");
    return value;
}); }
const original=make(originalJS),checked=make(checkedJS);
const controls=[];
for(const start of [undefined,0,1,-3]) for(const end of [undefined,0,2,8]) {
    const from=[undefined,23,,7];
    for(const destination of [undefined,[],[99]]) {
        const before=original(destination?.slice(),from,start,end),after=checked(destination?.slice(),from,start,end);
        assert.deepEqual(after,before);controls.push({start,end,destination,before});
    }
}
function getterState(fn,kind) {
    const from=[23],to=[];let reads=0;
    if(kind==="indexed getter") Object.defineProperty(from,"0",{get(){return ++reads===1?23:undefined;},configurable:true});
    else Object.defineProperty(to,"push",{get(){delete from[0];return Array.prototype.push;}});
    fn(to,from);
    return {reads,length:to.length,own:Object.hasOwn(to,"0"),value:String(to[0])};
}
const mutants=[];
for(const name of ["indexed getter","push getter"]){
    const old=getterState(original,name);assert.equal(old.own,true);assert.equal(old.value,"undefined");
    assert.throws(()=>getterState(checked,name),/parser addRange second read missing/);
    // Remove only the presence check: the same invalid state is accepted again.
    assert.deepEqual(getterState(make(checkedJS.replace("requirePresent(from[i])","from[i]")),name),old);
    mutants.push({name,original:old,checked:"throws: parser addRange second read missing",removed_check:"accepts own undefined again"});
}
fs.writeFileSync(output,JSON.stringify({controls:controls.length,mutants,model:"Only the asserted second read is wrapped; to.push is looked up before reading from[i], exactly as in source."},null,2)+"\n");
console.log("48 ordinary/sparse controls identical; both invalid getter states rejected only by presence check.");
