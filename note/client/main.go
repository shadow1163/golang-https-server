package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "../proto"
)

const (
	address     = "127.0.0.1:50051"
	title       = "test"
	description = "test123"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewNoteServiceClient(conn)

	// Contact the server and print out its response.
	r, err := c.Get(context.Background(), &pb.Message{Id: "1", Title: title, Description: description})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Geting: %s-%s-%s", r.Id, r.Title, r.Description)
}
