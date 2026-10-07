// Serialize the declared graph directly: ordinary JSON.stringify(object) would
// expose hidden fields and reorder numeric keys. V8 writes scalar JSON tokens.
export function encodeJson(value, descriptor) {
	const nameOf = field => field.nameUnits === undefined ? field.name : field.nameUnits.map(unit => String.fromCharCode(unit)).join('');
	function accepts(node, value) {
		if (node.kind === 'literal') {
			if (node.of === 3) return value === (node.literalUnits ?? []).map(unit => String.fromCharCode(unit)).join('');
			if (node.of === 1) return value === Number(node.numberText);
			return value === (node.boolean ?? false);
		}
		if (node.kind === 'array' || node.kind === 'tuple') return Array.isArray(value);
		if (node.kind === 'object') return value !== null && typeof value === 'object' && !Array.isArray(value);
		return typeof value === node.kind;
	}
	function write(index, value) {
		const node = descriptor.nodes[index];
		switch (node.kind) {
			case 'number': case 'boolean': case 'string': case 'literal':
				return JSON.stringify(value);
			case 'array':
				return '[' + value.map(element => write(node.children[0], element)).join(',') + ']';
			case 'tuple':
				return '[' + node.fields.map((field, index) => write(field.node, value[index])).join(',') + ']';
			case 'object': {
				const entries = [];
				for (const field of node.fields) {
					const name = nameOf(field);
					if (field.optional && value[name] === undefined) continue;
					entries.push(JSON.stringify(name) + ':' + write(field.node, value[name]));
				}
				return '{' + entries.join(',') + '}';
			}
			case 'union': {
				const objects = node.children.filter(index => descriptor.nodes[index].kind === 'object');
				for (const child of node.children) {
					const candidate = descriptor.nodes[child];
					if (!accepts(candidate, value)) continue;
					if (candidate.kind === 'object' && objects.length > 1) {
						const tag = node.discriminantUnits === undefined ? node.discriminant : node.discriminantUnits.map(unit => String.fromCharCode(unit)).join('');
						const field = candidate.fields.find(field => nameOf(field) === tag);
						if (!accepts(descriptor.nodes[field.node], value[tag])) continue;
					}
					return write(child, value);
				}
				throw new Error('encodeJson value does not match its proven union');
			}
			default: throw new Error(`Unsupported encodeJson descriptor kind: ${node.kind}`);
		}
	}
	return write(descriptor.root, value);
}
