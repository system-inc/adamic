const ts=require('typescript');
const file='/view-proof.ts',text='interface D { file: string | undefined }; function intersection<T extends Readonly<Pick<D, "file">> & D>(v:T) { v.file = undefined; } function replaced<T extends Omit<D,"file"> & Readonly<Pick<D,"file">>>(v:T) { v.file = undefined; }';
const options={strict:true,noEmit:true};const host=ts.createCompilerHost(options),read=host.readFile.bind(host),exists=host.fileExists.bind(host);host.readFile=f=>f===file?text:read(f);host.fileExists=f=>f===file||exists(f);const program=ts.createProgram([file],options,host);const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,start:d.start,text:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
console.log(JSON.stringify({version:ts.version,source:text,diagnostics},null,2));
if(diagnostics.length!==1||diagnostics[0].code!==2540||diagnostics[0].start!==214)throw Error('readonly control changed');
