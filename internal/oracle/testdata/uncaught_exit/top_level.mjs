setTimeout(() => console.log('pending timer'), 0);
throw new Error('uncaught exit');
