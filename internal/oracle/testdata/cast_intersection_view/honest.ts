// #gbr6vs2: tsc's toBuilderStateFileInfoForMultiEmit shape, an intersection source whose signature is
// string | false | undefined, cast to FileInfo (a structural view) after a string test, and an unguarded cast of an
// undefined signature, each value built where the source type is in force. Then a string and undefined written
// through the cast, each read back through both types.
interface FileInfo { readonly version: string; signature: string | undefined; readonly affectsGlobalScope: true | undefined }
type MultiInfo = Omit<FileInfo, "signature"> & { readonly signature: string | false | undefined };
function toState(info: MultiInfo): FileInfo {
  return typeof info.signature === "string" ? info as FileInfo : { version: info.version, signature: info.signature === false ? undefined : info.version, affectsGlobalScope: info.affectsGlobalScope };
}
function direct(info: MultiInfo): FileInfo { return info as FileInfo; }
const a: MultiInfo = { version: "v1", signature: "sig-" + "a", affectsGlobalScope: undefined };
const b: MultiInfo = { version: "v2", signature: false, affectsGlobalScope: true };
const c: MultiInfo = { version: "v3", signature: undefined, affectsGlobalScope: undefined };
for (const x of [a, b, c]) {
  const s = toState(x);
  console.log(s.version + " " + String(s.signature) + " " + String(s.affectsGlobalScope));
}
const d = direct(c);
console.log(d.version + " " + String(d.signature));
const e: MultiInfo = { version: "v5", signature: "s" + "0", affectsGlobalScope: undefined };
const w = direct(e);
w.signature = "written-" + "x";
console.log(String(w.signature) + " " + String(e.signature));
w.signature = undefined;
console.log(String(w.signature) + " " + String(e.signature));
