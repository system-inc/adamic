'use strict';
const fs=require('fs'),path=require('path'),cp=require('child_process');
const ts=require(process.env.TYPESCRIPT_API ? path.resolve(process.env.TYPESCRIPT_API) : path.resolve(__dirname, '../../api/node_modules/typescript'));
const tree=path.resolve(process.argv[2]), output=process.argv[3];
const names=new Set(['arrayFrom','arrayToMap','reduceLeft','sameMap']);
const files=cp.execFileSync('git',['-C',tree,'ls-files','-z'],{maxBuffer:20e6}).toString().split('\0').filter(f=>/\.[cm]?[jt]sx?$/.test(f)&&fs.existsSync(path.join(tree,f)));
// Generated namespace files must join the root closure too.
const src=ts.sys.readDirectory(path.join(tree,'src'),['.ts','.tsx','.js','.mjs'],undefined,['**/*']);
const index=JSON.parse(fs.readFileSync(process.env.CALLERS_INDEX));
const other=index.candidates.map(f=>path.join(tree,f));
const options={strict:true,target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,allowJs:true,checkJs:true,skipLibCheck:true,noEmit:true,types:[]};
const calls=[],owners={},scan=[];
function walk(root,fn){const stack=[root];while(stack.length){const n=stack.pop();fn(n);ts.forEachChild(n,c=>{stack.push(c);});}}
function run(roots, phase){
 const program=ts.createProgram(roots,options),checker=program.getTypeChecker();
 const core=program.getSourceFile(path.join(tree,'src/compiler/core.ts'));
 if(core) for(const node of core.statements) if(ts.isFunctionDeclaration(node)&&node.name&&names.has(node.name.text)) (owners[node.name.text]??=[]).push(node.getText(core));
 let count=0,callCount=0;
 for(const sf of program.getSourceFiles()){
  if(!sf.fileName.startsWith(tree+path.sep)||sf.fileName.includes('/node_modules/'))continue;
  if(phase!=='src'&&sf.fileName.startsWith(tree+'/src/'))continue;
  count++;
  const candidateNames=new Set(names);
  let changed=true;
  while(changed){changed=false;function aliases(n){
    let name,from;
    if(ts.isImportSpecifier(n)){name=n.name.text;from=(n.propertyName??n.name).text;}
    if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&n.initializer){name=n.name.text;const e=n.initializer;from=ts.isIdentifier(e)?e.text:ts.isPropertyAccessExpression(e)?e.name.text:undefined;}
    if(ts.isBindingElement(n)&&ts.isIdentifier(n.name)){name=n.name.text;from=n.propertyName?.text??n.name.text;}
    if(name&&candidateNames.has(from)&&!candidateNames.has(name)){candidateNames.add(name);changed=true;}
    
  }walk(sf,aliases);}
  function visit(n){
   if(ts.isCallExpression(n)||ts.isNewExpression(n)||ts.isTaggedTemplateExpression(n)){
    callCount++;
    const expression=ts.isTaggedTemplateExpression(n)?n.tag:n.expression;
    const spelling=ts.isIdentifier(expression)?expression.text:ts.isPropertyAccessExpression(expression)?expression.name.text:undefined;
    if(phase!=='src'&&!candidateNames.has(spelling)){return;}
    let symbol=checker.getSymbolAtLocation(ts.isTaggedTemplateExpression(n)?n.tag:n.expression);
    const seen=new Set();
    function refers(s){
      if(!s||seen.has(s))return false;seen.add(s);
      if(s.flags&ts.SymbolFlags.Alias)return refers(checker.getAliasedSymbol(s));
      return (s.declarations??[]).some(d=>{
        if(ts.isFunctionDeclaration(d)&&d.name&&names.has(d.name.text)&&path.resolve(d.getSourceFile().fileName)===path.join(tree,'src/compiler/core.ts'))return true;
        if(ts.isVariableDeclaration(d)&&d.initializer)return refers(checker.getSymbolAtLocation(d.initializer));
        return false;
      });
    }
    const sig=(phase==='src'||refers(symbol))?checker.getResolvedSignature(n):undefined,d=sig?.declaration;
    if(d&&ts.isFunctionDeclaration(d)&&d.name&&names.has(d.name.text)&&path.resolve(d.getSourceFile().fileName)===path.join(tree,'src/compiler/core.ts')){
     const loc=sf.getLineAndCharacterOfPosition(n.getStart(sf));
     let p=n.parent;while(ts.isParenthesizedExpression(p)||ts.isNonNullExpression(p)||ts.isAsExpression(p)){p=p.parent;}
     const cat=ts.isReturnStatement(p)?'return':ts.isVariableDeclaration(p)?'variable':ts.isBinaryExpression(p)&&[ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken,ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken].includes(p.operatorToken.kind)?'identity comparison':ts.isCallExpression(p)||ts.isNewExpression(p)?'argument':ts.isPropertyAccessExpression(p)?'member receiver':ts.isBinaryExpression(p)&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken?'assignment':ts.isForOfStatement(p)?'iteration':ts.isExpressionStatement(p)?'discarded':ts.SyntaxKind[p.kind];
     const fmt=ts.TypeFormatFlags.NoTruncation;
     calls.push({set:d.name.text,file:path.relative(tree,sf.fileName),line:loc.line+1,column:loc.character+1,category:cat,signature:checker.signatureToString(sig,undefined,fmt),result:checker.typeToString(checker.getTypeAtLocation(n),undefined,fmt),arguments:(n.arguments??[]).map(a=>({text:a.getText(sf),type:checker.typeToString(checker.getTypeAtLocation(a),undefined,fmt)})),call:n.getText(sf),context:p.parent?.getText(sf)});
    }
   }
   
  }walk(sf,visit);
 }
 scan.push({phase,roots:roots.length,files:count,calls:callCount});process.stderr.write(JSON.stringify(scan.at(-1))+'\n');
}
run(src,'src');fs.writeFileSync(output+'.src',JSON.stringify({owners,counts:calls.length,calls},null,2));
for(let i=0;!process.env.CALLERS_SRC_ONLY&&i<other.length;i+=400){run([...other.slice(i,i+400),path.join(tree,'src/compiler/core.ts')],'other-'+i);if(global.gc)global.gc();}
for(const name of Object.keys(owners)) owners[name]=[...new Set(owners[name])];
const counts={};for(const c of calls){const x=counts[c.set]??={total:0,categories:{},files:{},arity:{}};x.total++;x.categories[c.category]=(x.categories[c.category]||0)+1;x.files[c.file]=(x.files[c.file]||0)+1;x.arity[c.arguments.length]=(x.arity[c.arguments.length]||0)+1;}
fs.writeFileSync(output,JSON.stringify({version:ts.version,tree,trackedCodeFiles:files.length,index,scan,owners,counts,calls},null,2)+'\n');
