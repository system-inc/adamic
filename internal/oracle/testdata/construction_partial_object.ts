// A literal cast to a target it only partly writes (N4 head B on #b5w3ycg, generic over any construction):
// commandLineParser.ts:2291's { fileName, parseDiagnostics: [textOrDiagnostic] } as TsConfigSourceFile, with
// interfaces standing in for tsc's. The literal writes a string, an array and origin, which Big holds optional, so
// Big doesn't extend the literal's type and the cast is sideways. Big's number and parent are required and
// unwritten, and its other optional fields are absent. One path builds the whole Big. The program reads what the
// literal wrote and the optional fields it didn't, which are undefined, and prints what Node prints.
interface Owner {
	name: string;
}
interface Item {
	label: string;
}
interface Big {
	name: string;
	items: Item[];
	count: number;
	parent: Owner;
	origin?: Owner;
	note?: string;
	size?: number;
	extra?: Item;
}

function make(name: string, x: Item, origin: Owner, whole: boolean): Big {
	if (whole) {
		return { name, items: [x, x], count: 2, parent: origin, note: 'whole' };
	}
	return { name, items: [x], origin } as Big;
}

const boss: Owner = { name: 'boss' };
const partial = make('partial', { label: 'first' }, boss, false);
console.log(partial.name);
console.log(String(partial.items.length));
console.log(partial.items[0].label);
console.log(partial.origin === undefined ? 'no origin' : partial.origin.name);
console.log(String(partial.note));
console.log(partial.size === undefined ? 'no size' : String(partial.size));
console.log(partial.extra === undefined ? 'no extra' : partial.extra.label);
const whole = make('whole', { label: 'second' }, boss, true);
console.log(whole.name + ' ' + String(whole.count) + ' ' + whole.parent.name + ' ' + String(whole.note));
