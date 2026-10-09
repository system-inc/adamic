// Original pinned libraries disagree before Adamic is involved.
import {createRequire} from 'node:module';
import {join} from 'node:path';
const require=createRequire(join(process.argv[2],'package.json'));
const yaml=require('yaml');
const prettier=require('prettier');
if(require('yaml/package.json').version!=='2.9.0'||prettier.version!=='3.9.6')throw new Error('wrong pins');
for(const source of ['!!binary 2000-01-01\n','!!binary <<\n','!<toString> word\n']){
 const document=yaml.parseDocument(source,{prettyErrors:false,keepSourceTokens:true,uniqueKeys:false,merge:true});
 console.log(`yaml: ${document.errors[0]?.message??'accepted'}`);
 try{await prettier.format(source,{parser:'yaml'});console.log('prettier: accepted');}
 catch(error){console.log(`prettier: ${error.cause?.message??error.message}`);}
}
