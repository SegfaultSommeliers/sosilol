import "materialize-css/dist/js/materialize.min.js";
import hljs from "highlight.js";
import Alpine from "@alpinejs/csp";

Alpine.data("editor", () => ({
    text: "",
    saving: false,

    async save() {
        if (!this.text || this.text.length < 1) return;
        this.saving = true;
        try {
            const csrfToken = document.cookie
                .split("; ")
                .find((c) => c.startsWith("csrf_="))
                ?.split("=")[1] ?? "";

            const resp = await fetch("/save", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "X-Csrf-Token": csrfToken,
                },
                body: JSON.stringify({ text: this.text }),
                credentials: "same-origin",
            });

            if (!resp.ok) {
                alert("Ошибка при сохранении пасты");
                return;
            }

            const { id } = await resp.json();
            if (id) {
                await navigator.clipboard.writeText(
                    window.location.origin + "/view/" + id
                ).catch(() => {});
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
        hljs.highlightElement(codeview);
    }
    M.Sidenav.init(document.querySelectorAll(".sidenav"), {});
});
