package main

import(
    "fmt"
    "net/http"
    "encoding/json"
    "github.com/julienschmidt/httprouter"
)

type user struct {
    name string
    content userContent
}

type userContent map[string]string

var db map[string]user

func init() {
    db = make(map[string]user)
}

type topicMsg struct {
    Text    string     `json:"text"`
}

type topicListMsg struct {
    Topics []string   `json:"topics"`
}

func userExists(u string) bool {
    _ , ok := db[u]
    return ok
}

func topicExists(u, t string) bool {
    if !userExists(u) {
        return false
    }
    _ , ok := db[u].content[t]
    return ok
}

func createUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    u := ps.ByName("user")
    if _ , ok := db[u]; !ok {
        fmt.Printf("Create user content for %s\n", u)
        db[u] = user{ name: u , content: make(userContent) }
        w.WriteHeader(http.StatusCreated)
        return
    }
    fmt.Printf("User %s was created\n", u)
}

func createUserTopic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    u := ps.ByName("user")
    t := ps.ByName("topic")

    if !userExists(u) || topicExists(u, t) {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    var msg topicMsg

    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        fmt.Println(err.Error())
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    db[u].content[t] = msg.Text

    w.WriteHeader(http.StatusCreated)
}

func getUserTopic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    u := ps.ByName("user")
    t := ps.ByName("topic")

    if !userExists(u) {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    if !topicExists(u, t) {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    m := topicMsg{ db[u].content[t] }
    json.NewEncoder(w).Encode(&m)
}

func listUserTopic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    u := ps.ByName("user")

    if !userExists(u) {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    ts := []string{}
    for k := range db[u].content {
        ts = append(ts, k)
    }

    m := topicListMsg{ ts }

    json.NewEncoder(w).Encode(&m)
}

func main() {
    router := httprouter.New()
    router.POST("/wlog/:user", createUser)
    router.GET("/wlog/:user", listUserTopic)
    router.POST("/wlog/:user/:topic", createUserTopic)
    router.GET("/wlog/:user/:topic", getUserTopic)

    server := http.Server{
        Addr: ":8000",
        Handler: router,
    }

    fmt.Println("Start server on port 8000.")
    server.ListenAndServe()
}
