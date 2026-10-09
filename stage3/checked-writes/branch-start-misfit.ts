interface SourceFile { readonly name: string }
interface Diagnostic { file: SourceFile | undefined; start: number | undefined; length: number | undefined }
interface LocatedDiagnostic { file: SourceFile; start: number; length: number }
const file: SourceFile = { name: 'input'.repeat(2) };
const located: LocatedDiagnostic = { file, start: 1, length: 2 };
const plain: Diagnostic = { file: undefined, start: undefined, length: undefined };
function select(flag: boolean) { return flag ? located : plain; }
function store(view: Diagnostic): void { view.start = undefined; }
store(select(true));
console.log('stored');
