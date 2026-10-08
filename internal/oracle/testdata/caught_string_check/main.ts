function accepts(message: string): void { console.log(`[${message}]`); }
try { throw 'x'; } catch (e) { accepts(e.message); }
console.log('after');
