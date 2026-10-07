const flags = new Map<string, boolean>();
flags.set('important', true);
console.log(flags.get('important') ? 'important' : 'ordinary');
