package config

type Config struct {
	Database DatabaseConfig `json:"databaseConfig"`
	Slime    SlimeConfig `json:"slimeConfig"`
}

type DatabaseConfig struct {
	ConnectionPoolSize int `json:"connectionPoolSize"`
	User 			string `json:"user"`
	Password 		string `json:"password"`
	Dbname 			string `json:"dbname"`
	Sslmode 		string `json:"sslmode"`
	Host 			string `json:"host"`
}

type SlimeConfig struct {
	NotionBase64Key string `json:"notionBase64Key"`
	NotionRedirectUrl string `json:"notionRedirectUrl"`
	Mode string `json:"mode"`
}

