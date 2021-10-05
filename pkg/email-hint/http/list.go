package http

import "net/http"

func GetPhonesByEmailPrefix(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, from gopher-corp!"))
}
