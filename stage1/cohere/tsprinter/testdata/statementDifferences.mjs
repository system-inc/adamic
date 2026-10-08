import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const prettier=createRequire(process.argv[2]+'/package.json')('prettier');
if(prettier.version!=='3.9.6') throw Error('wrong Prettier version');
const records=JSON.parse(readFileSync(process.argv[3],'utf8'));
for(const record of records) {
 const text=await prettier.format(record.Source+';',{parser:'typescript',printWidth:80,tabWidth:4,singleQuote:true,semi:true});
 if(text!==record.Prettier || text===record.Go) throw Error('upstream difference changed: '+JSON.stringify(record.Source));
}
console.log(records.length+' exact npm differences pinned');
