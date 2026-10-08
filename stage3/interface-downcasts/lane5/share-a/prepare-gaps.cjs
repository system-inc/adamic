const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require(process.argv[3]),pin=process.argv[2];
const originalFile='src/compiler/types.ts',text=fs.readFileSync(path.join(pin,originalFile),'utf8');
const original=ts.createSourceFile(originalFile,text,ts.ScriptTarget.Latest,true);
const factory=[];function collect(n){if(ts.isInterfaceDeclaration(n)&&n.name.text==='NodeFactory')factory.push(...n.members);ts.forEachChild(n,collect);}collect(original);
const pairs=JSON.parse(fs.readFileSync(path.join(__dirname,'../unknown-callable-pairs-ranked.json')));
const rows=[];
for(const rank of [3,9,12,24,174,186,189,195]){
 const pair=pairs.find(p=>p.rank===rank);if(pair.type!=='NodeFactory')continue;
 const declarations=factory.filter(n=>n.name?.getText(original)===pair.field);if(!declarations.length)throw Error('missing '+rank);
 const sf=ts.createSourceFile(pair.witness.file,fs.readFileSync(path.join(pin,pair.witness.file),'utf8'),ts.ScriptTarget.Latest,true);let read;
 function visit(n){if(ts.isPropertyAccessExpression(n)&&n.name.text===pair.field){const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column)read=n;}ts.forEachChild(n,visit);}visit(sf);if(!read)throw Error('missing read '+rank);
 let carriers='',arg='"ok"',result='Identifier',producer='(text:string):Identifier=>({value:7})';
 if(rank===9){result='StringLiteral';producer='(text:string):StringLiteral=>({value:7})';}
 if(rank===12){result='Node';arg='{value:7}';producer='(node:Node):Node=>node';}
 if(rank===174){result='ModifierToken<ModifierSyntaxKind>';arg='1';producer='(kind:ModifierSyntaxKind):ModifierToken<ModifierSyntaxKind>=>({value:7})';carriers+='type ModifierSyntaxKind=number;\ninterface ModifierToken<T> {readonly value:number;}\n';}
 if(rank===186){result='ImportClause';arg='false,undefined,undefined';producer='(isTypeOnly:boolean,name:Identifier|undefined,namedBindings:NamedImportBindings|undefined):ImportClause=>({value:7})';}
 if(rank===189){result='YieldExpression';arg='undefined,undefined';producer='(asteriskToken:undefined,expression:Expression|undefined):YieldExpression=>({value:7})';}
 const decl=declarations.map(n=>n.getText(original)).join('\n');
 const refs=new Set();for(const n of declarations){function refsVisit(c){if(ts.isTypeReferenceNode(c)&&ts.isIdentifier(c.typeName))refs.add(c.typeName.text);ts.forEachChild(c,refsVisit);}refsVisit(n);}
 for(const name of refs){if(['T','TKind','ModifierToken','ModifierSyntaxKind'].includes(name))continue;if(['SyntaxKind','GeneratedIdentifierFlags','ImportPhaseModifierSyntaxKind'].includes(name))carriers+='type '+name+'=number;\n';else if(name==='Node')carriers+='interface Node {readonly value:number;}\n';else carriers+='interface '+name+' {readonly value:number;}\n';}
 if(!refs.has(result)&&!result.includes('<'))carriers+='interface '+result+' {readonly value:number;}\n';
 const receiver=read.expression.getText(sf);if(!/^[A-Za-z_$][\w$]*$/.test(receiver))throw Error('receiver '+rank);
 const directory='gap-'+rank;fs.mkdirSync(path.join(__dirname,directory),{recursive:true});
 fs.writeFileSync(path.join(__dirname,directory,'good.a'),'// Original overloaded/generic declarations and read; adjacent carriers reduced.\n'+carriers+'interface Base {readonly '+pair.field+':unknown;}\ninterface Target {'+decl+'}\nfunction probe(value:Base):void {const '+receiver+'=value as Target;console.log(`${'+read.getText(sf)+'('+arg+').value}`);}\nprobe({'+pair.field+':'+producer+'});\n');
 rows.push({...pair,directory,read:read.getText(sf),declarations:declarations.map(n=>n.getText(original)),sourceSha:'050880ce59e30b356b686bd3144efe24f875ebc8',declarationFile:originalFile,fileSha256:crypto.createHash('sha256').update(Buffer.from(sf.text)).digest('hex'),declarationSha256:crypto.createHash('sha256').update(Buffer.from(text)).digest('hex'),utf16Start:read.getStart(sf),utf16End:read.end,stdout:'7\n'});
}
const existing=JSON.parse(fs.readFileSync(path.join(__dirname,'gaps.json')));rows.push(...existing.members.filter(m=>m.rank===165));
fs.writeFileSync(path.join(__dirname,'gaps.json'),JSON.stringify({members:rows},null,2)+'\n');console.log('Prepared '+rows.length+' original overload/generic frontier controls.');
