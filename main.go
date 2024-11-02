package main

import (
	"fmt"
	"main/view"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
)

func handler(callback func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := callback(w, r)
		if err != nil {
			fmt.Println("failed to handle request: ", err)
		}
	}
}

type WrapperFn func(templ.Component) (templ.Component, error)

type Wrapper struct {
	Path string
	Fn   WrapperFn
}

func RenderView(
	w http.ResponseWriter,
	r *http.Request,
	wrappers []Wrapper,
	c templ.Component,
) error {
	urlParts := strings.Split(r.Header.Get("HX-Current-URL"), "/")
	skip := 0

	if len(urlParts) > 1 {
		for i := 3; i < len(urlParts); i++ {
			if wrappers[len(wrappers)-1-skip].Path == urlParts[i] {
				skip++
			}
		}
	}

	for i := len(wrappers) - 1 - skip; i >= 0; i-- {
		next, err := wrappers[i].Fn(c)
		if err != nil {
			return fmt.Errorf("failed to render view: %w", err)
		}

		c = next
	}

	return c.Render(r.Context(), w)
}

var layout = Wrapper{
	Path: "",
	Fn:   func(c templ.Component) (templ.Component, error) { return view.Layout(c), nil },
}

func main() {
	InitMail()

	router := http.NewServeMux()

	router.HandleFunc("/{$}", handler(func(w http.ResponseWriter, r *http.Request) error {
		return RenderView(w, r, []Wrapper{layout}, view.HeaderPage(view.Home()))
	}))

	router.HandleFunc("/contact", handler(func(w http.ResponseWriter, r *http.Request) error {
		return RenderView(w, r, []Wrapper{layout}, view.Contact())
	}))

	router.HandleFunc("/favicon.ico", handler(func(w http.ResponseWriter, r *http.Request) error {
		http.NotFound(w, r)
		return nil
	}))

	router.Handle(
		"GET /static/",
		http.StripPrefix("/static", http.FileServer(http.Dir("static"))),
	)

	router.HandleFunc("/404", handler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Header.Get("HX-Request") == "" {
			return RenderView(w, r, []Wrapper{layout}, view.E404())
		} else {
			w.Header().Add("HX-Redirect", "/404")
			w.Write([]byte{})

			return nil
		}
	}))

	router.HandleFunc("/", handler(func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		return nil
	}))

	if os.Getenv("ENV") == "PROD" {
		err := http.ListenAndServeTLS(":443", os.Getenv("CERT_PATH"), os.Getenv("KEY_PATH"), router)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		err := http.ListenAndServe(":80", router)
		if err != nil {
			fmt.Println(err)
		}
	}
}
