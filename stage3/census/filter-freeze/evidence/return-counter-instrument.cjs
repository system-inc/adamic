const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const root=process.argv[2], out=process.argv[3];
const old=JSON.parse(fs.readFileSync(path.join(__dirname,'evidence/previous-calls.json')));
const rows=JSON.parse(fs.readFileSync(path.join(__dirname,'evidence/resolved-calls.json')));
function narrow(r){return /:26[19]:1$/.test(r.declaration_where)}
const before=old.filter(narrow),after=rows.filter(narrow);
const key=r=>JSON.stringify([r.where,r.declaration_where,r.call]);
const diff={previous_total:old.length,current_total:rows.length,previous_narrowing:before.length,current_narrowing:after.length,removed:before.filter(r=>!after.some(s=>key(r)===key(s))),added:after.filter(r=>!before.some(s=>key(r)===key(s)))};
fs.writeFileSync(path.join(__dirname,'evidence/set-comparison.json'),JSON.stringify(diff,null,2)+'\n');
if(diff.removed.length||diff.added.length||rows.length!==157||after.length!==16)throw Error('STOP: changed set; report before proceeding');
fs.writeFileSync(path.join(__dirname,'evidence/sites.json'),JSON.stringify(after.map(r=>({site:r.where,call:r.call,argument:r.argument,overload:r.declaration_where.includes(':261:')?1:5})),null,2)+'\n');
const files=[...new Set(after.map(r=>r.where.split(':')[0]))],hashes=[];
for(const file of files){let text=fs.readFileSync(path.join(root,file),'utf8');const hash=t=>crypto.createHash('sha256').update(t).digest('hex');const original=hash(text);
for(const r of after.filter(r=>r.where.split(':')[0]===file).sort((a,b)=>b.start-a.start)){
if(text.slice(r.start,r.end)!==r.call)throw Error('span mismatch '+r.where);
// Capture the argument once in a local const, retaining its inferred type. The generic observer
// returns the original call type. The predicate is evaluated after the argument, as before.
const call=r.call;const arg=r.argument;const argStart=call.indexOf('(')+1;
if(call.slice(argStart,argStart+arg.length)!==arg)throw Error('argument span '+r.where);
const replacement=call.slice(0,argStart)+'__adamicFreezeArg'+call.slice(argStart+arg.length);
const wrapped='(() => { const __adamicFreezeArg = '+arg+'; return ((globalThis as any).__adamicFilterFreeze as <R>(site: string, result: R, argument: unknown) => R)('+JSON.stringify(r.where)+', '+replacement+', __adamicFreezeArg); })()';
text=text.slice(0,r.start)+wrapped+text.slice(r.end);
}
fs.writeFileSync(path.join(root,file),text);hashes.push({file,before:original,after:hash(text)});
}
fs.writeFileSync(out,JSON.stringify(hashes,null,2)+'\n');
console.log('Exact 16 confirmed; instrumented '+files.length+' source files; other 141 call expressions unchanged.');
