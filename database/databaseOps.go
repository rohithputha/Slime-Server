package database

import (
	"fmt"
	"Slime/Server/models"

	_ "github.com/lib/pq"
)

type Database struct {
	conPool *ConnectionPool
	dataChan chan models.DatabaseNote

}

func (dbp *Database) InitDatabase(){
	dbp.conPool = GetConnectionPool()
	dbp.dataChan = make(chan models.DatabaseNote)
}

func (dbp *Database) GetDataChan() chan models.DatabaseNote {
	return dbp.dataChan
}

func (dbp *Database) InsertData(testNumber int) {
	for {
		data := <-dbp.dataChan
		conn := dbp.conPool.GetConnection()
		
		if (data.Sink == "notion" || data.Sink == "") {
			notionNote, _:= data.Note.(models.SlimeNotionNote)
			_, err := conn.Exec("insert into notes (userid,sink,note) values ($1,$2,$3)",notionNote.User,"notion",notionNote.Note)
			if err != nil {
				fmt.Println(err)
			}
		}
		dbp.conPool.ReleaseConnection(conn)
	}
}	