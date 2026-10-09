// Build a scratch Node CLI witness; production/native builds remain external.
const fs=require('node:fs'),path=require('node:path');
const [dependencies,source,out]=process.argv.slice(2);
require(path.join(path.resolve(dependencies),'node_modules/esbuild')).buildSync({
 entryPoints:[path.join(path.resolve(source),'src/tsc/tsc.ts')],outfile:path.join(path.resolve(out),'tsc.cjs'),
 bundle:true,platform:'node',format:'cjs'});
for (const file of fs.readdirSync(path.join(dependencies,'built/local')).filter(f=>/^lib.*\.d\.ts$/.test(f)))
 fs.copyFileSync(path.join(dependencies,'built/local',file),path.join(out,file));
