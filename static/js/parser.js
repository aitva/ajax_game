"use strict";

const Parser = function() {
    let that = {};
    
    that.parse = function(str) {
        let req = {};
        let lines = str.split("\n");
        
        let toks = lines[0].split(" ")
        if (toks.length !== 3) {
            throw 'line 1: expects 3 tokens';
        }
        if (toks[2] !== "HTTP/1.1") {
            throw 'line 1: HTTP version is missing';
        }
        let url = toks[1];
        req.method = toks[0]

        req.headers = {};
        for (let i of lines.slice(1)) {
            toks = lines[i].split(":");
            if (toks.length !== 2) {
                throw `line ${i+1}: expects ":"`
            }
            req.headers[toks[0]] = toks[1];
        }

        new Request(url, req);
    }

    return Object.freeze(that);
};