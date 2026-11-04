const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),os=require('node:os'),crypto=require('node:crypto');
const adapter=require('./adapt.cjs');
assert.equal(process.argv.length,3,'usage: node contract.cjs BEFORE_TREE');
const before=path.resolve(process.argv[2]),root=fs.mkdtempSync(path.join(os.tmpdir(),'99j-contract-'));
const files=[...new Set(adapter.owners.map(o=>o.file))];
function tree(name){const dest=path.join(root,name);for(const file of files){const out=path.join(dest,'src/compiler',file);fs.mkdirSync(path.dirname(out),{recursive:true});fs.copyFileSync(path.join(before,'src/compiler',file),out);}return dest;}
function state(dest){return files.map(file=>{const p=path.join(dest,'src/compiler',file),stat=fs.statSync(p,{bigint:true});return [file,crypto.createHash('sha256').update(fs.readFileSync(p)).digest('hex'),String(stat.mtimeNs)];});}
const first=tree('first');let writes=0;const write=fs.writeFileSync;fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};
const firstResult=adapter.adapt(first);assert.equal(writes,6);assert.equal(firstResult.edits,12);assert.equal(firstResult.files,6);
const after=state(first);writes=0;const secondResult=adapter.adapt(first);assert.equal(writes,0);assert.equal(secondResult.edits,0);assert.equal(secondResult.files,0);assert.deepEqual(state(first),after);
fs.writeFileSync=write;
const fixtures=[];
function refused(name,mutate){const dest=tree(name);mutate(dest);const snapshot=state(dest);writes=0;fs.writeFileSync=function(...args){writes++;return write.apply(this,args);};assert.throws(()=>adapter.adapt(dest));assert.equal(writes,0);assert.deepEqual(state(dest),snapshot);fs.writeFileSync=write;fixtures.push(name);}
for(let i=0;i<adapter.sites.length;i++)refused('half-statement-'+i,dest=>{const site=adapter.sites[i],p=path.join(dest,'src/compiler',site.file);let text=fs.readFileSync(p,'utf8');const crlf=text.includes('\r\n');text=text.replace(/\r\n/g,'\n').replace(site.before,site.after);fs.writeFileSync(p,crlf?text.replace(/\n/g,'\r\n'):text);});
for(const file of files)refused('mixed-'+file.replaceAll('/','-'),dest=>fs.copyFileSync(path.join(first,'src/compiler',file),path.join(dest,'src/compiler',file)));
refused('changed-owner',dest=>{const p=path.join(dest,'src/compiler/sys.ts');const text=fs.readFileSync(p,'utf8');fs.writeFileSync(p,text.replace('(existing as string[]).push(...fileNames);','/* changed owner */ (existing as string[]).push(...fileNames);'));});
const unrelated=tree('unrelated-edit');const unrelatedPath=path.join(unrelated,'src/compiler/sys.ts');const addition='\r\nfunction unrelated99jFixture() { return 99; }\r\n';fs.appendFileSync(unrelatedPath,addition);const accepted=adapter.adapt(unrelated);assert.equal(accepted.edits,12);assert(fs.readFileSync(unrelatedPath,'utf8').endsWith(addition));const acceptedState=state(unrelated);const acceptedSecond=adapter.adapt(unrelated);assert.equal(acceptedSecond.edits,0);assert.deepEqual(state(unrelated),acceptedState);
refused('parse-error',dest=>fs.appendFileSync(path.join(dest,'src/compiler/transformers/generators.ts'),'\nfunction {'));
const left=[['serializeOptionBaseObject','result.set(name, (value as string[]).map(v => getRelativePathFromFile(pathOptions.configFilePath, getNormalizedAbsolutePath(v, getDirectoryPath(pathOptions.configFilePath)), getCanonicalFileName!)));'],['handleOptionConfigDirTemplateSubstitution','const listResult = getSubstitutedStringArrayWithConfigDirTemplate(value as string[], basePath);']];
const parser=fs.readFileSync(path.join(first,'src/compiler/commandLineParser.ts'),'utf8');for(const [owner,text] of left){assert(parser.includes(text));assert(!adapter.sites.some(s=>s.owner===owner&&s.before===text));}
assert.equal(crypto.createHash('sha256').update(fs.readFileSync(path.join(__dirname,'array-witness.json'))).digest('hex'), 'd02b1b08f6f2588385a897d700673f578f96efbba598f49ccbdbe3c4ef7b8ac7');
assert(fs.readFileSync(path.join(first,'src/compiler/builder.ts'),'utf8').includes('return relativeToBuildInfoEnsuringAbsolutePath(value as string);'));
console.log(JSON.stringify({first:firstResult,second:secondResult,noWriteSecondRun:true,refusalFixtures:fixtures,unrelatedEditAccepted:true,unrelatedSecondRun:acceptedSecond,leftStatementsUntouched:true,witnessByteIdentical:true,tree:first},null,2));
