try { String.prototype.valueOf.call(42); console.log('admitted'); } catch (error) { console.log(error.name); }
