// Encode UTF-16 code units directly; only an adjacent high/low pair is combined.
export function encodeWTF8(text) {
  const output = new Uint8Array(text.length * 3);
  let end = 0;
  for (let index = 0; index < text.length; index++) {
    const unit = text.charCodeAt(index);
    const next = text.charCodeAt(index + 1);
    if (unit >= 0xd800 && unit < 0xdc00 && next >= 0xdc00 && next < 0xe000) {
      const scalar = ((unit - 0xd800) << 10) + next - 0xdc00 + 0x10000;
      output[end++] = 0xf0 + (scalar >>> 18);
      output[end++] = 0x80 + ((scalar >>> 12) & 0x3f);
      output[end++] = 0x80 + ((scalar >>> 6) & 0x3f);
      output[end++] = 0x80 + (scalar & 0x3f);
      index++;
    } else if (unit <= 0x7f) {
      output[end++] = unit;
    } else if (unit <= 0x7ff) {
      output[end++] = 0xc0 + (unit >>> 6);
      output[end++] = 0x80 + (unit & 0x3f);
    } else {
      output[end++] = 0xe0 + (unit >>> 12);
      output[end++] = 0x80 + ((unit >>> 6) & 0x3f);
      output[end++] = 0x80 + (unit & 0x3f);
    }
  }
  return output.subarray(0, end);
}

export function decodeWTF8(bytes) {
  const units = new Uint16Array(bytes.length);
  let count = 0;
  let index = 0;
  function continuation() {
    const byte = bytes[index++];
    if (byte === undefined || byte < 0x80 || byte > 0xbf) throw new Error('Invalid WTF-8 continuation');
    return byte - 0x80;
  }
  while (index < bytes.length) {
    const first = bytes[index++];
    let scalar;
    if (first <= 0x7f) scalar = first;
    else if (first >= 0xc2 && first <= 0xdf) {
      scalar = (first - 0xc0) * 64 + continuation();
    } else if (first >= 0xe0 && first <= 0xef) {
      const middle = continuation();
      if (first === 0xe0 && middle < 32) throw new Error('Noncanonical WTF-8');
      scalar = (first - 0xe0) * 4096 + middle * 64 + continuation();
    } else if (first >= 0xf0 && first <= 0xf4) {
      const second = continuation();
      if ((first === 0xf0 && second < 16) || (first === 0xf4 && second > 15)) throw new Error('Invalid WTF-8 scalar');
      scalar = (first - 0xf0) * 262144 + second * 4096 + continuation() * 64 + continuation();
    } else throw new Error('Invalid WTF-8 lead');
    if (scalar <= 0xffff) units[count++] = scalar;
    else {
      const pair = scalar - 0x10000;
      units[count++] = 0xd800 + (pair >>> 10);
      units[count++] = 0xdc00 + (pair & 1023);
    }
  }
  const chunks = [];
  for (let start = 0; start < count; start += 8192) chunks.push(String.fromCharCode(...units.subarray(start, Math.min(start + 8192, count))));
  return chunks.join('');
}
