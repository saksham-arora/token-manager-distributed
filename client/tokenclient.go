package main

import (
	"context"
	"flag"
	"log"
	"strconv"
	"time"

	pb "example.com/token_client_server_rpc/token_management"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	port_input := flag.Int("port", 50051, "Port Number")
	host_input := flag.String("host", "localhost", "Host Name")
	create_input := flag.Bool("create", false, "Create Operation")
	drop_input := flag.Bool("drop", false, "Drop Operation")
	write_input := flag.Bool("write", false, "Write Operation")
	read_input := flag.Bool("read", false, "Read Operation")
	name_input := flag.String("name", "abc", "Token Name")
	low_input := flag.Int("low", 0, "Domain Low")
	mid_input := flag.Int("mid", 10, "Domain Mid")
	high_input := flag.Int("high", 100, "Domain High")
	id_input := flag.Int("id", 1001, "Token ID")
	flag.Parse()

	// Setting up the client side to connect from the server.
	var address = *host_input + ":" + strconv.Itoa(*port_input)
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Not able to Connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTokenManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if *create_input {
		create_req, err := c.CreateNewToken(ctx, &pb.Token{Id: int32(*id_input)})
		if err != nil {
			log.Fatalf("Couldn't Create Token: %v", err)
		}
		log.Printf(`Response from the Server: %v`, create_req.GetCreateResponse())
	}

	if *drop_input {
		drop_req, err := c.DropToken(ctx, &pb.Token{Id: int32(*id_input)})
		if err != nil {
			log.Fatalf("Couldn't Drop Token: %v", err)
		}
		log.Printf(`Response from the Server: %v`, drop_req.GetCreateResponse())
	}

	if *write_input {
		write_req, err := c.WriteToken(ctx, &pb.WriteTokenMsg{Id: int32(*id_input), Name: *name_input, Low: uint64(*low_input), Mid: uint64(*mid_input), High: uint64(*high_input)})
		if err != nil {
			log.Fatalf("Couldn't Write Token: %v", err)
		}
		log.Printf(`Response from the Server: %v`, write_req.GetCreateWriteResponse())
	}

	if *read_input {
		read_req, err := c.ReadToken(ctx, &pb.Token{Id: int32(*id_input)})
		if err != nil {
			log.Fatalf("Couldn't Read Token: %v", err)
		}
		log.Printf(`Response from the Server: %v`, read_req.GetCreateWriteResponse())
	}
}
