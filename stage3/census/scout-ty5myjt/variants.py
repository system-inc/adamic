import pathlib,shutil,re
root=pathlib.Path('/workspace/scratch/scout-ty5myjt')
base=root/'adapted'
for name in ['diag-intersection','diag-readonly','diag-exact','diag-overloads','resolution-readonly','resolution-exact','resolution-result-view']:
 tree=root/name
 shutil.copytree(base/'src',tree/'src')
 file='program.ts' if name.startswith('diag') else 'resolutionCache.ts'
 p=tree/'src/compiler'/file
 text=p.read_bytes().decode()
 if name=='diag-intersection':text=text.replace('getDiagnosticsHelper<T extends Diagnostic>','getDiagnosticsHelper<T extends Readonly<Pick<Diagnostic, "file">> & Diagnostic>')
 if name=='diag-readonly':
  text=text.replace('    function getDiagnosticsHelper<T extends Diagnostic>', '    interface DiagnosticReadView extends Diagnostic { readonly file: SourceFile | undefined; }\r\n    function getDiagnosticsHelper<T extends DiagnosticReadView>')
 if name in ['diag-exact','diag-overloads']:
  a=text.index('    function getDiagnosticsHelper<T extends Diagnostic>');b=text.index('\r\n    function getSyntacticDiagnostics',a)
  body=text[a:b]
  exact=body.replace('<T extends Diagnostic>','').replace('readonly T[]','readonly DiagnosticWithLocation[]')
  if name=='diag-exact':text=text[:a]+exact+text[b:]
  else:
   signature=body[:body.index(' {')]
   overloads=''
   for typ in ['DiagnosticWithLocation','Diagnostic']:
    overloads+=signature.replace('<T extends Diagnostic>','').replace('readonly T[]','readonly '+typ+'[]')+';\r\n'
   implementation=body.replace('<T extends Diagnostic>','').replace('readonly T[]','readonly Diagnostic[]')
   text=text[:a]+overloads+implementation+text[b:]
 if name=='resolution-readonly':
  text='interface ResolutionReadView extends ResolutionWithResolvedFileName { readonly resolvedFileName: string | undefined; }\r\n'+text
  text=text.replace('R extends ResolutionWithResolvedFileName','R extends ResolutionReadView').replace('= ResolutionWithResolvedFileName>','= ResolutionReadView>')
 if name=='resolution-exact':
  text=text.replace('    type ResolvedModuleWithFailedLookupLocations,','    type ResolvedModuleFull,\r\n    type ResolvedModuleWithFailedLookupLocations,')
  text=text.replace(', R extends ResolutionWithResolvedFileName','')
  text=text.replace('<T, R>','<T, ResolvedModuleFull>').replace('<Entry, SourceFile, T, R>', '<Entry, SourceFile, T>')
  # GetResolutionWithResolvedFileName itself retains R; only functions lose result binder.
  original=re.search(r'export type GetResolutionWithResolvedFileName[^\r]+',p.read_text()).group()
  start=text.index('export type GetResolutionWithResolvedFileName'); end=text.index('\r\n',start)
  text=text[:start]+original+text[end:]
 if name=='resolution-result-view':
  text='interface ResolutionReadView extends ResolutionWithResolvedFileName { readonly resolvedFileName: string | undefined; }\r\n'+text
  text=text.replace(', R extends ResolutionWithResolvedFileName','')
  text=text.replace('GetResolutionWithResolvedFileName<T, R>', '(resolution: T) => ResolutionReadView | undefined').replace('<Entry, SourceFile, T, R>', '<Entry, SourceFile, T>')
  original=re.search(r'export type GetResolutionWithResolvedFileName[^\r]+',p.read_text()).group()
  start=text.index('export type GetResolutionWithResolvedFileName');end=text.index('\r\n',start)
  text=text[:start]+original+text[end:]
 p.write_bytes(text.encode())
