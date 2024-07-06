package server

import (
	"Slime/Server/database"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type UserAuth struct {
	connPool *database.ConnectionPool
}

func InitUserAuth() *UserAuth {
	return &UserAuth{connPool: database.GetConnectionPool()}
}

func (ua *UserAuth) UserLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		fmt.Println("User Login")
		var loginInfo map[string]string
		json.NewDecoder(r.Body).Decode(&loginInfo)
		fmt.Println("received login request:"+loginInfo["username"])
		
		conn := ua.connPool.GetConnection()
		defer ua.connPool.ReleaseConnection(conn)

		var username, userid string
		err := conn.QueryRow(`SELECT username, userid FROM users WHERE username=$1 AND passwordhash=$2 limit 1`, loginInfo["username"], loginInfo["passwordhash"]).Scan(&username, &userid)
		if err != nil {
			fmt.Println(`SELECT username, userid FROM users WHERE username=$1 AND passwordhash=$2 limit 1`, loginInfo["username"], loginInfo["passwordhash"])
			fmt.Println(err)
			json.NewEncoder(w).Encode(map[string]string{"error": "Error in login"})
			return	
		}
		http.SetCookie(w, &http.Cookie{
			Name: "userSession",
			Value : generateJwtToken(userid),
			Path: "/",
			HttpOnly: true,
		})
		json.NewEncoder(w).Encode(map[string]string{"message": "Login Successful"})
	}
}

func (ua *UserAuth) UserSignup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		conn := ua.connPool.GetConnection()
		defer ua.connPool.ReleaseConnection(conn)

		var signupInfo map[string]string
		json.NewDecoder(r.Body).Decode(&signupInfo)
		genUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(signupInfo["username"])).String()
		_, err := conn.Exec("INSERT INTO users (userid, username, passwordhash, name) VALUES ($1,$2,$3, $4)", genUUID, signupInfo["username"], signupInfo["passwordhash"],signupInfo["name"])
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": "Error in signup"})
			return	
		}
		http.SetCookie(w, &http.Cookie{
			Name: "userSession",
			Value : generateJwtToken(genUUID),
			Path: "/",
			HttpOnly: true,
		})
		json.NewEncoder(w).Encode(map[string]string{"message": "Signup Successful"})
	}
}

func (ua *UserAuth) UserIn() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request){
		_, err := r.Cookie("userSession")

		if err != nil {
			fmt.Println("User not logged in")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

}

	

