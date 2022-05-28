import {Controller} from "@hotwired/stimulus"

export default class extends Controller {
    static targets = ["modal"]
    static classes = ["active"]

    goto(e) {
        if (e.currentTarget.dataset.goto) {
            window.location = e.currentTarget.dataset.goto;
        }
    }

    goback(e) {
        window.history.back();
    }

    openModal(e) {
        const targetModal = this.modalTargets.find(i => i.id === e.params.modalId);
        targetModal.classList.add(this.activeClass)
        e.preventDefault();
    }

    closeModal(e) {
        if (e.type === "click") {
            const targetModal = this.modalTargets.find(i => i.id === e.params.modalId);
            targetModal.classList.remove(this.activeClass)
            e.preventDefault();
            return;
        }
    }

    keyDown(e) {
        if (e.keyCode === 27) {
            this.modalTargets.forEach(item => {
                item.classList.remove(this.activeClass)
            })
        }

        if (e.keyCode === 37) {
            window.history.back();
        }

        if (e.keyCode === 39) {
            window.history.forward();
        }
    }

    toggle(e) {
        if (!e.currentTarget.dataset.toggleIds) {
            return;
        }
        if (!e.currentTarget.dataset.toggleClass) {
            return;
        }
        const targetToggleIds = e.currentTarget.dataset.toggleIds.split(",");
        const targetToggleClass = e.currentTarget.dataset.toggleClass;
        targetToggleIds.forEach(item => {
            document.getElementById(item).classList.toggle(targetToggleClass);
        })
    }
}

