# DevLoop 🚀

> Manage. Run. Track your developer scripts easily.

DevLoop is a desktop and web app for managing and executing scripts across multiple languages, with an intuitive UI to configure script arguments dynamically.

## ✨ Features

- Add folders with your scripts (JS, TS, SH, PY, GO supported).
- Use metadata to configure script parameters and descriptions.
- Execute scripts with configurable arguments.
- Track execution history, success/failure, and rerun scripts with the same parameters.
- Sort and search scripts by category, tags, or name.
- Runs as Desktop App (Electron) or as Web App.

## 📂 Script Metadata Example

```js
// @name: Hello
// @description: A simple script that prints "Hello, {name}" to the console
// @author: Nir Adler
// @category: Testing
// @tags: ["hello", "test"]
// @inputs: [
//   { "name": "name", "description": "Your name", "type": "string", "default": "" }
// ]
console.log("Hello:", process.argv.slice(2)[0]);
```
