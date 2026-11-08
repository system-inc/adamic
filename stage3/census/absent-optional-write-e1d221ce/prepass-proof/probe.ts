interface Mut { name?: string; }
interface Other { name?: string; }
const v: Mut = {};
const o: Other = v;
o.name = 'w';
