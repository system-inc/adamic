// probe.mjs: oracle/node.mjs with Adamic's inserted checks asked of the source itself, on Node's own
// state, for review r2: an element write rewritten to a probe that panics, as adamicSetIndex does,
// where the index isn't one the array has; map, find and findIndex panicking where the array shrank
// under them. A check that fires here fired for a reason Node can see.
import { readFileSync } from 'node:fs';
import { registerHooks, stripTypeScriptTypes } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';

const runtimeUrl = new URL('./adamic.mjs', import.meta.url).href;
const { panic } = await import(runtimeUrl);

globalThis.__probeSet = (array, index, value) => {
	if (!(Number.isInteger(index) && index >= 0 && index < array.length)) panic(`index ${index} is outside an array of length ${array.length}`);
	array[index] = value;
};
const shrinking = (name, searching) => function (callback, thisArgument) {
	const count = this.length;
	const mapped = [];
	for (let index = 0; index < count; index++) {
		if (index >= this.length) panic(searching ? `${name}: the array shrank while it was being searched` : 'map: the array shrank while it was being mapped');
		const element = this[index];
		const answer = callback.call(thisArgument, element, index, this);
		if (name === 'map') mapped.push(answer);
		else if (answer) return name === 'find' ? element : index;
	}
	return name === 'map' ? mapped : name === 'find' ? undefined : -1;
};
Array.prototype.map = shrinking('map', false);
Array.prototype.find = shrinking('find', true);
Array.prototype.findIndex = shrinking('findIndex', true);

// rewrite turns each statement `target[index] = value;` on a line of its own into __probeSet(target,
// index, value), the target a dotted name, the brackets matched.
function rewrite(source) {
	return source.split('\n').map((line) => {
		const head = /^(\s*)([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)*)\[/.exec(line);
		if (head === null || !line.trimEnd().endsWith(';')) return line;
		let depth = 0;
		let close = -1;
		for (let at = head[0].length - 1; at < line.length; at++) {
			if (line[at] === '[') depth++;
			else if (line[at] === ']' && --depth === 0) { close = at; break; }
		}
		if (close < 0 || !line.startsWith(' = ', close + 1) || line.startsWith(' = =', close + 1)) return line;
		const index = line.slice(head[0].length, close);
		const value = line.trimEnd().slice(close + 4, -1);
		return `${head[1]}__probeSet(${head[2]}, ${index}, ${value});`;
	}).join('\n');
}

registerHooks({
	resolve(specifier, context, nextResolve) {
		if (specifier === 'adamic') return { url: runtimeUrl, shortCircuit: true };
		return nextResolve(specifier, context);
	},
	load(url, context, nextLoad) {
		if (url.endsWith('.a') || url.endsWith('.ts')) {
			const source = readFileSync(fileURLToPath(url), 'utf8');
			return { format: 'module', source: rewrite(stripTypeScriptTypes(source)), shortCircuit: true };
		}
		return nextLoad(url, context);
	},
});

const program = process.argv[2];
process.argv.splice(1, 1);
await import(pathToFileURL(program).href);
