# Create an Alpha-compliant Node.js function

This guide will help you to set up your first Node.js function and execute it through Alpha.

## Requirements

- **Git repository** to host your function code (or similar, you will simply need something where you can store files and download them)

## Create the function

Once you've cloned / created your git repository, open it locally into your favourite code editor, and create an **index.js** file inside with the following placeholder content :

> **The name of the file must be exactly `index.js`**.

```javascript
exports.handler = function (ctx, params) {
  ctx.logger.log("Hello from my first JS function !")

  let name = params["name"] || "world";
  return { message: `Hello ${name} !`};
};
```

- First, we export a function called **handler**. The name, like the file name, is important and can't be changed for now.
- Your function takes 2 parameters : `ctx` and `params`. `ctx` is an object provided by the runtime that contains helpers like logger. `params` is simply a Javascript object that contains the variables given to the function. 
- The `return` statement at the end MUST return a valid JSON object representation.

