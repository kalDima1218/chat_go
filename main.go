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

type Item struct {
	k int
	sid Sid
}

func newItem(x Sid) Item{
	var tmp Item
	tmp.k = int(x.expiration.Unix())
	tmp.sid = x
	return tmp
}

type Node struct{
	i Item
	p int
	l, r, parent *Node
}

func newNode(x Item) *Node {
	var tmp Node
	tmp.p = rand.Int()
	tmp.i = x
	return &tmp
}

func _merge(l *Node, r *Node) *Node {
	if l == nil{
		return r
	}
	if r == nil{
		return l
	}
	if l.p > r.p{
		r.parent = l
		l.r = _merge(l.r, r)
		return l
	}else{
		l.parent = r
		r.l = _merge(l, r.l)
		return r
	}
}

func _split(p *Node, x int) (*Node, *Node){
	if p == nil{
		return nil, nil
	}
	if p.i.k <= x{
		var l, r = _split(p.r, x)
		if l != nil{
			l.parent = p
		}
		p.r = l
		return p, r
	}else{
		var l, r = _split(p.l, x)
		if r != nil{
			r.parent = p
		}
		p.l = r
		return l, p
	}
}

func _print(p *Node){
	if p.l != nil{
		_print(p.l)
	}
	fmt.Println(p.i.k)
	if p.r != nil{
		_print(p.r)
	}
}

type Treap struct{
	_root, _begin, _end *Node
}

func (t *Treap)_updBegin(){
	var p = t._root
	for p.l != nil{
		p = p.l
	}
	t._begin = p
}

func (t *Treap)_updEnd(){
	var p = t._root
	for p.r != nil{
		p = p.r
	}
	t._end = p
}

func (t *Treap)begin() *Node{
	return t._begin
}

func (t *Treap)end() *Node{
	return t._end
}

func (t *Treap)count(x Item) int{
	var p = t._root
	for p.i != x{
		if p.r != nil && p.i.k < x.k{
			p = p.r
			continue
		}
		if p.l != nil && x.k < p.i.k{
			p = p.l
			continue
		}
		break
	}
	if p.i == x{
		return 1
	}else{
		return 0
	}
}

func (t *Treap)find(x Item) (*Node, bool){
	var p = t._root
	for p.i != x{
		if p.r != nil && p.i.k < x.k{
			p = p.r
			continue
		}
		if p.l != nil && x.k < p.i.k{
			p = p.l
			continue
		}
		break
	}
	return p, p.i == x
}

func (t *Treap)insert(x Item){
	if t._root != nil && t.count(x) != 0{
		return
	}
	var l, r = _split(t._root, x.k)
	t._root = _merge(l, _merge(newNode(x), r))
	t._updBegin()
	t._updEnd()
}

func (t *Treap)erase(x Item){
	if t._root == nil || t.count(x) == 0{
		return
	}
	var l, r = _split(t._root, x.k)
	l, _ = _split(l, x.k-1)
	t._root = _merge(l, r)
	t._updBegin()
	t._updEnd()
}

func (t *Treap)print(){
	_print(t._root)
}



type Sid struct{
	id string
	expiration time.Time
}

type User struct{
	name string
	login, password string
	sidsByTime Treap
	sidsBySid map[string]time.Time
}

func newUser(name string, login string, password string) User {
	var tmp User
	tmp.name = name
	tmp.login = login
	tmp.password = password
	tmp.sidsBySid = make(map[string]time.Time)
	return tmp
}

func getMessagesUpdate(r []string, from int) string{
	var res = ""
	for i := from; i < len(r); i++{
		res+=r[i]
	}
	return res
}

func resetCookie(w http.ResponseWriter){
	http.SetCookie(w, &http.Cookie{Name: "name", Value: "", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "room", Value: "", MaxAge: -1})
}

func checkSids(usr string){
	user, ok := users[usr]
	if ok{
		for len(user.sidsBySid) > 0 && user.sidsByTime.begin().i.sid.expiration.Before(time.Now()){
			delete(user.sidsBySid, user.sidsByTime.begin().i.sid.id)
			user.sidsByTime.erase(user.sidsByTime.begin().i)
		}
	}
}

func checkUser(r *http.Request) bool{
	usr, errName := r.Cookie("name")
	sid, errSid := r.Cookie("sid")
	if errName != nil || errSid != nil{
		return false
	}
	checkSids(usr.Value)
	sidExpiration, okUserSid := users[usr.Value].sidsBySid[sid.Value]
	if errName != nil || errSid != nil || !okUserSid || sidExpiration.Before(time.Now()){
		return false
	}else{
		_, ok := users[usr.Value]
		return ok
	}
}

func newSid()Sid{
	var sid Sid
	sid.expiration = time.Now().Add(24 * time.Hour)
	sid.id = strconv.Itoa(rand.Int())
	return sid
}

func checkRoom(r *http.Request) bool{
	if _, err := r.Cookie("room"); err == nil{
		return true
	}else{
		return false
	}
}

func redirectToIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://" + url + ":" + port + "/", http.StatusSeeOther)
}

func index(w http.ResponseWriter, r *http.Request) {
	if checkUser(r){
		if checkRoom(r){
			page, _ := template.ParseFiles(path.Join("templates", "room.html"))
			page.Execute(w, "")
		}else{
			page, _ := template.ParseFiles(path.Join("templates", "enter_room.html"))
			page.Execute(w, "")
		}
	}else{
		resetCookie(w)
		page, _ := template.ParseFiles(path.Join("templates", "login_reg.html"))
		page.Execute(w, "")
	}
}

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
	if name == "" || login == "" || password == "" || okName || okLogin{
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

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		redirectToIndex(w, r)
		return
	}
	var login = r.FormValue("login")
	var password = r.FormValue("password")
	_, ok := logins[login]
	if !ok{
		redirectToIndex(w, r)
		return
	}
	var name = logins[login]
	var usr = users[name]
	if usr.password != password{
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
func logout(w http.ResponseWriter, r *http.Request) {
	resetCookie(w)
	redirectToIndex(w, r)
}

func enterRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var room = r.FormValue("room")
	http.SetCookie(w, &http.Cookie{Name: "room", Value: room})
	redirectToIndex(w, r)
}

func leaveRoom(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "room", Value: "", MaxAge: -1})
	redirectToIndex(w, r)
}

func post(w http.ResponseWriter, r *http.Request){
	var cookieName, _ = r.Cookie("name")
	var cookieRoom, _ = r.Cookie("room")
	if checkUser(r) && checkRoom(r){
		data[cookieRoom.Value]+=cookieName.Value + ": " + r.FormValue("text")+"<br>"
	}
	redirectToIndex(w, r)
}

func get(w http.ResponseWriter, r *http.Request){
	var cookie, err = r.Cookie("room")
	if err == nil{
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
	http.ListenAndServe(":" + port, nil)
}
