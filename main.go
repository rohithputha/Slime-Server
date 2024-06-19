package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"encoding/json"

	"Slime/Server/database"
	"Slime/Server/kvstore"
	"Slime/Server/models"
	"Slime/Server/server"
	"Slime/Server/config"
)

var notesStore *database.Database
var notionSink NotionSink

func setupSubRoutines(config config.Config){
	notionSink = NotionSink{
		notionSink: make(chan models.SlimeNotionNote),
	}
	
	notesStore = &(database.Database{})
	notesStore.InitDatabase()
	fmt.Println("Server started at port 8080")
	
	go notesStore.InsertData(1)
	go notesStore.InsertData(2)
	go notionSink.publishToSink()
}

func setupConfig() config.Config{
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file")
		panic(err)
	}

	defer file.Close()
	bytes, _ := io.ReadAll(file)
	var config config.Config
	json.Unmarshal(bytes, &config)
	fmt.Println(config.Database.User)
	return config
}

func main() {
	config := setupConfig()
	setupSubRoutines(config)
	
	notionAuth := server.InitNotionAuth(config.Slime.NotionBase64Key,
									 kvstore.InitKVStore[string, string]())

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/note", server.PostNoteHandlerFunc(notesStore.GetDataChan(), notionSink.notionSink)) //should this be notion specific end point..?
	mux.HandleFunc("GET /api/notion/public/pages/{userId}", server.GetNotionPublicPagesHandlerFunc())


	mux.HandleFunc("GET /api/notion/auth/state", notionAuth.GetAuthState())
	mux.HandleFunc("GET /api/notion/auth/redirect/",notionAuth.AuthRedirect())
	mux.HandleFunc("GET /api/notion/auth/status",notionAuth.GetAuthStatus())

	mux.HandleFunc("GET /api/heartbeat", server.GetHeartbeatHandlerFunc)
	fmt.Println("Server started at port 8080")
	http.ListenAndServe(":8080", mux)
}
