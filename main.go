package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var (
	templates = template.Must(template.ParseGlob("static/*.html"))
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "index.html", nil)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "hello.html", nil)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "form.html", nil)
	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		name := r.Form.Get("name")
		dob := r.Form.Get("dob")
		terms := r.Form.Get("terms")

		if name == "" || dob == "" || terms == "" {
			http.Error(w, "All fields are required and you must agree to the terms", http.StatusBadRequest)
			return
		}

		data := struct {
			Name string
			DOB  string
		}{
			Name: name,
			DOB:  dob,
		}

		renderTemplate(w, "success.html", data)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/form", formHandler)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
