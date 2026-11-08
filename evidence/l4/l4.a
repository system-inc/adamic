interface Slot { run(callback: (...args: string[]) => void, ...rest: unknown[]): void; }
class Provider { run<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>): void { callback(...rest); } }
const slot: Slot = new Provider();
slot.run((...args: string[]) => { for (const a of args) { console.log(a.toUpperCase()); } }, 5);
