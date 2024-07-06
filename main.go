package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"Slime/Server/config"
	"Slime/Server/database"
	"Slime/Server/kvstore"
	"Slime/Server/models"
	"Slime/Server/server"
	"Slime/Server/sinks"

	"github.com/rohithputha/DepReq"
	"golang.org/x/crypto/acme/autocert"
)

var notesStore *database.Database
var notionSink sinks.NotionSink

func setupSubRoutines(config config.Config){
	notionSink = sinks.NotionSink{
		NotionSinkChan: make(chan models.SlimeNotionNote),
		ConnPool: database.GetConnectionPool(),
	}
	notesStore = &(database.Database{})
	notesStore.InitDatabase()
	fmt.Println("Server started at port 8080")
	
	go notesStore.InsertData()
	go notesStore.InsertData()
	go notionSink.PublishToSink()
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
	return config
}

func main() {
	depReqApi := DepReq.GetDepReqApi()
	config := setupConfig()
	depReqApi.Put("Slime/Server/config", config)
	setupSubRoutines(config)
	
	notionAuth := server.InitNotionAuth(config.Slime.NotionBase64Key,
									 kvstore.InitKVStore[string, string]())
	notionOps := server.InitNotionOps()
									
	userAuth := server.InitUserAuth()						

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/notion/note", notionOps.PostNoteHandlerFunc(notesStore.GetDataChan(), notionSink.NotionSinkChan)) //should this be notion specific end point..?
	mux.HandleFunc("GET /api/notion/public/pages/", notionOps.GetNotionPublicPagesHandlerFunc())


	mux.HandleFunc("GET /api/notion/auth/state", notionAuth.GetAuthState())
	mux.HandleFunc("GET /api/notion/auth/redirect/",notionAuth.AuthRedirect())
	mux.HandleFunc("GET /api/notion/auth/in",notionAuth.GetNotionIn())

	mux.HandleFunc("POST /api/user/login",userAuth.UserLogin())
	mux.HandleFunc("POST /api/user/signup",userAuth.UserSignup())
	mux.HandleFunc("GET /api/user/in", userAuth.UserIn())

	mux.HandleFunc("GET /api/heartbeat", server.GetHeartbeatHandlerFunc)

	if(config.Slime.Mode == "prod"){
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("narad.online"),
			Cache: 	autocert.DirCache("."),
		}
		server := &http.Server{
			Addr: ":https",
			Handler: mux,
			TLSConfig: certManager.TLSConfig(),
		}
	
		go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	
		fmt.Println("Server started on HTTPS port 443")
		server.ListenAndServeTLS("", "")
	} else {
		http.ListenAndServe(":8080", mux)
	}
	
}
