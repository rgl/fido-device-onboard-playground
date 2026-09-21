package main

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func getFDOState() string {
	if _, err := os.Stat("/device/credential.blob"); err != nil {
		return "device not yet initialized"
	}
	output, err := exec.Command(
		"go-fdo-client",
		"print",
		"--blob", "/device/credential.blob",
	).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("ERROR: %v\n%s", err, output)
	}
	return string(output)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fdoState := getFDOState()

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<!doctype html><html><body><h1>Welcome to the Device Home Page!</h1>")
		if _, err := os.Stat("/device/admin-password.txt"); err == nil {
			fmt.Fprint(w, `<p><a href="/admin">admin</a></p>`)
		}
		fmt.Fprint(w, "<h2>Fido Device Onboarding (FDO) State<h2>")
		fmt.Fprintf(w, "<p><pre>%s</pre></p>", html.EscapeString(fdoState))
		fmt.Fprint(w, "</body></html>")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, fdoState)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	expectedUsername := "admin"
	expectedPassword, err := os.ReadFile("/device/admin-password.txt")
	if err != nil {
		http.Error(w, "unavailable", http.StatusNotFound)
		return
	}
	expectedPassword = bytes.TrimSpace(expectedPassword)

	username, password, ok := r.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(expectedUsername)) != 1 || subtle.ConstantTimeCompare([]byte(password), expectedPassword) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="device"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<!doctype html><html><body><h1>Welcome to the Device Admin Page!</h1></body></html>")
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Welcome to the Device Admin Page!")
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/admin", adminHandler)

	fmt.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
