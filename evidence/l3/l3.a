interface Slot { run(callback: (...args: string[]) => void, ...rest: unknown[]): void; }
class Provider { run<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>): void { callback(...rest); } }
const slot: Slot = new Provider();
slot.run((...args: string[]) => console.log(args[0].toUpperCase()), 5);
