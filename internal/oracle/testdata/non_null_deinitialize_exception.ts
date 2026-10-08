function f(): void {
 let value = 4;
 const clear = (): void => { value = undefined!; throw new Error('cleared'); };
 try { clear(); } catch { console.log('caught'); }
 console.log(`${value}`);
}
f();
