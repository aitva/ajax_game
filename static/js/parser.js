"use strict";

const Parser = function() {
    let that = {};
    
    that.parse = function(str) {
        let toks = str.split(" ")
        return new Request(toks[1], {
            method: toks[0]
        });
    }

    return Object.freeze(that);
};