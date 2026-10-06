// Both oracles are installed in scratch at the versions in the Go pin's notices.
import fs from 'node:fs';
import { pathToFileURL } from 'node:url';
const root = process.argv[2];
const mode = process.argv[3];
const manifest = process.argv[4];
const library = await import(pathToFileURL(`${root}/node_modules/${mode === 'raw' ? '@typescript-eslint/typescript-estree/dist/index.js' : 'prettier/index.mjs'}`).href);
const skip = new Set(['loc', 'range', 'comments', 'tokens', 'parent', 'start', 'end', 'extra']);
function dump(value) {
  if(value === null || value === undefined) return null;
  if(typeof value === 'bigint' || value instanceof RegExp) return null;
  if(Array.isArray(value)) return value.map(dump);
  if(typeof value !== 'object') return value;
  if(value.type === undefined) return Object.fromEntries(Object.keys(value).sort().map(key => [key, dump(value[key])]));
  const result = {type: value.type, range: value.range};
  if(value.__contentEnd !== undefined) result.__contentEnd = value.__contentEnd;
  for(const key of Object.keys(value).sort()) if(!skip.has(key) && key !== 'type' && !key.startsWith('__')) result[key] = dump(value[key]);
  if(result.type === 'Literal' && (value.bigint || value.regex)) result.value = null;
  return result;
}
for(const path of fs.readFileSync(manifest, 'utf8').split('\n').filter(Boolean)) {
  const source = fs.readFileSync(path, 'utf8');
  const text = source.startsWith('#!') ? '//' + source.slice(2) : source;
  let ast;
  try {
    ast = mode === 'raw' ? library.parse(text, {filePath: path, sourceType: 'module', range: true, loc: false, comment: true, tokens: false, warnOnUnsupportedTypeScriptVersion: false}) : (await library.__debug.parse(source, {parser: 'typescript', filepath: path})).ast;
    const comments = (ast.comments ?? []).map(comment => ({type:comment.type, range:comment.range, value:comment.value}));
    process.stdout.write(JSON.stringify({ast:dump(ast),comments})+'\n');
  } catch(error) { process.stdout.write(JSON.stringify({error:error.message.split('\n')[0]})+'\n'); }
}
