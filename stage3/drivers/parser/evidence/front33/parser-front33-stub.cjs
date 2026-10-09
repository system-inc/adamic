const fs=require('fs'),ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const [file,line,column]=process.argv.slice(2);const text=fs.readFileSync(file,'utf8');const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const pos=sf.getPositionOfLineAndCharacter(+line-1,+column-1);let candidates=[];
function walk(n){if(n.pos<=pos&&pos<n.end){if(ts.isFunctionDeclaration(n)||ts.isMethodDeclaration(n)||ts.isFunctionExpression(n)||ts.isArrowFunction(n))candidates.push(n);ts.forEachChild(n,walk);}}
walk(sf);const n=candidates.at(-1);if(!n){
 let variable;function find(x){if(x.pos<=pos&&pos<x.end){if(ts.isVariableDeclaration(x))variable=x;ts.forEachChild(x,find);}}find(sf);
 if(!variable?.initializer)throw Error('No function or variable stub for diagnostic; structural declaration stop');
 const init=variable.initializer;
 const start=variable.type ? init.getStart(sf) : variable.getStart(sf),end=init.end;
 let annotation=variable.type ? '' : variable.name.getText(sf)+': '+init.expression.getText(sf)+(init.typeArguments?.length ? '<'+init.typeArguments.map(t=>t.getText(sf)).join(', ')+ '>' : '')+' = ';
 const checker=ts.createProgram([file], {target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true,strict:true,noUncheckedIndexedAccess:true}).getTypeChecker();
 const live=checker.getProgram ? undefined : undefined;
 const program=ts.createProgram([file], {target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true,strict:true,noUncheckedIndexedAccess:true});
 const checked=program.getSourceFile(file);let checkedVariable;function seek(x){if(ts.isVariableDeclaration(x)&&x.name.getText(checked)===variable.name.getText(sf))checkedVariable=x;ts.forEachChild(x,seek);}seek(checked);
 const valueType=variable.type ? variable.type.getText(sf) : program.getTypeChecker().typeToString(program.getTypeChecker().getTypeAtLocation(checkedVariable),checkedVariable,ts.TypeFormatFlags.NoTruncation);
 annotation=variable.type ? '' : variable.name.getText(sf)+': '+valueType+' = ';
 const helper='parserProofMissing'+variable.name.getText(sf);
 const replacement=annotation+helper+'()'+ '\n'.repeat((text.slice(start,end).match(/\n/g)||[]).length);
 fs.writeFileSync(file,text.slice(0,start)+replacement+text.slice(end)+'\nfunction '+helper+'(): '+valueType+' { throw \"discovery-only initializer\"; }\n');
 console.log(JSON.stringify({file,line:+line,column:+column,name:variable.name.getText(sf),start,end,replacement}));process.exit(0);
}
let edits=[];function predicate(t){if(ts.isTypePredicateNode(t))edits.push({start:t.getStart(sf),end:t.end,replacement:'boolean'});else ts.forEachChild(t,predicate);}
for(const p of n.parameters)if(p.type)predicate(p.type);
if(n.body)edits.push({start:n.body.getStart(sf),end:n.body.end,replacement:'{ throw "scratch-only discovery stub";'+ '\n'.repeat((text.slice(n.body.getStart(sf),n.body.end).match(/\n/g)||[]).length)+'}'});
else edits.push({start:n.getStart(sf),end:n.end,replacement:'\n'.repeat((text.slice(n.getStart(sf),n.end).match(/\n/g)||[]).length)});
let output=text;for(const e of edits.sort((a,b)=>b.start-a.start))output=output.slice(0,e.start)+e.replacement+output.slice(e.end);
if(output===text)throw Error('Stub does not change source');fs.writeFileSync(file,output);console.log(JSON.stringify({file,line:+line,column:+column,name:n.name?.getText(sf),edits}));
