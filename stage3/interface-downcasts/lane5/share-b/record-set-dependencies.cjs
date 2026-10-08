// Assigned Set intrinsics depend on the user's completed receiver fix.
const fs=require('fs'),path=require('path');
const ts=require('/tmp/lane5-b-original/lib/typescript.js');
const ledger=require('../unknown-callable-pairs-ranked.json');
const dependencies=require('./code-dependencies.json');
const certified=new Set(require('./certified-pairs.json').map(p=>p.rank));
let count=0,reads=0;
for(const pair of ledger.filter(p=>p.rank%3===1&&!p.certified&&!certified.has(p.rank)&&/^(Readonly)?Set<|^ReadonlyMap.*ReadonlySet</.test(p.type))) {
 const source=ts.createSourceFile(pair.witness.file,fs.readFileSync(path.join('/tmp/lane5-b-original',pair.witness.file),'utf8'),ts.ScriptTarget.Latest,true);
 let read;
 function visit(node){if(ts.isPropertyAccessExpression(node)&&node.name.text===pair.field){const lc=source.getLineAndCharacterOfPosition(node.getStart(source));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column)read=node.getText(source);}ts.forEachChild(node,visit);}
 visit(source);if(!read)throw Error('missing Set read '+pair.rank);
 if(!dependencies.some(p=>p.rank===pair.rank))dependencies.push({rank:pair.rank,pair:pair.type+'.'+pair.field,reads:pair.reads,read,witness:pair.witness,stop:'Set receiver dependency absent from this delivery base; listed by user instruction, not rechecked against the fix',needed:'Set intrinsic receiver support from codex/views-set-receiver at 058635b9, then original-member fixtures and certification',dependencyBranch:'codex/views-set-receiver',dependencySha:'058635b9',basis:'user-provided dependency; no branch merge or certification claimed'});
 count++;reads+=pair.reads;
}
dependencies.sort((a,b)=>a.rank-b.rank);
fs.writeFileSync(path.join(__dirname,'code-dependencies.json'),JSON.stringify(dependencies,null,2)+'\n');
console.log(count+' assigned Set pairs / '+reads+' reads depend on 058635b9.');
