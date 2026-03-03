---
name: go-rmq
description: Comprehensive RabbitMQ skill with best practices from official documentation
---

# Go RabbitMQ Skill

## Overview

This skill provides comprehensive guidance for working with RabbitMQ in Go applications, following best practices from the official RabbitMQ documentation.

## Packages Used

The following packages are commonly used for RabbitMQ integration in Go applications:
1. `github.com/rabbitmq/amqp091-go v1.10.0` - Official RabbitMQ client
2. `github.com/wagslane/go-rabbitmq v0.15.0` - Higher-level wrapper providing easier-to-use APIs

## Core Concepts and Best Practices

### Connection Management

Proper connection management is critical for reliable messaging systems:

1. **Use connection pooling**: Reuse connections instead of creating new ones for each operation
2. **Implement reconnection logic**: Handle network interruptions gracefully
3. **Set appropriate timeouts**: Prevent hanging operations that can lead to resource exhaustion

```go
// Example connection configuration
type RMQConfig struct {
    URL               string        `validate:"required,url"`
    Concurrency       int           `default:"1" validate:"gt=0"`
    Prefetch          int           `default:"10" validate:"gt=0"`
    ReconnectInterval time.Duration `mapstructure:"reconnect_interval" default:"10s" validate:"min=1s,max=1m"`
    TTL               time.Duration `default:"1m" validate:"min=1s,max=1h"`
}
```

### Message Delivery Guarantees

Understanding and implementing appropriate delivery semantics:

1. **At-least-once delivery**: Messages are acknowledged after processing
2. **Idempotent processing**: Consumers should handle duplicate messages gracefully
3. **Dead letter exchanges**: Handle failed messages appropriately

### Consumer Implementation Pattern

A clean consumer pattern:

```go
func (h *Handler) ProcessMessage(
    delivery rabbitmq.Delivery, //nolint:gocritic // Передается по значению из пакета go-rabbitmq
) rabbitmq.Action {
    var msg message.RecordProcessed

    if err := json.Unmarshal(delivery.Body, &msg); err != nil {
        monitoring.LogAndCaptureError(h.log, "Failed to parse record processed message", err)
        return rabbitmq.Ack // Ack even on parsing error to avoid infinite retries
    }

    if err := msg.Validate(); err != nil {
        monitoring.LogAndCaptureError(h.log, "Failed to validate record processed message", err)
        return rabbitmq.Ack // Ack invalid messages to prevent retry loops
    }

    // Process message
    if err := h.processMessage(context.Background(), &msg); err != nil {
        monitoring.LogAndCaptureError(h.log, "Failed to process record processed message", err)
        // Determine if we should nack or ack based on the error type
        if errors.Is(err, pgx.ErrNoRows) {
            return rabbitmq.Ack
        }

        if delivery.Redelivered && config.Conf.MaxProcessDeadline != 0 &&
            time.Since(delivery.Timestamp) > config.Conf.MaxProcessDeadline {
            h.log.Warn(
                "Drop expired message",
                zap.Int("record_id", msg.ID),
                zap.Time("timestamp", delivery.Timestamp),
            )
            return rabbitmq.Ack
        }

        return rabbitmq.NackDiscard // Retry on transient failures
    }
    
    return rabbitmq.Ack
}
```

### Producer Implementation Pattern

For producers, proper error handling and confirmation mechanisms are important:

```go
// Producer with confirmation
func (p *Publisher) Publish(data any, routingKey string) error {
    body, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("marshal message: %w", err)
    }

    err = p.channel.PublishWithContext(
        context.Background(),
        p.exchange,
        routingKey,
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
            Timestamp:   time.Now(),
            Headers: amqp.Table{
                "x-delay": 0,
            },
        },
    )
    
    if err != nil {
        return fmt.Errorf("publish message: %w", err)
    }
    
    return nil
}
```

## Configuration Guidelines

### Connection Settings

Set these properly for production environments:

- **URL**: Use secure connection URLs (`amqps://`)
- **ReconnectInterval**: Set to 10-30 seconds for resilient systems
- **Prefetch**: Control concurrent message processing (10-100 recommended)
- **TTL**: Set appropriate time-to-live for messages

### Exchange and Queue Configuration

Standard configuration:
```json
{
  "exchanges": [
    {
      "name": "obelisk.events",
      "vhost": "/",
      "type": "topic",
      "durable": true,
      "auto_delete": false,
      "internal": false
    }
  ]
}
```

### Routing Keys

Use hierarchical routing keys for better organization:
- `record.inited`
- `record.processed` 
- `record.uploaded`
- `call.processed`

## Error Handling Best Practices

1. **Acknowledge successfully processed messages** immediately
2. **Nack with discard for invalid messages** to prevent retry loops
3. **Nack with requeue for transient failures** 
4. **Log errors appropriately** with context for debugging
5. **Handle timeout scenarios** for long-running processes

## Performance Considerations

1. **Batch message processing** where possible
2. **Limit concurrent consumers** to prevent resource exhaustion
3. **Use appropriate prefetch counts** to balance throughput and memory usage
4. **Monitor queue depths** to detect backpressure
5. **Implement circuit breakers** for external dependencies

## Monitoring and Observability

Implement these key metrics:
- Message processing rate
- Queue depth
- Processing latency
- Error rates
- Dead letter queue counts

## Security Recommendations

1. **Use TLS connections** (`amqps://`) in production
2. **Validate message content** before processing
3. **Implement proper authentication** with minimal required permissions
4. **Sanitize message headers** to prevent injection attacks
5. **Use durable exchanges and queues** for reliability

## Testing Patterns

For testing RabbitMQ consumers:

1. **Mock the message processing** layer
2. **Test different error scenarios** 
3. **Verify message acknowledgment** behavior
4. **Simulate network failures** and reconnections
5. **Test message routing** with different routing keys

## Common Pitfalls to Avoid

1. **Not acknowledging messages** leading to queue buildup
2. **Processing messages synchronously** causing backpressure
3. **Not handling re-deliveries** properly
4. **Overlooking message TTL** settings
5. **Poor error logging** that doesn't aid debugging
6. **Assuming AsyncAPI spec is automatically enforced** at runtime - requires manual synchronization with code
7. **Ignoring validation integration** - AsyncAPI validation should be part of CI pipeline

## Integration with Go Application Lifecycle

1. **Initialize connections during startup**
2. **Graceful shutdown** with connection cleanup
3. **Health check endpoints** for monitoring
4. **Configuration validation** before starting services
5. **Contract validation** - ensure AsyncAPI spec matches implementation details like exchange names

## Sample Configuration Structure

```yaml
rmq:
  url: amqps://user:pass@localhost:5671/
  concurrency: 10
  prefetch: 50
  reconnect_interval: 10s
  ttl: 1m
```

## Resource Management

Always implement proper resource cleanup:
```go
defer func() {
    if err := channel.Close(); err != nil {
        log.Printf("Error closing channel: %v", err)
    }
    if err := conn.Close(); err != nil {
        log.Printf("Error closing connection: %v", err)
    }
}()
```

## Advanced Features

1. **Message TTL**: Set time-based expiration
2. **Dead Letter Exchanges**: Route failed messages for analysis
3. **Priority Queues**: Handle urgent messages differently
4. **Message Compression**: Reduce bandwidth for large payloads
5. **Transactions**: Ensure atomic operations across multiple exchanges

This skill provides a foundation for building robust, scalable RabbitMQ integrations that follow industry best practices.