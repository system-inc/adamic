// Keep the published plan runnable by apply.sh while its semantic edits remain deferred.
// This adapter deliberately makes no source change. It is not a completed repair.
console.log(JSON.stringify({files: 0, status: 'deferred', reason: '13 scanner fallthrough continuations require individual control-flow review; no semantic rewrite validated'}));
