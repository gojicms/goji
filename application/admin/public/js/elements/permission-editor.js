import { LitElement, html, css } from '../vendor/lit-all.min.js';

export class PermissionEditor extends LitElement {
    static formAssociated = true;

    constructor(props) {
        super(props);
        this._value = 0;
        this._internals = this.attachInternals();

        // Pre-bind the toggle methods to maintain reference stability
        this.toggleRead = () => this.toggle(0);
        this.toggleWrite = () => this.toggle(1);
        this.toggleCreate = () => this.toggle(2);
        this.toggleDelete = () => this.toggle(3);
        this.toggleExecute = () => this.toggle(4);
    }

    // Form-associated element callbacks
    formAssociatedCallback(form) {
        // Called when the element is associated with a form
        console.log('Associated with form:', form);
    }

    formDisabledCallback(disabled) {
        // Handle disabled state
        this.toggleAttribute('disabled', disabled);
    }

    formResetCallback() {
        this.value = 0;
    }

    formStateRestoreCallback(state) {
        this.value = state ? parseInt(state, 10) : 0;
    }

    render() {
        return html`
            <button id="read" @click="${this.toggleRead}" aria-checked="${(this.value&1)===1}">Read</button>
            <button id="write" @click="${this.toggleWrite}" aria-checked="${(this.value&2)===2}">Write</button>
            <button id="create" @click="${this.toggleCreate}" aria-checked="${(this.value&4)===4}">Create</button>
            <button id="delete" @click="${this.toggleDelete}" aria-checked="${(this.value&8)===8}">Delete</button>
            <button id="execute" @click="${this.toggleExecute}" aria-checked="${(this.value&16)===16}">Execute</button>
        `;
    }

    get value() {
        return this._value;
    }

    set value(newValue) {
        this._value = newValue;
        this._internals?.setFormValue(newValue);
    }

    // Modify the toggle method to directly update the value
    toggle(index) {
        this.value = this.value ^ (1 << index);
    }

    static get properties() {
        return {
            value: { type: Number, reflect: true },
            name: { type: String, reflect: true }
        };
    }

    // Same styles as before
    static styles = css`
        :host {
            display: inline-flex;
            flex-direction: row;
            border: 1px solid var(--primary-color-60);
            border-radius: 6px;
            overflow: hidden;
        }
        
        button {
            font-size: 1.10rem;
            font-weight: bold;
            border: none;
            padding: 8px 16px;
            background: var(--primary-color-90);
            color: var(--primary-color-50);
            &[aria-checked="true"] {
                border-bottom: 4px solid var(--primary-color-30);
                color: var(--primary-color-10);
                background: var(--primary-color-80);
            }
        }
    `;
}

customElements.define('gc-permission-editor', PermissionEditor);
