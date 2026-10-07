"use strict";
const {createHash}=require("node:crypto");
const row=require("./watch-protocol.json");
function validate(ts,sf) {
 const found=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===row.function&&n.body);
 if(found.length!==1) throw new Error("watch protocol owner drift");
 const js=ts.transpileModule(found[0].getText(sf),{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,removeComments:true}}).outputText;
 if(createHash("sha256").update(js).digest("hex")!==row.runtimeSha256) throw new Error("watch key/callback protocol drift");
}
module.exports={validate};
