function number(value: number | undefined): number { return value!; }
function boolean(value: boolean | undefined): boolean { return value!; }
function text(value: string | undefined): string { return value!; }
function mixed(value: number | string | boolean | undefined): number | string | boolean { return value!; }
function reference(value: { readonly name: string } | undefined): string { return value!.name; }
let calls = 0;
function next(): number | undefined { calls += 1; return calls; }
console.log(`${number(0)} ${number(NaN)} ${boolean(false)} '${text('')}'`);
console.log(`${mixed(0)} ${mixed(false)} '${mixed('')}' ${mixed(NaN)}`);
console.log(reference({ name: `item${calls}` }));
console.log(`${next()!} ${calls}`);
console.log(`${'plain'!} ${42!} ${false!}`);
console.log('a'.match(/a/)![0]!);
function narrowed(value: number | string | undefined): number {
	if (typeof value === 'number') { return value!; }
	return 1;
}
console.log(`${narrowed(0)}`);
