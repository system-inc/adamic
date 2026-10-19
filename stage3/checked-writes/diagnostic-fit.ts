interface SourceFile { readonly name: string }
interface Diagnostic { file: SourceFile | undefined }
const file: SourceFile = { name: 'input'.repeat(2) };
const diagnostic: { file: SourceFile } = { file };
function store(view: Diagnostic, value: SourceFile | undefined): void { view.file = value; }
store(diagnostic, file);
console.log(diagnostic.file.name);
