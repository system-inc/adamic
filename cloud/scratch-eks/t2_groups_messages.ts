// V8's TypeError text for each Object.prototype method called on a RegExp match's groups, by callee form.
class Holder {
	groups: { [key: string]: string };
	constructor(groups: { [key: string]: string }) {
		this.groups = groups;
	}
	show(): string {
		return this.groups.toString();
	}
}
function attempt(f: () => unknown): string {
	try {
		return 'answered ' + String(f());
	} catch (error) {
		return error instanceof Error ? error.name + ': ' + error.message : 'thrown';
	}
}
const match = '2024'.match(/(?<year>[0-9]+)/);
if (match !== null && match.groups !== undefined) {
	const groups = match.groups;
	const box = { inner: { groups } };
	const list = [groups];
	console.log(attempt(() => match.groups!.toString()));
	console.log(attempt(() => match.groups!.toLocaleString()));
	console.log(attempt(() => match.groups!.valueOf()));
	console.log(attempt(() => match.groups!.hasOwnProperty('year')));
	console.log(attempt(() => match.groups!.propertyIsEnumerable('year')));
	console.log(attempt(() => match.groups!.isPrototypeOf(groups)));
	console.log(attempt(() => groups.toString()));
	console.log(attempt(() => (groups).toString()));
	console.log(attempt(() => (groups as object).toString()));
	console.log(attempt(() => groups['toString']()));
	console.log(attempt(() => groups?.toString()));
	console.log(attempt(() => box.inner.groups.toString()));
	console.log(attempt(() => list[0]!.toString()));
	console.log(attempt(() => new Holder(groups).show()));
	console.log(attempt(() => (match.groups)!.valueOf()));
	console.log(attempt(() => '2024'.match(/(?<year>[0-9]+)/)!.groups!.toString()));
	console.log(attempt(() => Object.prototype.toString.call(groups)));
	console.log(attempt(() => Object.prototype.hasOwnProperty.call(groups, 'year')));
	console.log(attempt(() => Object.prototype.valueOf.call(groups) === groups));
	console.log(attempt(() => Object.prototype.toLocaleString.call(groups)));
	console.log(attempt(() => `${groups}`));
	console.log(attempt(() => String(groups)));
	console.log(attempt(() => Object.getPrototypeOf(groups)));
}
const exec = /(?<year>[0-9]+)/.exec('2024');
if (exec !== null && exec.groups !== undefined) {
	console.log(attempt(() => exec.groups!.toString()));
}
for (const found of '2024'.matchAll(/(?<year>[0-9]+)/g)) {
	console.log(attempt(() => found.groups!.toString()));
}
