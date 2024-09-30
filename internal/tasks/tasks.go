package tasks

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"github.com/Masterminds/squirrel"
)

type Task struct {
	ID          int    `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Status      string `db:"status"`
}

type TaskStore struct {
	db    *sqlx.DB
	cache *redis.Client
	kafka *kafka.Writer
}

func NewTaskStore(db *sqlx.DB, cache *redis.Client, kafka *kafka.Writer) *TaskStore {
	return &TaskStore{db: db, cache: cache, kafka: kafka}
}

func (s *TaskStore) CreateTask(ctx context.Context, task *Task) error {
	// Создание SQL-запроса для вставки задачи
	query, args, err := squirrel.Insert("tasks").
		Columns("title", "description", "status").
		Values(task.Title, task.Description, task.Status).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return err
	}

	// Выполнение SQL-запроса и получение ID новой задачи
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&task.ID)
	if err != nil {
		return err
	}

	// Кэширование новой задачи в Redis
	cacheErr := s.cache.Set(ctx, generateRedisKey(task.ID), task.Title, 0).Err()
	if cacheErr != nil {
		log.Printf("Failed to set cache in Redis: %v", cacheErr)
	}

	// Отправка сообщения в Kafka о новой задаче
	msg := kafka.Message{
		Key:   []byte("CreateTask"),
		Value: []byte(task.Title),
	}
	kafkaErr := s.kafka.WriteMessages(ctx, msg)
	if kafkaErr != nil {
		log.Printf("Failed to send message to Kafka: %v", kafkaErr)
	}

	return nil
}

func (s *TaskStore) GetTasks(ctx context.Context) ([]Task, error) {
	// Создание SQL-запроса для получения всех задач
	query, args, err := squirrel.Select("*").From("tasks").ToSql()
	if err != nil {
		return nil, err
	}

	// Выполнение SQL-запроса и получение результата
	var tasks []Task
	err = s.db.SelectContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// Вспомогательная функция для генерации ключа Redis
func generateRedisKey(id int) string {
	return "task:" + string(id)
}