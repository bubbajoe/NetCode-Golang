package main

import (
	"encoding/base64"
	_ "encoding/json"
	"github.com/google/uuid"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/oauth2"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	_ "io"
	_ "io/ioutil"
	"log"
	"net/http"
	"os"
	_ "strconv"
	"strings"
	"time"
)

type User struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Username  string        `bson:"username"`
	Firstname string        `bson:"firstname"`
	Lastname  string        `bson:"lastname"`
	Password  string        `bson:"password"`
}

type Session struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Username   string        `bson:"username"`
	LastActive int64         `bson:"last_active"`
	SessionID  string        `bson:"sessionID"`
}

type Project struct {
	ID    string `bson:"_id,omitempty"`
	Title string
	Lang  string
	Desc  string
	Room  string `bson:"project_room,omitempty"`
	Text  string `bson:"text,omitempty"`
}

type TemplateData struct {
	Response string      `json:"response,omitempty"`
	Data     string      `json:"data,omitempty"`
	Lang     string      `json:"lang,omitempty"`
	File     string      `json:"file,omitempty"`
	Code     string      `json:"code,omitempty"`
	Tree     template.JS `json:"tree,omitempty"`
	Error    string      `json:"error,omitempty"`
	Username string      `json:"username,omitempty"`
}

// ProjectRoom
type Room struct {
	Users         map[string]socketio.Socket
	RecentUpdates []string
	Text          string
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var users *mgo.Collection

//var files *mgo.Collection
var projects *mgo.Collection
var sessions *mgo.Collection

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal("Socket Error: ", err)
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "127.0.0.1"
	}
	log.Println("Connecting to db at " + uri)
	session, err := mgo.Dial(uri)
	defer session.Close()

	if err != nil {
		log.Fatal("MongoDB Error: ", err)
	}
	dbName := os.Getenv("MONGODB")

	users = session.DB(dbName).C("users")
	//files = session.DB(dbName).C("files")
	projects = session.DB(dbName).C("projects")
	sessions = session.DB(dbName).C("sessions")

	activeUsers := make(map[string]socketio.Socket)
	rooms := make(map[string]map[string]socketio.Socket)
	recentUpdates := make(map[string][]string)

	server.On("connection", func(so socketio.Socket) {
		activeUsers[so.Id()] = so

		so.On("user:bind", func(data string) {
			log.Println("Socket.IO - room:bind " + data)
		})

		so.On("room:join", func(data string) {
			log.Println("Socket.IO - room:join " + data)
			var result Project
			projects.Find(bson.M{"project_name": data}).One(&result)
			room := result.Room
			rooms[room][so.Id()] = so
			so.Join(room)
			so.Emit("code:change", result.Text)
			for _, value := range recentUpdates[room] {
				so.Emit("code:update", value)
			}
		})

		so.On("room:leave", func(data string) {
			so.Leave(data)
			log.Println("Socket.IO - room:leave " + data)
		})

		so.On("code:update", func(data string) {
			for id, socket := range activeUsers {
				if id != so.Id() {
					socket.Emit("code:update", data)
					recentUpdates["default"] = append(recentUpdates["default"], data)
				}
			}
		})

		so.On("code:sync", func(data string) {

		})

		so.On("code:save", func(data string) {

		})

		so.On("code:check", func(data string) {

		})

		so.On("terminal:command", func(cmd string) {
			response := ""

			if cmd == "" {
				response = " "
			} else {
				switch response {
				case "login":
					so.Emit("terminal:response", "password: ")
					break
				case "":
					break
				default:
				}

				if id := getID(so.Request()); id != "" {
					var sess Session
					sessions.Find(bson.M{"sessionID": id}).One(&sess)
					if &sess != nil {
						switch cmd {
						case "whoami":
							response = sess.Username
							break
						case "room join":
							so.Join("default")
							response = "Room joined"
							break
						case "room leave":
							so.Join("default")
							response = "Room left"
							break
						case "send swarm":
							err := so.BroadcastTo("default", "swarm", "swarm from "+sess.Username)
							if err != nil {
								log.Print(err)
								response = "Could not send swarm"
							} else {
								response = "Swarm sent"
							}
						case "time":
							response = time.Now().String()
							break
						case "ping":
							so.Emit("pingpong", nil)
							break
						case "help":
							response = "whoami time"
							break
						default:
							response = "netcode: " + cmd + ": command not recognized"
						}
					}
				} else {
					response = "You need to log in"
				}
			}
			so.Emit("terminal:response", response)
		})

		so.On("ping", func(ms uint) {
			so.Emit("pong", ms)
		})

		so.On("user:log", func() {
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
	r.HandleFunc("/users/{username}", nil).Methods("GET")
	r.HandleFunc("/code", code).Methods("GET")
	r.HandleFunc("/projects", _projects).Methods("GET")
	r.HandleFunc("/projects/{p_name}", nil).Methods("GET")

	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/register", register).Methods("GET")
	r.HandleFunc("/login", _login).Methods("POST")
	r.HandleFunc("/register", _register).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	r.Handle("/socket.io/", server)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.Handle("/", r)

	log.Println("Serving at https://netcode-bubbajoe.c9users.io/ " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	// log.Println("Serving at https://localhost" )
	// log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", nil))
}

func tercon(z bool, a interface{}, b interface{}) interface{} {
	if z {
		return a
	}
	return b
}

func SetCookie(w http.ResponseWriter, name string, value string) {
	c := &http.Cookie{Name: name, Value: value}
	http.SetCookie(w, c)
}

func RemoveCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", MaxAge: -1})
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
			sessions.Find(bson.M{"sessionID": id}).One(&sess)
			username = sess.Username
		}
	}
	return username
}

func setSession(username string, w http.ResponseWriter) {
	id := uuid.New().String()
	sessions.Insert(&Session{
		SessionID:  id,
		LastActive: time.Now().Unix(),
		Username:   strings.ToLower(username),
	})
	value := map[string]string{"id": id}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: encoded, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "username", Value: username, Path: "/"})
	}
}

func clearSession(id string, w http.ResponseWriter) {
	sessions.Remove(bson.M{"sessionID": id})
	RemoveCookie(w, "session")
	RemoveCookie(w, "username")
}

// Routes
func homepage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/index.html",
		"src/partials/navbar.html",
		"src/partials/footer.html"}
	tmpl := template.Must(template.ParseFiles(paths...))

	homeItems := struct {
		Username string
		Projects []Project
	}{
		Username: getUsername(r),
	}

	err := tmpl.Execute(w, homeItems)
	if err != nil {
		log.Fatalf("template execution: %s", err)
		os.Exit(1)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func netcode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/netcode.html",
		"src/partials/navbar.html",
		"src/partials/footer.html",
	}

	tmpl := template.Must(template.ParseFiles(paths...))
	var data = TemplateData{
		Lang:     "HTML",
		File:     "netcode.html",
		Code:     `let welcomeTo = "Netcode"`,
		Tree:     `[{text:"Folder 1",nodes:[{text:"Folder 2",nodes:[{text:"File 1"}]},{text:"File 2"}]},{text:"Folder 3"},{text:"Folder 4"}]`,
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func code(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/code.html",
		"src/partials/themes.html",
		"src/partials/theme-options.html",
		"src/partials/navbar.html",
		"src/partials/footer.html"}
	tmpl := template.Must(template.ParseFiles(paths...))
	var data = TemplateData{
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func _projects(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/projects.html",
		"src/partials/navbar.html",
		"src/partials/footer.html"}
	tmpl := template.Must(template.ParseFiles(paths...))
	var data = TemplateData{
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/login.html",
		"src/partials/navbar.html",
		"src/partials/footer.html"}
	tmpl := template.Must(template.ParseFiles(paths...))

	var data = struct {
		Error    string
		Username string
	}{
		Error:    string(GetFlash(w, r, "error")),
		Username: getUsername(r),
	}
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	paths := []string{
		"src/register.html",
		"src/partials/navbar.html",
		"src/partials/footer.html"}
	tmpl := template.Must(template.ParseFiles(paths...))

	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func _login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.FormValue("Username")
	pass := r.FormValue("Password")
	redirectTarget := r.URL.Query().Get("redirectTarget")
	user := User{}
	if username != "" && pass != "" {
		err := users.Find(bson.M{"username": username}).One(&user)
		if err != nil {
			SetFlash(w, "error", "User does not exist")
			redirectTarget = r.URL.String()
		} else if CheckPasswordHash(pass, user.Password) {
			setSession(username, w)
		} else {
			SetFlash(w, "error", "Incorrect password")
			redirectTarget = r.URL.String()
		}
	}
	http.Redirect(w, r, redirectTarget, 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func _register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	username := r.FormValue("Username")
	pass := r.FormValue("Password")
	confirmpass := r.FormValue("ConfirmPassword")
	redirectTarget := "/register"
	if pass != confirmpass {
		SetFlash(w, "error", "Your passwords are different")
	} else if username != "" && pass != "" && confirmpass != "" {
		hash, _ := HashPassword(pass)
		err := users.Insert(bson.M{"username": username, "password": hash})
		if err != nil {
			log.Fatal(err)
		}
		SetFlash(w, "success", "Your account was created please log in")
		redirectTarget = "/login"
	}
	http.Redirect(w, r, redirectTarget, 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}

func logout(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clearSession(getID(r), w)
	http.Redirect(w, r, "/", 302)
	log.Printf(r.Method+" - "+r.URL.Path+" - %v\n", time.Now().Sub(start))
}
