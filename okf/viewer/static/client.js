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

  function isDark() {
    var explicit = document.documentElement.getAttribute("data-theme");
    if (explicit) return explicit === "dark";
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  }

  var toggle = document.getElementById("theme-toggle");
  if (toggle) {
    toggle.addEventListener("click", function () {
      var next = isDark() ? "light" : "dark";
      document.documentElement.setAttribute("data-theme", next);
      try { localStorage.setItem("okf-theme", next); } catch (e) {}
    });
  }

  if (window.mermaid) {
    mermaid.initialize({ startOnLoad: false, theme: isDark() ? "dark" : "default" });
    mermaid.run();
  }

  var MIN_SIDEBAR_WIDTH = 160;
  var MAX_SIDEBAR_WIDTH = 600;

  var resizer = document.getElementById("sidebar-resizer");
  var sidebar = document.getElementById("sidebar");
  if (resizer && sidebar) {
    resizer.addEventListener("mousedown", function (e) {
      e.preventDefault();
      resizer.classList.add("dragging");
      document.body.classList.add("resizing");

      function onMove(e) {
        var width = e.clientX - sidebar.getBoundingClientRect().left;
        width = Math.max(MIN_SIDEBAR_WIDTH, Math.min(MAX_SIDEBAR_WIDTH, width));
        document.documentElement.style.setProperty("--sidebar-width", width + "px");
      }

      function onUp() {
        document.removeEventListener("mousemove", onMove);
        document.removeEventListener("mouseup", onUp);
        resizer.classList.remove("dragging");
        document.body.classList.remove("resizing");
        var width = sidebar.getBoundingClientRect().width;
        try { localStorage.setItem("okf-sidebar-width", Math.round(width)); } catch (e) {}
      }

      document.addEventListener("mousemove", onMove);
      document.addEventListener("mouseup", onUp);
    });
  }
})();
