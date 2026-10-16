// Classify the observed sites and resolve literal placeholder slots through the compiler API.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const root=path.resolve(process.argv[2]);assert.equal(ts.version,'6.0.3');
const configPath=path.join(root,'src/compiler/tsconfig.json'),read=ts.readConfigFile(configPath,ts.sys.readFile),config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);
const p=ts.createProgram([path.join(root,'src/tsc/tsc.ts')],{...config.options,noEmit:true,composite:false,isolatedDeclarations:false}),checker=p.getTypeChecker();
const observed=JSON.parse(fs.readFileSync(path.join(__dirname,'runtime-sites.json'))).filter(r=>r.nullish),inventory=JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json')));
const nodes=new Map(),files=p.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/')&&!f.isDeclarationFile);
for(const file of files){function v(n){if(ts.isNonNullExpression(n))nodes.set(path.relative(root,file.fileName)+':'+n.getStart(file)+':'+n.end,n);ts.forEachChild(n,v);}v(file);}
const rows=[],slots=[],bindings=new Map();
function strip(n){while(ts.isParenthesizedExpression(n)||ts.isAsExpression(n)||ts.isTypeAssertionExpression(n))n=n.expression;return n;}
function loc(n){const f=n.getSourceFile(),pos=f.getLineAndCharacterOfPosition(n.getStart(f));return {file:path.relative(root,f.fileName),line:pos.line+1,column:pos.character+1,start:n.getStart(f),end:n.end};}
function binding(n){const symbol=checker.getSymbolAtLocation(n);assert.ok(symbol?.valueDeclaration,'unresolved placeholder binding');const d=symbol.valueDeclaration;const l=loc(d);const key=l.file+':'+l.start+':'+l.end;bindings.set(symbol,{id:key,...l,name:n.text});return key;}
for(const row of observed){
 const original=inventory.find(s=>s.file+':'+s.line+':'+s.column+'@'+s.start+'-'+s.end===row.site);assert.ok(original);
 const node=nodes.get(original.file+':'+original.start+':'+original.end);assert.ok(node);const operand=strip(node.expression);
 const literal=operand.kind===ts.SyntaxKind.NullKeyword||(ts.isIdentifier(operand)&&operand.text==='undefined'&&(checker.getTypeAtLocation(operand).flags&ts.TypeFlags.Undefined));
 const category=literal?'1-placeholder':ts.isPropertyAccessExpression(operand)||ts.isElementAccessExpression(operand)||ts.isCallExpression(operand)?'2-nullish-read-or-call':'3-other-nullish-unwrap';
 let owner=node;while(owner.parent&&(ts.isAsExpression(owner.parent)||ts.isParenthesizedExpression(owner.parent)))owner=owner.parent;
 let slot;
 if(literal){
  const parent=owner.parent;
  if(ts.isBinaryExpression(parent)&&parent.right===owner&&parent.operatorToken.kind===ts.SyntaxKind.EqualsToken){
   const target=strip(parent.left);
   if(ts.isIdentifier(target))slot={kind:'variable',binding:binding(target),target:target.getText(),initializer:loc(parent)};
   else if(ts.isPropertyAccessExpression(target))slot={kind:'field',key:target.name.text,target:target.getText(),initializer:loc(parent)};
   else if(ts.isElementAccessExpression(target))slot={kind:'element',target:target.getText(),initializer:loc(parent)};
  }else if(ts.isPropertyAssignment(parent))slot={kind:'object-field',key:ts.isIdentifier(parent.name)||ts.isStringLiteral(parent.name)?parent.name.text:parent.name.getText(),target:parent.name.getText(),initializer:loc(parent.parent),property:loc(parent)};
  else if(ts.isArrayLiteralExpression(parent))slot={kind:'array-element',key:String(parent.elements.indexOf(owner)),target:'array['+parent.elements.indexOf(owner)+']',initializer:loc(parent)};
  else if(ts.isVariableDeclaration(parent))slot={kind:'variable',binding:binding(parent.name),target:parent.name.getText(),initializer:loc(parent)};
  assert.ok(slot,'unhandled literal placeholder '+row.site);
  slots.push({...original,id:row.site,slot});
 }
 let fn=node.parent;while(fn&&!ts.isFunctionLike(fn))fn=fn.parent;
 rows.push({...original,id:row.site,category,type_lie:!literal,nullish_count:row.nullish,nullish_input_set:row.nullish_input_set,node_passes:row.null?'null and/or undefined':'undefined',owner:fn?.name?.getText()||'<closure or module>',...(slot?{slot}: {})});
}
const references=[];
for(const file of files){function v(n){
 if(ts.isIdentifier(n)){
  let symbol=ts.isShorthandPropertyAssignment(n.parent)?checker.getShorthandAssignmentValueSymbol(n.parent):checker.getSymbolAtLocation(n),b=bindings.get(symbol);
  if(b){
   const parent=n.parent;
   const declaration=(ts.isVariableDeclaration(parent)||ts.isParameter(parent)||ts.isBindingElement(parent))&&parent.name===n;
   let type=false;for(let a=n.parent;a&&!ts.isStatement(a)&&!ts.isExpression(a);a=a.parent)if(ts.isTypeNode(a))type=true;
   if(!declaration&&!type)references.push({...loc(n),binding:b.id,shorthand:ts.isShorthandPropertyAssignment(parent)});
  }
 }
 ts.forEachChild(n,v);
 }v(file);}
const counts=rows.reduce((a,r)=>(a[r.category]=(a[r.category]||0)+1,a),{});
const result={ruling:'literal undefined! and null! are placeholders, including resets; every observed nonliteral nullish unwrap is a type lie',sites:rows.length,counts,type_lies:rows.filter(r=>r.type_lie).length,bindings:[...bindings.values()],variable_references:references.length};
for(const [name,value] of Object.entries({'split.json':rows,'placeholder-slots.json':{slots,bindings:[...bindings.values()],references},'split-summary.json':result}))fs.writeFileSync(path.join(__dirname,name),JSON.stringify(value,null,2)+'\n');
fs.writeFileSync(path.join(__dirname,'SPLIT.tsv'),'category\tfile\tline\tcolumn\texpression\tNode value\tinput set\n'+rows.map(r=>[r.category,r.file,r.line,r.column,r.expression.replace(/\s+/g,' '),r.node_passes,r.nullish_input_set].join('\t')).join('\n')+'\n');
console.log(JSON.stringify(result,null,2));
