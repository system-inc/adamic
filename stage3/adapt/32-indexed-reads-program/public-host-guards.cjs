"use strict";
const fs=require("node:fs"),path=require("node:path"),{createHash}=require("node:crypto");
function validate(ts,tree){
 for(const row of require("./public-host-guards.json")){
  const file=path.join(tree,"src/compiler",row.file),sf=ts.createSourceFile(file,fs.readFileSync(file,"utf8"),ts.ScriptTarget.Latest,true),found=[];
  function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===row.function)found.push(n);ts.forEachChild(n,visit);}visit(sf);
  if(sf.parseDiagnostics.length||found.length!==1)throw new Error("host producer owner drift: "+row.function);
  const js=ts.transpileModule(found[0].getText(sf),{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,removeComments:true}}).outputText;
  if(createHash("sha256").update(js).digest("hex")!==row.runtimeSha256)throw new Error("host producer invariant drift: "+row.function);
 }
}
module.exports={validate};
