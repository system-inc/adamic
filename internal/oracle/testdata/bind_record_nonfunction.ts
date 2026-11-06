interface Dictionary { [key: string]: (() => void) | string; m(): void }
interface Caller { call(host: Record<string, (() => void) | string>): void }
const caller: Caller = {
  call(host: Dictionary): void { host.m.bind(host)(); }
};
const dictionary: Record<string, (() => void) | string> = { m: "not callable" };
caller.call(dictionary);
