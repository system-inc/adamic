// Build-check every selected stock site without running a compiler workload.
const fs=require('fs'),path=require('path'),cp=require('child_process');
const [stockArg,scratchArg]=process.argv.slice(2),stock=path.resolve(stockArg),scratch=path.resolve(scratchArg),out=__dirname;
if(fs.existsSync(scratch))throw Error('Scratch directory exists');
fs.mkdirSync(scratch,{recursive:true});fs.cpSync(path.join(stock,'src'),path.join(scratch,'src'),{recursive:true});
fs.copyFileSync(path.join(stock,'package.json'),path.join(scratch,'package.json'));
fs.symlinkSync(path.join(stock,'node_modules'),path.join(scratch,'node_modules'),'dir');
const result=cp.spawnSync(process.execPath,[path.join(out,'transform.cjs'),path.resolve(out,'../../api/node_modules/typescript/lib/typescript.js'),path.join(out,'stock.json'),scratch],{encoding:'utf8'});
fs.writeFileSync(path.join(out,'transform-validation.log'),(result.stdout||'')+(result.stderr||''));
console.log(`transform child exit code: ${result.status}`);if(result.status!==0)process.exit(result.status||1);
const esbuild=require(path.join(stock,'node_modules/esbuild'));
esbuild.buildSync({entryPoints:[path.join(scratch,'src/tsc/tsc.ts')],outfile:path.join(scratch,'tsc.js'),bundle:true,platform:'node',target:['es2020','node14.17'],format:'cjs',packages:'external',logLevel:'warning'});
const row={passed:true,...JSON.parse(result.stdout),bundleBytes:fs.statSync(path.join(scratch,'tsc.js')).size};
fs.writeFileSync(path.join(out,'transform-validation.json'),JSON.stringify(row,null,2)+'\n');console.log(JSON.stringify(row));
