import {LitElement, css, html, live } from '../../vendor/lit-all.min.js';
import '../../vendor/purify.min.js';

export class GcHtmlEditor extends LitElement {
    static formAssociated = true;

    firstUpdated() {
        this.value = this.innerHTML.trim();
        if (this.value.trim() === '') {
            this.value = '<p>&nbsp;</p>';
        }
    }

    constructor(props) {
        super(props);
        this._internals = this.attachInternals(); // Attach internals
        this._mode = 'html';

        // Allow custom elements
        DOMPurify.addHook('uponSanitizeElement', (node, data) => {
            if (node?.tagName?.includes?.('-')) {
                data.allowedTags[data.tagName] = true;
            }
        });

        // Optionally allow specific attributes for custom elements
        DOMPurify.addHook('uponSanitizeAttribute', (node, data) => {
            if (node?.tagName?.includes('-') && data.attrName === 'my-attribute') {
                data.allowedAttributes[data.attrName] = true;
            }
        });

        document.addEventListener('selectionchange', this.updateStyleSelect);
        
    }

    get value() {
        return this._value;
    }

    set value(val) {
        const oldVal = this._value;
        this._value = val;
        this.requestUpdate('value', oldVal);
        this._internals.setFormValue(this.sanitize(this._value));
    }

    static properties = {
        name: { type: String, reflect: true },
        _mode: { type: String, value: 'html' },
    }

    sanitize(newValue) {
        return DOMPurify.sanitize(newValue);
    }

    setMode(newValue) {
        return () => {
            this._mode = newValue;
            if (this._mode === 'html') {
                setTimeout(() => {
                    let editor = this.shadowRoot.querySelector('gc-html-edit-window');
                    editor?.focus?.();
                    document.execCommand('defaultParagraphSeparator', false,'P');
                    document.execCommand('enableAbsolutePositionEditor', true,true);
                    document.execCommand('enableInlineTableEditing', true,true);
                    document.execCommand('enableObjectResizing', true,true);
                },1);
            }
        }
    }

    handleRawInput(event) {
        this.value = event.target.value;
    }

    handleEditorInput(event) {
        this.value = event.target.innerHTML;
    }

    do(command, arg) {
        return function() {
            document.execCommand(command, false, arg);
        }
    }

    doAsk(command, promptText) {
        return function() {
            let arg = prompt(promptText)
            document.execCommand(command, false, arg);
        }
    }

    setInlineStyle(event) {
        let style = event.target.value;
        if (style === 'none') {
            document.execCommand('formatBlock', false, 'P');
        } else {
            document.execCommand('formatBlock', true, style);
        }
    }

    updateStyleSelect = () => {
        const selection = this.shadowRoot.getSelection();
        if (!selection.rangeCount) return;

        const range = selection.getRangeAt(0);
        let element = range.commonAncestorContainer;
        console.log(element, range);

        // If we're inside a text node, get its parent element
        if (element.nodeType === 3) {
            element = element.parentElement;
        }

        // Find the closest block level element
        const blockElement = element.closest('p,h1,h2,h3,h4,h5,h6,pre');
        
        // Get the select element
        const select = this.shadowRoot.querySelector('select');
        console.log(blockElement);
        if (!select) return;

        // Update the select value based on the tag name
        if (blockElement) {
            const tagName = blockElement.tagName.toLowerCase();
            select.value = tagName;
        } else {
            select.value = 'none';
        }
    }

    render() {
        return html`
            <gc-html-edit-bar>
                ${this._mode === 'html' ? html`
                    <div class="wrap">
                        <div class="group">
                            <button @click="${this.do('bold')}">B</button>
                            <button @click="${this.do('italic')}">I</button>
                        </div>
                        <div class="group">
                            <button @click="${this.do('formatBlock', 'P')}">¶</button>
                            <button @click="${this.doAsk('createLink', 'Where do you wish the link to go to?')}">A</button>
                        </div>
                        <div class="group">
                            <select @change="${this.setInlineStyle}">
                                <option value="none">None</option>
                                <option value="p">Paragraph</option>
                                <option value="h1">H1</option>
                                <option value="h2">H2</option>
                                <option value="h3">H3</option>
                                <option value="h4">H4</option>
                                <option value="h5">H5</option>
                                <option value="h6">H6</option>
                                <option value="pre">Preformatted</option>
                            </select>
                        </div>
                        <div class="group">
                            <button @click="${this.do('insertUnorderedList')}">
                                <img src="/admin/public/js/elements/editor/el-ul.svg" width="16" height="16" alt="Unordered List">
                            </button>
                            <button @click="${this.do('insertOrderedList')}">
                                <img src="/admin/public/js/elements/editor/el-ol.svg" width="16" height="16" alt="Ordered List">
                            </button>
                        </div>
                    </div>
                `: ``}
                <gc-html-edit-gap></gc-html-edit-gap>
                <div class="tabs">
                    <button @click="${this.setMode('html')}" class="${this._mode === 'html' ? 'active' : ''}">Visual</button>
                    <button @click="${this.setMode('raw')}" class="${this._mode === 'raw' ? 'active' : ''}">HTML</button>
                </div>
            </gc-html-edit-bar>
            <input type="hidden" name="${this.name}" value="${this._value}"></input>
            ${this._mode === 'html' ? html`<gc-html-edit-window @input="${this.handleEditorInput}" contentEditable .innerHTML="${live(this.value)}"></gc-html-edit-window>` : ``}
            ${this._mode === 'raw' ? html`<textarea @input="${this.handleRawInput}">${this.value}</textarea>` : ``}
        `
    }

    static styles = css`
        :host {
            margin-top: 8px
        }
        
        gc-html-edit-gap {
            flex-grow: 1;
        }
        
        gc-html-edit-bar {
            min-height: 40px;
            display: flex;
            gap: 12px;
            flex-direction: row;
            align-items: center;
            border: 1px solid var(--neutral-color-60);
            border-radius: 6px 6px 0 0;
            border-bottom: 0;
            background: linear-gradient(to top, var(--primary-color-90), white 8px);
            .wrap {
                display: flex;
                flex-direction: row;
                gap: 12px;
                flex-wrap: wrap;
            }
            .group {
                display: flex;
                gap: 2px;
                flex-wrap: wrap;
                align-items: center;
                button {
                    background: transparent;
                    border: 0;
                    font-size: 14px;
                    font-weight: 600;
                    padding: 10px 14px;
                    &:hover {
                        color: var(--primary-color-20);
                    }
                }
                select {
                    background: transparent;
                    border: 0;
                    font-size: 14px;
                    font-weight: 600;
                    padding: 10px 14px;
                    
                }
            }
            .tabs {
                display: flex;         
                flex-shrink: 0;
                height: fit-content;
                gap: 1px;
                overflow: hidden;
                button {
                    border: 0;
                    background: transparent;
                    font-size: 15px;
                    font-weight: 600;
                    padding: 6px 8px;
                    &.active {
                        color: var(--primary-color-30);
                    }
                }
            }
        }
        
        gc-html-edit-window, textarea {
            box-sizing: border-box;
            display: block;
            min-height: 300px;
            min-width: 100%;
            font-size: 18px;
            padding: 16px;
            background: linear-gradient(to bottom, var(--primary-color-80), white 8px);
        
            border-top: 0;
            border: 1px solid var(--neutral-color-60);
            border-radius: 0 0 6px 6px;

            /* Focus handling */
            outline: 0 solid transparent;
            outline-offset: 1px;
            transition: outline-color 0.3s, border-color 0.3s, outline-width 0.3s;
            &:focus-within {
                outline-width: 4px;
                outline-color: var(--primary-color-70);
                border-color: var(--secondary-color-30);
            }
        
            &:focus {
                outline: none;
            }
        }
    `
}
customElements.define('gc-html-editor', GcHtmlEditor);

