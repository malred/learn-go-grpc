package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
	"rpc-learn/calculator/calculatorpb"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	log.Printf("Received Sum RPC: %v", req)
	firstNumber := req.FirstNumber
	secondNumber := req.SecondNumber
	sum := firstNumber + secondNumber
	res := &calculatorpb.SumResponse{
		SumResult: sum,
	}
	return res, nil
}

func (*server) PrimeNumberDecomposition(
	req *calculatorpb.PrimeNumberDecompositionRequest,
	stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	log.Printf("Received PrimeNumberDecomposition RPC: %v", req)

	number := req.GetNumber()
	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			number = number / divisor
		} else {
			divisor++
			log.Printf("Divisor has increased to %v", divisor)
		}
	}
	return nil
}

func (*server) ComputeAverage(
	stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	log.Println("Received ComputeAverage RPC")

	sum := int32(0)
	count := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			average := float64(sum) / float64(count)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: average,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		sum += req.GetNumber()
		count++
	}
}

func (*server) SquareRoot(
	ctx context.Context, req *calculatorpb.SquareRootRequest) (
	*calculatorpb.SquareRootResponse, error) {
	log.Println("Received SquareRoot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func (*server) FindMaxium(stream calculatorpb.CalculatorService_FindMaxiumServer) error {
	log.Println("Received FindMaxium RPC")
	maxium := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}
		number := req.GetNumber()
		if number > maxium {
			maxium = number
			err = stream.Send(&calculatorpb.FindMaxiumResponse{
				Maxium: maxium,
			})
			if err != nil {
				log.Fatalf("Error while sending client stream: %v", err)
				return err
			}
		}
	}
}

func main() {
	fmt.Println("Calculator Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// 调用grpc,注册服务
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	reflection.Register(s)
	
	// 开启服务器
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
