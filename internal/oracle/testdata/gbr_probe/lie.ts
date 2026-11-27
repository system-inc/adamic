// Probe lie for #gbr6vs2 (never merged): a false signature cast as FileInfo with no test, then read.
interface FileInfo { readonly version: string; readonly signature: string | undefined; readonly affectsGlobalScope: true | undefined }
type MultiInfo = Omit<FileInfo, "signature"> & { readonly signature: string | false | undefined };
function direct(info: MultiInfo): FileInfo { return info as FileInfo; }
const b: MultiInfo = { version: "v2", signature: false, affectsGlobalScope: true };
console.log("casting");
console.log(String(direct(b).signature));
