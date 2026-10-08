interface Metadata { readonly flags: 16 | 32 }
interface SourceFile { readonly metadata: Metadata }
const file: SourceFile = { metadata: { flags: 16 } };
const diagnostic: { readonly file: SourceFile } = { file };
function store(view: { readonly file: { readonly metadata: { flags: number } } }, flags: number): void { view.file.metadata.flags = flags; }
store(diagnostic, 32);
console.log(diagnostic.file.metadata.flags.toString());
