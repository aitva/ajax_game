"use strict";

const Parser = function() {
    let that = {};
    let allSpace = /^\s+$/;
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
        req.method = toks[0];
        
        req.headers = {};
        let i = 1;
        let nbreak = 0;
        for (let header of lines.slice(1)) {
            header = header.trim();
            if (header === "") {
                i++;
                nbreak++
                if (nbreak === 1) {
                    break;
                }
                continue;
            }
            toks = header.split(":");
            if (toks.length !== 2) {
                throw `line ${i}: colon are mandatory in header`;
            }
            req.headers[toks[0]] = toks[1].trim();
            i++;
        }
        
        console.log(req);
        return new Request(url, req);
    }

    return Object.freeze(that);
};