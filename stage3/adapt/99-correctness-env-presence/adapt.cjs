"use strict";
const fs=require("node:fs"),path=require("node:path"),crypto=require("node:crypto"),assert=require("node:assert/strict"),ts=require("typescript");
const owners=require("./owners.json");
const digest=text=>crypto.createHash("sha256").update(text).digest("hex");
function locate(text) {
    const source=ts.createSourceFile("sys.ts",text,ts.ScriptTarget.Latest,true);
    assert.equal(source.parseDiagnostics.length,0,"sys.ts: parse error");
    const found=source.statements.filter(n=>ts.isIfStatement(n)&&ts.isBlock(n.thenStatement)&&n.thenStatement.statements.some(s=>ts.isExpressionStatement(s)&&ts.isCallExpression(s.expression)&&ts.isIdentifier(s.expression.expression)&&s.expression.expression.text==="setCustomPollingValues"));
    assert.equal(found.length,1,"environment initialization: missing or duplicate owner");
    const start=found[0].getStart(source),end=found[0].end,spelling=text.slice(start,end),hash=digest(spelling),pin=owners[0];
    assert([pin.ownerBeforeSha256,pin.ownerAfterSha256].includes(hash),"environment initialization: drift or half-adapted owner");
    const adapted=hash===pin.ownerAfterSha256;
    assert.equal(spelling,adapted?pin.after:pin.before,"environment initialization: exact text drift");
    return {start,end,adapted};
}
function transpile(text,removeComments=false) {
    return ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
}
function rewrite(tree,restoring=false) {
    assert.equal(ts.version,"6.0.3");
    assert.equal(owners.length,1);
    const pin=owners[0];
    for(const side of ["Before","After"]) assert.equal(digest(pin[side.toLowerCase()]),pin["owner"+side+"Sha256"],"invalid owner pin");
    const file=path.join(tree,pin.file),raw=fs.readFileSync(file,"utf8"),text=raw.replace(/\r\n/g,"\n"),crlf=raw.includes("\r\n");
    assert(!text.includes("\r"),"unsupported line ending");
    assert(!crlf||!raw.replace(/\r\n/g,"").includes("\n"),"mixed line endings");
    const state=locate(text),wanted=!restoring;
    if(state.adapted===wanted) return {files:0,edits:0,javascriptChanging:true};
    let after=text.slice(0,state.start)+(wanted?pin.after:pin.before)+text.slice(state.end);
    assert.equal(locate(after).adapted,wanted,"invalid planned owner");
    // This approved change alters JavaScript. Only the dropped condition test may differ.
    for(const removeComments of [false,true]) {
        const original=transpile(restoring?after:text,removeComments),adapted=transpile(restoring?text:after,removeComments);
        const needle="if (sys && sys.getEnvironmentVariable) {";
        assert.equal(original.split(needle).length,2,"JavaScript target count drift");
        assert.equal(adapted,original.replace(needle,"if (sys) {"),"unexpected JavaScript change");
    }
    if(crlf) after=after.replace(/\n/g,"\r\n");
    assert.equal(fs.readFileSync(file,"utf8"),raw,"source changed while planning");
    fs.writeFileSync(file,after);
    return {files:1,edits:1,javascriptChanging:true};
}
const adapt=tree=>rewrite(tree);
module.exports={adapt,rewrite,locate,transpile,owners,digest};
if(require.main===module) { assert.equal(process.argv.length,3,"usage: node adapt.cjs TREE");console.log(JSON.stringify(adapt(path.resolve(process.argv[2])))); }
