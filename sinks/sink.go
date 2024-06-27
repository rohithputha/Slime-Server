package sinks

import (
	"Slime/Server/database"
	"Slime/Server/models"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type Sink interface {
	PublishToSink()
}

type NotionSink struct {
	NotionSinkChan chan models.SlimeNotionNote
	ConnPool *database.ConnectionPool
}

func (ns *NotionSink) PublishToSink() {
	for {
		d := <-ns.NotionSinkChan
		fmt.Println(d)
		d.Note = strings.ReplaceAll(d.Note, `"`, `\"`)
		data := bytes.NewBuffer([]byte(`{
			"children":[
				{
					"object":"block",
					"type":"paragraph",
					"paragraph":{
						"rich_text":[	
							{
								"type":"text",	
								"text":{
									"content":"`+d.Note+`",
									"link":null
								}
							}
						]
					}
				}
			]
		}`))
        conn:= ns.ConnPool.GetConnection()
		var notionAccessToken string
		err := conn.QueryRow(`SELECT accesstk FROM notionaccess WHERE userid=$1`,d.User).Scan(&notionAccessToken)
		if err != nil {
			fmt.Println("Error in getting notion access token")
			return
		}


		fmt.Println("https://api.notion.com/v1/blocks/"+d.PageID+"/children")
		req, _:= http.NewRequest("PATCH","https://api.notion.com/v1/blocks/"+d.PageID+"/children",data)
		req.Header.Add("Authorization","Bearer "+notionAccessToken)
		req.Header.Add("Content-Type","application/json")
		req.Header.Add("Notion-Version","2022-06-28")
		client := &http.Client{}
		resp, _ := client.Do(req)
		fmt.Println(resp)
		fmt.Println("Note saved into Notion")
	}
}
