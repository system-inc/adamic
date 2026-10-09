import { panic } from './adamic.mjs';

// Only scalar quoting/printing is delegated to V8. Containers are written from the
// declared graph, without invoking toJSON or enumerating runtime fields.
export function encodeJson(value, schema) {
	if (schema === undefined) throw new Error('encodeJson requires its generated type descriptor');
	const units = (array, fallback) => array?.map(unit => String.fromCharCode(unit)).join('') ?? fallback;
	const literal = (value, node) => node.of === 1 ? value === Number(node.numberText)
		: node.of === 2 ? value === (node.boolean ?? false) : value === units(node.literalUnits, '');
	function write(value, index, depth) {
		const node = schema.nodes[index];
		if (node.kind === 'union') {
			const discriminant = units(node.discriminantUnits, node.discriminant);
			for (const child of node.children) {
				const member = schema.nodes[child];
				const matches = member.kind === 'literal' ? literal(value, member)
					: member.kind === 'object' ? value !== null && typeof value === 'object' && !Array.isArray(value)
						&& (discriminant === undefined || literal(value[discriminant], schema.nodes[member.fields.find(field => units(field.nameUnits, field.name) === discriminant).node]))
					: member.kind === typeof value;
				if (matches) return write(value, child, depth);
			}
			panic('encodeJson: value does not match its declared union');
		}
		if (node.kind === 'array' || node.kind === 'tuple' || node.kind === 'object') {
			if (node.kind === 'array') return '[' + value.map(element => write(element, node.children[0], depth + 1)).join(',') + ']';
			if (node.kind === 'tuple') return '[' + node.fields.map((field, index) => write(value[index], field.node, depth + 1)).join(',') + ']';
			const fields = [];
			for (const field of node.fields) {
				const name = units(field.nameUnits, field.name);
				if (field.optional && (!Object.hasOwn(value, name) || value[name] === undefined)) continue;
				fields.push(JSON.stringify(name) + ':' + write(value[name], field.node, depth + 1));
			}
			return '{' + fields.join(',') + '}';
		}
		return JSON.stringify(value);
	}
	return write(value, schema.root, 0);
}
