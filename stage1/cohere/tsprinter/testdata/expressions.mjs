import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const require=createRequire(process.argv[2]+'/package.json');
const prettier=require('prettier');
if(prettier.version!=='3.9.6') throw Error('wrong Prettier version');
const escape=text=>text.replaceAll('\\','\\\\').replaceAll('\n','\\n').replaceAll('\r','\\r').replaceAll('\t','\\t');
const cases=JSON.parse(readFileSync(process.argv[3],'utf8'));
// Only the separately named unported proving corpus can contain an upstream difference.
// Accepted cases.json always remains a strict byte comparison.
const gapProof=process.argv[3].endsWith('/gaps.json');
const known=gapProof ? JSON.parse(readFileSync(new URL('./prettier-differences.json',import.meta.url),'utf8')) : [];
const seen=new Set();
for(const [index,item] of cases.entries()) {
 const formatted=await prettier.format(item.Source+';',{parser:'typescript',printWidth:process.argv[4]===undefined ? 80 : Number.parseInt(process.argv[4],10),tabWidth:4,singleQuote:true,semi:true});
 if(formatted!==item.Want) {
  const difference=known.find(record=>record.Label===item.Label && record.Source===item.Source && record.Go===item.Want && record.Prettier===formatted);
  if(!difference) throw Error(`case ${index} ${item.Label}: source ${JSON.stringify(item.Source)}, Go ${JSON.stringify(item.Want)}, Prettier ${JSON.stringify(formatted)}`);
  seen.add(difference.Source);
 }
 process.stdout.write('ok\t'+escape(formatted)+'\n');
}

if(gapProof && seen.size!==known.length) throw Error('known upstream difference changed or disappeared');
