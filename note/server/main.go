package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"strings"

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

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		log.Println(fmt.Sprintf("Not found: %s", r.URL.Path))
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join("proto", p)
	http.ServeFile(w, r, p)
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

	r := http.NewServeMux()
	fs := http.FileServer(http.Dir("../swagger-ui"))

	r.HandleFunc("/swagger/", serveSwagger)
	r.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))
	r.Handle("/", mux)
	http.ListenAndServe(address, r)
}
