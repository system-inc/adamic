import ts = require("./typescript");
declare const file: ts.SourceFile;
ts.getLineAndCharacterOfPosition(file, 0);
