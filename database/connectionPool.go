package database

import (
	"database/sql"
	"sync"
)

var instance *ConnectionPool
var once sync.Once

type ConnectionPool struct {
	conn chan *sql.DB
}

func GetConnectionPool() *ConnectionPool{
	once.Do(func(){
		instance = &ConnectionPool{}
		instance.conn = make(chan *sql.DB,3)
		for i:=0; i<3; i++ {
			conn, err := sql.Open("postgres","user=postgres password=hvyam319 dbname=postgres sslmode=disable")
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
	for i:=0; i<3; i++ {
		conn := <-cp.conn
		conn.Close()
	}
}
	