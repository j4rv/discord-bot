package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// currently unused, it needs some kind of auth
func initHTTPServer() {
	addSimpleCommand := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		key := strings.TrimSpace(r.FormValue("key"))
		response := strings.TrimSpace(r.FormValue("response"))
		if key == "" || response == "" {
			http.Error(w, "key and response must be set", http.StatusBadRequest)
			return
		}

		err := commandDS.addSimpleCommand(key, response)
		if err != nil {
			http.Error(w, fmt.Sprintf("addSimpleCommand() err: %v", err), http.StatusBadRequest)
			return
		}
		w.Write([]byte("Ok!"))
	}
	http.HandleFunc("/simpleCommand", addSimpleCommand)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println("ListenAndServe err:", err)
	}
}
