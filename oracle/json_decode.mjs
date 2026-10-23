// The value parser here is V8's JSON.parse, independently of the native parser. The recognizer
// supplies stable diagnostics and a shared depth policy, never numbers or decoded strings.
export function decodeJson(text, schema) {
 let parsed, parseFailed = false;
 try { parsed = JSON.parse(text); } catch { parseFailed = true; /* Adamic owns the wording below. */ }
 const syntax = recognize(text);
 if (syntax !== undefined) return { kind: 'Error', message: syntax };
 if (parseFailed) throw new Error('the diagnostic recognizer accepted text JSON.parse refused');
 if (schema === undefined) throw new Error('decodeJson requires its generated type descriptor');
 try {
  return { kind: 'Ok', value: validate(parsed, schema, schema.root, '$') };
 } catch (error) {
  if (typeof error === 'string') return { kind: 'Error', message: error };
  throw error;
 }
}
function recognize(text) {
 let position = 0;
 const peek = () => text[position];
 const space = () => { while (peek() === ' ' || peek() === '\t' || peek() === '\r' || peek() === '\n') position++; };
 const fail = (message) => { throw message; };
 const digit = () => peek() >= '0' && peek() <= '9';
 function string() {
  position++;
  while (position < text.length) {
   const c = text[position++];
   if (c === '"') return;
   if (c.charCodeAt(0) < 32) { position--; fail('control character in string'); }
   if (c === '\\') {
    if (position === text.length) fail('unterminated string');
    const escape = text[position++];
    if (escape === 'u') {
     for (let i = 0; i < 4; i++) {
      if (position === text.length || !/[0-9a-fA-F]/.test(peek())) fail('expected four hexadecimal digits');
      position++;
     }
    } else if (!'"\\/bfnrt'.includes(escape)) { position--; fail('invalid string escape'); }
   }
  }
  fail('unterminated string');
 }
 function value(depth) {
  space();
  const c = peek();
  if (c === '[' || c === '{') {
   if (depth >= 128) fail('nesting depth exceeds 128');
   const object = c === '{', end = object ? '}' : ']';
   position++; space();
   if (peek() === end) { position++; return; }
   for (;;) {
    if (object) {
     if (peek() !== '"') fail('expected object key');
     string(); space();
     if (peek() !== ':') fail("expected ':'");
     position++; space();
    }
    value(depth + 1); space();
    if (peek() === end) { position++; return; }
    if (peek() !== ',') fail(object ? "expected ',' or '}'" : "expected ',' or ']'");
    position++; space();
   }
  }
  if (c === '"') { string(); return; }
  if (c === '-' || digit()) {
   if (c === '-') position++;
   if (peek() === '0') position++;
   else if (peek() >= '1' && peek() <= '9') { while (digit()) position++; }
   else fail('expected digit');
   if (peek() === '.') { position++; if (!digit()) fail('expected digit'); while (digit()) position++; }
   if (peek() === 'e' || peek() === 'E') {
    position++; if (peek() === '+' || peek() === '-') position++;
    if (!digit()) fail('expected digit'); while (digit()) position++;
   }
   return;
  }
  const word = c === 't' ? 'true' : c === 'f' ? 'false' : c === 'n' ? 'null' : undefined;
  if (word === undefined) fail('expected JSON value');
  for (const letter of word) { if (peek() !== letter) fail('invalid keyword'); position++; }
 }
 try {
  value(0); space(); if (position !== text.length) fail('expected end of input');
 } catch (message) {
  const before = text.slice(0, position);
  const line = before.split('\n').length;
  const column = position - before.lastIndexOf('\n');
  return `invalid JSON at line ${line} column ${column}: ${message}`;
 }
}
function kind(value) { return value === null ? 'null' : Array.isArray(value) ? 'array' : typeof value; }
function literal(value, node) {
 return node.of === 1 ? typeof value === 'number' && value === Number(node.numberText ?? 0)
  : node.of === 2 ? typeof value === 'boolean' && value === (node.boolean ?? false)
  : typeof value === 'string' && value === (node.literalUnits ?? []).map(unit => String.fromCharCode(unit)).join('');
}
function validate(value, schema, index, path) {
 const node = schema.nodes[index];
 const discriminant = node.discriminantUnits?.map(unit => String.fromCharCode(unit)).join('') ?? node.discriminant;
 const mismatch = (at = path, found = value) => { throw `at ${at}: expected ${node.expected}, found ${kind(found)}`; };
 if (node.kind === 'union') {
  let tag;
  if (discriminant !== undefined && kind(value) === 'object') {
   if (!Object.hasOwn(value, discriminant)) throw `at ${path}: missing field ${discriminant}`;
   tag = value[discriminant];
  }
  for (const child of node.children) {
   const member = schema.nodes[child];
   const match = member.kind === 'literal' ? literal(value, member)
    : member.kind === 'object' && kind(value) === 'object'
     ? discriminant === undefined || literal(tag, schema.nodes[member.fields.find(f => fieldName(f) === discriminant).node])
     : member.kind === kind(value);
   if (match) return validate(value, schema, child, path);
  }
  mismatch(tag === undefined ? path : `${path}.${discriminant}`, tag === undefined ? value : tag);
 }
 if (!(node.kind === 'literal' ? literal(value, node) : node.kind === 'tuple' ? Array.isArray(value) : node.kind === kind(value))) mismatch();
 if (node.kind === 'array') return value.map((element, index) => validate(element, schema, node.children[0], `${path}[${index}]`));
 if (node.kind === 'tuple') {
  if (value.length !== node.fields.length) mismatch();
  return node.fields.map((field, index) => validate(value[index], schema, field.node, `${path}[${index}]`));
 }
 if (node.kind === 'object') {
  const result = {};
  for (const field of node.fields) {
   const name = fieldName(field);
   if (!Object.hasOwn(value, name)) {
    if (field.optional) continue;
    throw `at ${path}: missing field ${name}`;
   }
   Object.defineProperty(result, name, { value: validate(value[name], schema, field.node, `${path}.${name}`), enumerable: true, writable: true, configurable: true });
  }
  return result;
 }
 return value;
}

function fieldName(field) { return field.nameUnits?.map(unit => String.fromCharCode(unit)).join('') ?? field.name; }
