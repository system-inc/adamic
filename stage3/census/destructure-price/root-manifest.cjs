const ts=require('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root='/tmp/destructure-price-scratch/adapted',cfg=ts.getParsedCommandLineOfConfigFile(root+'/src/compiler/tsconfig.json',{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'))}});
if(ts.version!=='6.0.3'||cfg.fileNames.length!==79)throw Error('wrong stock/root set');
fs.writeFileSync('/workspace/scratch/destructure-price-f2989762/r2/root-manifest.json',JSON.stringify({typescript:ts.version,rootCount:79,files:cfg.fileNames.map(f=>path.relative(root,f)),options:Object.fromEntries(Object.entries(cfg.options).filter(([key])=>key!=='configFile'))},null,2));
