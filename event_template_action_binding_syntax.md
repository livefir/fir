
# Event Template Binding Syntax

The event template binding syntax in the provided code allows associating specific templates or actions with events triggered by the frontend. This syntax is used to define how events (like `ok`, `error`, `pending`, etc.) interact with templates or blocks of HTML. Here's an explanation of the syntax:

## General Syntax

```html
@fir:<event>:<state>::<template-name>="javascript code or function call"
```

### Components of the Syntax

1. **`@fir:` or `x-on:fir:`**:
   - Prefix indicating that the attribute is related to a `fir` event binding.
   - `@fir:` is shorthand, while `x-on:fir:` is compatible with Alpine.js.

2. **`<event>`**:
   - The name of the event being handled (e.g., `create`, `delete`, `toggle-done`).

3. **`<state>`**:
   - The state of the event. Common states include:
     - `ok`: Event succeeded.
     - `error`: Event failed.
     - `pending`: Event is in progress.
     - `done`: Event is completed.
   - States can also have a `.nohtml` modifier to indicate no HTML should be rendered.

4. **`::<template-name>`** (Optional):
   - Specifies the name of the template or block to render when the event occurs.
   - If omitted, the default behavior is used.

5. **`="javascript code or function call"`** (Optional):
   - JavaScript code or function to execute when the event occurs.
   - This can be used for custom actions or to manipulate the DOM.
6. **`[ ]`** (Optional):
   - Square brackets can be used to group multiple events and states together.
   - Example: `@fir:[create:ok,delete:error]` binds to both `create:ok` and `delete:error`.
7. **`<key>`** (Optional):
   - A unique identifier for the element, used for tracking and managing state.
   - Example: `fir-key="{{ .ID }}"` assigns a unique key to the element.
8. **`<modifier>`** (Optional):
   - Modifiers can be added to the event binding to control behavior.
   - Common modifiers include `.prevent`, `.stop`, `.self`, etc.


## Alternative Declarative Syntax

The declarative syntax allows you to bind events to templates or actions in a more readable and maintainable way. This is particularly useful for managing dynamic updates in the frontend.

```html
<div x-fir-render="inc,dec">
</div>
   
```html
<div 
   x-fir-render-task="[create:ok,update-later:ok]->task" 
   x-fir-action-task="$fir.prependEl()" 
   x-fir-render-later-task="update-now:ok->later-task" 
   x-fir-action-later-task="$fir.prependEl()" 
   x-fir-render-query-more="query-more:ok" 
   x-fir-action-query-more="$fir.appendEl()" 
   x-fir-render-query="query:ok" 
   x-fir-action-query="$fir.replace()">
</div>

<div x-fir-render="
create:ok,update-later->task=>fir.prependEl;
update-now=>fir.prependEl; 
query-more=>fir.appendEl;
query=>replace;
delete-task.nohtml=>fir.removeEl">
    <!-- This will be replaced with the result of the `query-more` template -->
</div>

    <!-- This will be replaced with the result of the `create` or `update-later` template -->

<div x-fir-render="delete-task=>fir.removeEl">
    <!-- This will be removed from the DOM when the `remove-task` event is triggered -->
</div>
<div x-fir-render="query=>replace;
delete-task.nohtml=>fir.removeEl">

</div>



```html


```string
create.pending
create:ok
create:error
create:pending
create:done
create:pending.window
create:ok.window=>fir.replace
create:ok=>myfunction
create:ok,delete:error=>fir.replace
create:ok,update:ok->todo=>fir.replace
create,update:ok->todo=>fir.replace
create,update->todo=>fir.replace
create:ok->todo
create:ok.window,delete:error.nohtml=>fir.replace;create:ok,delete:error.nohtml=>fir.replace;create:ok,delete:error=>fir.replace
```


### Examples

1. **Basic Event Binding**:

   ```html
   <div @fir:create:ok="js">
       <!-- Call `js` code when the `create` event succeeds -->
   </div>
   ```

2. **Multiple States**:

   ```html
   <div @fir:[create:ok,delete:error]="js">
       <!-- Call `js` code for `create:ok` or `delete:error` -->
   </div>
   ```

3. **State with `.nohtml` Modifier**:

   ```html
   <div @fir:create:ok="js">
       <!-- Call `js` code when the `create` event succeeds -->
   </div>
   ```

4. **State with `.nohtml` Modifier**:

   ```html
   <div @fir:create:ok.nohtml>
       <!-- Handle `create:ok` without rendering any HTML -->
   </div>
   ```

### Notes

- **Modifiers**: States can include modifiers like `.prevent`, `.stop`, `.self`, etc., to control event propagation.
- **Dynamic Class Names**: The system automatically generates CSS class names like `fir-create-ok--<key>` for elements, which can be used for styling or targeting.
- **Error Handling**: Events like `error` can be used to display validation or server-side errors.

This syntax provides a declarative way to bind events to templates, making it easier to manage dynamic updates in the frontend.
