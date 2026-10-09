const ts=require('typescript'),assert=require('node:assert/strict');
const printer=ts.createPrinter({removeComments:true});
function owner(n){const a=[];for(let p=n.parent;p;p=p.parent)if(ts.isFunctionLike(p)&&p.name&&ts.isIdentifier(p.name))a.unshift(p.name.text);return a.join('/');}
function canonical(n,source){return printer.printNode(ts.EmitHint.Expression,n,source);}
function removeAssertions(file,text,rules){
 const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),edits=[],matches=[];
 const groups=new Map();for(const rule of rules){const k=rule.owner+'|'+rule.before;if(!groups.has(k))groups.set(k,rule);}
 function visit(n){if(ts.isNonNullExpression(n)){const rule=groups.get(owner(n)+'|'+canonical(n,source));if(rule){const bang=n.getLastToken(source);assert.equal(bang.kind,ts.SyntaxKind.ExclamationToken);edits.push({start:bang.getStart(source),end:bang.end,text:''});matches.push(rule.id);}}ts.forEachChild(n,visit);}visit(source);
 let next=text;for(const e of edits.sort((a,b)=>b.start-a.start))next=next.slice(0,e.start)+e.text+next.slice(e.end);
 for(const removeComments of [false,true]){const compilerOptions={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};assert.equal(ts.transpileModule(next,{compilerOptions}).outputText,ts.transpileModule(text,{compilerOptions}).outputText,'JavaScript must remain byte-identical');}
 return {text:next,edits:edits.length,matches};
}
module.exports={owner,canonical,removeAssertions};
