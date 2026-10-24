import { realpathSync } from 'node:fs';
import { readFile, writeFile, copyFile, mkdir } from 'node:fs/promises';
import { basename, dirname, resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const quote = JSON.stringify;
function validateType(type, depth = 0, result = false) {
  if (!type || typeof type !== 'object') throw new Error('Missing ABI type');
  if (type.kind === 'record') {
    if (depth >= 2) throw new Error('ABI record depth exceeds 2');
    const fields = type.fields ?? [];
    if (!Array.isArray(fields)) throw new Error('ABI fields must be an array');
    const names = new Set();
    for (const field of fields) {
      if (typeof field.name !== 'string' || names.has(field.name)) throw new Error('Invalid ABI field name');
      names.add(field.name); validateType(field.type, depth + 1);
    }
  } else if (!['number', 'boolean', 'string', 'number[]', 'string[]', ...(result ? ['void'] : [])].includes(type.kind)) {
    throw new Error(`Unsupported ABI kind: ${type.kind}`);
  }
}
export function generateSource(table) {
  if (table.version !== 1 || !Array.isArray(table.exports)) throw new Error('Unsupported Adamic ABI table');
  const names = new Set();
  for (const signature of table.exports) {
    if (typeof signature.name !== 'string' || names.has(signature.name) || !Array.isArray(signature.parameters)) throw new Error('Invalid ABI export');
    names.add(signature.name);
    for (const parameter of signature.parameters) validateType(parameter.type);
    validateType(signature.returns, 0, true);
  }
  let serial = 0;
  const fresh = () => `temporary${serial++}`;
  function encode(type, value, writer) {
    switch (type.kind) {
      case 'record': return (type.fields ?? []).map(field => encode(field.type, `${value}[${quote(field.name)}]`, writer)).join('\n');
      case 'number[]': {
        const item = fresh();
        return `${writer}.count(${value}.length);\n${writer}.align();\nfor (const ${item} of ${value}) ${writer}.number(${item});`;
      }
      case 'string[]': {
        const item = fresh();
        return `${writer}.count(${value}.length);\nfor (const ${item} of ${value}) ${writer}.string(${item});`;
      }
      default: return `${writer}.${type.kind}(${value});`;
    }
  }
  function decode(type, target, reader) {
    if (type.kind === 'record') {
      return `${target} = {};\n` + (type.fields ?? []).map(field => {
        const temporary = fresh();
        return `let ${temporary};\n${decode(field.type, temporary, reader)}\nObject.defineProperty(${target}, ${quote(field.name)}, { value: ${temporary}, enumerable: true, writable: true, configurable: true });`;
      }).join('\n');
    }
    if (type.kind === 'number[]' || type.kind === 'string[]') {
      const count = fresh(), index = fresh();
      const kind = type.kind.slice(0, -2);
      return `const ${count} = ${reader}.count();\n${kind === 'number' ? `${reader}.align();` : ''}\n${target} = [];\nfor (let ${index} = 0; ${index} < ${count}; ${index}++) ${target}.push(${reader}.${kind}());`;
    }
    return `${target} = ${reader}.${type.kind}();`;
  }
  const wrappers = table.exports.map((signature, index) => {
    const parameters = signature.parameters.map((_, index) => `argument${index}`);
    const owned = [], arguments_ = [], statements = [];
    signature.parameters.forEach(({ type }, index) => {
      const value = parameters[index];
      if (type.kind === 'number') { arguments_.push(value); return; }
      if (type.kind === 'boolean') { arguments_.push(`${value} ? 1 : 0`); return; }
      const bytes = `bytes${index}`, pointer = `input${index}`;
      owned.push(pointer);
      if (type.kind === 'string') statements.push(`const ${bytes} = encodeWTF8(${value});`);
      else if (type.kind === 'number[]') statements.push(`const ${bytes} = directNumbers(${value});`);
      else {
        const writer = `encoder${index}`;
        statements.push(`const ${writer} = new Encoder();\n${encode(type, value, writer)}\nconst ${bytes} = ${writer}.finish();`);
      }
      statements.push(`${pointer} = context.invoke('adamic_alloc', ${bytes}.length) >>> 0;\nnew Uint8Array(api.memory.buffer, ${pointer}, ${bytes}.length).set(${bytes});`);
      arguments_.push(pointer, type.kind === 'number[]' ? `${value}.length` : `${bytes}.length`);
    });
    const kind = signature.returns.kind;
    let result;
    if (kind === 'void') result = 'return undefined;';
    else if (kind === 'number') result = 'return returned;';
    else if (kind === 'boolean') result = 'return returned !== 0;';
    else {
      result = `handle = returned >>> 0;\nconst pointer = context.invoke('adamic_result_bytes', handle) >>> 0;\nconst length = context.invoke('adamic_result_length', handle) >>> 0;\nconst bytes = new Uint8Array(api.memory.buffer, pointer, length);\n`;
      if (kind === 'string') result += 'return decodeWTF8(bytes);';
      else result += `const decoder = new Decoder(bytes);\nlet result;\n${decode(signature.returns, 'result', 'decoder')}\ndecoder.finish();\nreturn result;`;
    }
    return `[${quote(signature.name)}]: function crossing${index}(${parameters.join(', ')}) {
if (arguments.length !== ${parameters.length}) throw new TypeError('ABI argument count mismatch');
return lifecycle.run(context => {
const api = context.api;
${owned.map(pointer => `let ${pointer};`).join('\n')}
let handle;
try {
${statements.join('\n')}
const returned = context.invoke(${quote('adamic_export_' + signature.name)}${arguments_.length ? ', ' + arguments_.join(', ') : ''});
${result}
} finally {
if (context.alive) {
if (handle !== undefined) context.invoke('adamic_result_release', handle);
${owned.map(pointer => `if (${pointer} !== undefined) context.invoke('adamic_free', ${pointer});`).join('\n')}
}
}
});
}`;
  });
  return `// Generated from Adamic ABI version 1. Do not edit.
import { encodeWTF8, decodeWTF8 } from './wtf8.mjs';
import { Encoder, Decoder, directNumbers, createLifecycle } from './crossing-runtime.mjs';
export function createCrossing(module, { logger = console } = {}) {
const lifecycle = createLifecycle(module, ${quote([...names].map(name => 'adamic_export_' + name))}, logger);
return {
${wrappers.join(',\n')}
};
}
`;
}

export async function generateCrossing(abiPath, outputPath) {
  const table = JSON.parse(await readFile(abiPath, 'utf8'));
  const source = generateSource(table);
  const output = resolve(outputPath), directory = dirname(output);
  const template = fileURLToPath(new URL('.', import.meta.url));
  if (['wtf8.mjs', 'crossing-runtime.mjs', 'wasi.mjs', 'generate-crossing.mjs'].includes(basename(output))) throw new Error('Output would overwrite a generator dependency');
  await mkdir(directory, { recursive: true });
  if (resolve(directory) !== resolve(template)) {
    for (const name of ['wtf8.mjs', 'crossing-runtime.mjs', 'wasi.mjs']) await copyFile(join(template, name), join(directory, name));
  }
  await writeFile(output, source);
}
if (process.argv[1] && realpathSync(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 4) throw new Error('Usage: generate-crossing.mjs <abi.json> <out.mjs>');
  await generateCrossing(process.argv[2], process.argv[3]);
}
