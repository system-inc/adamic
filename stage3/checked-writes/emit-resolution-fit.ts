interface ResolvedModuleFull { fileName: string }
interface ResolutionWithResolvedFileName { fileName: string | undefined }
const original: ResolvedModuleFull = { fileName: 'inputinput' };
function resolve(): ResolvedModuleFull { return original; }
function invoke(callback: () => ResolutionWithResolvedFileName): ResolutionWithResolvedFileName { return callback(); }
const view: ResolutionWithResolvedFileName = invoke(resolve);
view.fileName = 'outputoutput';
console.log((view.fileName === undefined).toString());
