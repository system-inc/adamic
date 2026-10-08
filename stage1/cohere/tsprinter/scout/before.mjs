// Snapshot of formatFile at area/stage1-lint 9156bf5c; actual TSX filename supplied.
import {readFileSync} from 'node:fs';
import {Parser} from '../../../typescript/parser/parser.ts';
import {formatProgram} from '../expressions.ts';
import {hasBlankLine} from '../syntax.ts';
for(const {source,path} of JSON.parse(readFileSync(process.argv[2],'utf8'))) {
 let result;
 if(source.startsWith('#!') || source.startsWith('\ufeff#!') || source.includes('/*') || source.includes('//')) result={kind:'NotYet',reason:'comment-attachment'};
 else if(hasBlankLine(source) || source.includes('\r')) result={kind:'NotYet',reason:'source-trivia'};
 else result=formatProgram(new Parser(source,path),source,{printWidth:120,tabWidth:4,useTabs:false});
 process.stdout.write(JSON.stringify(result)+'\n');
}
