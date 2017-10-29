"use strict";

// Notif creates a new notification on the page.
// The type can be any of: success, info, warning, danger.
const Notif = function(type, msg) {
    let that = {
        parent: null,
        node: null,
        type: type,
        msg: msg,
    };
    // render renders the alert as a DOM object and attaches it to the given node.
    that.render = function(node) {
        
        let div = document.createElement("div");
        div.classList.add("alert", "alert-"+type);
        div.setAttribute("role", "alert");
        if (typeof msg === 'object') {
            let str = "";
            for (let key of Object.keys(msg)) {
                str += `<strong>${key}:</strong> ${msg[key]}</br>`;
            }
            div.innerHTML = str;
        } else {
            div.textContent = msg;
        }
        
        let button = document.createElement("button");
        button.type = "button";
        button.classList.add("close");
        button.dataset.dismiss = "alert";
        button.innerHTML = "<span>&times;</span>"
        button.addEventListener("click", that.close);
        div.insertBefore(button, div.firstElementChild);
        
        that.node = div;
        that.parent = node;
        node.appendChild(div);
    }
    // close closes the notification.
    that.close = function() {
        that.parent.removeChild(that.node);
    }

    return that;
}