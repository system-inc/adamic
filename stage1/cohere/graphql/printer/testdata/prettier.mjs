import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
import {join} from 'node:path';
const [directory,casesPath,mode,engine]=process.argv.slice(2);
const require=createRequire(join(directory,'package.json'));
const prettier=engine==='embedded'?require('./standalone.js'):require('prettier');
const plugins=engine==='embedded'?[require('./plugins/graphql.js')]:[];
if(prettier.version!=='3.9.6') throw Error(`expected Prettier 3.9.6, got ${prettier.version}`);
const escape=text=>text.replaceAll('\\','\\\\').replaceAll('\n','\\n').replaceAll('\r','\\r').replaceAll('\t','\\t');
const unescape=text=>text.replace(/\\([\\nrt])/g,(_,c)=>({'\\':'\\',n:'\n',r:'\r',t:'\t'})[c]);
const output=[];
for(const line of readFileSync(casesPath,'utf8').split('\n')) {
 if(!line)continue;
 try {output.push('ok\t'+escape(await prettier.format(unescape(line.slice(1)),{parser:'graphql',plugins,printWidth:mode==='narrow'?80:120,tabWidth:mode==='narrow'?2:4,useTabs:mode==='tabs',bracketSpacing:mode!=='tight'})));}
 catch(error) {output.push('error\t'+escape(error.message));}
}
process.stdout.write(output.join('\n')+'\n');
