package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
)

type room struct{
	messages []string
	
}

type user struct{
	name string
	login, password string
	sid, expiration int
}

func (r *room) getMessagesUpdate(from int) string{
	var res = ""
	for i := from; i < len(r.messages); i++{
		res+=r.messages[i]
	}
	return res
}

func index(w http.ResponseWriter, r *http.Request) {
	var _, err = r.Cookie("name")
	if err == nil{
		_, err = r.Cookie("room")
		if err == nil{
			page, _ := template.ParseFiles(path.Join("templates", "room.html"))
			page.Execute(w, "")
		}else{
			page, _ := template.ParseFiles(path.Join("templates", "index.html"))
			page.Execute(w, "")
		}
	}else{
		page, _ := template.ParseFiles(path.Join("templates", "login.html"))
		page.Execute(w, "")
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	var name = r.FormValue("name")
	http.SetCookie(w, &http.Cookie{Name: "name", Value: name})
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "name", Value: "", MaxAge: -1})
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
	var cookieName, errName = r.Cookie("name")
	var cookieRoom, errRoom = r.Cookie("room")
	if errName == nil && errRoom == nil{
		data[cookieRoom.Value]+=cookieName.Value + ": " + r.FormValue("text")+"<br>"
	}
	http.Redirect(w, r, "http://127.0.0.1:8080/", http.StatusSeeOther)
}

func get(w http.ResponseWriter, r *http.Request){
	//var _, err_name = r.Cookie("name")
	var cookieRoom, errRoom = r.Cookie("room")
	if errRoom == nil{
		fmt.Fprint(w, data[cookieRoom.Value])
	}
}

func getRoomsList(w http.ResponseWriter, r *http.Request){

}

var data = make(map[string]string)

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/enter", enterRoom)
	http.HandleFunc("/leave", leaveRoom)
	http.HandleFunc("/post", post)
	http.HandleFunc("/get", get)
	http.ListenAndServe(":8080", nil)
}