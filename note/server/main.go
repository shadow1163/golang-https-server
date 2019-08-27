package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	pb "../proto"
)

var (
	noteEndpoint = "localhost:50051"
	address      = ":8080"
)

type noteServer struct {
	ID          int    `jsion:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func newNoteServer() pb.NoteServiceServer {
	return new(noteServer)
}

func (s *noteServer) Get(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Get: %s", msg))
	return msg, nil
}

func (s *noteServer) Post(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Post: %s", msg))
	return msg, nil
}

func (s *noteServer) Put(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Put: %s", msg))
	return msg, nil
}

func (s *noteServer) Delete(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	log.Println(fmt.Sprintf("Delete: %s", msg))
	return msg, nil
}

func main() {
	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln(err)
	}
	server := grpc.NewServer()
	pb.RegisterNoteServiceServer(server, newNoteServer())
	go server.Serve(listen)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterNoteServiceHandlerFromEndpoint(ctx, mux, noteEndpoint, dialOpts)
	if err != nil {
		log.Fatalln(err)
	}

	http.ListenAndServe(address, mux)
}
