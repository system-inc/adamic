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

export function decodeJson(text, descriptor) {
	const units = (encoded, fallback = '') => encoded === undefined ? fallback : encoded.map(unit => String.fromCharCode(unit)).join('');
	const jsonKind = value => value === null ? 'null' : Array.isArray(value) ? 'array' : typeof value;
	const missing = (path, name) => { throw new WorkerJsonError(`at ${path}: missing field ${name}`); };
	const mismatch = (node, path, value) => { throw new WorkerJsonError(`at ${path}: expected ${node.expected}, found ${jsonKind(value)}`); };
	function matches(node, value) {
		if (node.kind === 'literal') {
			if (node.of === 3) return typeof value === 'string' && value === units(node.literalUnits);
			if (node.of === 1) return typeof value === 'number' && value === Number(node.numberText);
			return typeof value === 'boolean' && value === (node.boolean ?? false);
		}
		return node.kind === jsonKind(value);
	}
	function validate(index, value, path) {
		const node = descriptor.nodes[index];
		switch (node.kind) {
			case 'number': case 'boolean': case 'string': case 'literal':
				if (!matches(node, value)) mismatch(node, path, value);
				return value;
			case 'array':
				if (!Array.isArray(value)) mismatch(node, path, value);
				return value.map((element, index) => validate(node.children[0], element, `${path}[${index}]`));
			case 'tuple':
				if (!Array.isArray(value) || value.length !== node.fields.length) mismatch(node, path, value);
				return node.fields.map((field, index) => validate(field.node, value[index], `${path}[${index}]`));
			case 'object': {
				if (jsonKind(value) !== 'object') mismatch(node, path, value);
				const result = {};
				for (const field of node.fields) {
					const name = units(field.nameUnits, field.name);
					if (!Object.hasOwn(value, name)) {
						if (field.optional) continue;
						missing(path, name);
					}
					Object.defineProperty(result, name, { value: validate(field.node, value[name], `${path}.${name}`), enumerable: true, writable: true, configurable: true });
				}
				return result;
			}
			case 'union': {
				const children = node.children.map(index => ({ index, schema: descriptor.nodes[index] }));
				if (jsonKind(value) === 'object') {
					const objects = children.filter(child => child.schema.kind === 'object');
					if (objects.length === 1) return validate(objects[0].index, value, path);
					if (objects.length > 1) {
						const name = units(node.discriminantUnits, node.discriminant);
						if (!Object.hasOwn(value, name)) missing(path, name);
						for (const child of objects) {
							const field = child.schema.fields.find(field => units(field.nameUnits, field.name) === name);
							if (matches(descriptor.nodes[field.node], value[name])) return validate(child.index, value, path);
						}
						mismatch(node, `${path}.${name}`, value[name]);
					}
				}
				for (const child of children) if (matches(child.schema, value)) return validate(child.index, value, path);
				mismatch(node, path, value);
			}
			default: throw new Error(`Unsupported decodeJson descriptor kind: ${node.kind}`);
		}
	}
	try {
		return { kind: 'Ok', value: validate(descriptor.root, parseWorkerJson(text), '$') };
	} catch (error) {
		if (!(error instanceof WorkerJsonError)) throw error;
		return { kind: 'Error', message: error.message };
	}
}
