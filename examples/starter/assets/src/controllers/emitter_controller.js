import { Controller } from 'stimulus'

export default class extends Controller {
    event(e) {
        e.preventDefault();
        e.stopPropagation();
        const {eventId, ...rest} = e.params
        this.dispatch(eventId,  { detail: {...rest}} )
    }
}