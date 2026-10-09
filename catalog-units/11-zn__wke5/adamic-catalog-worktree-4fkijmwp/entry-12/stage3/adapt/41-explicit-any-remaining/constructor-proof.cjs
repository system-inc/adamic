"use strict";
const fs=require("node:fs"),path=require("node:path"),vm=require("node:vm"),assert=require("node:assert/strict"),ts=require("typescript");
const [beforeTree,afterTree,output]=process.argv.slice(2);
const read=(tree,file)=>fs.readFileSync(path.join(tree,"src/compiler",file),"utf8");
const parse=(text,file)=>ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const declaration=(source,name,predicate)=>{const found=[]; const visit=n=>{if(predicate(n)&&n.name?.text===name)found.push(n);n.forEachChild(visit);};visit(source);assert.equal(found.length,1,name);return found[0];};
const required=(tree,name)=>declaration(parse(read(tree,"types.ts"),"types.ts"),name,ts.isInterfaceDeclaration).members.filter(n=>ts.isPropertySignature(n)&&!n.questionToken).map(n=>n.name.getText());
const snippets=tree=>{
 const u=parse(read(tree,"utilities.ts"),"utilities.ts"),c=parse(read(tree,"checker.ts"),"checker.ts");
 return ["Symbol","Signature","SourceMapSource"].map(n=>declaration(u,n,ts.isFunctionDeclaration).getText(u)).concat(declaration(c,"createSignature",ts.isFunctionDeclaration).getText(c)).join("\n");
};
const run=text=>{
 const js=ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText;
 return vm.runInNewContext(js+`
({symbol:new Symbol(0,"name"), source:new SourceMapSource("input.a","text"), signature:createSignature(undefined,undefined,undefined,[],undefined,undefined,0,0), header:new Signature(checker,0)})`,{Debug:{isDebugging:false},tracing:undefined,checker:{}});
};
const a=run(snippets(beforeTree)),b=run(snippets(afterTree));
for(const k of Object.keys(a)){
 assert.equal(JSON.stringify(b[k]),JSON.stringify(a[k]),"actual constructor observations "+k);
 assert.deepEqual(Object.keys(b[k]),Object.keys(a[k]),"own field presence "+k);
}
const cases={symbol:"Symbol",source:"SourceMapSource",signature:"Signature"},evidence={required:{},beforeMissing:{},afterMissing:{},mutants:[]};
const check=(value,name)=>{for(const key of required(afterTree,name))assert(Object.hasOwn(value,key),name+" missing "+key);};
for(const [key,name] of Object.entries(cases)){
 evidence.required[name]=required(afterTree,name);
 evidence.beforeMissing[name]=required(beforeTree,name).filter(p=>!Object.hasOwn(a[key],p));
 evidence.afterMissing[name]=required(afterTree,name).filter(p=>!Object.hasOwn(b[key],p));
 check(b[key],name);
}
assert.equal(b.symbol.flags,0);assert.equal(b.symbol.escapedName,"name");assert.equal(b.symbol.id,0);assert.equal(b.symbol.mergeId,0);assert.equal(b.symbol.constEnumOnlyModule,undefined);
assert.equal(b.source.fileName,"input.a");assert.equal(b.source.text,"text");assert.equal(b.source.skipTrivia(23),23);
assert.equal(b.signature.flags,0);assert.equal(b.signature.minArgumentCount,0);assert.equal(b.signature.parameters.length,0);
assert(!Object.hasOwn(b.header,"parameters"),"bare header remains incomplete");
for(const [line,key,name] of [["this.id = 0;","symbol","Symbol"],["this.text = text;","source","SourceMapSource"],["sig.parameters = parameters;","signature","Signature"]]){
 const text=snippets(afterTree);assert.equal(text.split(line).length,2);
 const mutant=run(text.replace(line,""));assert.throws(()=>check(mutant[key],name),assert.AssertionError);
 evidence.mutants.push({owner:name,removed:line,caughtBy:"required own fields from actual interface"});
}
fs.writeFileSync(output,JSON.stringify(evidence,null,2)+"\n");console.log(JSON.stringify(evidence));
