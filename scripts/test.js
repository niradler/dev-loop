// @name: Hello
// @description: A simple script that prints "Hello, {name}" to the console
// @author: Nir Adler
// @category: Testing
// @tags: ["hello", "test"]
// @inputs: [
//   { "name": "name", "description": "Your name", "type": "string", "default": "" }
// ]
console.log("Hello:", process.argv.slice(2)[0]);
console.log("DEV_LOOP version:",process.env.DEV_LOOP)