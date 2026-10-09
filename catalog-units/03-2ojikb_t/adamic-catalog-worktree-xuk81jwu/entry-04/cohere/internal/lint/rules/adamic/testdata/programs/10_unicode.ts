const greeting = 'héllo, 世界 🌍';

console.log(`${greeting.length} UTF-16 code units, ${[...greeting].length} code points`);
console.log(greeting.slice(7, 9));
console.log(`${greeting.codePointAt(10) ?? -1} ${greeting.charCodeAt(11)}`);
console.log(`[${greeting.slice(0, 11)}]`);
console.log(`${'b' < 'é'} ${'\uFFFF' < '🌍'}`);
console.log(`${'ab'.padStart(5, '-')}|${'  trim me \n'.trim()}|${'na'.repeat(3)}`);
console.log(`${greeting.indexOf('世')} ${greeting.includes('🌍')} ${greeting.startsWith('hé')}`);
