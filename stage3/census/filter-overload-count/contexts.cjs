'use strict';
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root=path.resolve(process.argv[2]),out=process.argv[3];
const names=ts.sys.readDirectory(root+'/src',['.ts'],[]);
const opts={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true};
const p=ts.createProgram(names,opts),c=p.getTypeChecker(),rows=[],all=[];
function where(n){const f=n.getSourceFile(),lc=f.getLineAndCharacterOfPosition(n.getStart());return path.relative(root,f.fileName)+':'+(lc.line+1)+':'+(lc.character+1)}
const writable=t=>t?.isUnion?.()?t.types.some(writable):!!(t&&c.getPropertyOfType(t,'push'));
const type=n=>c.typeToString(c.getTypeAtLocation(n),n,ts.TypeFormatFlags.NoTruncation);
function context(n){const a=[];for(let x=n.parent;x&&!ts.isSourceFile(x);x=x.parent){a.push({kind:ts.SyntaxKind[x.kind],where:where(x),type:type(x),text:x.getText().slice(0,700)});if(ts.isStatement(x)||ts.isArrowFunction(x))break;}return a;}
for(const f of p.getSourceFiles()){if(!f.fileName.startsWith(root+'/src/'))continue;
function visit(n){if(ts.isCallExpression(n)){const s=c.getResolvedSignature(n),d=s?.declaration;if(d&&d.getSourceFile().fileName===root+'/src/compiler/core.ts'&&d.name?.getText()==='filter'){
const decls=d.parent.statements.filter(x=>ts.isFunctionDeclaration(x)&&x.name?.text==='filter'&&!x.body),overload=decls.indexOf(d)+1;
all.push({where:where(n),overload,call:n.getText()});
if([1,2,5,6].includes(overload)){
let x=n;while(ts.isParenthesizedExpression(x.parent)||ts.isAsExpression(x.parent)||ts.isConditionalExpression(x.parent)||ts.isNonNullExpression(x.parent))x=x.parent;
let kind='read',reason='',slot;
if(ts.isVariableDeclaration(x.parent)&&x.parent.initializer===x){kind='store';slot=x.parent;reason='variable '+x.parent.name.getText()+' : '+type(x.parent.name);if(!writable(c.getTypeAtLocation(x.parent.name)))kind='read';}
else if(ts.isReturnStatement(x.parent)||(ts.isArrowFunction(x.parent)&&x.parent.body===x)){let fn=x.parent;while(fn&&!ts.isFunctionLike(fn))fn=fn.parent;const sig=fn&&c.getSignatureFromDeclaration(fn);reason='return : '+(sig?c.typeToString(c.getReturnTypeOfSignature(sig)): '?');kind=sig&&writable(c.getReturnTypeOfSignature(sig))?'store':'read';}
else if(ts.isBinaryExpression(x.parent)&&x.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken){kind='store';reason='assignment : '+type(x.parent.left);if(!writable(c.getTypeAtLocation(x.parent.left)))kind='read';}
else if(ts.isPropertyAssignment(x.parent)){kind='store';reason='property : '+type(x.parent.name);if(!writable(c.getTypeAtLocation(x.parent.name)))kind='read';}
else if(ts.isCallExpression(x.parent)){const call=x.parent,ix=call.arguments.indexOf(x),sig=c.getResolvedSignature(call),param=sig?.parameters[Math.min(ix,sig.parameters.length-1)];if(param){const pt=c.getTypeOfSymbolAtLocation(param,call);reason='parameter : '+c.typeToString(pt);if(writable(pt))kind='write';}}
if(ts.isAsExpression(x)&&writable(c.getTypeAtLocation(x))){kind='store';reason='writable type assertion : '+type(x);}
const refs=[],refNodes=[];if(slot&&ts.isIdentifier(slot.name)){const sym=c.getSymbolAtLocation(slot.name);function scan(z){if(ts.isIdentifier(z)&&c.getSymbolAtLocation(z)===sym&&z!==slot.name){refNodes.push(z);refs.push({where:where(z),parent:z.parent.getText().slice(0,700),context:context(z)});};ts.forEachChild(z,scan);}scan(f);}
const flags={store:kind==='store',write:kind==='write',identity:false},uses=[];
for(const z of [n,...refNodes]){
 let q=z;while(ts.isNonNullExpression(q.parent)||ts.isParenthesizedExpression(q.parent))q=q.parent;
 const par=q.parent;
 if(ts.isBinaryExpression(par)&&[ts.SyntaxKind.EqualsEqualsToken,ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(par.operatorToken.kind)){
  const other=par.left===q?par.right:par.left;if(other.getText()===n.arguments[0].getText()){flags.identity=true;uses.push({kind:'identity',where:where(par),text:par.getText()});}}
 if(ts.isCallExpression(par)){
  const sig=c.getResolvedSignature(par),ix=par.arguments.indexOf(q),param=ix>=0&&sig?.parameters[Math.min(ix,sig.parameters.length-1)];
  if(param&&writable(c.getTypeOfSymbolAtLocation(param,par))){flags.write=true;uses.push({kind:'writable parameter',where:where(par),text:par.getText().slice(0,1000),parameter:c.typeToString(c.getTypeOfSymbolAtLocation(param,par))});}}
 if(ts.isPropertyAccessExpression(par)&&['push','pop','shift','unshift','splice','sort','reverse','fill','copyWithin'].includes(par.name.text)&&ts.isCallExpression(par.parent)&&par.parent.expression===par){flags.write=true;uses.push({kind:'mutation',where:where(par.parent),text:par.parent.getText()});}
 if((ts.isElementAccessExpression(par)||(ts.isPropertyAccessExpression(par)&&par.name.text==='length'))&&ts.isBinaryExpression(par.parent)&&par.parent.left===par&&par.parent.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&par.parent.operatorToken.kind<=ts.SyntaxKind.LastAssignment){flags.write=true;uses.push({kind:'assignment',where:where(par.parent),text:par.parent.getText()});}
}
rows.push({flags,uses,where:where(n),overload,call:n.getText(),argument:n.arguments[0].getText(),argument_type:type(n.arguments[0]),result_type:type(n),kind,reason,context:context(n),refs});
}}}ts.forEachChild(n,visit);}visit(f);}
fs.writeFileSync(out,JSON.stringify({options:opts,roots:names,all,rows},null,2));console.log(JSON.stringify({all:all.length,writable:rows.length,counts:rows.reduce((a,r)=>(a[r.overload]=(a[r.overload]||0)+1,a),{})}));
