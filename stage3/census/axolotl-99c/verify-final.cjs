const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto'),cp=require('child_process');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const adapter=require('/workspace/adamic/stage3/adapt/99-correctness-axolotl-c/adapt.cjs');
const a='/workspace/scratch/before',b='/workspace/scratch/after-final';
const files=ts.sys.readDirectory(a+'/src',['.ts','.tsx','.json']);
const changed=files.filter(f=>!fs.readFileSync(f).equals(fs.readFileSync(b+f.slice(a.length))));
assert.deepEqual(changed.map(f=>f.slice(a.length)).sort(),['/src/compiler/commandLineParser.ts','/src/compiler/types.ts']);
const options={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments:false};
const report={changedFiles:changed.map(f=>f.slice(a.length+1)),javascript:[]};
const original=new Map();
for(const f of changed){const relative=f.slice(a.length);const before=fs.readFileSync(f,'utf8'),after=fs.readFileSync(b+relative,'utf8');original.set(relative,after);
 const js=x=>ts.transpileModule(x,{compilerOptions:options,fileName:relative}).outputText;
 assert.equal(js(before),js(after));report.javascript.push({file:relative.slice(1),bytes:Buffer.byteLength(js(before)),sha256:crypto.createHash('sha256').update(js(before)).digest('hex')});}
assert.deepEqual(adapter.adapt(b),{files:0,declarations:0,javascriptChanging:false});
for(const [f,text] of original)assert.equal(fs.readFileSync(b+f,'utf8'),text);
report.idempotence=true;
const tmp=fs.mkdtempSync('/workspace/scratch/contracts-final-');fs.mkdirSync(tmp+'/src/compiler',{recursive:true});
const fixture=new Map([...original].map(([f])=>[f,fs.readFileSync(a+f,'utf8')]));
const variants={partial:['/src/compiler/types.ts','project?: string;','project?: string | undefined;'],lastFileDrift:['/src/compiler/commandLineParser.ts','include?: readonly string[];','include?: readonly number[];'],lastFileDuplicate:['/src/compiler/commandLineParser.ts','', '\ninterface TSConfig {}\n']};
report.refusals={};
for(const [name,[file,needle,replacement]] of Object.entries(variants)){
 for(const [f,text] of fixture)fs.writeFileSync(tmp+f,text);
 const text=fixture.get(file),mutated=needle?text.replace(needle,replacement):text+replacement;fs.writeFileSync(tmp+file,mutated);
 const snapshots=new Map([...fixture].map(([f])=>[f,fs.readFileSync(tmp+f)]));
 assert.throws(()=>adapter.adapt(tmp));for(const [f,bytes] of snapshots)assert(fs.readFileSync(tmp+f).equals(bytes));report.refusals[name]=true;
}
for(const [f] of fixture)fs.writeFileSync(tmp+f,fs.readFileSync('/workspace/scratch/upstream'+f));
assert.deepEqual(adapter.adapt(tmp),{files:2,declarations:12,javascriptChanging:false});assert.deepEqual(adapter.adapt(tmp),{files:0,declarations:0,javascriptChanging:false});report.pristineAndIdempotence=true;
const expected=JSON.parse(fs.readFileSync('/workspace/scratch/api-expected.json')).expected;
assert.equal(fs.readFileSync(b+'/built/local/typescript.d.ts','utf8'),expected);
const diff=cp.spawnSync('diff',['-u',a+'/built/local/typescript.d.ts',b+'/built/local/typescript.d.ts'],{encoding:'utf8'});fs.writeFileSync('/workspace/scratch/api-99c-final.diff',diff.stdout);
report.apiReplacedLines=diff.stdout.split('\n').filter(l=>l.startsWith('+')&&!l.startsWith('+++')).length;assert.equal(report.apiReplacedLines,2);
const arts=fs.readdirSync(a+'/built/local').filter(f=>f.endsWith('.js'));report.bundledJavascript=[];
for(const f of arts){assert(fs.readFileSync(a+'/built/local/'+f).equals(fs.readFileSync(b+'/built/local/'+f)),f);report.bundledJavascript.push({file:f,bytes:fs.statSync(a+'/built/local/'+f).size});}
console.log(JSON.stringify(report,null,2));
