interface Narrow {
    [key: string]: string | undefined;
}
interface Wide {
    [key: string]: string | number | undefined;
}
function readOption(options: Readonly<Wide>, key: string) {
    return options[key];
}
export function probe(options: Narrow, key: string) {
    return readOption(options, key);
}
