import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = [ "toggled" ]
    static values = {
        toggleClass: String
    }

    it(e){
        e.preventDefault()
        this.toggledTargets.forEach(item => {
            item.classList.toggle(this.toggleClassValue)
        })
    }

}