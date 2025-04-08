import {LitElement, css, html } from '../vendor/lit-all.min.js';

export class GcAlert extends LitElement {
    render() {
        let classes = [`${this.classList}`];
        if (this.inline) { classes.push('inline'); }
        if (this.type) { classes.push(this.type); }
        if (this.autoClose) { classes.push('autoClose'); }

        return html`<div id="element" class="${classes.join(' ')}">
            <slot></slot>
            ${this.dismissible || this.autoClose ? html`<button @click="${this.close}" type="button" class="close" data-dismiss="alert">&times;</button>` : html``}
        </div>`;
    }

    firstUpdated() {
        if (this.autoClose) {
            setTimeout(() => this.close(), 5000);
        }
    }

    close() {
        this.shadowRoot.querySelector('#element').classList.add('hiding');
        setTimeout(() => this.remove(), 300);
    }

    static properties = {
        inline: { type: Boolean, reflect: true },
        type: { type: String, reflect: true },
        class: { type: String, reflect: true },
        autoClose: { type: Boolean, reflect: true },
        dismissible: { type: Boolean, reflect: true },
    }


    static styles = css`
        :host {
            display: block;
            position: relative;
        }
    
        #element {
            display: flex;
            justify-content: space-between;
            padding: 16px 24px;
            margin: 0;
            opacity: 0.9;
            transition: opacity 300ms;
            font-size: 18px;
            font-family: sans-serif;
            left: 0;
            right: 0;
                        
            .close {
                background: none;
                border: none;
                color: white;
                font-size: 20px;
            }
            
            &.autoClose {
                position: absolute;
            }
            
            &.hiding {
                opacity: 0;
            }
            
            &.inline {
                opacity: 1;
                border-radius: 0;
                margin: 8px 0;
                border-radius: 4px;
                display: inline-flex;
            }
        
            &.error {
                --alert-color: var(--danger);
                color: white;
            }
        
            &.success {
                --alert-color: var(--success);
                color: white;
            }
        
            &.warning {
                --alert-color: var(--warning);
                color: white;
            }
            
            background: linear-gradient(to top, var(--alert-color), color-mix(in srgb, var(--alert-color), white 7%) 14%);
        }`
}
customElements.define('gc-alert', GcAlert)