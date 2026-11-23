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
function isString(text: unknown): text is string {
    return typeof text === "string";
}
function decode(fileInfo: string | EncodedFileInfo): FileInfo {
    return isString(fileInfo) ?
        { version: fileInfo, signature: fileInfo, affectsGlobalScope: undefined } :
        isString(fileInfo.signature) ?
        { version: fileInfo.version, signature: fileInfo.signature, affectsGlobalScope: fileInfo.affectsGlobalScope } :
        { version: fileInfo.version, signature: fileInfo.signature === false ? undefined : fileInfo.version, affectsGlobalScope: fileInfo.affectsGlobalScope };
}
function show(fileInfo: string | EncodedFileInfo): void {
    const decoded = decode(fileInfo);
    console.log(decoded.version + " " + String(decoded.signature) + " " + String(decoded.affectsGlobalScope));
}
show("v1");
show({ version: "v2", signature: "s2", affectsGlobalScope: undefined });
show({ version: "v3", signature: false, affectsGlobalScope: true });
show({ version: "v4", signature: undefined, affectsGlobalScope: undefined });
console.log(String(table["count"]));
