interface Slot { run: (callback: (...args: string[]) => void, ...rest: unknown[]) => void; }
function provide<T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>): void { callback(...rest); }
const slot: Slot = { run: provide };
slot.run((...args: string[]) => { for (const a of args) { console.log(a.toUpperCase()); } }, 5);
