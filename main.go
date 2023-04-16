package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"time"
)

// Item represents an item with a key and a session ID
type Item struct {
	k   int
	sid Sid
}

// newItem creates a new Item with the given Sid and expiration time
func newItem(x Sid) Item {
	var tmp Item
	tmp.k = int(x.expiration.Unix())
	tmp.sid = x
	return tmp
}

type Node struct {
	// i is the Item stored in the Node
	i Item
	// p is the priority of the Node
	p int
	// l is the left child of the Node
	l *Node
	// r is the right child of the Node
	r *Node
	// parent is the parent of the Node
	parent *Node
}

// newNode creates a new Node with the given Item
func newNode(x Item) *Node {
	var tmp Node
	tmp.p = rand.Int()
	tmp.i = x
	return &tmp
}

// _merge merges two Nodes and returns the resulting Node
func _merge(l *Node, r *Node) *Node {
	if l == nil {
		return r
	}
	if r == nil {
		return l
	}
	if l.p > r.p {
		r.parent = l
		l.r = _merge(l.r, r)
		return l
	} else {
		l.parent = r
		r.l = _merge(l, r.l)
		return r
	}
}

// _split splits a Node into two Nodes based on the given key
func _split(p *Node, x int) (*Node, *Node) {
	if p == nil {
		return nil, nil
	}
	if p.i.k <= x {
		var l, r = _split(p.r, x)
		if l != nil {
			l.parent = p
		}
		p.r = l
		return p, r
	} else {
		var l, r = _split(p.l, x)
		if r != nil {
			r.parent = p
		}
		p.l = r
		return l, p
	}
}

// Treap is a binary search tree with randomized priorities
type Treap struct {
	_root, _begin, _end *Node
}

// _updBegin updates the beginning Node of the Treap
func (t *Treap) _updBegin() {
	t._begin = t._root
	for t._begin != nil && t._begin.l != nil {
		t._begin = t._begin.l
	}
}

// _updEnd updates the ending Node of the Treap
func (t *Treap) _updEnd() {
	t._end = t._root
	for t._end != nil && t._end.r != nil {
		t._end = t._end.r
	}
}

// begin returns the beginning Node of the Treap
func (t *Treap) begin() *Node {
	return t._begin
}

// end returns the ending Node of the Treap
func (t *Treap) end() *Node {
	return t._end
}

// count returns the number of occurrences of the given Item in the Treap
func (t *Treap) count(x Item) int {
	var p = t._root
	for p.i != x {
		if p.r != nil && p.i.k < x.k {
			p = p.r
			continue
		}
		if p.l != nil && x.k < p.i.k {
			p = p.l
			continue
		}
		break
	}
	if p.i == x {
		return 1
	} else {
		return 0
	}
}

// insert inserts the given Item into the Treap
func (t *Treap) insert(x Item) {
	if t._root != nil && t.count(x) != 0 {
		return
	}
	var l, r = _split(t._root, x.k)
	t._root = _merge(l, _merge(newNode(x), r))
	t._updBegin()
	t._updEnd()
}

// erase removes the given Item from the Treap
func (t *Treap) erase(x Item) {
	if t._root == nil || t.count(x) == 0 {
		return
	}
	var l, r = _split(t._root, x.k)
	l, _ = _split(l, x.k-1)
	t._root = _merge(l, r)
	t._updBegin()
	t._updEnd()
}

// Sid represents a session ID with an expiration time
type Sid struct {
	id         string
	expiration time.Time
}

// newSid generates a new session ID with a 24-hour expiration time
func newSid() Sid {
	var sid Sid
	sid.expiration = time.Now().Add(24 * time.Hour)
	sid.id = strconv.Itoa(rand.Int())
	return sid
}

// User represents a user with a name, login, password, and session IDs stored in a Treap and map
type User struct {
	name            string
	login, password string
	sidsByTime      Treap
	sidsBySid       map[string]time.Time
}

// newUser creates a new User with the given name, login, and password
func newUser(name string, login string, password string) User {
	var tmp User
	tmp.name = name
	tmp.login = login
	tmp.password = password
	tmp.sidsBySid = make(map[string]time.Time)
	return tmp
}

// getMessagesUpdate concatenates all messages in the given slice starting from the specified index
func getMessagesUpdate(r []string, from int) string {
	var res = ""
	for i := from; i < len(r); i++ {
		res += r[i]
	}
	return res
}

// resetCookie removes all cookies related to the chat application
func resetCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "name", Value: "", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "room", Value: "", MaxAge: -1})
}

// checkSids removes expired session IDs from the given user's Treap and map
func checkSids(usr string) {
	user, ok := users[usr]
	if ok {
		for len(user.sidsBySid) > 0 && user.sidsByTime.begin().i.sid.expiration.Before(time.Now()) {
			delete(user.sidsBySid, user.sidsByTime.begin().i.sid.id)
			user.sidsByTime.erase(user.sidsByTime.begin().i)
		}
		users[usr] = user
	}
}

// checkUser checks if the current request has a valid session ID for a logged-in user
func checkUser(r *http.Request) bool {
	usr, errName := r.Cookie("name")
	sid, errSid := r.Cookie("sid")
	if errName != nil || errSid != nil {
		return false
	}
	checkSids(usr.Value)
	sidExpiration, okUserSid := users[usr.Value].sidsBySid[sid.Value]
	if errName != nil || errSid != nil || !okUserSid || sidExpiration.Before(time.Now()) {
		return false
	} else {
		_, ok := users[usr.Value]
		return ok
	}
}

// checkRoom checks if the current request has a valid room cookie
func checkRoom(r *http.Request) bool {
	if _, err := r.Cookie("room"); err == nil {
		return true
	} else {
		return false
	}
}

// redirectToIndex redirects the user to the chat application's index page
func redirectToIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://"+url+":"+port+"/", http.StatusSeeOther)
}

// index handles requests to the chat application's index page
func index(w http.ResponseWriter, r *http.Request) {
	if checkUser(r) {
		if checkRoom(r) {
			page, _ := template.ParseFiles(path.Join("templates", "room.html"))
			page.Execute(w, "")
		} else {
			page, _ := template.ParseFiles(path.Join("templates", "enter_room.html"))
			page.Execute(w, "")
		}
	} else {
		resetCookie(w)
		page, _ := template.ParseFiles(path.Join("templates", "login_reg.html"))
		page.Execute(w, "")
	}
}

// reg handles requests to register a new user
func reg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		redirectToIndex(w, r)
		return
	}
	var name = r.FormValue("name")
	var login = r.FormValue("login")
	var password = r.FormValue("password")
	_, okName := names[name]
	_, okLogin := logins[login]
	if name == "" || login == "" || password == "" || okName || okLogin {
		redirectToIndex(w, r)
		return
	}
	users[name] = newUser(name, login, password)
	names[name] = true
	logins[login] = name
	var usr = users[name]
	var sid = newSid()
	usr.sidsBySid[sid.id] = sid.expiration
	usr.sidsByTime.insert(newItem(sid))
	users[name] = usr
	http.SetCookie(w, &http.Cookie{Name: "name", Value: name})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid.id})
	redirectToIndex(w, r)
}

// login handles requests to log in an existing user
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		redirectToIndex(w, r)
		return
	}
	var login = r.FormValue("login")
	var password = r.FormValue("password")
	_, ok := logins[login]
	if !ok {
		redirectToIndex(w, r)
		return
	}
	var name = logins[login]
	var usr = users[name]
	if usr.password != password {
		redirectToIndex(w, r)
		return
	}
	var sid = newSid()
	usr.sidsBySid[sid.id] = sid.expiration
	usr.sidsByTime.insert(newItem(sid))
	users[name] = usr
	http.SetCookie(w, &http.Cookie{Name: "name", Value: name})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid.id})
	redirectToIndex(w, r)
}

// logout handles requests to log out the current user
func logout(w http.ResponseWriter, r *http.Request) {
	resetCookie(w)
	redirectToIndex(w, r)
}

// enterRoom handles requests to enter a chat room
func enterRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var room = r.FormValue("room")
	http.SetCookie(w, &http.Cookie{Name: "room", Value: room})
	redirectToIndex(w, r)
}

// leaveRoom handles requests to leave the current chat room
func leaveRoom(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "room", Value: "", MaxAge: -1})
	redirectToIndex(w, r)
}

// post handles requests to post a message in the current chat room
func post(w http.ResponseWriter, r *http.Request) {
	var cookieName, _ = r.Cookie("name")
	var cookieRoom, _ = r.Cookie("room")
	if checkUser(r) && checkRoom(r) {
		data[cookieRoom.Value] += cookieName.Value + ": " + r.FormValue("text") + "<br>"
	}
	redirectToIndex(w, r)
}

// get handles requests to retrieve the chat messages for the current room
func get(w http.ResponseWriter, r *http.Request) {
	var cookie, err = r.Cookie("room")
	if err == nil {
		fmt.Fprint(w, data[cookie.Value])
	}
}

var data = make(map[string]string)
var users = make(map[string]User)

var logins = make(map[string]string)
var names = make(map[string]bool)

var url = "127.0.0.1"
var port = "8080"

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/reg", reg)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/enter", enterRoom)
	http.HandleFunc("/leave", leaveRoom)
	http.HandleFunc("/post", post)
	http.HandleFunc("/get", get)
	http.ListenAndServe(":"+port, nil)
}
