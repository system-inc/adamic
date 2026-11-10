"use strict";
const fs=require("node:fs"),assert=require("node:assert/strict"),crypto=require("node:crypto"),{pathToFileURL}=require("node:url");
const base=process.argv[2],candidate=process.argv[3],unit=__dirname;
const hash=b=>crypto.createHash("sha256").update(b).digest("hex");
(async()=>{
 const rows=[];
 for(const root of [base,candidate]){
  const file=root+"/built/local/typescript.js",cjs=require(file),mod=await import(pathToFileURL(file).href);
  rows.push({root,Debug:Object.keys(cjs.Debug).sort(),module:Object.keys(mod).sort(),ts:Object.keys(cjs).sort()});
 }
 const checks=[];
 for(const key of ["Debug","module","ts"]){
  const before=rows[0][key],after=rows[1][key],removed=before.filter(k=>!after.includes(k)),added=after.filter(k=>!before.includes(k));
  assert.deepEqual(removed,key==="Debug"?["formatControlFlowGraph","printControlFlowGraph"]:[]);assert.deepEqual(added,[]);
  checks.push({object:key,before_keys:before.length,after_keys:after.length,removed,added});
 }
 const declarations=[];
 for(const file of ["typescript.d.ts","tsserverlibrary.d.ts","typescript.internal.d.ts"]){
  const a=fs.readFileSync(base+"/built/local/"+file),b=fs.readFileSync(candidate+"/built/local/"+file);
  if(file==="typescript.internal.d.ts"){
   let expected=a.toString();
   for(const line of ["        function printControlFlowGraph(flowNode: FlowNode): void;\n","        function formatControlFlowGraph(flowNode: FlowNode): string;\n"]){assert.equal(expected.split(line).length-1,1);expected=expected.replace(line,"");}
   assert.equal(b.toString(),expected,"internal declarations changed beyond graph pair");
  } else assert(a.equals(b),file+" changed");
  declarations.push({file,before_sha256:hash(a),after_sha256:hash(b),identical:a.equals(b)});
 }
 const info=JSON.parse(fs.readFileSync(candidate+"/built/local/deprecatedCompat/tsconfig.tsbuildinfo","utf8"));
 assert(info.fileNames.some(n=>n.endsWith("/controlFlowGraph.ts")));
 fs.writeFileSync(unit+"/debug-display-2-exports-and-declarations.json",JSON.stringify({status:"pass",node:process.version,checks,declarations,new_module_typechecked:true,rows},null,2)+"\n");
 console.log(JSON.stringify({status:"pass",checks,declarations}));
})().catch(e=>{console.error(e);process.exitCode=1});
