import fast, { burn } from './base.mjs';
export default {
  fetch(request) {
    if (new URL(request.url).pathname !== '/health') burn(4000000);
    return fast.fetch(request);
  },
};
