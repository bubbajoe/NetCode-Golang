package main

import (
	// "io"
	"os"
	"log"
	"time"
	"strconv"
	"net/http"
	// "io/ioutil"
	"html/template"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/securecookie"
	"github.com/googollee/go-socket.io"
)

type User struct {
	ID bson.ObjectId	`bson:"_id,omitempty"`
	username string 	`bson:"username"`
	email string		`bson:"email,omitempty"`
	password string		`bson:"pass,omitempty"`
	//extra string		`bson:"extra,omitempty"`
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

var DB mgo.Database

func main() {
	port := "80"
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal("Socket Error: ",err)
	}

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		log.Fatal("MongoDB Error: ", err)
		
	}
	DB := session.DB("Netcode")

	defer session.Close()
	defer session.Close()

	activeUsers := make(map[string]socketio.Socket)
	server.On("connection", func(so socketio.Socket) {
		activeUsers[so.Id()] = so
		// so.Join("netcode")
		so.On("code:update", func(s socketio.Socket, data string) {
			// Send to all active users
			for id, socket := range activeUsers {
				if id != s.Id() {
					socket.Emit("code:update", data)
				}
			}
		})

		so.On("pong",func(ms int) {
			log.Printf("ms: %d", ms)
		})

		so.On("users",func() {
			log.Printf("active users: %d", activeUsers)
		})

		so.On("disconnection", func() {
			delete(activeUsers, so.Id())
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		paths := []string{
			"src/index.html",
			"src/partials/navbar.html",
			"src/partials/footer.html" }
		tmpl := template.Must(template.ParseFiles(paths...))
		
		err := tmpl.Execute(w, nil)
		if err != nil {
				log.Fatalf("template execution: %s", err)
				os.Exit(1)
		}
		log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
	})

	r.HandleFunc("/netcode", netcode).Methods("GET")
	r.HandleFunc("/code", code).Methods("GET")
	r.HandleFunc("/projects", projects).Methods("GET")

	// Authentication
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/register", register).Methods("GET")
	r.HandleFunc("/login", _login).Methods("POST")
	r.HandleFunc("/register", _register).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	r.Handle("/socket.io/", server)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.Handle("/",r)

	log.Println("Serving at http://localhost:"+port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	// log.Println("Serving at https://localhost" )
	// log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", nil))
}

// Password to hash
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

// Check if password and hash matvh
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//
func getUsername(request *http.Request) (username string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			username = cookieValue["name"]
		}
	}
	return username
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}
// Routes
func netcode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/netcode.html",
		"src/partials/navbar.html",
		"src/partials/footer.html",
	}
	tmpl := template.Must(template.ParseFiles(paths...))
	
	err := tmpl.Execute(w, nil)
	if err != nil {
			log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func code(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/code.html",
		"src/partials/theme.html",
		"src/partials/theme-options.html",
		"src/partials/navbar.html",
		"src/partials/footer.html" }
	tmpl := template.Must(template.ParseFiles(paths...))
	
	err := tmpl.Execute(w, nil)
	if err != nil {
			log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func projects(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/projects.html",
		"src/partials/navbar.html",
		"src/partials/footer.html" }
	tmpl := template.Must(template.ParseFiles(paths...))
	
	err := tmpl.Execute(w, nil)
	if err != nil {
			log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/login.html",
		"src/partials/navbar.html",
		"src/partials/footer.html" }
	tmpl := template.Must(template.ParseFiles(paths...))
	
	err := tmpl.Execute(w, nil)
	if err != nil {
			log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/register.html",
		"src/partials/navbar.html",
		"src/partials/footer.html" }
	tmpl := template.Must(template.ParseFiles(paths...))
	
	err := tmpl.Execute(w, nil)
	if err != nil {
			log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func _login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.FormValue("Username")
	pass := r.FormValue("Password")
	redirectTarget := "/"
	user := User{}
	if username != "" && pass != "" {
		f := bson.M{"username":username}
		err := DB.C("users").Find(f).One(&user)
		if err != nil {
			log.Print(err)
		}
		log.Println(user)
		log.Println(HashPassword(pass))
		setSession(username, w)
		redirectTarget = "/"
	}
	http.Redirect(w, r, redirectTarget, 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func _register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.FormValue("Username")
	pass := r.FormValue("Password")
	confirmpass := r.FormValue("ConfirmPassword")
	log.Println(username + " - " + strconv.FormatBool(confirmpass == pass))
	if pass != confirmpass {
		http.Redirect(w, r, "register", 302)
	} else {
	hash, err := HashPassword(pass)
	if err != nil {
		log.Println("Hash Failed")
	}
	log.Println(hash)
	http.Redirect(w, r, "register", 302)
	//DB.C("users").Insert(bson.M{"username":username,"password":hash})
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func logout(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clearSession(w)
	http.Redirect(w, r, "/", 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}