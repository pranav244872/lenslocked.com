package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// home page handler
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Welcome to my awesome site!</h1>")
}

// contact page handler
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "To get in touch, please send an email "+
		"to <a href=\"mailto:support@lenslocked.com\">"+
		"support@lenslocked.com</a>.")
}

// faq page handler
func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<h1>Frequently Asked Questions (Shhh... it’s a secret!)</h1>

<details>
  <summary>Q: Why did the chicken cross the road?</summary>
  <p>A: To get to the other side... of this website! 🐔</p>
</details>

<details>
  <summary>Q: Can I have unlimited coffee while coding?</summary>
  <p>Sure! But remember: too much caffeine may cause spontaneous bug fixing at 3 AM.</p>
</details>

<details>
  <summary>Q: What\’s the meaning of life, the universe, and everything?</summary>
  <p>42. But don\’t quote me on that — I’m just a website.</p>
</details>

<details>
  <summary>Q: How do I unlock the secret mode?</summary>
  <p class="secret">Try typing <code>sudo make me a sandwich</code> in your terminal. No guarantees!</p>
</details>

<details>
  <summary>Q: Can you tell me a joke?</summary>
  <p>Why do programmers prefer dark mode? Because light attracts bugs! 🐞</p>
</details>
`)

}

// 404 page
func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>Oops! This page does not exist 🛸</h1>")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/faq", faq)
	r.NotFoundHandler = http.HandlerFunc(notFound)
	http.ListenAndServe(":3000", r)
}
