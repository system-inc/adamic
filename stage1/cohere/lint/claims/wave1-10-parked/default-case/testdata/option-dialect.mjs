// External Node oracle for the required option constructor, not Adamic source.
for (const [pattern, text] of [['\\s', '\u00a0'], ['todo', 'TODO'], ['(?i)todo', 'TODO']]) {
    try { console.log(new RegExp(pattern, 'u').test(text)); }
    catch (error) { console.log(error.name); }
}
