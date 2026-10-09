type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface SourceFile { readonly name: string; readonly statements: readonly number[]; readonly line: () => number }
interface SynthesizedSourceFile extends SourceFile { readonly statements: readonly 16[]; readonly line: () => 16 }
interface Diagnostic { readonly file: SourceFile }
function store<T extends Diagnostic>(view: T, file: SourceFile): void { (view as Mutable<T>).file = file; }
const original: SynthesizedSourceFile = { name: 'old'.repeat(2), statements: [16], line: () => 16 };
const replacement: SourceFile = { name: 'new'.repeat(2), statements: [0], line: () => 0 };
const diagnostic: { readonly file: SynthesizedSourceFile } = { file: original };
store(diagnostic, replacement);
console.log(diagnostic.file.line().toString());
