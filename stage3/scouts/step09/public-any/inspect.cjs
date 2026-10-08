'use strict';
// Resolve source owners and symbol references before any runtime instrumentation.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),ts=require('typescript');
if(ts.version!=='6.0.3')throw Error('requires TypeScript 6.0.3');
const [tree,output]=process.argv.slice(2);
const residue=fs.readFileSync(path.join(__dirname,'evidence/RESIDUE.md'),'utf8');
const locations=[...residue.matchAll(/\| (src\/compiler\/[^:]+):(\d+):(\d+) \| ([^\n]+) \|/g)].map((m,i)=>({id:'S'+String(i+1).padStart(2,'0'),file:m[1],line:+m[2],column:+m[3],residue:m[4]}));
if(locations.length!==20)throw Error('expected 20 observations including timer anomaly');
const program=ts.createProgram(ts.sys.readDirectory(path.join(tree,'src'),['.ts']),{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,noEmit:true,types:[]});
const checker=program.getTypeChecker();
function loc(n){const s=n.getSourceFile(),p=s.getLineAndCharacterOfPosition(n.getStart());return {file:path.relative(tree,s.fileName),line:p.line+1,column:p.character+1,text:n.getText()};}
function symbol(n){let s=ts.isShorthandPropertyAssignment(n.parent)?checker.getShorthandAssignmentValueSymbol(n.parent):checker.getSymbolAtLocation(n);if(s?.flags&ts.SymbolFlags.Alias)s=checker.getAliasedSymbol(s);return s;}
function at(n,p){let r=n;ts.forEachChild(n,c=>{if(c.getStart()<=p&&p<c.end)r=at(c,p)});return r;}
const all=[];
for(const s of program.getSourceFiles())if(s.fileName.startsWith(tree+'/src/')){function walk(n){all.push(n);ts.forEachChild(n,walk)}walk(s);}
const publicNames=['convertToObject','parseJsonConfigFileContent','convertCompilerOptionsFromJson','convertTypeAcquisitionFromJson','readConfigFile','parseConfigFileTextToJson','ParsedCommandLine','System','WatchHost','ObjectAllocator','Node','Type','CompilerOptionsValue','JsonConfigValue','JsonConfigObject','readJson','readJsonOrUndefined'];
const publicOwners=all.filter(n=>(ts.isFunctionDeclaration(n)||ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n))&&publicNames.includes(n.name?.text)).map(n=>({name:n.name.text,...loc(n),signature:n.body?n.getText().slice(0,n.body.getStart()-n.getStart()):n.getText(),symbol:symbol(n.name)}));
const rows=locations.map(r=>{const sf=program.getSourceFile(path.join(tree,r.file)),p=sf.getPositionOfLineAndCharacter(r.line-1,r.column-1);let n=at(sf,p);while(n.parent&&!(ts.isFunctionLike(n)||ts.isVariableDeclaration(n)||ts.isPropertyDeclaration(n)||ts.isPropertyAssignment(n)||ts.isMethodSignature(n)))n=n.parent;const owner=n,ownerName=n.name?.getText()||(ts.isPropertyAssignment(n.parent)?n.parent.name.getText():'<anonymous>');const s=n.name&&symbol(n.name);const references=s?all.filter(x=>ts.isIdentifier(x)&&symbol(x)===s&&x!==n.name).map(x=>({...loc(x),use:ts.isCallExpression(x.parent)?'direct-call':ts.isNewExpression(x.parent)?'construct':'reference',enclosing:(()=>{let q=x.parent;while(q&&!ts.isFunctionLike(q))q=q.parent;return q?.name?.getText()||'<module-or-callback>'})()})):[];return {...r,owner:loc(owner),owner_name:ownerName,signature:owner.body?owner.getText().slice(0,owner.body.getStart()-owner.getStart()):owner.getText(),references};});
for(const p of publicOwners){p.references=all.filter(x=>ts.isIdentifier(x)&&symbol(x)===p.symbol&&x.parent!==p).map(x=>({...loc(x),use:ts.isCallExpression(x.parent)?'direct-call':'reference'}));delete p.symbol;}
const memberNames=['setTimeout','clearTimeout','getNodeConstructor','getTokenConstructor','getIdentifierConstructor','getPrivateIdentifierConstructor','getSourceFileConstructor','getTypeConstructor','timerToBuildInvalidatedProject','timerToUpdateProgram','timerToInvalidateFailedLookupResolutions','timerToUpdateChildWatches','pollScheduled'];
const memberReferences=Object.fromEntries(memberNames.map(name=>[name,all.filter(n=>ts.isIdentifier(n)&&n.text===name).map(n=>({...loc(n),context:n.parent.getText().slice(0,500),domain:n.getSourceFile().fileName.includes('/testRunner/')||n.getSourceFile().fileName.includes('/harness/')?'test':'production'}))]));
const hashes=Object.fromEntries(program.getSourceFiles().filter(s=>s.fileName.startsWith(tree+'/src/compiler/')).map(s=>[path.relative(tree,s.fileName),crypto.createHash('sha256').update(fs.readFileSync(s.fileName)).digest('hex')]));
fs.writeFileSync(output,JSON.stringify({typescript:ts.version,adaptation_sha:'1e920a31b601422c75a26ead083e9c1423c34b66',source_pin:'050880ce59e30b356b686bd3144efe24f875ebc8',rows,publicOwners,memberReferences,source_hashes:hashes},null,2)+'\n');
console.log(JSON.stringify(rows.map(r=>({id:r.id,source:r.file+':'+r.line,owner:r.owner_name,references:r.references.length}))));
