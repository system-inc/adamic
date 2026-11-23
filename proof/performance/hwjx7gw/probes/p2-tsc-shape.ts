const table: Record<string, number> = {};
table["count"] = 1;

interface FileInfo {
    readonly version: string;
    signature: string | undefined;
    affectsGlobalScope: true | undefined;
}
type EncodedFileInfo = Omit<FileInfo, "signature"> & {
    signature: string | false | undefined;
};
type EncodedFileInfoOrString = string | EncodedFileInfo;
function isString(text: unknown): text is string {
    return typeof text === "string";
}
function decode(fileInfo: EncodedFileInfoOrString): FileInfo {
    return isString(fileInfo) ?
        { version: fileInfo, signature: fileInfo, affectsGlobalScope: undefined } :
        isString(fileInfo.signature) ?
        fileInfo as FileInfo :
        { version: fileInfo.version, signature: fileInfo.signature === false ? undefined : fileInfo.version, affectsGlobalScope: fileInfo.affectsGlobalScope };
}
function show(fileInfo: EncodedFileInfoOrString): void {
    const decoded = decode(fileInfo);
    console.log(decoded.version + " " + String(decoded.signature) + " " + String(decoded.affectsGlobalScope));
}
show("v1");
show({ version: "v2", signature: "s2", affectsGlobalScope: undefined });
show({ version: "v3", signature: false, affectsGlobalScope: true });
show({ version: "v4", signature: undefined, affectsGlobalScope: undefined });
console.log(String(table["count"]));
