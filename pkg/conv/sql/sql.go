package sql

import (
	sql "database/sql"
	"fmt"
	"log"
	"main/pkg/conv/subscriber"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Sql struct {
	//Empty for localhost
	Host     string
	User     string
	Password string
	DBName   string
}

func New(host string, user string, password string, dbName string) *Sql {
	return &Sql{
		Host:     host,
		User:     user,
		Password: password,
		DBName:   dbName,
	}
}

// Opens sql connection
//
// Dont forget to use db.Close() [defer db.Close()]
func (s *Sql) open() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", s.User, s.Password, s.Host, s.DBName))
	if err != nil {
		log.Fatalf("Cannot open db connection. %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot ping to db. %v", err)
	}
	return db
}

func (s *Sql) AddImportantEvent(datetime int64, name string, description string, one_time bool, notify_before_days int, notify_time time.Time) error {
	panic("implement me")
}

func (s *Sql) GetClientResponses(telegram_id int64) ([]string, error) {

	db := s.open()
	defer db.Close()

	rows, err := db.Query("SELECT resp_id, response_text FROM clients_responses WHERE client_telegram_id = ?", telegram_id)
	if err != nil {
		log.Printf("No responses for client with id %d. Error: %v", telegram_id, err)
		return nil, err
	}

	type Resp struct {
		resp_id int

		resp_text string
	}

	responses := make([]Resp, 0)

	for {
		if !rows.Next() {
			break
		}
		resp := Resp{}
		if err = rows.Scan(&resp.resp_id, &resp.resp_text); err != nil {
			log.Printf("Cannot receive data about client's responses. Error: %v", err)
			continue
		}
		responses = append(responses, resp)
	}

	// TODO:

	sort.Slice(responses, func(i, j int) bool {
		return responses[i].resp_id < responses[j].resp_id
	})

	return func() []string {
		r := make([]string, len(responses))
		for idx, v := range responses {
			r[idx] = v.resp_text
		}
		return r
	}(), nil
}

func (s *Sql) AddClientResponse(telegram_id int64, response_text string) error {
	db := s.open()
	defer db.Close()

	stmtInsert, err := db.Prepare("INSERT INTO clients_responses VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("Cannot add client's response to db. tid: %d, resp_text: %s. Err: %v", telegram_id, response_text, err)
		return err
	}

	defer stmtInsert.Close()

	_, err = stmtInsert.Exec(0, telegram_id, response_text)
	if err != nil {
		log.Printf("Cannot add client's response to db. tid: %d, resp_text: %s. Err: %v", telegram_id, response_text, err)
		return err
	}
	return nil
}

func (s *Sql) GetClient(telegram_id int64) (*subscriber.Client, error) {
	db := s.open()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM clients WHERE telegram_id = ?", telegram_id)
	client := subscriber.Client{}
	if err := row.Scan(&client.ChatId); err != nil {
		log.Printf("In db no client with id: %d.\nError: %v", telegram_id, err)
		return nil, err
	}

	if responses, err := s.GetClientResponses(telegram_id); err == nil {
		client.Responses = responses
	}
	return &client, nil
}

func (s *Sql) AddClient(telegram_id int64, full_name string, domain *string, commandHandler *string) error {
	db := s.open()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO clients VALUES (?,?,?,?)")
	if err != nil {
		log.Printf("Cannot add client with id: %d, name: %s, domain: %s.\nErr: %v",
			telegram_id, full_name, *domain, err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(telegram_id, full_name, *domain, *commandHandler)
	if err != nil {
		log.Printf("Cannot add client with id: %d, name: %s, domain: %s.\nErr: %v",
			telegram_id, full_name, *domain, err)
		return err
	}

	log.Printf("Client with id %d, successfully added [Fullname: %s, Domain: %s]",
		telegram_id, full_name, *domain)
	return nil
}

func (s *Sql) UpdateClientHandler(telegram_id int64, handlerCommand string) error {
	db := s.open()

	updStmt, err := db.Prepare("UPDATE clients SET current_handler = ? WHERE telegram_id = ?")
	if err != nil {
		log.Printf("Cannot update client's current handler. telegram_id: %d, handler: %s.\nError: %v", telegram_id, handlerCommand, err)
		return err
	}

	defer updStmt.Close()
	_, err = updStmt.Exec(handlerCommand, telegram_id)

	if err != nil {
		log.Printf("Error on execution update client handler. Error: %v", err)
		return err
	}

	return nil
}

func (s *Sql) DeleteClientResponses(telegram_id int64) error {
	db := s.open()
	deleteStmt, err := db.Prepare("DELETE FROM clients_responses WHERE client_telegram_id = ?")
	if err != nil {
		log.Printf("Cannot prepare query for delete all client's responses for client with telegram_id: %d. Error: %v", telegram_id, err)
		return err
	}

	defer deleteStmt.Close()
	_, err = deleteStmt.Exec(telegram_id)
	if err != nil {
		log.Printf("Cannot execute delete all client's responses for client with telegram_id: %d. Error: %v", telegram_id, err)
		return err
	}
	return nil
}

//
//	INITIALIZE DATABASE START

func Init(s *Sql) {
	db := s.open()
	defer db.Close()
	init_Events(db)
	init_Clients(db)
	init_ClientsResponses(db)
}

func init_Events(db *sql.DB) {
	exec(db, `CREATE TABLE IF NOT EXISTS important_events (
		event_id INT AUTO_INCREMENT PRIMARY KEY,
		datetime DATETIME,
		name VARCHAR(256),
		description VARCHAR(256),
		one_time BOOLEAN,
		notify_before_days INT,
		notify_daytime TIME
	);`)
}
func init_Clients(db *sql.DB) {
	exec(db, `CREATE TABLE IF NOT EXISTS clients (

		telegram_id BIGINT UNIQUE,
		fullname VARCHAR(128),
		domain VARCHAR(32)
	
	);`)
}
func init_ClientsResponses(db *sql.DB) {
	exec(db, `CREATE TABLE IF NOT EXISTS clients_responses (

		resp_id INT AUTO_INCREMENT PRIMARY KEY,
		client_telegram_id BIGINT, 
		reponse_number INT,
		response_text VARCHAR(256)

	);`)
}
func exec(db *sql.DB, query string) {
	stmt, err := db.Prepare(query)

	if err != nil {
		log.Fatalf("Error in preparation of create `events` table query. %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		log.Fatalf("Error in executing creaate `events` table. %v", err)
	}
	log.Print("Database initialized successfully")
}

//	INITIALIZE DATABASE END
//

func (s *Sql) GetClientsFromDB() *map[int64]*subscriber.Client {
	db := s.open()
	rows, err := db.Query("SELECT telegram_id, current_handler FROM clients")
	if err != nil {
		log.Printf("Cannot receive list of clients id's. Err: %v", err)
	}
	defer rows.Close()

	clients := make(map[int64]*subscriber.Client)

	for {
		if !rows.Next() {
			break
		}
		var tid int64
		var current_handler string
		if err = rows.Scan(&tid, &current_handler); err != nil {
			log.Printf("Cannot read info about client. Err: %v", err)
			continue
		}

		clients[tid] = &subscriber.Client{ChatId: tid, HandlerCommand: current_handler}
		// TODO: associate client with Handler [Requires ConversationManager object]
	}

	responsesCount := 0
	for _, client := range clients {
		if client.Responses, err = s.GetClientResponses(client.ChatId); err != nil {
			log.Printf("Cannot receive client's responses. Tid: %d. Err: %v", client.ChatId, err)
			continue
		}
		responsesCount += len(client.Responses)
	}

	log.Printf("Loaded %d responses for %d clients", responsesCount, len(clients))

	return &clients
}
