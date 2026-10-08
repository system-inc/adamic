// Original Map receiver and member declarations; adjacent interfaces are reduced.
const fs=require('fs'),path=require('path'),cp=require('child_process');
const root=process.argv[2],ts=require(path.join(root,'lib/typescript.js'));
if(cp.execFileSync('git',['-C',root,'rev-parse','HEAD'],{encoding:'utf8'}).trim()!=='050880ce59e30b356b686bd3144efe24f875ebc8')throw Error('wrong source pin');
const home=path.resolve(__dirname,'../..'),members=JSON.parse(fs.readFileSync(home+'/original-members.json')).members,c=JSON.parse(fs.readFileSync('/tmp/lane5-c-carrier-declarations.json'));
const ranks=[419,851,854,1637,1664,1667,1673,1697,1706,1724,1763,1769,1796,1805,1823,1835,1508,1511,1514,1793];
const parse=s=>ts.createSourceFile('type.a',s,ts.ScriptTarget.Latest,true).statements[0].type;
const rows=[],blocked=[];
for(const rank of ranks){
 const m=members.find(x=>x.rank===rank),map=parse('type M='+m.type+';'),[key,value]=map.typeArguments,defs=new Map(),free=new Set();
 function named(n){if(defs.has(n)||['Map','Array','ReadonlyArray','RegExp'].includes(n))return;if(['K','V','K2','V2','T'].includes(n)){free.add(n);return;}const d=c[n];if(!d)throw Error('missing declaration '+n);if(d.kind==='TypeAliasDeclaration'){const t=parse(d.text);defs.set(n,d.text);walk(t);}else if(d.kind==='EnumDeclaration')defs.set(n,'const enum '+n+' {'+Object.entries(d.values).map(([n,v])=>n+'='+JSON.stringify(v)).join(',')+'}');else if(d.kind==='InterfaceDeclaration')defs.set(n,'interface '+n+' {readonly value:number;'+(d.kindValue===undefined?'':'readonly kind:'+d.kindValue+';')+'}');else throw Error('carrier '+n);}
 function walk(t){if(ts.isTypeReferenceNode(t)){if(ts.isIdentifier(t.typeName))named(t.typeName.text);else named(t.typeName.left.getText());for(const a of t.typeArguments||[])walk(a);return;}ts.forEachChild(t,walk);}
 function literal(t){if(ts.isParenthesizedTypeNode(t)||ts.isTypeOperatorNode(t))return literal(t.type);if(ts.isUnionTypeNode(t))return literal(t.types.find(n=>n.kind!==ts.SyntaxKind.UndefinedKeyword&&n.kind!==ts.SyntaxKind.NullKeyword));if(ts.isArrayTypeNode(t))return '[]';if(ts.isLiteralTypeNode(t))return t.getText();if(ts.isTypeReferenceNode(t)){const n=t.typeName.getText();if(n==='__String')return 'InternalSymbolName.Call';if(free.has(n))return n===key.getText()?'key':'incoming';if(n==='Map')return 'new Map<'+t.typeArguments.map(t=>t.getText()).join(',')+'>()';if(['Array','ReadonlyArray'].includes(n))return '[]';const d=c[n];if(d?.kind==='TypeAliasDeclaration')return literal(parse(d.text));if(d?.kind==='EnumDeclaration')return n+'.'+Object.keys(d.values)[0];if(d?.kind==='InterfaceDeclaration')return '{value:1'+(d.kindValue===undefined?'':',kind:'+d.kindValue)+'}';if(ts.isQualifiedName(t.typeName))return t.typeName.getText();throw Error('value '+n);}if(t.kind===ts.SyntaxKind.StringKeyword)return '"missing"';if(t.kind===ts.SyntaxKind.NumberKeyword)return '1';if(t.kind===ts.SyntaxKind.BooleanKeyword)return 'true';throw Error('value '+t.getText());}
 try{
  walk(key);walk(value);const params=[...free],generic=params.length?'<'+params.join(',')+'>':'',special=params.length?'<'+params.map(n=>n===key.getText()?'string':'number').join(',')+'>':'';
  const parts=m.read.split('.');parts.pop();let owner='(value as Target'+generic+').items';for(let i=parts.length-1;i>0;i--)owner='{'+parts[i]+':'+owner+'}';
  const argumentsText=literal(key)+(m.field==='set'?','+literal(value):'');
  const prefix='// Original Map receiver, declaration, aliases and read; adjacent interfaces reduced.\n'+[...defs.values()].join('\n')+'\ninterface OriginalMember<K,V> {'+m.declarations[0].text+'}\ninterface Base {readonly items:unknown;}\ninterface Target'+generic+' {readonly items:'+m.type+';}\nfunction probe'+generic+'(value:Base'+(params.length?',key:'+key.getText()+',incoming:'+value.getText():'')+'):void {const '+parts[0]+'='+owner+';'+m.read+'('+argumentsText+');console.log("completed");}\n';
  const actualKey=free.has(key.getText())?'string':key.getText(),actualValue=free.has(value.getText())?'number':value.getText(),wrong=actualValue==='boolean'?'number':'boolean';
  const dir=path.join(home,'families','rank-'+rank);fs.mkdirSync(dir,{recursive:true});
  for(const [variant,v] of [['good',actualValue],['wrong-map',wrong]])fs.writeFileSync(path.join(dir,variant+'.a'),(prefix+'const sourceMap:Map<'+actualKey+','+v+'>=new Map<'+actualKey+','+v+'>();probe'+special+'({items:sourceMap}'+(params.length?',"missing",2':'')+');\n').replace(/\r\n/g,'\n'));
  rows.push({rank,reads:m.reads,type:m.type,read:m.read,declaration:m.declarations[0].text});
 }catch(e){blocked.push({rank,reads:m.reads,read:m.read,preparation:e.message});}
}
fs.writeFileSync(path.join(__dirname,'candidates.json'),JSON.stringify(rows,null,2)+'\n');fs.writeFileSync(path.join(__dirname,'preparation-blocked.json'),JSON.stringify(blocked,null,2)+'\n');console.log('Prepared '+rows.length+' families; '+blocked.length+' preparation boundaries.');
