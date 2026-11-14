const ts=require('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript'),fs=require('fs'),path=require('path');
const base='/tmp/destructure-price-scratch/adapted',out='/workspace/scratch/destructure-price-f2989762';
const sites=JSON.parse(fs.readFileSync('/tmp/destructure-price-scratch/ast-sites.json')),pub=JSON.parse(fs.readFileSync('/tmp/destructure-price-scratch/public-types.json')).names.filter(x=>x.typescript).map(x=>x.name);
const cfg=ts.parseJsonConfigFileContent(ts.readConfigFile(base+'/src/compiler/tsconfig.json',ts.sys.readFile).config,ts.sys,base+'/src/compiler');
const program=ts.createProgram(cfg.fileNames,cfg.options),checker=program.getTypeChecker(),printer=ts.createPrinter();
for(const road of ['properties','hybrid']){const edits=new Map(),groups=new Map(),originals=new Map();
 const add=(file,e)=>{if(!edits.has(file))edits.set(file,[]);edits.get(file).push(e);};
 for(const site of sites){const sf=program.getSourceFile(path.join(base,site.file)),isPublic=pub.includes(site.type.split('<')[0]);let pattern;
 function find(n){if(ts.isObjectBindingPattern(n)&&sf.getLineAndCharacterOfPosition(n.getStart()).line+1===site.line)pattern=n;ts.forEachChild(n,find);}find(sf);if(!pattern)throw Error('missing site');
 const type=checker.getTypeAtLocation(pattern.parent.initializer||pattern);
 if(road==='hybrid'&&isPublic){const init=pattern.parent.initializer;if(!init||!ts.isVariableDeclaration(pattern.parent))throw Error('unexpected public binding');const receiver='__destructureReceiver'+site.line,parts=[receiver+' = '+init.getText(sf)];
 for(const [index,e] of pattern.elements.entries()){const name=(e.propertyName||e.name).getText(sf),local=e.name.getText(sf),member=site.members.find(m=>m.name===name),get=receiver+'.'+name,original='__destructureOriginal'+site.line+'_'+index;
 const optional=!!type.getProperty(name)?.declarations?.some(d=>!!d.questionToken);
 const originalValue=e.initializer?`${get} === undefined ? ${e.initializer.getText(sf)} : ${get}`:get;
 parts.push(original+' = '+originalValue);let value=original;
 if(member.method){originals.set(checker.getSymbolAtLocation(e.name),original);const params=`Parameters<NonNullable<typeof ${original}>>`,callable=`NonNullable<typeof ${original}>`;
 const supplied=`((...args: ${params}) => ${get}!(...args)) as ${callable}`;
 if(e.initializer){const fallback=`((...args: ${params}) => ${e.initializer.getText(sf)}(...args)) as ${callable}`;value=`${get} === undefined ? (${fallback}) : (${supplied})`;}
 else value=optional?`${original} === undefined ? undefined : (${supplied})`:supplied;
 }
 parts.push(local+' = '+value);}
 add(sf,{start:pattern.parent.getStart(sf),end:pattern.parent.end,text:parts.join(',\n')});
 }else{for(const member of site.members.filter(m=>m.method)){const sym=type.getProperty(member.name);for(const d of sym.declarations||[])if(ts.isMethodSignature(d)){const key=d.getSourceFile().fileName+':'+d.parent.pos+':'+member.name;if(!groups.has(key))groups.set(key,[]);if(!groups.get(key).includes(d))groups.get(key).push(d);}}}
 }
 for(const sf of program.getSourceFiles())if(sf.fileName.startsWith(base+'/src/compiler/')){function identity(n){if(ts.isBinaryExpression(n)&&[ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(n.operatorToken.kind))for(const operand of [n.left,n.right])if(ts.isIdentifier(operand)){const original=originals.get(checker.getSymbolAtLocation(operand));if(original)add(sf,{start:operand.getStart(),end:operand.end,text:original});}ts.forEachChild(n,identity);}identity(sf);}
 for(const declarations of groups.values()){const ds=declarations.sort((a,b)=>a.pos-b.pos),first=ds[0],sf=first.getSourceFile(),last=ds.at(-1);if(ds.some(d=>d.parent!==first.parent))throw Error('crossparent overload');const range=first.parent.members.filter(d=>d.pos>=first.pos&&d.pos<=last.pos);if(range.length!==ds.length)throw Error('noncontiguous overload');
 let type;if(ds.length===1)type=printer.printNode(ts.EmitHint.Unspecified,ts.factory.createFunctionTypeNode(first.typeParameters,first.parameters,first.type),sf);
 else{let inner='';for(let i=0;i<ds.length;i++){const d=ds[i];if(i)inner+=sf.text.slice(ds[i-1].end,d.getStart(sf));inner+=d.getText(sf).slice((d.questionToken||d.name).end-d.getStart(sf));}type='{ '+inner+' }';}
 add(sf,{start:first.getStart(sf),end:last.end,text:first.name.getText(sf)+(first.questionToken?'?':'')+': '+type+';'});
 }
 for(const [sf,es] of edits){let text=sf.text;const unique=[...new Map(es.map(e=>[e.start,e])).values()].sort((a,b)=>b.start-a.start);for(const e of unique)text=text.slice(0,e.start)+e.text+text.slice(e.end);if(road==='properties'){const options={target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.ESNext,removeComments:true};if(ts.transpileModule(sf.text,{compilerOptions:options}).outputText!==ts.transpileModule(text,{compilerOptions:options}).outputText)throw Error('property road changed JavaScript '+sf.fileName);}
 fs.writeFileSync(path.join(out,road,path.relative(base,sf.fileName)),text);}
 console.log(JSON.stringify({road,files:edits.size,propertyGroups:groups.size}));
}
