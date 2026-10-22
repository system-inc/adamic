interface CommentRange { pos: number }
interface SynthesizedComment { pos: -1 }
const original: SynthesizedComment = { pos: -1 };
function invoke(callback: (value: SynthesizedComment) => boolean, value: SynthesizedComment): boolean { return callback(value); }
function visit(view: CommentRange): boolean { view.pos = 0; return view.pos === -1; }
console.log(invoke(visit, original).toString());
