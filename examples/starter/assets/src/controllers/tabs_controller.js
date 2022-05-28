import { Controller } from "@hotwired/stimulus";

export default class extends Controller {

    static targets = [ "tab", "tabPanel" ]
    static values = {
        defaultTabkey: String,
        disableHistory: Boolean
    }
    static classes = [ "active", "hidden" ]

    connect() {
        let activeTabKey = undefined;
        if (window.location.hash && !this.disableHistoryValue){
            activeTabKey = window.location.hash.substring(1);
        }else if (!this.disableHistoryValue){
            activeTabKey = localStorage.getItem('tabkey')
        }
        if (!activeTabKey){
            activeTabKey = this.defaultTabkeyValue;
        }

        this.activateTab(activeTabKey)
    }

    activateTab(tabkey){
        this.tabPanelTargets.forEach((el, i) => {
            if(el.dataset.tabkey === tabkey){
                el.classList.remove(this.hiddenClass)
            } else {
                el.classList.add(this.hiddenClass)
            }
        })

        this.tabTargets.forEach((el, i) => {
            if(el.dataset.tabkey === tabkey){
                el.classList.add(this.activeClass)
                if (!this.disableHistoryValue){
                    history.replaceState(null,null,`#${el.dataset.tabkey}`)
                    localStorage.setItem("tabkey",el.dataset.tabkey)
                }
            } else {
                el.classList.remove(this.activeClass)
            }
        })
    }

    activate(e){
        let tabkey = undefined;
        if(e.currentTarget.dataset.tabkey){
            tabkey = e.currentTarget.dataset.tabkey;
        } else  if(e.currentTarget.parentElement.dataset.tabkey){
            tabkey = e.currentTarget.parentElement.dataset.tabkey
        }
        this.activateTab(tabkey);
    }

}