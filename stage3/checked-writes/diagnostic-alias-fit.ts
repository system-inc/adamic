type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface SourceFile { flags: number }
interface Diagnostic { readonly file: SourceFile }
function widen<T extends SourceFile>(file: T): SourceFile { return file; }
function store<T extends Diagnostic>(view: T, file: SourceFile): void { (view as Mutable<T>).file = file; }
function change(wide: SourceFile, value: number): void { wide.flags = value; }
const file: { readonly flags: 16 } = { flags: 16 };
const diagnostic: { readonly file: { readonly flags: 16 } } = { file };
const wide = widen(file);
store(diagnostic, wide);
change(wide, 16);
console.log(diagnostic.file.flags.toString());
