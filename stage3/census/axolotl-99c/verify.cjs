const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto'),cp=require('child_process');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const adapter=require('/workspace/adamic/stage3/adapt/99-correctness-axolotl-c/adapt.cjs');
const a='/workspace/scratch/before',b='/workspace/scratch/after';
const files=ts.sys.readDirectory(a+'/src',['.ts','.tsx','.json']);
const changed=files.filter(f=>!fs.readFileSync(f).equals(fs.readFileSync(b+f.slice(a.length))));
assert.deepEqual(changed.map(f=>f.slice(a.length)),['/src/compiler/types.ts']);
const options={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments:false};
const before=fs.readFileSync(a+'/src/compiler/types.ts','utf8'),after=fs.readFileSync(b+'/src/compiler/types.ts','utf8');
const js=x=>ts.transpileModule(x,{compilerOptions:options,fileName:'types.ts'}).outputText;
assert.equal(js(before),js(after));
const report={changedFiles:changed.map(f=>f.slice(a.length+1)),javascriptBytes:Buffer.byteLength(js(before)),javascriptSha256:crypto.createHash('sha256').update(js(before)).digest('hex')};
assert.deepEqual(adapter.adapt(b),{files:0,declarations:0,javascriptChanging:false});assert.equal(fs.readFileSync(b+'/src/compiler/types.ts','utf8'),after);
report.idempotence=true;
const tmp=fs.mkdtempSync('/workspace/scratch/contract-');fs.mkdirSync(tmp+'/src/compiler',{recursive:true});
const fp=tmp+'/src/compiler/types.ts';
const cases={partial:before.replace('project?: string;', 'project?: string | undefined;'),drift:before.replace('project?: string;', 'project?: number;'),duplicate:before+'\ninterface CompilerOptions {}\n'};
report.refusals={};
for(const [name,text] of Object.entries(cases)){fs.writeFileSync(fp,text);assert.throws(()=>adapter.adapt(tmp));assert.equal(fs.readFileSync(fp,'utf8'),text);report.refusals[name]=true;}
const pristine=fs.readFileSync('/workspace/scratch/upstream/src/compiler/types.ts','utf8');fs.writeFileSync(fp,pristine);adapter.adapt(tmp);assert.deepEqual(adapter.adapt(tmp),{files:0,declarations:0,javascriptChanging:false});report.pristineAndIdempotence=true;
const expected=JSON.parse(fs.readFileSync('/workspace/scratch/api-expected.json')).expected;
assert.equal(fs.readFileSync(b+'/built/local/typescript.d.ts','utf8'),expected);
const ab=fs.readFileSync(a+'/built/local/typescript.d.ts','utf8'),bb=fs.readFileSync(b+'/built/local/typescript.d.ts','utf8');
const diff=cp.spawnSync('diff',['-u',a+'/built/local/typescript.d.ts',b+'/built/local/typescript.d.ts'],{encoding:'utf8'});fs.writeFileSync('/workspace/scratch/api-99c.diff',diff.stdout);
report.apiReplacedLines=diff.stdout.split('\n').filter(l=>l.startsWith('+')&&!l.startsWith('+++')).length;assert.equal(report.apiReplacedLines,2);
const arts=fs.readdirSync(a+'/built/local').filter(f=>f.endsWith('.js'));report.bundledJavascript=[];
for(const f of arts){assert(fs.readFileSync(a+'/built/local/'+f).equals(fs.readFileSync(b+'/built/local/'+f)),f);report.bundledJavascript.push({file:f,bytes:fs.statSync(a+'/built/local/'+f).size});}
console.log(JSON.stringify(report,null,2));
