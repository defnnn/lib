import './index.css'
import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

class ExtElement extends LitElement {
    createRenderRoot() {
        return this;
    }
}

@customElement("my-element")
export class MyElement extends ExtElement {
    @property() name = "Earthly";

    render() {
        return html`
            <button type="button"
                class="m-2 inline-flex items-center rounded-md border
                border-transparent bg-gray-600 px-6 py-3 text-base font-medium
                text-white shadow-sm hover:bg-indigo-700 focus:outline-none
                focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
            ${this.name}
            <!-- Heroicon name: mini/envelope -->
            <svg class="ml-3 -mr-1 h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path d="M3 4a2 2 0 00-2 2v1.161l8.441 4.221a1.25 1.25 0 001.118 0L19 7.162V6a2 2 0 00-2-2H3z" />
                <path d="M19 8.839l-7.77 3.885a2.75 2.75 0 01-2.46 0L1 8.839V14a2 2 0 002 2h14a2 2 0 002-2V8.839z" />
            </svg>
            </button>
        `;
    }
}
