setTimeout(() => console.log('pending timer'), 0);
Promise.resolve().then(() => { throw new Error('uncaught exit'); });
