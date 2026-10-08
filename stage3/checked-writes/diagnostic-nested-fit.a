interface Metadata { readonly flags: number }
interface SourceFile { readonly name: string; readonly metadata: Metadata }
interface SynthesizedMetadata { readonly flags: 16 }
interface SynthesizedSourceFile { readonly name: string; readonly metadata: SynthesizedMetadata }
interface Diagnostic { file: SourceFile | undefined }
const original: SynthesizedSourceFile = { name: 'old'.repeat(2), metadata: { flags: 16 } };
const replacement: SynthesizedSourceFile = { name: 'new'.repeat(2), metadata: { flags: 16 } };
const throughWideView: SourceFile = replacement;
const diagnostic: { file: SynthesizedSourceFile } = { file: original };
function store(view: Diagnostic, file: SourceFile | undefined): void { view.file = file; }
store(diagnostic, throughWideView);
console.log(diagnostic.file.name + ' ' + diagnostic.file.metadata.flags.toString());
