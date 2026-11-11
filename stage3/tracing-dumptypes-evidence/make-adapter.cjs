const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),assert=require('node:assert/strict'),ts=require('typescript');
const dir='/workspace/adamic/stage3/adapt/99-correctness-tracing-dumptypes';
assert.equal(JSON.parse(fs.readFileSync('/workspace/scratch/dp24r5m-evidence/step1-results.json')).writtenValueDifferences,0);
function owner(file){const text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),found=[];function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='dumpTypes')found.push(n);ts.forEachChild(n,walk);}walk(sf);assert.equal(found.length,1);return found[0].getText(sf);}
const before=owner('/workspace/scratch/dp24r5m-base/src/compiler/tracing.ts'),after=owner('/workspace/scratch/dp24r5m-flags-road/src/compiler/tracing.ts');
const sha=s=>crypto.createHash('sha256').update(s).digest('hex');fs.mkdirSync(dir);
fs.writeFileSync(path.join(dir,'pins.json'),JSON.stringify({version:'6.0.3',before,beforeSha256:sha(before),after,afterSha256:sha(after)},null,2)+'\n');
