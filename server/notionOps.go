package server
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"Slime/Server/models"
	
)


func PostNoteHandlerFunc(databaseChan chan models.DatabaseNote, notionSink chan models.SlimeNotionNote) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		var data models.SlimeNotionNote
		err := json.NewDecoder(r.Body).Decode(&data)

		if err != nil {
			http.Error(w, err.Error(),http.StatusInternalServerError)	
			return	
		}
		fmt.Println(data)
		fmt.Println("Note saved")
		databaseChan <- models.DatabaseNote{Sink: data.SinkType, Note: data}
		if data.SinkType == "notion" || data.SinkType == "" {
			notionSink <- data
		}	
	})
}

func GetNotionPublicPagesHandlerFunc() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		fmt.Println("Get Notion Public Pages")
		userID := r.PathValue("userId")
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

		req.Header.Add("Authorization","Bearer secret_UtusC7jombTNwll5OFplbIpAawkE5Hma8sjhvRtWFb4")
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


