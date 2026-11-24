interface Info { readonly version: string; signature: string | undefined }
interface Encoded { readonly version: string; signature: string | false | undefined }
function decode(x: Encoded): Info {
    return typeof x.signature === "string" ? x as Info : { version: x.version, signature: undefined };
}
console.log(decode({ version: "v", signature: "s" }).signature);
console.log(String(decode({ version: "v", signature: false }).signature));
