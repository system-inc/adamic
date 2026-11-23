const table: Record<string, number> = {};
table["count"] = 1;

interface FileInfo {
    readonly version?: string;
    readonly signature?: string | false;
}
function decode(fileInfo: string | FileInfo): string {
    return typeof fileInfo === "string" ? "plain " + fileInfo :
        typeof fileInfo.signature === "string" ? "signed " + fileInfo.signature :
        fileInfo.signature === false ? "unsigned " + String(fileInfo.version) : "version " + String(fileInfo.version);
}
console.log(decode("v1"));
console.log(decode({ version: "v2", signature: "s2" }));
console.log(decode({ version: "v3", signature: false }));
console.log(decode({ version: "v4" }));
console.log(String(table["count"]));
