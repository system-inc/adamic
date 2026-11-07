function replace(host: Record<string, (() => void) | string>): void {
  host["m"] = "not callable";
}
const dictionary: Record<string, (() => void) | string> = { m: (): void => {} };
if (typeof dictionary.m === "function") {
  replace(dictionary);
  dictionary.m.bind(dictionary)();
}
