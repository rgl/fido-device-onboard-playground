package main

import (
	"bufio"
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
	output, err := exec.Command(
		"tpm2",
		"getcap",
		"handles-nv-index",
	).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("ERROR: %v\n%s", err, output)
	}
	found := false
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "- 0x1D10001" {
			found = true
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Sprintf("ERROR scanning output: %v", err)
	}
	if !found {
		return "device not yet initialized"
	}
	output, err = exec.Command(
		"go-fdo-client",
		"print",
		"--tpm", "/dev/tpmrm0",
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
