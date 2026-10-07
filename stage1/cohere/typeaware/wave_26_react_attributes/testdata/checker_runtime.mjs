export * from '../../../../../oracle/adamic.mjs';
import {panic} from '../../../../../oracle/adamic.mjs';
import {openSync,readSync,writeSync,closeSync} from 'node:fs';
const input=openSync(process.env.ADAMIC_REACT_INPUT,'r+');
const output=openSync(process.env.ADAMIC_REACT_OUTPUT,'r+');
let live=false;
function ask(value) {
    writeSync(input,JSON.stringify(value)+'\n');
    const byte=Buffer.alloc(1);const parts=[];
    while(readSync(output,byte,0,1,null)===1) {if(byte[0]===10) {break;}parts.push(byte[0]);}
    const answer=JSON.parse(Buffer.from(parts).toString('utf8'));
    if(answer.Error) {panic(answer.Error);}
    return answer.Value;
}
export function tsgoProgram(config,files) {ask({Command:'open',Config:config,Files:files});live=true;return 1;}
export function tsgoInspect(program,file,start,end,kind,question) {
    if(program!==1||!live) {panic('invalid or released checker handle');}
    return ask({Command:'inspect',File:file,Start:start,End:end,Kind:kind,Question:question});
}
export function tsgoRelease(program) {
    if(program!==1||!live) {panic('invalid or released checker handle');}
    ask({Command:'release'});live=false;closeSync(input);closeSync(output);
}
