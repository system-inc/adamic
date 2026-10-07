const capital = /^\p{Lu}/u;
for(let point = 0; point <= 0x10ffff; point++) {
    if(capital.test(String.fromCodePoint(point))) { console.log(point); }
}
