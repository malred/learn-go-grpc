package main

import (
	"context"
	"fmt"
	"io"

	// "io"
	"log"
	"rpc-learn/blog/blogpb"

	// "time"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/credentials"
	// "google.golang.org/grpc/status"
)

func main() {
	fmt.Println("blog client")

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)

	// create
	// blog := &blogpb.Blog{
	// 	AuthorId: "Stephane2",
	// 	Title:    "My First Blog2",
	// 	Content:  "Content of the first blog2",
	// }
	// create(c, blog)

	// find
	// findOne(c, "64a5f45f94466f3e02f74f79")

	// update
	// blog := &blogpb.Blog{
	// 	Id:       "64a5f45f94466f3e02f74f79",
	// 	AuthorId: "changer",
	// 	Title:    "My change Blog",
	// 	Content:  "Content of the change blog",
	// }
	// update(c, blog)

	// delete
	// delete(c, "64a627df7303d9f27d4f19a8")

	// list
	list(c)
}

func create(c blogpb.BlogServiceClient, blog *blogpb.Blog) {
	log.Println("Creating the blog")
	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: blog,
	})
	if err != nil {
		log.Fatalf("Unexpected error creating blog: %v", err)
	}
	log.Println("Blog has been created:", res)
}

func findOne(c blogpb.BlogServiceClient, id string) {
	log.Println("Reading the blog")

	res, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: id,
	})
	if err != nil {
		log.Fatalf("Unexpected error reading blog: %v", err)
	}
	log.Println("Blog has been read:", res)
}

func update(c blogpb.BlogServiceClient, blog *blogpb.Blog) {
	log.Println("Updating the blog")
	uptRes, uptErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{
		Blog: blog,
	})
	if uptErr != nil {
		log.Fatalf("Error while updating: %v", uptErr)
	}
	log.Println("Blog was read:", uptRes)
}

func delete(c blogpb.BlogServiceClient, id string) {
	log.Println("Deleting the blog")
	delRes, delErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{
		BlogId: id,
	})
	if delErr != nil {
		log.Fatalf("Error while deleting: %v", delErr)
	}
	log.Println("Blog was read:", delRes)
}

func list(c blogpb.BlogServiceClient) {
	stream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("error while calling ListBlog RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading ListBlog RPC: %v", err)
		}
		log.Println(res.GetBlog())
	}
}
