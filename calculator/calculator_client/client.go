package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"rpc-learn/calculator/calculatorpb"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Calculator Client")
	// WithInsecure: 不安全连接
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer conn.Close()

	c := calculatorpb.NewCalculatorServiceClient(conn)
	// fmt.Printf("Created client: %f", c)

	// doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	// doBiDiStreaming(c)

	doErrorUnary(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	log.Println("Starting to do a Sum Unary RPC...")
	req := &calculatorpb.SumRequest{
		FirstNumber:  56,
		SecondNumber: 8,
	}

	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Sum RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.SumResult)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	log.Println("Starting to do a PrimeNumberDecomposition RPC...")
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 222141,
	}

	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeNumberDecomposition RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Println(res.GetPrimeFactor())
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	log.Println("Starting to do a ComputeAverage Client Streaming RPC...")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream: %v", err)
	}

	numbers := []int32{3, 5, 9, 54, 23}

	for _, number := range numbers {
		log.Printf("Sending number: %v\n", number)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while receiving response: %v", err)
	}

	log.Println("The Average is:", res.GetAverage())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	log.Println("Starting to do a FindMaxium BiDi Streaming RPC...")

	stream, err := c.FindMaxium(context.Background())

	if err != nil {
		log.Fatalf("Error while opening stream and calling FindMaxium: %v", err)
	}

	waitc := make(chan struct{})
	// send go routine
	go func() {
		numbers := []int32{4, 7, 2, 19, 4, 6, 32}
		for _, number := range numbers {
			log.Println("Sending number:", number)
			stream.Send(&calculatorpb.FindMaxiumRequest{
				Number: number,
			})
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	// receive go routine
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Problem while reading server stream: %v", err)
				break
			}
			maxium := res.GetMaxium()
			log.Println("Received a new maxium of...:", maxium)
		}
		close(waitc)
	}()
	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	log.Println("Starting to be a SquareRoot Unary RPC...")

	doErrorCall(c, 10)
	doErrorCall(c, -2)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: n})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from grpc (user error)
			log.Println("Error message from server:", respErr.Message())
			log.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				log.Println("We probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Big Error calling SquareRoot: %v", err)
			return
		}
	}
	log.Println("Result of square root of %v: %v", n, res.GetNumberRoot())
}
