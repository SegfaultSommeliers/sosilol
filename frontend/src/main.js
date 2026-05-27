import "materialize-css/dist/js/materialize.min.js";
import hljs from "highlight.js";
import Alpine from "alpinejs";
import "htmx.org";

window.hljs = hljs;

Alpine.data("editor", () => ({
    text: "",
    saving: false,

    async save() {
        if (!this.text || this.text.length < 1) return;
        this.saving = true;
        try {
            const resp = await fetch("/save", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: "text=" + encodeURIComponent(this.text),
                redirect: "follow",
                credentials: "same-origin",
            });
            const id = resp.url.split("/view/")[1];
            if (id) {
                const newUrl = window.location.origin + "/view/" + id;
                await navigator.clipboard.writeText(newUrl).catch(() => {});
                window.location.href = "/view/" + id;
            }
        } finally {
            this.saving = false;
        }
    },

    handleTab(event) {
        const el = event.target;
        const s = el.selectionStart;
        const e = el.selectionEnd;
        this.text = this.text.substring(0, s) + "\t" + this.text.substring(e);
        this.$nextTick(() => {
            el.selectionStart = el.selectionEnd = s + 1;
        });
    },
}));

window.Alpine = Alpine;
Alpine.start();

function onReady(fn) {
    if (document.readyState !== "loading") {
        fn();
    } else {
        document.addEventListener("DOMContentLoaded", fn);
    }
}

onReady(() => {
    const codeview = document.querySelector(".code-view");
    if (codeview) {
        codeview.innerHTML = hljs.highlightAuto(codeview.innerText).value;
    }
    M.Sidenav.init(document.querySelectorAll(".sidenav"), {});
});
