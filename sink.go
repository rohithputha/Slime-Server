package main

import (
	"Slime/Server/models"
	"bytes"
	"fmt"
	"net/http"
)

type Sink interface {
	publishToSink()
}

type NotionSink struct {
	notionSink chan models.SlimeNotionNote
}

func (ns *NotionSink) publishToSink() {
	for {
		d := <-ns.notionSink
		fmt.Println(d)
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

		fmt.Println("https://api.notion.com/v1/blocks/"+d.PageID+"/children")
		req, _:= http.NewRequest("PATCH","https://api.notion.com/v1/blocks/"+d.PageID+"/children",data)
		req.Header.Add("Authorization","Bearer secret_UtusC7jombTNwll5OFplbIpAawkE5Hma8sjhvRtWFb4")
		req.Header.Add("Content-Type","application/json")
		req.Header.Add("Notion-Version","2022-06-28")
		client := &http.Client{}
		resp, _ := client.Do(req)
		fmt.Println(resp)
		fmt.Println("Note saved into Notion")
	}
}
