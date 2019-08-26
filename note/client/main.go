package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "../proto"
)

const (
	address     = "172.17.0.3:50051"
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
	r, err := c.Get(context.Background(), &pb.Message{Title: title, Description: description})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Geting: %s-%s", r.Title, r.Description)
}
