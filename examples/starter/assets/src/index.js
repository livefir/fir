import {Application} from "@hotwired/stimulus"
import {definitionsFromContext} from "@hotwired/stimulus-webpack-helpers"
import {Dispatcher} from 'goliveview';
import "./styles.scss";

const application = Application.start()
const context = require.context("./controllers", true, /\.js$/)
application.load(definitionsFromContext(context))
application.register("glv", Dispatcher)
window.Stimulus = application
