package functions

import (
	"html/template"
	"net/http"
)

type ErrorPageData struct {
	ErrorCode    int
	ErrorMessage string
}

func errorHandler(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	tmpl, _ := template.ParseFiles("tmpl/error.html")
	data := ErrorPageData{
		ErrorCode:    status,
		ErrorMessage: err.Error(),
	}
	tmpl.Execute(w, data)
}
