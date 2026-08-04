(function () {
  "use strict";

  // Keep one websocket open for as long as this page is. The server uses
  // connect/disconnect events to decide when nobody's looking anymore and
  // shut itself down; a brief reconnect loop absorbs transient drops
  // without the page (and thus the server) treating them as "closed".
  var unloading = false;
  var retry = 0;

  function connect() {
    var proto = location.protocol === "https:" ? "wss:" : "ws:";
    var ws = new WebSocket(proto + "//" + location.host + "/ws");
    ws.onopen = function () { retry = 0; };
    ws.onclose = function () {
      if (unloading) return;
      var delay = Math.min(1000 * Math.pow(2, retry), 5000);
      retry++;
      setTimeout(connect, delay);
    };
  }

  window.addEventListener("beforeunload", function () { unloading = true; });
  connect();

  if (window.mermaid) {
    var dark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
    mermaid.initialize({ startOnLoad: false, theme: dark ? "dark" : "default" });
    mermaid.run();
  }
})();
