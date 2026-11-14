'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict'),ts=require('typescript'),crypto=require('node:crypto');
const root='/workspace/scratch/binding-views',repo=root+'/repo',base=root+'/baseline';
const {patch}=require(root+'/patch.cjs');
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
const pin=(file,owner,before,after)=>({file,owner,before,after,ownerBeforeSha256:hash(before),ownerAfterSha256:hash(after)});
const json=(f,x)=>fs.writeFileSync(f,JSON.stringify(x,null,2)+'\n');
const h=repo+'/stage3/adapt/95-checker-axolotl-h',dir=repo+'/stage3/adapt/99-correctness-binding-views';fs.mkdirSync(dir,{recursive:true});
const res=fs.readFileSync(base+'/src/compiler/resolutionCache.ts','utf8').replace(/\r\n/g,'\n');
const prog=fs.readFileSync(base+'/src/compiler/program.ts','utf8').replace(/\r\n/g,'\n');
const resAfter=patch(res,'resolution-readonly'),progAfter=patch(prog,'diag-overloads');
const resolutionPin=pin('resolutionCache.ts','SourceFile (all edited owners and original interface)',res,resAfter);
function declaration(text,name){const sf=ts.createSourceFile('resolutionCache.ts',text,ts.ScriptTarget.Latest,true);const found=sf.statements.filter(n=>n.name?.text===name);assert.equal(found.length,1);return found[0].getText(sf);}
resolutionPin.editedOwners=['ResolutionCache','GetResolutionWithResolvedFileName','createResolutionCache'].map(name=>pin('resolutionCache.ts',name,declaration(res,name),declaration(resAfter,name)));
resolutionPin.insertedOwner=pin('resolutionCache.ts','ResolutionReadView','',declaration(resAfter,'ResolutionReadView'));
assert.equal(res.split('type ResolutionWithResolvedFileName,').length,2);
assert.equal(resAfter.split('type ResolutionWithResolvedFileName,').length,2);
json(dir+'/owners.json',[resolutionPin]);
json(dir+'/mutant-owners.json',[pin('resolutionCache.ts','SourceFile',res,resAfter),pin('program.ts','SourceFile',prog,progAfter)]);
fs.copyFileSync(root+'/engine.cjs',dir+'/engine.cjs');
fs.writeFileSync(dir+'/adapt.cjs',`'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const e = require('./engine.cjs');
const owners = require('./owners.json');
const programPin = require('./mutant-owners.json')[1];
function adapt(tree) {
    // createProgram's prior whole-owner pin makes diagnostics an amendment to 95h.
    assert(e.inspect(programPin, e.normalize(fs.readFileSync(path.join(tree, 'src/compiler/program.ts'), 'utf8'))), 'missing 95h diagnostics overload amendment');
    return e.rewrite(tree, owners);
}
module.exports = {adapt, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
`);
fs.writeFileSync(dir+'/mutant.cjs',`'use strict';
const assert = require('node:assert/strict');
const path = require('node:path');
const {rewrite} = require('./engine.cjs');
assert.equal(process.argv.length, 3, 'usage: node mutant.cjs TREE');
console.log(JSON.stringify(rewrite(path.resolve(process.argv[2]), require('./mutant-owners.json'), true), null, 2));
`);
function parse(text){const sf=ts.createSourceFile('program.ts',text,ts.ScriptTarget.Latest,true);assert.equal(sf.parseDiagnostics.length,0);return sf;}
function owner(text){const sf=parse(text),nodes=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='createProgram'&&n.body);assert.equal(nodes.length,1);return nodes[0].getText(sf);}
function helper(text){const sf=parse(text);let results=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getDiagnosticsHelper')results.push(n);ts.forEachChild(n,visit);}visit(sf);assert([1,3].includes(results.length));return text.slice(results[0].getStart(sf),results.at(-1).end);}
const git=ref=>cp.execFileSync('git',['-C',base,'show',ref+':src/compiler/program.ts'],{encoding:'utf8',maxBuffer:20e6});
const hBefore=git('refs/stage3/95-checker-axolotl-h-before'),hAfter=git('refs/stage3/95-checker-axolotl-h-after');
const newHAfter=patch(hAfter,'diag-overloads');
const inventory=JSON.parse(fs.readFileSync(h+'/sites.json')),row=inventory.literals.find(r=>r.owner==='createProgram');
assert.equal(hash(owner(hBefore)),row.ownerBeforeSha256);assert.equal(hash(owner(hAfter)),row.ownerAfterSha256);
const pins={owner:pin('program.ts','createProgram',owner(hBefore),owner(newHAfter)),helper:pin('program.ts','createProgram.getDiagnosticsHelper (overload group)',helper(hBefore),helper(newHAfter))};
json(h+'/diagnostics-overloads.json',pins);fs.copyFileSync(root+'/diagnostics-overloads.cjs',h+'/diagnostics-overloads.cjs');
row.ownerAfterSha256=pins.owner.ownerAfterSha256;json(h+'/sites.json',inventory);
let adapter=fs.readFileSync(h+'/adapt.cjs','utf8');
const needle="if(row.owner==='createProgram'){assert.equal(mm.state,state,'half-adapted MultiMap calls');undo.push(...mm.undo);}";
assert.equal(adapter.split(needle).length,2);
adapter=adapter.replace(needle,`if(row.owner==='createProgram'){
     assert.equal(mm.state,state,'half-adapted MultiMap calls');undo.push(...mm.undo);
     const amendment=require('./diagnostics-overloads.cjs');
     const binding=amendment.inspect(ts,source,owner);
     assert.equal(binding.state,state,'half-adapted diagnostics overload amendment');
     amendment.ownerText(owner.getText(source),state);
     if(state==='after')undo.push(binding.undo);
     else {assert(!check,'missing diagnostics overload amendment');plan(source,binding.edit);}
    }`);
fs.writeFileSync(h+'/adapt.cjs',adapter);
fs.writeFileSync(root+'/candidate-program.ts',progAfter.replace(/\n/g,'\r\n'));fs.writeFileSync(root+'/candidate-resolutionCache.ts',resAfter.replace(/\n/g,'\r\n'));
console.log(JSON.stringify({resolution:[hash(res),hash(resAfter)],program:[hash(prog),hash(progAfter)],hOwner:[pins.owner.ownerBeforeSha256,pins.owner.ownerAfterSha256]},null,2));
