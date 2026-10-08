const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const {owner,canonical}=require('./plan.cjs'),{rules}=require('./adapt.cjs');
function check(root){
 const sites=[],expressions=new Set();
 for(const file of new Set(rules.map(r=>r.file))){const source=ts.createSourceFile(file,fs.readFileSync(path.join(root,file),'utf8'),99,true);function visit(n){if(ts.isExpressionNode(n))expressions.add(file+'|'+owner(n)+'|'+canonical(n,source));if(ts.isNonNullExpression(n))sites.push({file,owner:owner(n),text:canonical(n,source)});ts.forEachChild(n,visit);}visit(source);}
 for(const r of rules){assert.ok(expressions.has(r.file+'|'+r.owner+'|'+r.after),'missing owning expression: '+r.id);assert.ok(!sites.some(s=>s.file===r.file&&s.owner===r.owner&&s.text===r.before),'restored type lie: '+r.id);}
 return {repaired_observed_sites:rules.length,remaining_observed_sites:22-rules.length};
}
if(require.main===module)console.log(JSON.stringify(check(path.resolve(process.argv[2])),null,2));
module.exports=check;
