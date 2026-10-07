// Serialize the declared graph directly: ordinary JSON.stringify(object) would
// expose hidden fields and reorder numeric keys. V8 writes scalar JSON tokens.
const workerEncodeWalkers = new WeakMap();

function compileWorkerEncode(descriptor) {
	const nameOf = field => field.nameUnits === undefined ? field.name : field.nameUnits.map(unit => String.fromCharCode(unit)).join('');
	const writers = [];
	function compile(index) {
		if (writers[index]) return writers[index];
		let write;
		writers[index] = value => write(value);
		const node = descriptor.nodes[index];
		switch (node.kind) {
			case 'number': case 'boolean': case 'string': case 'literal':
				write = value => { return JSON.stringify(value); };
				break;
			case 'array': {
				const element = compile(node.children[0]);
				write = value => '[' + value.map(element).join(',') + ']';
				break;
			}
			case 'tuple': {
				const fields = node.fields.map(field => compile(field.node));
				write = value => '[' + fields.map((field, index) => field(value[index])).join(',') + ']';
				break;
			}
			case 'object': {
				const fields = node.fields.map(field => {
					const name = nameOf(field);
					return { name, prefix: JSON.stringify(name) + ':', optional: field.optional, write: compile(field.node) };
				});
				write = value => {
					const entries = [];
					for (const field of fields) {
						const name = field.name;
						if (field.optional && value[name] === undefined) continue;
						entries.push(field.prefix + field.write(value[name]));
					}
					return '{' + entries.join(',') + '}';
				};
				break;
			}
			case 'union': {
				const objects = node.children.filter(index => descriptor.nodes[index].kind === 'object');
				const branches = node.children.map(index => {
					const candidate = descriptor.nodes[index];
					let test;
					switch (candidate.kind) {
						case 'literal': {
							const wanted = candidate.of === 3 ? (candidate.literalUnits ?? []).map(unit => String.fromCharCode(unit)).join('') : candidate.of === 1 ? Number(candidate.numberText) : candidate.boolean ?? false;
							test = value => value === wanted; break;
						}
						case 'number': test = value => typeof value === 'number'; break;
						case 'string': test = value => typeof value === 'string'; break;
						case 'boolean': test = value => typeof value === 'boolean'; break;
						case 'array': case 'tuple': test = Array.isArray; break;
						case 'object': {
							const object = value => value !== null && typeof value === 'object' && !Array.isArray(value);
							if (objects.length === 1) test = object;
							else {
								const tag = node.discriminantUnits === undefined ? node.discriminant : node.discriminantUnits.map(unit => String.fromCharCode(unit)).join('');
								const field = candidate.fields.find(field => nameOf(field) === tag);
								const literal = descriptor.nodes[field.node];
								const wanted = literal.of === 3 ? (literal.literalUnits ?? []).map(unit => String.fromCharCode(unit)).join('') : literal.of === 1 ? Number(literal.numberText) : literal.boolean ?? false;
								test = value => object(value) && value[tag] === wanted;
							}
							break;
						}
						default: throw new Error(`Unsupported encodeJson union child: ${candidate.kind}`);
					}
					return { test, write: compile(index) };
				});
				write = value => {
					for (const branch of branches) if (branch.test(value)) return branch.write(value);
					throw new Error('encodeJson value does not match its proven union');
				};
				break;
			}
			default: throw new Error(`Unsupported encodeJson descriptor kind: ${node.kind}`);
		}
		writers[index] = write;
		return write;
	}
	return compile(descriptor.root);
}

export function encodeJson(value, descriptor) {
	let walker = workerEncodeWalkers.get(descriptor);
	if (!walker) {
		walker = compileWorkerEncode(descriptor);
		workerEncodeWalkers.set(descriptor, walker);
	}
	return walker(value);
}
