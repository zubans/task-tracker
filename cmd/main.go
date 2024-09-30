package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"github.com/zubans/task-tracker/internal/config"
	"github.com/zubans/task-tracker/internal/tasks"
	pb "github.com/zubans/task-tracker/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTaskServiceServer
	store *tasks.TaskStore
}

func (s *server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	task := &tasks.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := s.store.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return &pb.CreateTaskResponse{
		Task: &pb.Task{
			Id:          int32(task.ID),
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
		},
	}, nil
}

func (s *server) GetTasks(ctx context.Context, req *pb.GetTasksRequest) (*pb.GetTasksResponse, error) {
	tasks, err := s.store.GetTasks(ctx)
	if err != nil {
		return nil, err
	}

	pbTasks := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		pbTasks[i] = &pb.Task{
			Id:          int32(task.ID),
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
		}
	}

	return &pb.GetTasksResponse{Tasks: pbTasks}, nil
}

func main() {
	port := ":50051"

	// Загружаем конфигурацию и подключаемся к базе данных
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST"),
	})

	// Подключение к Kafka
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic:   "tasks",
	})

	// Инициализация хранилища задач
	store := tasks.NewTaskStore(cfg.DB, rdb, kafkaWriter)

	// Настройка gRPC сервера
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &server{store: store})

	log.Printf("Server is listening on %v", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}