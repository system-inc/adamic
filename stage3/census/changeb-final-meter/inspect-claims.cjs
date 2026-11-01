// Supplementary stock-TypeScript source evidence; run only after both timed censuses finish.
const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const fs = require('fs'), path = require('path');
const out=__dirname, root='/workspace/scratch/tree';
const names=Object.keys(JSON.parse(fs.readFileSync(path.join(out,'source-sha256.json'))));
const program=ts.createProgram(names.map(s=>root+'/src/compiler/'+s),{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,skipLibCheck:true,strict:true});
const checker=program.getTypeChecker();
const where=n=>{const sf=n.getSourceFile(),p=sf.getLineAndCharacterOfPosition(n.getStart(sf)); return sf.fileName.replace(root+'/','')+':'+(p.line+1)+':'+(p.character+1)};
function branded(t){
 const kind=checker.getPropertyOfType(t,'kind'); if(!kind)return false;
 const kt=checker.getTypeOfSymbolAtLocation(kind,kind.valueDeclaration||kind.declarations[0]);
 if(!(kt.flags&ts.TypeFlags.Union) && (kt.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.EnumLiteral)))return false;
 return checker.getPropertiesOfType(t).some(p=>!(p.flags&ts.SymbolFlags.Optional)&&!(p.escapedName+'').startsWith('__@')&&!!(checker.getTypeOfSymbolAtLocation(p,p.valueDeclaration||p.declarations[0]).flags&(ts.TypeFlags.Undefined|ts.TypeFlags.Void)));
}
function members(t,seen=new Set()){
 if(!t||seen.has(t))return []; seen.add(t);
 if(t.isUnion())return t.types.flatMap(t=>members(t,seen));
 if(checker.isArrayType(t))return members(checker.getTypeArguments(t)[0],seen);
 return branded(t)?[checker.typeToString(t)]:[];
}
const inventory=[];
for(const sf of program.getSourceFiles()) if(sf.fileName.startsWith(root+'/src/compiler/')){
 function visit(n){if(ts.isInterfaceDeclaration(n)){const t=checker.getTypeAtLocation(n);if(branded(t))inventory.push({name:n.name.text,where:where(n)});}ts.forEachChild(n,visit);} visit(sf);
}
const rows=JSON.parse(fs.readFileSync(path.join(out,'unknown-instance-units.json')));
for(const r of rows){
 const m=r.stop_where.match(/^(.*):(\d+):(\d+)$/),sf=program.getSourceFile(root+'/'+m[1]);
 const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);
 let best;
 function visit(n){if(n.pos<=pos&&pos<=n.end){if(ts.isCallExpression(n))best=n;ts.forEachChild(n,visit);}}visit(sf);
 if(!best)throw Error('no call at '+r.stop_where);
 const sig=checker.getResolvedSignature(best),pred=sig&&checker.getTypePredicateOfSignature(sig);
 r.call_source=best.getText(sf);
 r.static_claim=pred&&pred.type?checker.typeToString(pred.type):null;
 r.static_claim_branded_members=pred&&pred.type?members(pred.type):[];
 r.predicate_signature=sig&&sig.declaration?sig.declaration.getText().slice(0,1000):null;
 r.predicate_declaration=sig&&sig.declaration?where(sig.declaration):null;
 r.program_declares_branded_supertypes=inventory.length>0;
}
fs.writeFileSync(path.join(out,'branded-supertype-source-inventory.json'),JSON.stringify(inventory,null,2)+'\n');
fs.writeFileSync(path.join(out,'unknown-instance-source-evidence.json'),JSON.stringify(rows,null,2)+'\n');
console.log(JSON.stringify({unknown_units:rows.length,declared_branded_supertypes:inventory.length}));
