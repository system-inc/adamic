'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const spec=require('./proofs/watch-directory.json');
function rewrite(tree,mode='restore') {
 const file=path.join(tree,'src/compiler/sys.ts');let text=fs.readFileSync(file,'utf8');assert.equal(text.split(spec.after).length,2,'expected exactly one accepted forwarder');
 if(mode==='restore')text=text.replace(spec.after,spec.before);
 else {assert(['unnormalized','optional-contract'].includes(mode));text=text.replace(spec.after,spec.after.replace('!!recursive','recursive'));}
 fs.writeFileSync(file,text);
 if(mode==='optional-contract')for(const rel of ['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts','src/harness/watchUtils.ts']){
  const f=path.join(tree,rel);const raw=fs.readFileSync(f,'utf8');fs.writeFileSync(f,raw.replaceAll('recursive: boolean','recursive: boolean | undefined'));
 }
 return {mode,normalizationRemoved:mode!=='restore',wideReaderContracts:mode==='optional-contract'};
}
module.exports={rewrite};
if(require.main===module)console.log(JSON.stringify(rewrite(path.resolve(process.argv[2]),process.argv[3]||'restore'),null,2));
