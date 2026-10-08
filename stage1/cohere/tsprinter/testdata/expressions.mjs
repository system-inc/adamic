import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const require=createRequire(process.argv[2]+'/package.json');
const prettier=require('prettier');
if(prettier.version!=='3.9.6') throw Error('wrong Prettier version');
const escape=text=>text.replaceAll('\\','\\\\').replaceAll('\n','\\n').replaceAll('\r','\\r').replaceAll('\t','\\t');
const cases=JSON.parse(readFileSync(process.argv[3],'utf8'));
// Port results remain byte-identical to Go. External printer differences and
// parser refusals must match the separately pinned exact upstream outcomes.
const gapProof=process.argv[3].endsWith('/gaps.json');
const known=gapProof ? JSON.parse(readFileSync(new URL('./prettier-differences.json',import.meta.url),'utf8')) : [];
const upstream=gapProof ? [] : JSON.parse(readFileSync(new URL('./tsc-upstream-differences.json',import.meta.url),'utf8')).Records;
const seen=new Set();
for(const [index,item] of cases.entries()) {
 const pinned=upstream.find(record=>(item.Label ?? '').endsWith(record.Label) && record.Source===item.Source && record.Go===item.Want);
 let formatted;
 try {
  formatted=await prettier.format(item.Source+';',{parser:'typescript',printWidth:process.argv[4]===undefined ? 80 : Number.parseInt(process.argv[4],10),tabWidth:4,singleQuote:true,semi:true});
 } catch(error) {
  if(!pinned || pinned.PrettierError!==String(error)) throw Error(`case ${index} ${item.Label}: ${error}`);
  process.stdout.write('error\t'+escape(String(error))+'\n');
  continue;
 }
 if(formatted!==item.Want) {
  const difference=known.find(record=>record.Label===item.Label && record.Source===item.Source && record.Go===item.Want && record.Prettier===formatted);
  if(!difference && (!pinned || pinned.Prettier!==formatted)) throw Error(`case ${index} ${item.Label}: source ${JSON.stringify(item.Source)}, Go ${JSON.stringify(item.Want)}, Prettier ${JSON.stringify(formatted)}`);
  if(difference) seen.add(difference.Source);
 }
 process.stdout.write('ok\t'+escape(formatted)+'\n');
}

if(gapProof && seen.size!==known.length) throw Error('known upstream difference changed or disappeared');
