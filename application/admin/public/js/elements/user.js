import {LitElement, css, html } from '../vendor/lit-all.min.js';

export class GcUser extends LitElement {
    render() {
        return html`<img src="${this.avatar}" />`;
    }


    static properties = {
        avatar: { type: String, reflect: true },
    }


    static styles = css`
        :host {
            display: block;
            width: 48px;
            height: 48px;
            border-radius: 50%;
            overflow: hidden;
            box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
        }
        
        img {
            width: 100%;
            height: 100%;
        }
    `
}
customElements.define('gc-user', GcUser)