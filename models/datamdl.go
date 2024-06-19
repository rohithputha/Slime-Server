package models

type Page struct {
    ID    string `json:"id"`
	Object string `json:"object"`
	Properties struct{
		Title struct{
			Title []struct{
				Text struct{
					Content string `json:"content"`
				} `json:"text"`
			} `json:"title"`
		} `json:"title"`
	} `json:"properties"`
}

type NotionPageSearchResponse struct {
    Results []Page `json:"results"`
}

type SlimeNotionPage struct{
	ID string
	Title string
	User string
}

type SlimeNotionNote struct{
	User string
	Note string
	PageID string
	SinkType string
}

type DatabaseNote struct{
	Sink string
	Note interface{}
}

