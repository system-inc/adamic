import {readFileSync} from 'node:fs';
import {formatFile} from '../files.ts';
for(const {source,path} of JSON.parse(readFileSync(process.argv[2],'utf8'))) {
 const result=formatFile(source,{printWidth:120,tabWidth:4,useTabs:false},path);
 process.stdout.write(JSON.stringify(result)+'\n');
}
