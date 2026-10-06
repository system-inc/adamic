class Callback {
    call<Result>(callback: () => Result): Result { return callback(); }
}
console.log(new Callback().call(() => 'value'));
