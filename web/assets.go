package web

import "embed"

// Assets contains the frontend so the deployed application is one Go binary.
//
//go:embed index.html styles.css app.js
var Assets embed.FS
