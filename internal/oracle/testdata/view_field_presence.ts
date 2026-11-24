// A presence-only read of a field through a checked view answers what Node answers (Adamic's ruling on #y5n0x0j):
// !!, a whole condition, !, and === or !== undefined, on a value of the cast's member, of another plain member, of
// a member holding the field as another type or as undefined, and of a Date. tsbuildPublic.ts:1371's guard first.
type FileWatcherCallback = (fileName: string, eventKind: number, modifiedTime?: Date) => void;
interface FileWatcher { close(): void }
interface FileWatcherWithModifiedTime { callbacks: FileWatcherCallback[]; watcher: FileWatcher; modifiedTime?: Date }
function isFileWatcherWithModifiedTime(value: FileWatcherWithModifiedTime | Date): boolean {
	return !!(value as FileWatcherWithModifiedTime).watcher;
}
const watched = new Map<string, FileWatcherWithModifiedTime | Date>();
watched.set("a", { callbacks: [], watcher: { close: () => {} } });
watched.set("b", new Date(5));
for (const key of ["a", "b"]) {
	const value = watched.get(key);
	if (value !== undefined) console.log(key + " " + String(isFileWatcherWithModifiedTime(value)));
}

interface Entry { readonly kind: number; readonly watcher: string }
interface Counted { readonly kind: number; readonly watcher: number }
interface Unset { readonly kind: number; readonly watcher: string | undefined }
interface Other { readonly kind: number; readonly label: string }
type Held = Entry | Counted | Unset | Other | Date;
function forms(value: Held): string {
	const truthy = !!(value as Entry).watcher;
	const negated = !(value as Entry).watcher;
	const tested = (value as Entry).watcher ? "yes" : "no";
	let looped = 0;
	while ((value as Entry).watcher && looped < 1) looped++;
	const absent = (value as Entry).watcher === undefined;
	const present = undefined !== (value as Entry).watcher;
	const both = looped >= 0 && (value as Entry).watcher ? "and" : "not";
	return [String(truthy), String(negated), tested, String(looped), String(absent), String(present), both].join(" ");
}
const held: Held[] = [
	{ kind: 1, watcher: "w" },
	{ kind: 2, watcher: "" },
	{ kind: 3, watcher: 0 },
	{ kind: 4, watcher: 7 },
	{ kind: 5, watcher: undefined },
	{ kind: 6, label: "l" },
	new Date(9),
];
for (const value of held) console.log(forms(value));

// A plain member declaring the name native's Date once kept its time under: a Date has no such field.
interface Timed { readonly _fsFileTime: number; readonly at: string }
function timed(value: Timed | Date): string {
	return String(!!(value as Timed)._fsFileTime) + " " + String((value as Timed)._fsFileTime !== undefined);
}
console.log(timed({ _fsFileTime: 3, at: "t" }) + " | " + timed(new Date(5)));
