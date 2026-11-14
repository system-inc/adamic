const ts=require('typescript'),fs=require('fs'),path=require('path'),crypto=require('crypto'),assert=require('assert/strict');
const root=__dirname;let rows=[];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const base=path.join(root,'baseline'),change=path.join(root,'candidate');
const files=ts.sys.readDirectory(path.join(base,'src'),undefined).sort();let changed=[];
for(const f of files){const rel=path.relative(base,f),a=fs.readFileSync(f),b=fs.readFileSync(path.join(change,rel));if(!a.equals(b))changed.push(rel);}
assert.deepEqual(changed,['src/compiler/sys.ts','src/harness/watchUtils.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts']);
for(const f of ts.sys.readDirectory(path.join(base,'src'),['.ts'])) {if(f.endsWith('.d.ts'))continue;const rel=path.relative(base,f);const emit=tree=>ts.transpileModule(fs.readFileSync(path.join(tree,rel),'utf8'),{fileName:rel,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,verbatimModuleSyntax:true,sourceMap:false}}).outputText;const a=emit(base),b=emit(change);assert.equal(a,b,rel);rows.push({file:rel,sha256:hash(a),byteIdentical:true});}
fs.writeFileSync(path.join(root,'javascript.json'),JSON.stringify({sourceFilesCompared:files.length,changedSourceFiles:changed,allSourceOutputs:rows,otherSources:'byte-identical source inputs, therefore identical deterministic emissions'},null,2)+'\n');
