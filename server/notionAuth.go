package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"Slime/Server/kvstore"
	"Slime/Server/database"
)

type OAuth interface{

}

type NotionAuth struct {
	clientIdSecretEncode string
	stateKvStore *kvstore.KVStore[string,string]
	connPool *database.ConnectionPool
}

func InitNotionAuth(clientIdSecretEncode string, stateKvStore *kvstore.KVStore[string,string]) *NotionAuth{
	return &NotionAuth{
		clientIdSecretEncode: clientIdSecretEncode,
		stateKvStore: stateKvStore,
		connPool: database.GetConnectionPool(),
	}
}


func (na *NotionAuth) AuthRedirect() http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request){
		queryParams := r.URL.Query()
		state:= queryParams.Get("state")
		code := queryParams.Get("code")
		na.stateKvStore.Set(state,"InProgress")

		jwtDecodedTk, _ := jwt.Parse(state, func(token *jwt.Token) (interface{}, error) {
			return []byte("hvyam319"), nil
		})
		userid := jwtDecodedTk.Claims.(jwt.MapClaims)["user"]
		
		authPayload := map[string]string{
			"code":code,
			"grant_type":"authorization_code",
			"redirect_uri":"http://localhost:8080/api/notion/auth/redirect/",
		}
		jsonBytes, _ := json.Marshal(authPayload)
		req, _ :=http.NewRequest("POST","https://api.notion.com/v1/oauth/token",bytes.NewBuffer(jsonBytes))
		req.Header.Add("Authorization", `Basic "`+na.clientIdSecretEncode+`"`)
		req.Header.Add("Content-Type","application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Authentication Failed</title>
			</head>
			<body> <p> Authentication Failed. You can close this window. <p> </body>
			</html>
			`)
			return 
		}


		jsonDecoder := json.NewDecoder(resp.Body)
		var authResp map[string]string
		jsonDecoder.Decode(&authResp)
		if authResp["error"] != "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Authentication Failed</title>
			</head>
			<body> <p> Authentication Failed. You can close this window. <p> </body>
			</html>
			`)
			return 
		}

		dbConn := na.connPool.GetConnection()
		_, ierr := dbConn.Exec("INSERT INTO  notionaccess (userid,accesstk) VALUES ($1,$2) ON CONFLICT (userid) DO NOTHING" ,userid,authResp["access_token"])
		if (ierr != nil) {
			http.Error(w,ierr.Error(),http.StatusInternalServerError)
			return
		}
		na.stateKvStore.Delete(state)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Authentication Complete</title>
		</head>
		<body> <p> Authentication Complete. You can close this window. <p> </body>
		</html>
		`)
	}
}

func(na *NotionAuth) GetAuthState() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		var user map[string]string
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		jwt := generateJwtToken(user["user"])
		cookie:= &http.Cookie{
			Name: "state",
			Value: jwt,
			MaxAge: 86400,
			HttpOnly: true,
			Secure: false,
			Path: "/",
		}
		na.stateKvStore.Set(jwt,"InProgress")
		http.SetCookie(w,cookie)
		w.Header().Set("Content-Type","application/json")
		json.NewEncoder(w).Encode(map[string]string{"notion-auth-state-set":"success"})	
	}
}

func (na *NotionAuth) GetAuthStatus() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		stateToken, err := r.Cookie("state")
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type","application/json")
		if progress:= na.stateKvStore.Get(stateToken.Value); progress == "" {
			json.NewEncoder(w).Encode(map[string]string{"notion-auth-status":"Success"})
		}else{
			json.NewEncoder(w).Encode(map[string]string{"notion-auth-status":progress})
		}
		
	}
}


func generateJwtToken(user string) string{
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"user": user,
	})
	tokenString, _ := token.SignedString([]byte("hvyam319"))
	return tokenString
}

func GetHeartbeatHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
