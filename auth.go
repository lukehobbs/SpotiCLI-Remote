package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"

	"golang.org/x/oauth2"
)

func startAuth() {
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	log.Fatal(http.ListenAndServe(":8080", nil))

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	usr, err := client.CurrentUser()
	checkErr(err)
	fmt.Println("You are logged in as:", usr.ID)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	var err error
	tok, err = auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	_, err = fmt.Fprintln(w, "Login Completed!")
	checkErr(err)
	ch <- &client
	err = saveToken(tok)
	checkErr(err)
}

func saveToken(t *oauth2.Token) error {
	tok := &t
	usr, err := user.Current()
	checkErr(err)
	if _, err = os.Stat(usr.HomeDir + tokenDir); os.IsNotExist(err) {
		err = os.Mkdir(usr.HomeDir+tokenDir, 0600)
		checkErr(err)
	}
	path := usr.HomeDir + tokenFile
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err == nil {
		e := gob.NewEncoder(file)
		err = e.Encode(tok)
		checkErr(err)
	}
	err = file.Close()
	checkErr(err)
	return err
}

func loadToken() error {
	usr, err := user.Current()
	checkErr(err)
	path := usr.HomeDir + tokenFile
	file, err := os.Open(path)
	if err == nil {
		d := gob.NewDecoder(file)
		err = d.Decode(&tok)
		checkErr(err)
	}
	err = file.Close()
	checkErr(err)
	return err
}
