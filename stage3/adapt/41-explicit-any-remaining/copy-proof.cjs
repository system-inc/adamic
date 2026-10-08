"use strict";
const fs=require("node:fs"),path=require("node:path"),assert=require("node:assert/strict"),vm=require("node:vm"),ts=require("typescript");
const [beforeTree,afterTree,output]=process.argv.slice(2);
const parse=(tree,file)=>ts.createSourceFile(file,fs.readFileSync(path.join(tree,"src",file),"utf8"),ts.ScriptTarget.Latest,true);
const find=(source,name,predicate)=>{const found=[];const visit=n=>{if(predicate(n)&&n.name?.text===name)found.push(n);n.forEachChild(visit);};visit(source);assert.equal(found.length,1,name);return found[0];};
const core=tree=>{const source=parse(tree,"compiler/core.ts");return ["clone","extend","copyProperties"].map(name=>find(source,name,ts.isFunctionDeclaration).getText(source).replace(/^export /,"")).join("\n");};
const source=parse(afterTree,"testRunner/unittests/helpers/virtualFileSystemWithWatch.ts");
const methods=["toFsEntry","toFsFileOrLibFile","toFsSymLink","toFsFolder"].map(name=>find(source,name,ts.isMethodDeclaration).getText(source)).join("\n");
const required=name=>{const n=find(source,name,ts.isInterfaceDeclaration),base=n.heritageClauses?.flatMap(h=>h.types.map(t=>t.expression.getText(source)))||[];return [...base.flatMap(required),...n.members.filter(ts.isPropertySignature).filter(p=>!p.questionToken).map(p=>p.name.getText(source))];};
const run=text=>{
 const input=text+`
class ActualEntryFactory { currentDirectory="/"; toPath(p: string){return p;} now(){return new Date(1);} ${methods} }
 const factory=new ActualEntryFactory();
 const entries=[factory.toFsFileOrLibFile({path:"/file",content:"body"}),factory.toFsFileOrLibFile({path:"/lib",libFile:true}),factory.toFsSymLink({path:"/link",symLink:"/file"}),factory.toFsFolder("/folder")];
 ({entries,copies:entries.map(clone),merged:extend({strict:true},{strict:false,noEmit:true})});`;
 return vm.runInNewContext(ts.transpileModule(input,{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText,{hasOwnProperty:Object.prototype.hasOwnProperty,getNormalizedAbsolutePath:p=>p,getDirectoryPath:p=>p.substring(0,p.lastIndexOf("/"))});
};
const a=run(core(beforeTree)),b=run(core(afterTree));assert.equal(JSON.stringify(b),JSON.stringify(a),"actual caller entry construction and copies");
const names=["FsFile","FsLibFile","FsSymLink","FsFolder"];
const check=result=>{names.forEach((name,i)=>{for(const key of required(name)){assert(Object.hasOwn(result.entries[i],key),"completed input "+name+"."+key);assert(Object.prototype.propertyIsEnumerable.call(result.entries[i],key),"enumerable required input");assert(Object.hasOwn(result.copies[i],key),"completed copy "+name+"."+key);assert.strictEqual(result.copies[i][key],result.entries[i][key],"shallow identity");}});};
check(b);assert.equal(b.merged.strict,true);assert.equal(b.merged.noEmit,true);
const text=core(afterTree),mutation="hasOwnProperty.call(object, id)";assert.equal(text.split(mutation).length,2);
const mutant=run(text.replace(mutation,mutation+' && id !== "path"'));assert.throws(()=>check(mutant),assert.AssertionError);
const evidence={actualCaller:"cloneFsMap",constructors:["toFsEntry","toFsFileOrLibFile","toFsSymLink","toFsFolder"],required:Object.fromEntries(names.map(n=>[n,required(n)])),shallowIdentity:true,firstPrecedence:true,mutants:[{change:"omit actual required path from clone",caughtBy:"required own enumerable input and output fields"}]};
fs.writeFileSync(output,JSON.stringify(evidence,null,2)+"\n");console.log(JSON.stringify(evidence));
