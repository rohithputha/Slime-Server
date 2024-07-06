package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"Slime/Server/config"
	"Slime/Server/database"
	"Slime/Server/htmltemplate"
	"Slime/Server/kvstore"

	"github.com/golang-jwt/jwt"

	"github.com/rohithputha/DepReq"
)


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
		w.Header().Set("Content-Type", "text/html")
		depReqApi := DepReq.GetDepReqApi()
		c, err:=depReqApi.Get("Slime/Server/config")
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		config := c.(config.Config)

		queryParams := r.URL.Query()
		state:= queryParams.Get("state")
		code := queryParams.Get("code")
		e := queryParams.Get("error")
		na.stateKvStore.Set(state,"InProgress")
		defer na.stateKvStore.Delete(state)
		defer http.SetCookie(w, &http.Cookie{
			Name: "state",
			Value: "",
			Expires: time.Unix(0,0),
			Path: "/",
		})


		if e !="" {
			na.stateKvStore.Delete(state)
			http.Error(w,e,http.StatusInternalServerError)
			return
		}
		jwtDecodedTk, _ := jwt.Parse(state, func(token *jwt.Token) (interface{}, error) {
			return []byte("hvyam319"), nil
		})
		userid := jwtDecodedTk.Claims.(jwt.MapClaims)["user"]
		
		authPayload := map[string]string{
			"code":code,
			"grant_type":"authorization_code",
			"redirect_uri": config.Slime.NotionRedirectUrl,
		}
		jsonBytes, _ := json.Marshal(authPayload)
		req, _ :=http.NewRequest("POST","https://api.notion.com/v1/oauth/token",bytes.NewBuffer(jsonBytes))
		req.Header.Add("Authorization", `Basic "`+config.Slime.NotionBase64Key+`"`)
		req.Header.Add("Content-Type","application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			
			fmt.Fprint(w,htmltemplate.GetSimpleHtmlMessagePage("Authentication Failed", "Authentication Failed. You can close this window.",err))
			return 
		}


		jsonDecoder := json.NewDecoder(resp.Body)
		var authResp map[string]string
		jsonDecoder.Decode(&authResp)
		if authResp["error"] != "" {
			fmt.Println(authResp["error"])
			fmt.Fprint(w, htmltemplate.GetSimpleHtmlMessagePage("Authentication Failed", "Authentication Failed. You can close this window.",nil))
			return 
		}

		dbConn := na.connPool.GetConnection()
		defer na.connPool.ReleaseConnection(dbConn)
		
		_, ierr := dbConn.Exec("INSERT INTO  notionaccess (userid,accesstk) VALUES ($1,$2) ON CONFLICT (userid) DO NOTHING" ,userid,authResp["access_token"])
		if (ierr != nil) {
			fmt.Fprint(w,htmltemplate.GetSimpleHtmlMessagePage("Authentication Failed", "Authentication Failed. You can close this window.",ierr))
			return
		}

		fmt.Fprint(w, htmltemplate.GetSimpleHtmlMessagePage("Authentication Success", "Authentication Success. You can close this window.", nil))
	}
}

func(na *NotionAuth) GetAuthState() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		userSessionCookie, err := r.Cookie("userSession")
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		userSessionToken ,err :=jwt.Parse(userSessionCookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte("hvyam319"), nil
		})
		userID:= userSessionToken.Claims.(jwt.MapClaims)["user"].(string)
		jwt := generateJwtToken(userID)
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
		json.NewEncoder(w).Encode(map[string]string{"notion-auth-state-set":jwt})	
	}
}

func (na *NotionAuth) GetNotionIn() http.HandlerFunc{
	
	return func (w http.ResponseWriter, r *http.Request){
		stateCookie, err := r.Cookie("state")
		if err==nil{
			fmt.Println("State cookie found")
			fmt.Println(stateCookie.Value)
			fmt.Println(na.stateKvStore.Get(stateCookie.Value))
			if na.stateKvStore.Get(stateCookie.Value) == "InProgress" {
				fmt.Println("sending processing")
				w.WriteHeader(http.StatusCreated)  	
				return
			}
		}


		userSession, err := r.Cookie("userSession")
		if err != nil{
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		jwtDecodedTk, _ := jwt.Parse(userSession.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte("hvyam319"), nil
			})
		userid := jwtDecodedTk.Claims.(jwt.MapClaims)["user"]
		
		dbConn := na.connPool.GetConnection()
		defer na.connPool.ReleaseConnection(dbConn)

		var accesstoken string
		err = dbConn.QueryRow("SELECT accesstk FROM notionaccess WHERE userid=$1",userid).Scan(&accesstoken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized) 
			return
		}
		
		w.WriteHeader(http.StatusOK) 
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



