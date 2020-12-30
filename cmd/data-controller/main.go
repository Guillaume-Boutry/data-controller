package main

import (
	"context"
	"fmt"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gocql/gocql"
	"github.com/kelseyhightower/envconfig"
)

type Receiver struct {
	client       cloudevents.Client
	CassandraURI string `envconfig:"CASSANDRA"`
	Keyspace     string `envconfig:"KEYSPACE"`
	session      *gocql.Session
}

func main() {
	log.Println("Starting data-controller")
	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	r := Receiver{client: client}
	if err := envconfig.Process("", &r); err != nil {
		log.Fatal(err.Error())
	}
	cluster := gocql.NewCluster(r.CassandraURI)
	cluster.Timeout = 20 * time.Second
	session, err := cluster.CreateSession()
	defer session.Close()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	r.session = session
	if err := createCassandraEnvironment(session, r.Keyspace); err != nil {
		log.Fatal(err.Error())
		return
	}
	cluster.Keyspace = r.Keyspace
	if err := client.StartReceiver(context.Background(), r.ReceiveAndReply); err != nil {
		log.Fatal(err.Error())
	}
}

func createCassandraEnvironment(session *gocql.Session, keyspace string) error {
	err := session.Query(fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
    WITH replication = {
        'class' : 'NetworkTopologyStrategy',
        'dc1' : 1
    };`, keyspace)).Exec()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	err = session.Query(fmt.Sprintf("CREATE COLUMNFAMILY IF NOT EXISTS %s.tbl_Face (Email VARCHAR PRIMARY KEY, Template VARCHAR);", keyspace)).Exec()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

// Request is the structure of the event we expect to receive.
type RequestInsert struct {
	Id         string `json:"id"`
	Embeddings string `json:"embeddings"`
}

// Response is the structure of the event we send in response to requests.
type Response struct {
	Message string `json:"message,omitempty"`
}

// ReceiveAndReply is invoked whenever we receive an event.
func (recv *Receiver) ReceiveAndReply(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {

	switch event.Type() {
	case "insert":
		return recv.processInsert(ctx, event)
	case "get":
		return recv.processGet(ctx, event)
	}

	resp := &Response{Message: "Error, no supported event type given"}

	r := cloudevents.NewEvent(cloudevents.VersionV1)
	r.SetType("error")
	r.SetSource("data-controller")
	if err := r.SetData("application/json", resp); err != nil {
		log.Println(err)
		return nil, cloudevents.NewHTTPResult(500, "failed to set response data: %s", err)
	}

	return &r, nil
}

func (recv *Receiver) processInsert(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {
	req := RequestInsert{}
	if err := event.DataAs(&req); err != nil {
		log.Println(err)
		return nil, cloudevents.NewHTTPResult(400, "failed to convert data: %s", err)
	}

	if err := recv.session.Query(fmt.Sprintf("INSERT INTO %s.tbl_Face (Email,Template) VALUES ('%s', '%s')", recv.Keyspace, req.Id, req.Embeddings)).Exec(); err != nil {
		log.Println(err)
		return nil, cloudevents.NewHTTPResult(500, "error while inserting data in database: %s", err)
	}
	log.Printf("Successfuly inserted %s", req.Id)

	r := cloudevents.NewEvent(cloudevents.VersionV1)
	r.SetType("inserter")
	r.SetSource("data-controller")
	resp := &Response{Message: "Success"}
	if err := r.SetData("application/json", resp); err != nil {
		log.Println(err)
		return nil, cloudevents.NewHTTPResult(500, "failed to set response data: %s", err)
	}

	return &r, nil
}

func (recv *Receiver) processGet(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {

	return nil, nil
}
