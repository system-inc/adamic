// Workers decoder, derived from the JSON grammar and compiler schema graph.
// Parsing validates every container, including values omitted by the schema.
class WorkerJsonError extends Error {}

function parseWorkerJson(text) {
	let position = 0;
	function fail(reason) {
		const before = text.slice(0, position);
		const line = before.split('\n').length;
		const column = position - before.lastIndexOf('\n');
		throw new WorkerJsonError(`invalid JSON at line ${line} column ${column}: ${reason}`);
	}
	function space() {
		while (position < text.length && ' \t\r\n'.includes(text[position])) position++;
	}
	function string() {
		const start = position++;
		while (position < text.length) {
			const unit = text.charCodeAt(position);
			if (unit === 34) {
				position++;
				return JSON.parse(text.slice(start, position));
			}
			if (unit < 32) fail('control character in string');
			if (unit === 92) {
				position++;
				if (position === text.length) fail('unterminated string');
				const escape = text[position];
				if (escape === 'u') {
					position++;
					for (let digit = 0; digit < 4; digit++) {
						if (position === text.length || !/[0-9a-fA-F]/.test(text[position])) fail('expected four hexadecimal digits');
						position++;
					}
					continue;
				}
				if (!'"\\/bfnrt'.includes(escape)) fail('invalid string escape');
			}
			position++;
		}
		fail('unterminated string');
	}
	function number() {
		const start = position;
		if (text[position] === '-') position++;
		const digit = () => position < text.length && text[position] >= '0' && text[position] <= '9';
		if (!digit()) fail('expected digit');
		if (text[position] === '0') position++;
		else while (digit()) position++;
		if (text[position] === '.') {
			position++;
			if (!digit()) fail('expected digit');
			while (digit()) position++;
		}
		if (text[position] === 'e' || text[position] === 'E') {
			position++;
			if (text[position] === '+' || text[position] === '-') position++;
			if (!digit()) fail('expected digit');
			while (digit()) position++;
		}
		return Number(text.slice(start, position));
	}
	function value(depth) {
		space();
		const first = text[position];
		if (first === '"') return string();
		if (first === '-' || (first >= '0' && first <= '9')) return number();
		if (first === 't' || first === 'f' || first === 'n') {
			const keyword = first === 't' ? 'true' : first === 'f' ? 'false' : 'null';
			for (const unit of keyword) {
				if (text[position] !== unit) fail('invalid keyword');
				position++;
			}
			return first === 'n' ? null : first === 't';
		}
		if (first !== '[' && first !== '{') fail('expected JSON value');
		if (depth >= 128) fail('nesting depth exceeds 128');
		position++;
		space();
		if (first === '[') {
			const array = [];
			if (text[position] !== ']') {
				for (;;) {
					array.push(value(depth + 1));
					space();
					if (text[position] === ']') break;
					if (text[position] !== ',') fail("expected ',' or ']'");
					position++;
				}
			}
			position++;
			return array;
		}
		const object = Object.create(null);
		if (text[position] !== '}') {
			for (;;) {
				space();
				if (text[position] !== '"') fail('expected object key');
				const key = string();
				space();
				if (text[position] !== ':') fail("expected ':'");
				position++;
				object[key] = value(depth + 1);
				space();
				if (text[position] === '}') break;
				if (text[position] !== ',') fail("expected ',' or '}'");
				position++;
			}
		}
		position++;
		return object;
	}
	const parsed = value(0);
	space();
	if (position !== text.length) fail('expected end of input');
	return parsed;
}

// JSON.parse has already checked grammar. Count every raw container, including
// overwritten duplicate values and fields that schema validation will drop.
function withinWorkerJsonDepth(text) {
	let depth = 0;
	let quoted = false;
	for (let index = 0; index < text.length; index++) {
		const unit = text.charCodeAt(index);
		if (quoted) {
			if (unit === 92) index++;
			else if (unit === 34) quoted = false;
		} else if (unit === 34) quoted = true;
		else if (unit === 123 || unit === 91) {
			if (++depth > 128) return false;
		} else if (unit === 125 || unit === 93) depth--;
	}
	return true;
}

function parseWorkerJsonFast(text) {
	let parsed;
	try {
		parsed = JSON.parse(text);
	} catch {
		return parseWorkerJson(text);
	}
	if (!withinWorkerJsonDepth(text)) return parseWorkerJson(text);
	return parsed;
}

const workerDecodeWalkers = new WeakMap();
const workerJsonUnits = (encoded, fallback = '') => encoded === undefined ? fallback : encoded.map(unit => String.fromCharCode(unit)).join('');
const workerJsonKind = value => value === null ? 'null' : Array.isArray(value) ? 'array' : typeof value;
function workerJsonMissing(path, name) { throw new WorkerJsonError(`at ${path}: missing field ${name}`); }
function workerJsonMismatch(expected, path, value) { throw new WorkerJsonError(`at ${path}: expected ${expected}, found ${workerJsonKind(value)}`); }

function compileWorkerDecode(descriptor) {
	const readers = [];
	const matches = [];
	const object = value => value !== null && typeof value === 'object' && !Array.isArray(value);
	function literal(node) {
		if (node.of === 3) return workerJsonUnits(node.literalUnits);
		if (node.of === 1) return Number(node.numberText);
		return node.boolean ?? false;
	}
	function compile(index) {
		if (readers[index]) return readers[index];
		let read;
		// Install a forward reference before descending, for recursive graphs.
		readers[index] = (value, path) => read(value, path);
		const node = descriptor.nodes[index];
		const expected = node.expected;
		switch (node.kind) {
			case 'number': matches[index] = value => typeof value === 'number'; break;
			case 'string': matches[index] = value => typeof value === 'string'; break;
			case 'boolean': matches[index] = value => typeof value === 'boolean'; break;
			case 'literal': { const wanted = literal(node); matches[index] = value => value === wanted; break; }
			case 'array': case 'tuple': matches[index] = Array.isArray; break;
			case 'object': matches[index] = object; break;
		}
		switch (node.kind) {
			case 'number': case 'string': case 'boolean': case 'literal': {
				const accepts = matches[index];
				read = (value, path) => { if (!accepts(value)) workerJsonMismatch(expected, path, value); return value; };
				break;
			}
			case 'array': {
				const element = compile(node.children[0]);
				read = (value, path) => {
					if (!Array.isArray(value)) workerJsonMismatch(expected, path, value);
					return value.map((value, index) => element(value, `${path}[${index}]`));
				};
				break;
			}
			case 'tuple': {
				const fields = node.fields.map(field => compile(field.node));
				read = (value, path) => {
					if (!Array.isArray(value) || value.length !== fields.length) workerJsonMismatch(expected, path, value);
					return fields.map((field, index) => field(value[index], `${path}[${index}]`));
				};
				break;
			}
			case 'object': {
				const fields = node.fields.map(field => {
					const name = workerJsonUnits(field.nameUnits, field.name);
					const set = name === '__proto__'
						? (result, value) => Object.defineProperty(result, name, { value, enumerable: true, writable: true, configurable: true })
						: (result, value) => { result[name] = value; };
					return { name, optional: field.optional, read: compile(field.node), set };
				});
				read = (value, path) => {
					if (!object(value)) workerJsonMismatch(expected, path, value);
					const result = {};
					for (const field of fields) {
						const name = field.name;
						if (!Object.hasOwn(value, name)) {
							if (field.optional) continue;
							workerJsonMissing(path, name);
						}
						field.set(result, field.read(value[name], `${path}.${name}`));
					}
					return result;
				};
				break;
			}
			case 'union': {
				const children = node.children.map(index => ({ read: compile(index), accepts: matches[index] }));
				const objects = node.children.filter(index => descriptor.nodes[index].kind === 'object');
				const objectRead = objects.length === 1 ? compile(objects[0]) : undefined;
				const tag = workerJsonUnits(node.discriminantUnits, node.discriminant);
				const branches = new Map();
				if (objects.length > 1) for (const index of objects) {
					const field = descriptor.nodes[index].fields.find(field => workerJsonUnits(field.nameUnits, field.name) === tag);
					branches.set(literal(descriptor.nodes[field.node]), compile(index));
				}
				read = (value, path) => {
					if (object(value)) {
						if (objectRead) return objectRead(value, path);
						if (branches.size) {
							if (!Object.hasOwn(value, tag)) workerJsonMissing(path, tag);
							const branch = branches.get(value[tag]);
							if (branch) return branch(value, path);
							workerJsonMismatch(expected, `${path}.${tag}`, value[tag]);
						}
					}
					for (const child of children) if (child.accepts(value)) return child.read(value, path);
					workerJsonMismatch(expected, path, value);
				};
				matches[index] = value => children.some(child => child.accepts(value));
				break;
			}
			default: throw new Error(`Unsupported decodeJson descriptor kind: ${node.kind}`);
		}
		readers[index] = read;
		return read;
	}
	return compile(descriptor.root);
}

export function decodeJson(text, descriptor) {
	try {
		const value = parseWorkerJsonFast(text);
		let walker = workerDecodeWalkers.get(descriptor);
		if (!walker) {
			walker = compileWorkerDecode(descriptor);
			workerDecodeWalkers.set(descriptor, walker);
		}
		return { kind: 'Ok', value: walker(value, '$') };
	} catch (error) {
		if (!(error instanceof WorkerJsonError)) throw error;
		return { kind: 'Error', message: error.message };
	}
}
