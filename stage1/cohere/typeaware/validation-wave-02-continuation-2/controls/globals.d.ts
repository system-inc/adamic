export {};import * as console from 'node:console';declare global {interface Console extends console.Console{}var console:Console;
function setTimeout<TArgs extends any[]>(callback:(...args:TArgs)=>void,delay?:number,...args:TArgs):NodeJS.Timeout;
namespace setTimeout{const __promisify__:unknown;}function clearTimeout(timeout:NodeJS.Timeout|undefined):void;}
