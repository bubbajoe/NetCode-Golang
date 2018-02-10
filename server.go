package main

import (
	// "io"
	"os"
	"log"
	"time"
	// "tree"
	"strings"
	// "strconv"
	"net/http"
	// "io/ioutil"
	"html/template"
	// "encoding/json"
	"encoding/base64"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/securecookie"
	"github.com/googollee/go-socket.io"
)

type User struct {
	ID bson.ObjectId 	`bson:"_id,omitempty"`
	Username string		`bson:"username"`
	Firstname string	`bson:"firstname"`
	Lastname string		`bson:"lastname"`
	Password string		`bson:"password"`
}

type Session struct {
	ID bson.ObjectId 	`bson:"_id,omitempty"`
	Username string		`bson:"username"`
	LastActive int64		`bson:"last_active"`
	SessionID string	`bson:"sessionID"`
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var users *mgo.Collection
var sessions *mgo.Collection

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

	users = session.DB("Netcode").C("users")
	sessions = session.DB("Netcode").C("sessions")

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

	r.HandleFunc("/", homepage)

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

func SetFlash(w http.ResponseWriter, name string, value string) {
	c := &http.Cookie{Name: name, Value: encode([]byte(value))}
	http.SetCookie(w, c)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) []byte {
	c, err := r.Cookie(name)
	if err != nil {
		switch err {
			case http.ErrNoCookie:
				return nil
			default:
				log.Fatal(err)
				return nil
		}
	}

	value, err := decode(c.Value)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
	return value
}

func encode(src []byte) string {
  	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
  	return base64.URLEncoding.DecodeString(src)
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

func getID(request *http.Request) (id string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			id = cookieValue["id"]
		}
	}
	return id
}

func getUsername(request *http.Request) (username string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			id := cookieValue["id"]
			var sess Session
			sessions.Find(bson.M{"sessionID":id}).One(&sess)
			username = sess.Username
		}
	}
	return username
}

func setSession(username string, response http.ResponseWriter) {
	id := uuid.New().String()
	sessions.Insert(&Session{
		SessionID: id,
		LastActive: time.Now().Unix(),
		Username: strings.ToLower(username),
	})
	value := map[string]string{"id": id}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(id string, response http.ResponseWriter) {
	sessions.Remove(bson.M{"sessionID":id})
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

type Project struct {
	ID string
	Title string
	Lang string
	Desc string
}

type TemplateData struct {
	Lang string
	File string
	Code string
	Tree template.JS
	Error string
	Username string
}

// Routes
func homepage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/index.html",
		"src/partials/navbar.html",
		"src/partials/footer.html" }
	tmpl := template.Must(template.ParseFiles(paths...))

	homeItems := struct {
		Username string
		Projects []Project
	} {
		Username: getUsername(r),
	}
	
	err := tmpl.Execute(w, homeItems)
	if err != nil {
			log.Fatalf("template execution: %s", err)
			os.Exit(1)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func netcode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/netcode.html",
		"src/partials/navbar.html",
		"src/partials/footer.html",
	}

	tmpl := template.Must(template.ParseFiles(paths...))
	var data = TemplateData {
		Lang: "HTML",
		File: "netcode.html",
		Code: `let welcomeTo = "Netcode"`,
		Tree: `[{text:"Folder 1",nodes:[{text:"Folder 2",nodes:[{text:"File 1"}]},{text:"File 2"}]},{text:"Folder 3"},{text:"Folder 4"}]`,
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
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
	var data = TemplateData {
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
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
	var data = TemplateData {
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
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

	var data = struct{
		Error string
		Username string
	} {
		Error: string(GetFlash(w,r,"error")),
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
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
	redirectTarget := "/login"
	user := User{}
	if username != "" && pass != "" {
		err := users.Find(bson.M{"username":username}).One(&user)
		if err != nil {
			SetFlash(w, "error", "User does not exist")
		} else if CheckPasswordHash(pass,user.Password) {
			setSession(username, w)
			redirectTarget = "/"
		} else {
			SetFlash(w, "error", "Incorrect password")
		}
	}
	http.Redirect(w, r, redirectTarget, 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func _register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.FormValue("Username")
	pass := r.FormValue("Password")
	confirmpass := r.FormValue("ConfirmPassword")

	if pass != confirmpass {
		
	} else if username != "" && pass != "" && confirmpass != ""  {
		hash, _ := HashPassword(pass)
		users.Insert(bson.M{"username":username,"password":hash})
		SetFlash(w, "success","Your account was created please log in")
	}
	http.Redirect(w, r, "register", 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}

func logout(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clearSession(getID(r),w)
	http.Redirect(w, r, "/", 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n",time.Now().Sub(start))
}