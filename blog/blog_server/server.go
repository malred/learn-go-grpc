package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"rpc-learn/blog/blogpb"

	// "io"
	"log"
	"net"

	// "strconv"
	// "time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

var coll *mongo.Collection

type server struct{}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := coll.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			fmt.Sprintf("Internal error: %v", err))
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal,
			fmt.Sprintf("Cannot convert to OID"))
	}
	data.ID = oid

	return &blogpb.CreateBlogResponse{
		Blog: dataToBlogpb(&data),
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Println("Read blog request")

	blogId := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"))
	}

	// create a empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}

	res := coll.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID %v", err))
	}

	return &blogpb.ReadBlogResponse{
		Blog: dataToBlogpb(data),
	}, nil
}

func (*server) UpdateBlog(
	ctx context.Context, req *blogpb.UpdateBlogRequest,
) (
	*blogpb.UpdateBlogResponse, error,
) {
	log.Println("Updating blog request")
	blog := req.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"))
	}

	// create a empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}

	res := coll.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID %v", err),
		)
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updateErr := coll.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object:%v", updateErr),
		)
	}

	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogpb(data),
	}, nil
}

func (*server) DeleteBlog(
	ctx context.Context, req *blogpb.DeleteBlogRequest,
) (
	*blogpb.DeleteBlogResponse, error,
) {
	log.Println("Deleting blog request")
	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	filter := bson.D{{"_id", oid}}
	res, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object: %v", err),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot find object: %v", err),
		)
	}

	return &blogpb.DeleteBlogResponse{BlogId: req.GetBlogId()}, nil
}

func (*server) ListBlog(
	req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer,
) error {
	log.Println("List blog request")

	// 查询所有要给一个空的filter
	cur, err := coll.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data: %v", err),
			)
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlogpb(data),
		})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	return nil
}

func dataToBlogpb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog server started")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	mongoOpts := options.Client().ApplyURI("mongodb://localhost:27017").SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), mongoOpts)
	if err != nil {
		{
			panic(err)
		}
	}
	// 数据库 sample_mflix,表 blogs
	coll = client.Database("sample_mflix").Collection("blogs")

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	// 调用grpc,注册服务
	blogpb.RegisterBlogServiceServer(s, &server{})
	// 开启反射服务
	reflection.Register(s) 

	go func() {
		// 开启服务器
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// wait for ctrl-c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is received
	<-ch
	log.Println("Stopping the server")
	s.Stop()
	log.Println("Stopping the listener")
	lis.Close()
	log.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	log.Println("End of Program")
}
