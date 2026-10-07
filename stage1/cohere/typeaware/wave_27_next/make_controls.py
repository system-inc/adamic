from pathlib import Path
import json,sys
root=Path(sys.argv[1]);root.mkdir(parents=True,exist_ok=True);files=[]
prelude='''declare namespace NodeJS { interface Process { exit(code?:number):never; stdout:{write(text:string,cb?:()=>void):void};stderr:{write(text:string):void};exitCode:number;on(name:string,fn:()=>void):void;} }
declare var process:NodeJS.Process;
declare const flag:boolean;declare const value:string;declare function work():Promise<string>;declare function use(value:unknown):void;declare function run():void;declare const ms:number;
declare module 'node:process' {export = process;}
declare module 'NodeJS' {export interface Process {exit(n:number):never;}export const fake:Process;}
'''
(root/'prelude.d.ts').write_text(prelude)
(root/'source/system').mkdir(parents=True,exist_ok=True)
(root/'source/system/StandardStreams.a').write_text('export function blockStandardStreams():void {}\n')
alias=root/'source/system/StandardStreams.ts'
if not alias.exists():alias.symlink_to('StandardStreams.a')
(root/'writer.a').write_text('export function print(){console.log("writer");}\n')
alias=root/'writer.ts'
if not alias.exists():alias.symlink_to('writer.a')
(root/'loading.a').write_text('import {blockStandardStreams} from "./source/system/StandardStreams.js";blockStandardStreams();\n')
alias=root/'loading.ts'
if not alias.exists():alias.symlink_to('loading.a')
def add(body):
 p=root/f'control-{len(files):03}.a';p.write_text(body+'\nexport {};\n');files.append(str(p))
for body in [
 'console.log("x");process.exit(0);',
 'process.exit(0);console.log("x");process.exit(1);',
 'if(flag){console.log("x");}process.exit(0);',
 'if(flag){process.exit(0);}console.log("x");',
 'while(flag){process.exit(0);console.log("x");}',
 'while(flag){if(value){process.exit(0);}console.log("x");}',
 'for(let index=0;index<3;index++){console.log("x");}process.exit(0);',
 'try{console.log("x");}catch{process.exit(0);}',
 'console.log("x");try{run();}catch{process.exit(0);}',
 'try{run();}catch{console.error("x");process.exit(0);}',
 'try{console.log("x");}finally{process.exit(0);}',
 'console.log(process.exit(0));process.exit(1);',
 'process.stdout.write("x",()=>process.exit(0));',
 'process.stderr.write("x");process.exit(0);',
 'console["log"]("x");process.exit(0);',
 'console["\\x6cog"]("x");process.exit(0);',
 'function help(){console.log("x");}help();process.exit(0);',
 'function help(){console.log("x");process.exit(0);}help();process.exit(0);',
 'function* help(){console.log("x");}help();process.exit(0);',
 'async function help(){console.log("x");}help();process.exit(0);',
 'async function main(){async function help(){console.log("x");}await help();process.exit(0);}main();',
 'function help(){console.log("x");}function main(){help();process.exit(0);}main();',
 'export function unused(){console.log("x");process.exit(0);}',
 'const local={log(x:string){}};local.log("x");process.exit(0);',
 'const console={log(x:string){}};console.log("x");process.exit(0);',
 'const process={exit(n:number){},stdout:{write(x:string){}}};console.log("x");process.exit(0);',
 'const main=()=>{console.log("x");process.exit(0);};main();',
 'function* main(){console.log("x");process.exit(0);}use(main());',
 'process.on("signal",()=>{console.log("x");process.exit(0);});',
 'switch(value){case "a":console.log("x");case "b":process.exit(0);}',
 'outer:while(flag){console.log("x");break outer;}process.exit(0);',
 'console.log("x");function cb(){process.exit(0);}use(cb);',
 'import {print} from "./writer.js";print();process.exit(0);',
 'import * as NodeProcess from "node:process";console.log("x");NodeProcess.exit(0);',
 'import {exit,stdout} from "node:process";stdout.write("x");exit(0);',
 '/* 世界🌍 */\r\nconsole.log("é");process.exit(0);\r\n',
 'console.log("x");if(flag){process.exit(0);}else{process.exit(1);}',
 '(()=>{console.log("x");process.exit(0);})();',
 'class A {static {console.log("x");process.exit(0);}}',
 'class A {x=(()=>{console.log("x");process.exit(0);})();}',
]:add(body)
for method in ['info','debug','warn','error','trace','table','dir','dirxml']:add(f'console.{method}("x");process.exit(0);')
for body in [
 'console.log("x");process.exit(0);blockStandardStreams();',
 'blockStandardStreams();console.log("x");process.exit(0);',
 'function main(){console.log("x");process.exit(0);blockStandardStreams();}main();',
 'function main(){blockStandardStreams();console.log("x");process.exit(0);}main();',
 'if(flag){console.log("x");process.exit(0);}blockStandardStreams();',
 'const main=()=>{console.log("x");process.exit(0);};main();blockStandardStreams();',
 'process.on("signal",()=>{console.log("x");process.exit(0);});blockStandardStreams();',
 'async function main(){await work();console.log("x");process.exit(0);}main();',
 'async function main(){console.log("x");process.exit(0);await work();}main();',
 'declare function unknown():void;unknown();console.log("x");process.exit(0);',
 'function main(){console.log("x");process.exit(0);}unknown(main);',
 'function main(){console.log("x");process.exit(0);}function unknown(fn:()=>void){}unknown(main);',
]:add('import {blockStandardStreams} from "./source/system/StandardStreams.js";\n'+body)
add('import "./loading.js";console.log("x");process.exit(0);')
add('#!/usr/bin/env tsx\nconsole.log("x");process.exit(0);')
for timer in ['setTimeout(()=>{},ms);','const timer=setTimeout(()=>{},ms);','let timer;timer=setTimeout(()=>{},ms);','void setTimeout(()=>{},ms);','globalThis.setTimeout(()=>{},ms);','window.setTimeout(()=>{},ms);','setTimeout(()=>{},ms).unref();','const timer=setTimeout(()=>{},ms);use(timer);','let timer;timer=setTimeout(()=>{},ms);clearTimeout(timer);','let timer;use(timer=setTimeout(()=>{},ms));','return setTimeout(()=>{},ms);','use(setTimeout(()=>{},ms));','use(()=>{setTimeout(()=>{},ms);});','setInterval(()=>{},ms);','const {timer}=setTimeout(()=>{},ms);','const timer=setTimeout(()=>{},ms);use({timer});','const timer=setTimeout(()=>{},ms);{const timer=1;use(timer);}','const timer=setTimeout(()=>{},ms);timer=2;']:
 add('Promise.race([work(),new Promise((_resolve,reject)=>{'+timer+'})]);')
add('Promise.race([work(),new Promise((_resolve,reject)=>setTimeout(reject,ms))]);')
add('const timeout=new Promise((_resolve,reject)=>{setTimeout(reject,ms);});Promise.race([work(),timeout]);')
add('let timeout=new Promise((_resolve,reject)=>{setTimeout(reject,ms);});Promise.race([work(),timeout]);')
add('function setTimeout(callback:()=>void,ms:number){return 1;}Promise.race([new Promise(()=>{setTimeout(()=>{},ms);})]);')
add('const pool={race(x:unknown[]){}};pool.race([new Promise(()=>{setTimeout(()=>{},ms);})]);')
add('Promise.all([new Promise(()=>{setTimeout(()=>{},ms);})]);')
add('const timeout=new Promise(()=>{setTimeout(()=>{},ms);});Promise.race([timeout,timeout]);')
add('const {console}={console:{log(x:string){}}};console.log("x");process.exit(0);')
add('const {pool}={pool:{race(x:unknown[]){return x;}}};pool.race([new Promise(()=>{setTimeout(()=>{},ms);})]);')
add('console[("log")]("x");process.exit(0);')
add('console[`log`]("x");process.exit(0);')
add('import {fake as process} from "NodeJS";console.log("x");process.exit(0);')
add('async function load(){await using timeout=new Promise(()=>{setTimeout(()=>{},ms);});await Promise.race([work(),timeout]);}load();')
(root/'controls.manifest').write_text('\n'.join(files)+'\n')
(root/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'NodeNext','moduleResolution':'NodeNext','lib':['ES2022','DOM'],'allowImportingTsExtensions':True,'noEmit':True},'files':['prelude.d.ts']}))

print(len(files),'controls')
