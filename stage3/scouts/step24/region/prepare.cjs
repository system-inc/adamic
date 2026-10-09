// Instrument only an external scratch tree, using the stock compiler's AST.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const ts=require(process.env.STAGE3_TYPESCRIPT||'typescript');
if(ts.version!=='6.0.3')throw Error('Expected stock TypeScript 6.0.3');
const tree=path.resolve(process.argv[2]),out=path.resolve(process.argv[3]);
const pin='050880ce59e30b356b686bd3144efe24f875ebc8';
if(cp.execFileSync('git',['-C',tree,'rev-parse','HEAD'],{encoding:'utf8'}).trim()!==pin)throw Error('Wrong source pin');
if(fs.existsSync(out))throw Error('Scratch output must be new');
fs.mkdirSync(out,{recursive:true});
fs.cpSync(path.join(tree,'src'),path.join(out,'src'),{recursive:true});
const selected=['parser.ts','factory/nodeFactory.ts','factory/baseNodeFactory.ts'];
const sites=[],parserFunctionSites=[];
for(const file of selected){
 const input=path.join(tree,'src/compiler',file),text=fs.readFileSync(input,'utf8');
 if(text!==cp.execFileSync('git',['-C',tree,'show',pin+':src/compiler/'+file],{encoding:'utf8',maxBuffer:4000000}))throw Error('Modified input '+file);
 const sf=ts.createSourceFile(input,text,ts.ScriptTarget.Latest,true);const edits=[];
 const loc=n=>'src/compiler/'+file+':'+(sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1);
 const insert=(pos,text)=>edits.push({pos,text});
 const walk=n=>{
  if(ts.isFunctionLike(n)&&n.body&&ts.isBlock(n.body)){
   const name=n.name?.getText(sf)||'(anonymous)';const site=loc(n)+':'+name;
   let inParser=false;for(let parent=n.parent;parent;parent=parent.parent)if(ts.isModuleDeclaration(parent)&&parent.name.getText(sf)==='Parser')inParser=true;
   if(inParser)parserFunctionSites.push(site);
   insert(n.body.getStart(sf)+1,`globalThis.__region.enter(${JSON.stringify(site)},${inParser});try{`);
   insert(n.body.end-1,'}finally{globalThis.__region.leave();}');
  }
  if(ts.isNewExpression(n)&&(['NodeConstructor','TokenConstructor','IdentifierConstructor','PrivateIdentifierConstructor','SourceFileConstructor'].some(name=>n.expression.getText(sf).includes(name)))){
   const site=loc(n);sites.push({site,expression:n.getText(sf)});
   insert(n.getStart(sf), 'globalThis.__region.record(');insert(n.end,','+JSON.stringify(site)+')');
  }
  ts.forEachChild(n,walk);
 };walk(sf);
 edits.sort((a,b)=>b.pos-a.pos);let changed=text;for(const e of edits)changed=changed.slice(0,e.pos)+e.text+changed.slice(e.pos);
 fs.writeFileSync(path.join(out,'src/compiler',file),changed);
}
// An independent constructor-entry census catches an omitted allocation seam.
for(const file of ['compiler/utilities.ts','services/services.ts']){
 const input=path.join(tree,'src',file),text=fs.readFileSync(input,'utf8');
 if(text!==cp.execFileSync('git',['-C',tree,'show',pin+':src/'+file],{encoding:'utf8',maxBuffer:4000000}))throw Error('Modified audit input '+file);
 const sf=ts.createSourceFile(input,text,ts.ScriptTarget.Latest,true);const positions=[];
 const visit=n=>{
  if(ts.isFunctionDeclaration(n)&&['Node','Token','Identifier'].includes(n.name?.text)&&n.body)positions.push(n.body.end-1);
  if(ts.isClassDeclaration(n)&&['NodeObject','TokenOrIdentifierObject'].includes(n.name?.text))for(const member of n.members)if(ts.isConstructorDeclaration(member)&&member.body)positions.push(member.body.end-1);
  ts.forEachChild(n,visit);
 };visit(sf);let changed=text;for(const pos of positions.sort((a,b)=>b-a))changed=changed.slice(0,pos)+'globalThis.__region.audit(this);'+changed.slice(pos);
 fs.writeFileSync(path.join(out,'src',file),changed);
}
const esbuild=require(path.join(tree,'node_modules/esbuild'));
const build=(entry,filename)=>esbuild.buildSync({entryPoints:[entry],outfile:path.join(out,filename),bundle:true,platform:'node',format:'cjs',target:'node24',sourcemap:true,logLevel:'silent'});
build(path.join(out,'src/typescript/typescript.ts'),'instrumented.cjs');
build(path.join(tree,'src/typescript/typescript.ts'),'control.cjs');
const stockLib=path.dirname(require.resolve(process.env.STAGE3_TYPESCRIPT||'typescript'));
for(const name of fs.readdirSync(stockLib))if(name.startsWith('lib.')&&name.endsWith('.d.ts'))fs.copyFileSync(path.join(stockLib,name),path.join(out,name));
fs.writeFileSync(path.join(out,'sites.json'),JSON.stringify({pin,sites,parserFunctionSites},null,2)+'\n');
console.log(`Instrumented ${sites.length} constructor expressions in ${selected.length} files; built scratch and control bundles.`);
