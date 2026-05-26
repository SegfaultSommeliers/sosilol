import "materialize-css/dist/js/materialize.min.js";
import hljs from "highlight.js";
import Alpine from "alpinejs";
import "htmx.org";

window.hljs = hljs;

// Syntax highlighting — run after DOM is ready
hljs.highlightAll();

// Alpine.js — must be started manually when bundled (not CDN)
window.Alpine = Alpine;
Alpine.start();
