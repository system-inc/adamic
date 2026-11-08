setTimeout(() => { throw new Error('uncaught exit'); }, 0);
setTimeout(() => console.log('pending timer'), 0);
