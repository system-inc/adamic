module.exports = text => text.replace(/declare const console: \{[\s\S]*?\n\};/, block => block.replace(/\t/g, '    '));
