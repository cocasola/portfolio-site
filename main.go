package main

import (
	"fmt"
	"main/view"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	c "main/view/components"

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

type contactFields struct {
	name    c.Field
	email   c.Field
	message c.Field
}

func initContactFields() contactFields {
	return contactFields{
		name: c.Field{
			Name:        "name",
			Placeholder: "Your name",
			Limit:       16,
			Validate: func(value string) string {
				if len(value) < 2 {
					return "Please enter a name."
				} else {
					return ""
				}
			},
		},
		email: c.Field{
			Name:        "email",
			Placeholder: "Your email",
			Limit:       254,
			Validate: func(value string) string {
				if len(value) < 4 {
					return "Please enter an email."
				} else {
					return ""
				}
			},
		},
		message: c.Field{
			Name:        "message",
			Placeholder: "Write your message here...",
			Rows:        5,
			Limit:       1024,
			Validate: func(value string) string {
				if len(value) < 8 {
					return "Please enter a proper message."
				} else {
					return ""
				}
			},
		},
	}
}

func (fields *contactFields) ptrArray() []*c.Field {
	return []*c.Field{&fields.name, &fields.email, &fields.message}
}

func main() {
	InitMail()

	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", handler(func(w http.ResponseWriter, r *http.Request) error {
		return RenderView(w, r, []Wrapper{layout}, view.HeaderPage(view.MainPages()))
	}))

	router.HandleFunc("GET /contact", handler(func(w http.ResponseWriter, r *http.Request) error {
		fields := initContactFields()
		return RenderView(w, r, []Wrapper{layout}, view.Contact(fields.ptrArray()))
	}))

	lastMessages := &struct {
		Map map[string]int64
		Mu  sync.Mutex
	}{
		Map: make(map[string]int64),
	}

	router.HandleFunc("POST /contact", handler(func(w http.ResponseWriter, r *http.Request) error {
		fields := initContactFields()

		err := c.ReadBodyIntoFields(r, fields.ptrArray())
		if err != nil {
			return c.Fields(fields.ptrArray()).Render(r.Context(), w)
		}

		{
			lastMessages.Mu.Lock()
			defer lastMessages.Mu.Unlock()

			now := time.Now().Unix()

			last, ok := lastMessages.Map[r.RemoteAddr]
			if ok {
				if now-last < 60 {
					return c.FieldsError(
						"You just sent a message!",
						fields.ptrArray(),
					).Render(r.Context(), w)
				}
			}
			lastMessages.Map[r.RemoteAddr] = now
		}

		message := fmt.Sprintf(
			"Name: %s\n"+
				"Email: %s\n"+
				"-------------------------------------------------------\n"+
				"%s",
			fields.name.Value, fields.email.Value, fields.message.Value,
		)

		err = SendEmail(message)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		return c.FieldsSuccess(
			"Successfully sent message!",
			fields.ptrArray(),
		).Render(r.Context(), w)
	}))

	router.HandleFunc("GET /favicon.ico", handler(func(w http.ResponseWriter, r *http.Request) error {
		http.NotFound(w, r)
		return nil
	}))

	router.Handle(
		"GET /static/",
		http.StripPrefix("/static", http.FileServer(http.Dir("static"))),
	)

	router.HandleFunc("GET /404", handler(func(w http.ResponseWriter, r *http.Request) error {
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
