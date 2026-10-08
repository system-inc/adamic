// Capability ablations of the final printer, preserving all other refusals.
// This is measurement code on Node; it is not an Adamic production fallback.
import {readFileSync} from 'node:fs';
import {formatFile} from '../files.ts';
import {Parser} from '../../../typescript/parser/parser.ts';
const stages=['default','namespace','named','type','attributes'];
const level=stages.indexOf(process.argv[3]);
if(level<0) throw Error('unknown import stage');
for(const {source,path} of JSON.parse(readFileSync(process.argv[2],'utf8'))) {
 let result=formatFile(source,{printWidth:120,tabWidth:4,useTabs:false},path);
 if(result.kind==='Ok') {
  const parser=new Parser(source,path);parser.file();
  for(const node of parser.nodes) {
   if((node.kind==='NamespaceImport' && level<1) ||
      (node.kind==='NamedImports' && level<2) ||
      (node.kind==='ImportClause' && node.semantic==='TypeKeyword' && level<3) ||
      (node.kind==='ImportSpecifier' && node.semantic==='1' && level<3) ||
      (node.kind==='ImportAttributes' && level<4)) {
    result={kind:'NotYet',reason:'ImportDeclaration'};break;
   }
  }
 }
 process.stdout.write(JSON.stringify(result)+'\n');
}
