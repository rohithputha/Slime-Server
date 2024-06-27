package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"Slime/Server/database"
	"Slime/Server/models"

	"github.com/golang-jwt/jwt"
)

type NotionOps struct {
	dbConnectionPool database.ConnectionPool
}

func InitNotionOps() *NotionOps {
	return &NotionOps{dbConnectionPool: *database.GetConnectionPool()}
}


func(no *NotionOps) PostNoteHandlerFunc(databaseChan chan models.DatabaseNote, notionSink chan models.SlimeNotionNote) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		var data models.SlimeNotionNote
		err := json.NewDecoder(r.Body).Decode(&data)
       
		if err != nil {
			http.Error(w, err.Error(),http.StatusInternalServerError)	
			return	
		}
		userSessionCookie, err := r.Cookie("userSession")
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		userSessionToken ,err :=jwt.Parse(userSessionCookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte("hvyam319"), nil
		})
		userID:= userSessionToken.Claims.(jwt.MapClaims)["user"].(string)
        data.User = userID

		fmt.Println(data)
		fmt.Println("Note saved")
		databaseChan <- models.DatabaseNote{Sink: data.SinkType, Note: data}
		if data.SinkType == "notion" || data.SinkType == "" {
			notionSink <- data
		}	
	})
}

func(no *NotionOps) GetNotionPublicPagesHandlerFunc() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		fmt.Println("Get Notion Public Pages")
		userSessionCookie, err := r.Cookie("userSession")
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		userSessionToken ,err :=jwt.Parse(userSessionCookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte("hvyam319"), nil
		})
		userID:= userSessionToken.Claims.(jwt.MapClaims)["user"].(string)
		var notionAccessToken string
        
		conn := no.dbConnectionPool.GetConnection()
		defer no.dbConnectionPool.ReleaseConnection(conn)

		err = conn.QueryRow(`SELECT accesstk FROM notionaccess WHERE userid=$1`,userID).Scan(&notionAccessToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		client := &http.Client{}
		data := bytes.NewBuffer([]byte(`{
			"query":"",
			"filter": {
      				"value": "page",
      				"property": "object"
    		},
			"sort":{
				"direction":"ascending",
				"timestamp":"last_edited_time"
			}		
    	}`))
		req, qerr := http.NewRequest("POST","https://api.notion.com/v1/search/",data)
		if qerr != nil {
			http.Error(w,qerr.Error(),http.StatusInternalServerError)
			return		
		}
		
		req.Header.Add("Authorization","Bearer "+notionAccessToken)
		req.Header.Add("Content-Type","application/json")
		req.Header.Add("Notion-Version","2022-06-28")
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}

		jsonDecoder := json.NewDecoder(resp.Body)
		var notionPageResp models.NotionPageSearchResponse
		jsonDecoder.Decode(&notionPageResp)

		slimePages := make([]models.SlimeNotionPage,0)
		for _, page := range notionPageResp.Results {
			if len(page.Properties.Title.Title) == 0{
				continue
			}
			slimePage := models.SlimeNotionPage{
				ID: page.ID,
				Title: page.Properties.Title.Title[0].Text.Content,
				User: userID,
			}
			slimePages = append(slimePages,slimePage)
		}
		json.NewEncoder(w).Encode(slimePages)
		defer resp.Body.Close()
	}
}

/*
goroutines 
1. Insert into notion
2. Insert into database

*/



