// Classify the supplied static candidate reads with the pinned stock checker.
// The source tree is referenced through a cohere/TypeScript detached worktree.
// This neither compiles the program nor measures allocation reachability.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('typescript');assert.equal(ts.version,'6.0.3');
const [root,input,output]=process.argv.slice(2);
const inventory=JSON.parse(fs.readFileSync(input));
const configPath=path.join(root,'src/compiler/tsconfig.json');
const read=ts.readConfigFile(configPath,ts.sys.readFile);assert.equal(read.error,undefined);
const config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);
const program=ts.createProgram(config.fileNames,config.options),checker=program.getTypeChecker();
const nullable=type=>!!(type.flags&(ts.TypeFlags.Null|ts.TypeFlags.Undefined));
const scalarContract=type=>{
 const parts=type.isUnion()?type.types:[type];
 const symbol=type.symbol || parts[0]?.symbol?.parent;
 if(symbol?.flags&ts.SymbolFlags.Enum){
  const whole=checker.getDeclaredTypeOfSymbol(symbol),all=whole.isUnion()?whole.types:[whole];
  if(parts.length===all.length && parts.every(t=>all.includes(t)) && all.every(t=>typeof t.value==='number')) return {type:'number',values:[all[0].value],open_enum:true};
 }
 if(parts.every(t=>t.flags&ts.TypeFlags.NumberLiteral)) return {type:parts.map(t=>String(t.value)).join(' | '),values:parts.map(t=>t.value)};
 if(parts.every(t=>t.flags&ts.TypeFlags.StringLiteral)) return {type:parts.map(t=>JSON.stringify(t.value)).join(' | '),values:parts.map(t=>t.value)};
 if(parts.every(t=>t.flags&ts.TypeFlags.BooleanLiteral)) return {type:parts.map(t=>t.intrinsicName).join(' | '),values:parts.map(t=>t.intrinsicName==='true')};
 if(type.flags&ts.TypeFlags.Number) return {type:'number',values:[0]};
 if(type.flags&ts.TypeFlags.String) return {type:'string',values:['sample']};
 return null;
};
const finite=type=>{
 const parts=type.isUnion()?type.types:[type];
 return parts.length>0 && parts.every(t=>!!(t.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.BooleanLiteral)));
};
for(const row of inventory.ranked_candidates){
 const file=program.getSourceFile(path.join(root,row.witness.file));let node;
 if(file){const visit=n=>{if(n.getStart(file)===row.witness.start && n.end===row.witness.end) node=n;else ts.forEachChild(n,visit);};visit(file);}
 if(!node || node.getText(file)!==row.witness.text){row.classification='unresolved witness';continue;}
 const type=checker.getTypeAtLocation(node),members=(type.isUnion()?type.types:[type]).filter(t=>!nullable(t));
 row.member_types=members.map(t=>checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation));
 if(members.length<2 || members.some(t=>!(t.flags&ts.TypeFlags.Object) && !t.isIntersection())) {row.classification='other contract or consumer';continue;}
 const tags=checker.getPropertiesOfType(members[0]).filter(symbol=>{const seen=new Set();return members.every(member=>{
  const own=checker.getPropertyOfType(member,symbol.name);
  if(!own || own.flags&ts.SymbolFlags.Optional) return false;
  const declared=checker.getTypeOfSymbolAtLocation(own,node);if(!finite(declared)) return false;
  const values=(declared.isUnion()?declared.types:[declared]).map(t=>JSON.stringify([t.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.BooleanLiteral),t.value??t.intrinsicName]));
  if(values.some(value=>seen.has(value))) return false;for(const value of values) seen.add(value);return true;
 });}).map(s=>s.name);
 row.shared_discriminants=tags;
 row.classification=tags.length?'shared discriminant':'untagged object union candidate';
 row.member_selectors=members.map(member=>{
  const own=checker.getPropertyOfType(member,'kind');
  if(!own || own.flags&ts.SymbolFlags.Optional) return null;
  const scalar=scalarContract(checker.getTypeOfSymbolAtLocation(own,node));
  return scalar===null?null:{field:'kind',...scalar};
 });
 row.member_own_tags=members.map(member=>checker.getPropertiesOfType(member).filter(symbol=>!(symbol.flags&ts.SymbolFlags.Optional)&&finite(checker.getTypeOfSymbolAtLocation(symbol,node))).map(symbol=>symbol.name));
}
inventory.classification_checker_diagnostic_count=ts.getPreEmitDiagnostics(program).length;
inventory.classification_note='Stock type inspection only; diagnostics are reported, never bypassed for production.';
inventory.groups={};
for(const row of inventory.ranked_candidates){const group=inventory.groups[row.classification]??={pairs:0,reads:0};group.pairs++;group.reads+=row.read_count;}
fs.writeFileSync(output,JSON.stringify(inventory,null,2)+'\n');console.log(JSON.stringify(inventory.groups));
