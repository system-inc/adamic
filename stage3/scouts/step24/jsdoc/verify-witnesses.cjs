// Source function bodies are held unchanged; only the dependencies are bounded.
const fs=require('node:fs'),path=require('node:path');
const ts=require(process.env.JSDOC_TYPESCRIPT);const tree=process.argv[2];const unit=__dirname;
function functions(text){const sf=ts.createSourceFile('file.ts',text,ts.ScriptTarget.Latest,true),found=new Map();function walk(n){if(ts.isFunctionDeclaration(n)&&n.name)found.set(n.name.text,n.body?.getText(sf).replaceAll('\r\n','\n'));ts.forEachChild(n,walk);}walk(sf);return found;}
const original=functions(fs.readFileSync(path.join(tree,'src/compiler/parser.ts'),'utf8'));
for(const row of JSON.parse(fs.readFileSync(path.join(unit,'fixtures.json')))){
 const text=fs.readFileSync(path.join(unit,row.file),'utf8');const actual=functions(text).get(row.source_function);
 if(actual!==original.get(row.source_function))throw Error('source body changed: '+row.source_function);
 // A body edit must fail this exact comparison, independently of the output oracle.
 const mutant=functions(text.replace(row.before,row.after)).get(row.source_function);
 if(mutant===original.get(row.source_function))throw Error('source-body mutant survived');
 console.log('PASS body unchanged; mutant caught:',row.source_function);
}
