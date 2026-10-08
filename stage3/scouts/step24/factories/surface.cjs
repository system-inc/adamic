// Public factory members are listed separately from their implementation bodies and helpers.
const ts=require(process.env.TYPESCRIPT_API||'/workspace/cache/tsc-census/npm/node_modules/typescript');
const fs=require('node:fs'),path=require('node:path');
const root=path.resolve(process.argv[2]),config=path.join(root,'src/compiler/tsconfig.json');
const read=ts.readConfigFile(config,ts.sys.readFile),parsed=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(config),undefined,config);
const program=ts.createProgram(parsed.fileNames,parsed.options),checker=program.getTypeChecker();
const file=program.getSourceFile(path.join(root,'src/compiler/types.ts'));
function location(n){const p=n.getSourceFile().getLineAndCharacterOfPosition(n.getStart());return {file:path.relative(root,n.getSourceFile().fileName),line:p.line+1,column:p.character+1};}
const declaration=file.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='NodeFactory');
const type=checker.getTypeAtLocation(declaration),members=checker.getPropertiesOfType(type).map(p=>{const d=p.valueDeclaration||p.declarations[0],t=checker.getTypeOfSymbolAtLocation(p,d);return {name:p.name,location:location(d),type:checker.typeToString(t),signatures:checker.getSignaturesOfType(t,ts.SignatureKind.Call).map(s=>({returnType:checker.typeToString(checker.getReturnTypeOfSignature(s)),declaration:s.declaration?location(s.declaration):null}))};});
fs.writeFileSync(path.join(__dirname,'data/node_factory_surface.json'),JSON.stringify(members,null,2)+'\n');console.log('PASS '+members.length+' public NodeFactory members, including aliases and overload signatures');
