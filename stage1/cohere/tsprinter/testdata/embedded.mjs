// The actual Prettier fork cohere embeds, independently parsing the original source on V8.
import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const require=createRequire(process.argv[2]+'/standalone.js');
const prettier=require('./standalone.js');
if(prettier.version!=='3.9.6') throw Error('wrong embedded Prettier version');
const plugins=[require('./plugins/typescript.js'),require('./plugins/estree.js')];
const escape=text=>text.replaceAll('\\','\\\\').replaceAll('\n','\\n').replaceAll('\r','\\r').replaceAll('\t','\\t');
const cases=JSON.parse(readFileSync(process.argv[3],'utf8'));
for(const [index,item] of cases.entries()) {
 const formatted=await prettier.format(item.Source+';',{parser:'typescript',plugins,printWidth:process.argv[4]===undefined ? 80 : Number.parseInt(process.argv[4],10),tabWidth:4,singleQuote:true,semi:true});
 if(formatted!==item.Want) throw Error(`case ${index} ${item.Label}: source ${JSON.stringify(item.Source)}, Go ${JSON.stringify(item.Want)}, embedded Prettier ${JSON.stringify(formatted)}`);
 process.stdout.write('ok\t'+escape(formatted)+'\n');
}
