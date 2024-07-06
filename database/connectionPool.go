package database

import (
	"database/sql"
	"fmt"
	"sync"

	"Slime/Server/config"

	"github.com/rohithputha/DepReq"
)

var instance *ConnectionPool
var once sync.Once

type ConnectionPool struct {
	conn chan *sql.DB
}

func GetConnectionPool() *ConnectionPool{
	once.Do(func(){
		depReqApi := DepReq.GetDepReqApi()
		c, err := depReqApi.Get("Slime/Server/config")
		if err != nil {
			panic(err)
		}
		config := c.(config.Config)
		instance = &ConnectionPool{}
		instance.conn = make(chan *sql.DB,config.Database.ConnectionPoolSize)
		for i:=0; i<config.Database.ConnectionPoolSize; i++ {
			conn, err := sql.Open("postgres", "host="+config.Database.Host+"  user="+config.Database.User+" password="+config.Database.Password+" dbname="+config.Database.Dbname+" sslmode=disable")
			if err != nil {
				panic(err)
			}
		instance.conn <- conn	
		}
	})
	return instance
}

func (cp *ConnectionPool) GetConnection() *sql.DB {
	return <-cp.conn
}

func (cp *ConnectionPool) ReleaseConnection(conn *sql.DB) {
	cp.conn <- conn
}

func (cp *ConnectionPool) CloseConnectionPool() {
	depReqApi := DepReq.GetDepReqApi()
	c, err := depReqApi.Get("Slime/Server/config")
		if err != nil {
			panic(err)
		}
	config := c.(config.Config)

	for i:=0; i<config.Database.ConnectionPoolSize; i++ {
		fmt.Println("Closing connection :"+fmt.Sprintf("%d",i))
		conn := <-cp.conn
		conn.Close()
	}
}
	