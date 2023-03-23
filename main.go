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

type user struct{
	name string
	login, password string
	sid string
	expiration time.Time
}

func newUser(name string, login string, password string) user{
	var tmp user
	tmp.name = name
	tmp.login = login
	tmp.password = password
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

func checkUser(r *http.Request) bool{
	usr, errName := r.Cookie("name")
	sid, errSid := r.Cookie("sid")
	if errName != nil || errSid != nil || users[usr.Value].expiration.Before(time.Now()) || users[usr.Value].sid != sid.Value{
		return false
	}else{
		_, ok := users[usr.Value]
		return ok
	}
}

func checkRoom(r *http.Request) bool{
	if _, err := r.Cookie("room"); err == nil{
		return true
	}else{
		return false
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if checkUser(r){
		if checkRoom(r){
			page, _ := template.ParseFiles(path.Join("templates", "room.html"))
			page.Execute(w, "")
		}else{
			page, _ := template.ParseFiles(path.Join("templates", "index.html"))
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
		http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
		return
	}
	var name = r.FormValue("name")
	var login = r.FormValue("login")
	var password = r.FormValue("password")
	_, okName := names[name]
	_, okLogin := logins[login]
	if name == "" || login == "" || password == "" || okName || okLogin{
		http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
		return
	}
	users[name] = newUser(name, login, password)
	names[name] = true
	logins[login] = name
	var usr = users[name]
	usr.sid = strconv.Itoa(rand.Int())
	usr.expiration = time.Now().Add(24 * time.Hour)
	users[name] = usr
	http.SetCookie(w, &http.Cookie{Name: "name", Value: name})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: users[name].sid})
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
		return
	}
	var login = r.FormValue("login")
	var password = r.FormValue("password")
	_, ok := logins[login]
	if !ok{
		http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
		return
	}
	var name = logins[login]
	var usr = users[name]
	if usr.password != password{
		http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
		return
	}
	usr.sid = strconv.Itoa(rand.Int())
	usr.expiration = time.Now().Add(24 * time.Hour)
	users[name] = usr
	http.SetCookie(w, &http.Cookie{Name: "name", Value: name})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: users[name].sid})
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}
func logout(w http.ResponseWriter, r *http.Request) {
	resetCookie(w)
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func enterRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var room = r.FormValue("room")
	http.SetCookie(w, &http.Cookie{Name: "room", Value: room})
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func leaveRoom(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "room", Value: "", MaxAge: -1})
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func post(w http.ResponseWriter, r *http.Request){
	var cookieName, _ = r.Cookie("name")
	var cookieRoom, _ = r.Cookie("room")
	if checkUser(r) && checkRoom(r){
		data[cookieRoom.Value]+=cookieName.Value + ": " + r.FormValue("text")+"<br>"
	}
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func get(w http.ResponseWriter, r *http.Request){
	var cookie, err = r.Cookie("room")
	if err == nil{
		fmt.Fprint(w, data[cookie.Value])
	}
}

func getRoomsList(w http.ResponseWriter, r *http.Request){

}

var data = make(map[string]string)
var users = make(map[string]user)

var logins = make(map[string]string)
var names = make(map[string]bool)

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/reg", reg)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/enter", enterRoom)
	http.HandleFunc("/leave", leaveRoom)
	http.HandleFunc("/post", post)
	http.HandleFunc("/get", get)
	http.ListenAndServe(":8080", nil)
}
