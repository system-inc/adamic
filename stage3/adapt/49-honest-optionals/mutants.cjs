const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),assert=require('node:assert/strict'),ts=require('typescript');
const {owner,canonical}=require('./plan.cjs'),{rules}=require('./adapt.cjs'),check=require('./check.cjs');
const root=path.resolve(process.argv[2]),scratch=fs.mkdtempSync(path.join(os.tmpdir(),'49-mutants-')),rows=[];
for(const file of new Set(rules.map(r=>r.file))){fs.mkdirSync(path.dirname(path.join(scratch,file)),{recursive:true});fs.copyFileSync(path.join(root,file),path.join(scratch,file));}
check(scratch);
for(const r of rules){const file=path.join(scratch,r.file),original=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,original,99,true);let at,count=0;function visit(n){if(ts.isExpressionNode(n)&&owner(n)===r.owner&&canonical(n,source)===r.after&&require('./site-anchor.cjs')(n,r)){at=n.end;count++;}ts.forEachChild(n,visit);}visit(source);assert.equal(count,1,r.id);assert.ok(at,r.id);fs.writeFileSync(file,original.slice(0,at)+'!'+original.slice(at));assert.throws(()=>check(scratch),/restored type lie/);fs.writeFileSync(file,original);rows.push({site:r.id,mutation:'restore the removed nullish unwrap',caught:'owning-expression census'});}
console.log(JSON.stringify({mutants:rows},null,2));
