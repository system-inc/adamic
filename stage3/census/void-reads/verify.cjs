'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const [repo,tree,table]=process.argv.slice(2),data=JSON.parse(fs.readFileSync(table,'utf8'));
const ts=require(path.join(repo,'stage3/api/node_modules/typescript'));
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
const parsed=ts.convertCompilerOptionsFromJson(data.options,tree);
if(parsed.errors.length)throw Error('options');
if(data.roots.length!==79||data.sites.length!==491||data.unionWithVoid.length!==16)throw Error('count drift');
for(const root of data.roots)if(hash(fs.readFileSync(path.join(tree,root.file)))!==root.sha256)throw Error('root drift '+root.file);
let files=0;
for(const [file,proof] of Object.entries(data.adaptationProof.files)){
 const source=fs.readFileSync(path.join(tree,file),'utf8');let changed=source;
 for(const edit of [...proof.edits].sort((a,b)=>b.start-a.start))changed=changed.slice(0,edit.start)+edit.replacement+changed.slice(edit.end);
 const emit=text=>ts.transpileModule(text,{fileName:file,compilerOptions:parsed.options}).outputText;
 const original=emit(source),rewritten=emit(changed),mutant=emit(changed+'\n0;\n');
 if(original!==rewritten||hash(original)!==proof.beforeSha256||hash(rewritten)!==proof.afterSha256)throw Error('emission proof failed '+file);
 if(mutant===original)throw Error('emission mutant escaped '+file);
 files++;
}
const count={};for(const site of data.sites)count[site.judgment.kind]=(count[site.judgment.kind]||0)+1;
if(JSON.stringify(Object.entries(count).sort())!==JSON.stringify(Object.entries(data.countsByJudgment).sort()))throw Error('judgment totals');
console.log(JSON.stringify({verifiedRootHashes:79,verifiedEmissionFiles:files,emissionMutantsCaught:files,emissionMutation:'Append the runtime expression statement 0; to each rewritten file, one at a time.',judgments:count}));
