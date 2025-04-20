# 🧠 Developer Loop

A simple REST API built with Go and Gin to manage, store, and execute script snippets (e.g., Bash or Shell scripts). Supports metadata, parameterized execution, and history tracking. Public UI, and Swagger is provided for documentation.

---

## 🚀 Features

- Store and retrieve script metadata
- Execute `.sh` or `.bash` scripts with dynamic environment variables
- Auto-detects `/bin/bash` or `/bin/sh` based on script file extension
- Stores execution history
- Public UI served from `/public`
- Auto-generated Swagger docs available at `/swagger/index.html`

---

Open your browser at:  
- Swagger: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)  
- UI (if exists): [http://localhost:8080/public](http://localhost:8080/public)

---
