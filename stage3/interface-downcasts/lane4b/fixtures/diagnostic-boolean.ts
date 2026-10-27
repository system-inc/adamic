interface Chain { readonly messageText: string; }
interface Base { readonly kind: string; }
interface Box extends Base { readonly kind: 'Box'; readonly messageText: string | Chain; }
function read(value: Box): string {
 const member = value.messageText;
 return typeof member === 'string' ? member : member.messageText;
}
const concrete0: { readonly kind: string; readonly messageText: string | boolean | Chain } = { kind: 'Box', messageText: true };
const source0: Base = concrete0;
console.log(read(source0 as Box));
