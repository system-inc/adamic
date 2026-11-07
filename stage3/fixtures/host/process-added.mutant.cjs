module.exports = text => text + '\ndeclare const process: {\n\texit(code: number): never;\n};\n';
