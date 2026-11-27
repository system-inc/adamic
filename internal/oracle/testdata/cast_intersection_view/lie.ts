// The lie: a false signature cast as FileInfo with no test (#gbr6vs2), then read. It stops at the read, in the view's
// words. Node runs on and prints false.
interface FileInfo { readonly version: string; readonly signature: string | undefined; readonly affectsGlobalScope: true | undefined }
type MultiInfo = Omit<FileInfo, "signature"> & { readonly signature: string | false | undefined };
function direct(info: MultiInfo): FileInfo { return info as FileInfo; }
const b: MultiInfo = { version: "v2", signature: false, affectsGlobalScope: true };
console.log("casting");
console.log(String(direct(b).signature));
