import fast from './base.mjs';
const retained = [];
export default {
  fetch(request) {
    if (new URL(request.url).pathname !== '/health') {
      const bytes = new Uint8Array(1024 * 1024);
      bytes.fill(17);
      retained.push(bytes);
      if (retained.length > 24) retained.shift();
    }
    return fast.fetch(request);
  },
};
