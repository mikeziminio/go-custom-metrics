package memstorage

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
)

func ExampleStorage_Update_counter() {
	ctx := context.Background()

	storage, err := New(false, "", zap.L())
	if err != nil {
		fmt.Println("Error creating storage:", err)
		return
	}

	metric := model.Metric{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: new(int64(100)),
	}

	updated, err := storage.Update(ctx, metric)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Updated counter value: %d\n", *updated.Delta)

	metric2 := model.Metric{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: new(int64(50)),
	}

	updated2, err := storage.Update(ctx, metric2)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Updated counter value after second update: %d\n", *updated2.Delta)
	// Output:
	// Updated counter value: 100
	// Updated counter value after second update: 150
}

func ExampleStorage_Update_gauge() {
	ctx := context.Background()

	storage, err := New(false, "", zap.L())
	if err != nil {
		fmt.Println("Error creating storage:", err)
		return
	}

	metric := model.Metric{
		ID:    "test_gauge",
		MType: model.Gauge,
		Value: new(float64(42.5)),
	}

	updated, err := storage.Update(ctx, metric)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Updated gauge value: %.1f\n", *updated.Value)

	metric2 := model.Metric{
		ID:    "test_gauge",
		MType: model.Gauge,
		Value: new(float64(100.7)),
	}

	updated2, err := storage.Update(ctx, metric2)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Updated gauge value after second update: %.1f\n", *updated2.Value)
	// Output:
	// Updated gauge value: 42.5
	// Updated gauge value after second update: 100.7
}

func ExampleStorage_Update_withTimeout() {
	metric := model.Metric{
		ID:    "timeout_test",
		MType: model.Counter,
		Delta: new(int64(10)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	storage, err := New(false, "", zap.L())
	if err != nil {
		fmt.Println("Error creating storage:", err)
		return
	}

	updated, err := storage.Update(ctx, metric)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Updated metric: %s, value: %d\n", updated.ID, *updated.Delta)
	// Output:
	// Updated metric: timeout_test, value: 10
}

func ExampleStorage_Update_errorContext() {
	metric := model.Metric{
		ID:    "test_metric",
		MType: model.Gauge,
		Value: new(float64(123.45)),
	}

	ctx := context.Background()

	storage, err := New(false, "", zap.L())
	if err != nil {
		fmt.Println("Error creating storage:", err)
		return
	}

	updated, err := storage.Update(ctx, metric)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Metric updated: ID=%s, Type=%s, Value=%.2f\n", updated.ID, updated.MType, *updated.Value)
	// Output:
	// Metric updated: ID=test_metric, Type=gauge, Value=123.45
}
