// A union of arrays and other members cast to one side (#9z2ksv7, under Adamic's casts principle on #b5w3ycg):
// Array.isArray at the cast tells the sides apart. tsc's two first stops, tsbuildPublic.ts's AnyBuildOrder cast
// to CircularBuildOrder and watchUtilities.ts's names-or-program cast to readonly string[], with interfaces standing
// in for tsc's, each declaring required properties no array has. Every arm's value goes through its cast and is
// read after it, undefined too where the target keeps it, and the program prints what Node prints.
type BuildOrder = readonly string[];
interface CircularBuildOrder {
	readonly buildOrder: BuildOrder;
	readonly circularDiagnostics: readonly string[];
}
type AnyBuildOrder = BuildOrder | CircularBuildOrder;
interface Program {
	readonly kind: 'program';
	readonly rootNames: readonly string[];
}
interface BuilderProgram {
	readonly kind: 'builder';
	readonly program: Program;
}
type NamesOrProgram = readonly string[] | Program | BuilderProgram | undefined;

function circularOf(order: AnyBuildOrder): CircularBuildOrder {
	return order as CircularBuildOrder;
}
function buildOrderOf(order: AnyBuildOrder): BuildOrder {
	return order as BuildOrder;
}
function describe(order: AnyBuildOrder, circular: boolean): string {
	if (circular) {
		const cycle = order as CircularBuildOrder;
		return 'circular ' + cycle.buildOrder.join(',') + ' ' + String(cycle.circularDiagnostics.length);
	}
	return 'order ' + (order as BuildOrder).join(',');
}
function rootNamesOf(source: NamesOrProgram): readonly string[] {
	return source as readonly string[];
}
function programOf(source: NamesOrProgram): Program | BuilderProgram {
	return source as Program | BuilderProgram;
}
function maybeNames(source: NamesOrProgram): readonly string[] | undefined {
	return source as readonly string[] | undefined;
}
function maybeProgram(source: NamesOrProgram): Program | BuilderProgram | undefined {
	return source as Program | BuilderProgram | undefined;
}
function nameOf(program: Program | BuilderProgram): string {
	return program.kind === 'program' ? program.rootNames.join(',') : 'builder of ' + program.program.rootNames.join(',');
}

const circular: CircularBuildOrder = { buildOrder: ['a', 'b'], circularDiagnostics: ['cycle'] };
const order: BuildOrder = ['c', 'd', 'e'];
console.log(circularOf(circular).buildOrder.join(','));
console.log(String(circularOf(circular).circularDiagnostics.length));
console.log(buildOrderOf(order).join(','));
const empty: BuildOrder = [];
console.log(String(buildOrderOf(empty).length));
console.log(describe(circular, true));
console.log(describe(order, false));
const program: Program = { kind: 'program', rootNames: ['main.ts', 'other.ts'] };
const builder: BuilderProgram = { kind: 'builder', program };
const names: readonly string[] = ['x.ts', 'y.ts'];
console.log(rootNamesOf(names).join(','));
console.log(nameOf(programOf(program)));
console.log(nameOf(programOf(builder)));
const none = maybeNames(undefined);
console.log(none === undefined ? 'no names' : none.join(','));
const one: readonly string[] = ['z.ts'];
const some = maybeNames(one);
console.log(some === undefined ? 'no names' : some.join(','));
const absent = maybeProgram(undefined);
console.log(absent === undefined ? 'no program' : nameOf(absent));
const present = maybeProgram(builder);
console.log(present === undefined ? 'no program' : nameOf(present));
