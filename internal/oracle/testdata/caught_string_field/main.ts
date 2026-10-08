try { throw 7; } catch (e) { const record: { message: string } = { message: e.message }; console.log(record.message); }
console.log('after');
