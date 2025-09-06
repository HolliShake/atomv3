


var r = 0;


function mutate() {
    console.log(r);
    (function() {
        r = 100;
    })()
}

mutate();
console.log(r);