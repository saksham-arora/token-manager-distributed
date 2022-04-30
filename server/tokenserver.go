package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"sync"

	pb "example.com/token_client_server_rpc/token_management"
	"google.golang.org/grpc"
)

type TokenManagerServer struct {
	pb.UnimplementedTokenManagerServer
}

type token struct {
	id     int32
	name   string
	domain [3]uint64
	state  [2]uint64
	mutex  sync.RWMutex
}

var TokenList []token

// var TokenList map[int32]token

func (s *TokenManagerServer) CreateNewToken(ctx context.Context, in *pb.Token) (*pb.Response, error) {
	log.Printf("Token ID Received to Create Token: %v", in.GetId())

	var create_res = "Server Response -> Token Created!"
	for _, token := range TokenList {
		if token.id == in.GetId() {
			token.mutex.Lock()
			token.state[0] = 0
			token.state[1] = 0
			token.mutex.Unlock()
			create_res = "Server Response -> Token Exists. State values of the Token Updated"
			log.Printf(create_res)
			return &pb.Response{CreateResponse: create_res}, nil
		}
	}

	created_token := token{id: in.GetId(), mutex: sync.RWMutex{}}
	TokenList = append(TokenList, created_token)
	// TokenList[in.GetId()] = created_token
	log.Printf("Server Response -> Token Created!")
	return &pb.Response{CreateResponse: create_res}, nil
}

//Reference - https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
func RemoveIndex(s []token, index int) []token {
	return append(s[:index], s[index+1:]...)
}

func (s *TokenManagerServer) DropToken(ctx context.Context, in *pb.Token) (*pb.Response, error) {
	log.Printf("Token ID received to Drop Token: %v", in.GetId())
	for i, token_desc := range TokenList {
		if token_desc.id == in.GetId() {
			token_desc.mutex.Lock()
			TokenList = RemoveIndex(TokenList, i)
			token_desc.mutex.Unlock()
			log.Printf("Server Response -> Token Deleted!")
			return &pb.Response{CreateResponse: "Server Response -> Success! Token Deleted!"}, nil
		}
	}
	return &pb.Response{CreateResponse: "Failed as token doesn't exists."}, nil
}

func Hash(name string, nonce uint64) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s %d", name, nonce)))
	return binary.BigEndian.Uint64(hasher.Sum(nil))
}

func (s *TokenManagerServer) WriteToken(ctx context.Context, in *pb.WriteTokenMsg) (*pb.WriteResponse, error) {
	log.Printf("Token ID Received to Write: %v", in.GetId())
	var flag_token_list int
	for _, token := range TokenList {
		if token.id == in.GetId() {
			flag_token_list = 1
			break
		}
	}
	var partial_val uint64 = math.MaxUint64
	if flag_token_list != 1 {
		log.Printf("Token ID doesn't exist!")
		return &pb.WriteResponse{CreateWriteResponse: partial_val}, nil
	}
	for x := in.GetLow(); x < in.GetMid(); x++ {
		var h_val = Hash(in.GetName(), x)
		if h_val < partial_val {
			partial_val = h_val
		}
	}
	for i, token := range TokenList {
		if token.id == in.GetId() {
			token.mutex.Lock()
			token_write := TokenList[i]
			token_write.name = in.GetName()
			token_write.domain[0] = in.GetLow()
			token_write.domain[1] = in.GetMid()
			token_write.domain[2] = in.GetHigh()
			token_write.state[0] = partial_val
			token_write.state[1] = 0
			TokenList[i] = token_write
			token.mutex.Unlock()
		}
	}
	log.Printf("Server Response -> Token Write Completed!")
	return &pb.WriteResponse{CreateWriteResponse: partial_val}, nil

}

func (s *TokenManagerServer) ReadToken(ctx context.Context, in *pb.Token) (*pb.WriteResponse, error) {
	log.Printf("Token ID Received to Read: %v", in.GetId())
	var final_val uint64 = math.MaxUint64
	var read_mid uint64
	var read_high uint64
	var read_partial_val uint64
	var read_name string
	var flag int
	for _, curr_token := range TokenList {
		if curr_token.id == in.GetId() {
			read_mid = curr_token.domain[1]
			read_high = curr_token.domain[2]
			read_name = curr_token.name
			read_partial_val = curr_token.state[0]
			flag = 1
		}
	}
	// print(read_name)
	if flag != 1 {
		return &pb.WriteResponse{CreateWriteResponse: final_val}, nil
	}

	for x := read_mid; x < read_high; x++ {
		var h_val = Hash(read_name, x)
		if h_val < final_val {
			final_val = h_val
		}
	}
	if final_val > read_partial_val {
		final_val = read_partial_val
	}
	for i, token := range TokenList {
		if token.id == in.GetId() {
			token.mutex.RLock()
			token_write := TokenList[i]
			token_write.state[1] = final_val
			TokenList[i] = token_write
			token.mutex.RUnlock()
		}
	}
	log.Printf("Server Response -> Token Read Completed!")
	return &pb.WriteResponse{CreateWriteResponse: final_val}, nil

}
func main() {
	port_input := flag.Int("port", 50051, "Port Number")
	flag.Parse()
	var port = ":" + strconv.Itoa(*port_input)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to Listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTokenManagerServer(s, &TokenManagerServer{})
	log.Printf("Server Listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to Serve: %v", err)
	}
}
