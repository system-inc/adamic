// Create virtual source edits only. The actual adapted tree stays untouched.
const fs=require('node:fs'),path=require('node:path');
const [tree,inventory]=process.argv.slice(2);const m=JSON.parse(fs.readFileSync(inventory));
const overlay={},lines={};
for(const file of new Set(m.casts.map(c=>c.location.split(':')[0]))){
 let text=fs.readFileSync(path.join(tree,file),'utf8');
 for(const c of m.casts.filter(c=>c.location.startsWith(file+':')).sort((a,b)=>b.start-a.start)){
  if(text.slice(c.start,c.end)!==c.expression)throw Error('cast source drift');
  text=text.slice(0,c.start)+c.directCastText+text.slice(c.end);
  const line=c.location.split(':')[1];lines[path.join(tree,file)+':'+line]=true;
 }
 overlay[path.join(tree,file)]=text;
}
console.log(JSON.stringify({overlay,lines}));
