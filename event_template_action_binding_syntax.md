
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
<div x-fir-refresh="inc,dec" >
</div>
<div x-fir-remove="delete" >
</div>
<div x-fir-append:todo="inc,dec" >
</div>
<div x-fir-prepend:todo="inc,dec" >
</div>
<div x-fir-replace:todo="inc,dec" >
</div>
<div x-fir-toggleClass:loading="inc,dec" >
</div>
<div x-fir-resetForm="submit" >
</div>
<div x-fir-trigger:resetForm="inc,dec" x-fir-action-resetForm="$fir.resetForm()" >
</div>
<div x-fir-dispatch:close-modal="inc,dec" >
</div>

````

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
