import { css, html, LitElement } from '../vendor/lit-all.min.js';

export class Table extends LitElement {
    constructor() {
        super();
        this.count = this.count || 10
    }

    render() {
        return this.hasPagination ? html`
            <slot></slot>
            <div id="pager">
                <span>Showing items ${this.rangeStart} to ${this.rangeEnd} of ${this.total}</span>
                <select  @change="${this.setPage}">
                    ${this.pageArray.map(page => html`<option value="${page-1}" ?selected="${this.page === page-1}">${page}</option>`)}
                </select>
            </div>` : html`<slot></slot>`
    }

    get hasPagination() {
        return !!(this.count > 0 && this.total > 0)
    }

    get rangeStart() {
        return this.offset + 1
    }

    get rangeEnd() {
        return Math.min(this.offset + this.count, this.total);
    }

    get pages() {
        return Math.ceil(this.total / this.count)
    }

    get page() {
        return Math.floor(this.offset / this.count)
    }

    get pageArray() {
        let pages = []
        for (let i = 1; i <= this.pages; i++) {
           pages.push(i)
        }
        return pages
    }

    setPage(event) {
        let page = event.target.value * this.count
        window.location = "?offset=" + page + "&count=" + this.count
    }

    static properties = {
        count: { type: Number, reflect: true },
        offset: { type: Number, reflect: true },
        total: { type: Number, reflect: true },
    }

    static styles = css`
        :host {
            display: block;
            width: 100%;
            border-radius: 6px;
            border: 1px solid var(--neutral-color-70);
            overflow: hidden;
        }       
        
        #pager {
            display: flex;
            justify-content: space-between;
            padding: 8px;
            background-color: var(--neutral-color-90);
            border-top: 2px solid var(--neutral-color-70);
            box-shadow: 0px -2px 0px 0px white;
        }
        
    `
}
customElements.define('gc-table', Table);