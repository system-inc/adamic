const fs=require('node:fs'),path=require('node:path'),ts=require('typescript'),assert=require('node:assert/strict');
const repo=path.resolve(__dirname,'../../../../..');
const [before,after]=process.argv.slice(2);
const rules=require(repo+'/stage3/adapt/40-explicit-any/class-rules.json')['enum-display'].filter(r=>r.parentKind==='AsExpression'&&r.id!==106);
const old=JSON.parse(require('node:child_process').execFileSync('git',['show','63e14a0b2b61cf251ad74292e42d5fd09af804b3:stage3/adapt/40-explicit-any/class-rules.json'],{cwd:repo,encoding:'utf8'}))['enum-display'];
const rows=rules.map(r=>({id:r.id,file:r.file,key:JSON.parse(r.replacement.match(/Record<("[^"]+")/)[1]),original_line:r.line}));
for(const [label,tree] of [['before',before],['after',after]]) {
 const consumed=new Set();
 for(let i=0;i<rules.length;i++) {
  const r=rules[i],expected=label==='before'?old.find(x=>x.id===r.id).replacement:r.replacement;
  const file=path.join(tree,r.file),s=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);let found;
  function visit(n){if(!found&&ts.isAsExpression(n)&&n.expression.getText(s)==='ts'&&n.type.getText(s)===expected&&!consumed.has(file+':'+n.pos)){found=n;return;}ts.forEachChild(n,visit);}visit(s);
  assert(found,'missing '+label+' '+r.id);consumed.add(file+':'+found.pos);
  const p=s.getLineAndCharacterOfPosition(found.expression.getStart(s));
  let owner=found.parent;while(owner&&!ts.isFunctionDeclaration(owner))owner=owner.parent;
  rows[i][label]={where:file+':'+(p.line+1)+':'+(p.character+1),owner:owner?.name?.text,type:expected};
 }
}
console.log(JSON.stringify(rows,null,2));
