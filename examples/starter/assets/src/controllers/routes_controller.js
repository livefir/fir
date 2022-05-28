import {Controller} from "@hotwired/stimulus"
import { URLPattern } from "urlpattern-polyfill";

export default class extends Controller {
    static targets = ["route"]
    static values = {
        activeClass: {type: String, default: "is-link"},
    }

    connect() {
        this.routeTargets.forEach(el => {
                let p = new URLPattern(`${el.href}*`);
                let href = window.location.href;
                if (href.includes("#")){
                    href = href.split("#")[0];
                }
                let r = p.exec(href);
                if (r) {
                    el.classList.add(this.activeClassValue)
                } else {
                    el.classList.remove(this.activeClassValue)
                }

            }
        )
    }

}
