import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
    static values = {
        id: String,
        activeClass: {type: String, default: "is-active"},
    }

    open({detail: obj}) {
        if (obj.modalId === this.idValue){
            this.element.classList.add(this.activeClassValue);
        }
    }

    close(e) {
        this.element.classList.remove(this.activeClassValue);
    }

    keyDown(e) {
        if (e.keyCode === 27) {
            this.close(e)
        }
        if (e.target.classList.contains("modal-background")) {
            this.close(e)
        }
    }
}
