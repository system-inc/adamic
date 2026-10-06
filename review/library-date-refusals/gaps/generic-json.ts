console.log(`${Date.prototype.toJSON.call({toISOString: (): string => "custom"})}`);
