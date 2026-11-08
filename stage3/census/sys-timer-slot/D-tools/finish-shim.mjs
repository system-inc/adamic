// Invoke only upstream's existing lssl writer. The failed typed build already
// emitted its services bundle. No checker or source file is changed here.
import { pathToFileURL } from 'node:url';
const root=process.argv[2];
const {local}=await import(pathToFileURL(root+'/Herebyfile.mjs').href);
const task=local.options.dependencies.find(t=>t.options.name==='lssl');
if(!task||typeof task.options.run!=='function')throw Error('missing upstream lssl task');
await task.options.run();
console.log('upstream lssl writer completed');
