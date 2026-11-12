let leak = () => {};
try {
 leak = () => { console.log(String(session)); };
 throw new Error('initializer failed');
 const session = 1;
} catch { console.log('require threw'); }
try { leak(); } catch (error) {
 if (error instanceof Error) console.log('caught ' + error.name);
}
