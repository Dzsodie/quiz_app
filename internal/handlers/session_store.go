package handlers

import "github.com/gorilla/sessions"

// SessionStore manages session-related operations for the app
var SessionStore *sessions.CookieStore
