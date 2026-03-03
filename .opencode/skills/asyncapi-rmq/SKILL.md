---
name: asyncapi-rmq
description: AsyncAPI skill for documenting RabbitMQ message broker interactions
---

# AsyncAPI for RabbitMQ Interactions Skill

This skill provides comprehensive guidance for creating AsyncAPI specifications to document RabbitMQ message broker interactions, following official AsyncAPI documentation and best practices.

## What is AsyncAPI?

AsyncAPI is an open-source initiative that provides a standard way to define asynchronous APIs, including message brokers like RabbitMQ, Apache Kafka, AWS SNS/SQS, and others. It enables:

- Clear documentation of message flows
- Automated code generation
- Contract testing
- Service discovery
- Integration with API gateways and monitoring tools

## AsyncAPI Specification Structure

### Basic Document Structure

```yaml
asyncapi: 3.0.0
info:
  title: Service Name
  version: 1.0.0
  description: Service description

servers:
  development:
    url: amqp://localhost:5672
    protocol: amqp
    description: Development server

channels:
  # Message channels and their definitions

operations:
  # Operations (publish/subscribe)

components:
  schemas:
    # Data schemas
  messages:
    # Message definitions
```

## RabbitMQ-Specific Patterns

### 1. Channel Definitions

Channels represent message queues or exchanges in RabbitMQ:

```yaml
channels:
  user.created:
    address: user.created
    messages:
      UserCreatedEvent:
        $ref: "#/components/messages/UserCreatedEvent"
    description: Queue for user creation events

  payment.processed:
    address: payment.processed
    messages:
      PaymentProcessedEvent:
        $ref: "#/components/messages/PaymentProcessedEvent"
    description: Exchange for payment processing events
```

### 2. Operation Definitions

Operations define how messages are published or consumed:

```yaml
operations:
  UserCreated:
    action: receive
    channel:
      $ref: "#/channels/user.created"
    messages:
      - $ref: "#/channels/user.created/messages/UserCreatedEvent"

  PaymentProcessed:
    action: send
    channel:
      $ref: "#/channels/payment.processed"
    messages:
      - $ref: "#/channels/payment.processed/messages/PaymentProcessedEvent"
```

### 3. Message Schema Definitions

Messages should be defined with proper payloads:

```yaml
components:
  schemas:
    UserId:
      type: integer
      description: Unique identifier for user
    
    UserPayload:
      type: object
      properties:
        id:
          $ref: "#/components/schemas/UserId"
        name:
          type: string
        email:
          type: string
      required:
        - id
        - name
        - email

  messages:
    UserCreatedEvent:
      summary: User creation event
      payload:
        $ref: "#/components/schemas/UserPayload"
      description: Event sent when a new user is created
```

## Best Practices for AsyncAPI with RabbitMQ

### 1. Naming Conventions

Use consistent, descriptive naming for channels and operations:
- Use dot notation for hierarchical organization: `service.event.type`
- Be specific about event content: `user.created`, `payment.confirmed`
- Use past tense for events: `user.created`, `order.shipped`

### 2. Channel Organization

Group related channels logically:
```yaml
channels:
  # User service events
  user.created:
    address: user.created
    messages:
      UserCreatedEvent: ...

  user.updated:
    address: user.updated
    messages:
      UserUpdatedEvent: ...

  # Payment service events  
  payment.processed:
    address: payment.processed
    messages:
      PaymentProcessedEvent: ...
```

### 3. Message Payload Structure

Define clear and consistent message structures:
```yaml
components:
  schemas:
    EventMetadata:
      type: object
      properties:
        eventId:
          type: string
        timestamp:
          type: string
          format: date-time
        source:
          type: string
        version:
          type: string
      required:
        - eventId
        - timestamp
        - source
        - version

    UserCreatedEventPayload:
      allOf:
        - $ref: "#/components/schemas/EventMetadata"
        - type: object
          properties:
            userId:
              type: integer
            userEmail:
              type: string
            userName:
              type: string
          required:
            - userId
            - userEmail
            - userName
```

### 4. Error Handling Documentation

Include error message schemas:
```yaml
components:
  schemas:
    ErrorMessage:
      type: object
      properties:
        errorId:
          type: string
        timestamp:
          type: string
          format: date-time
        message:
          type: string
        code:
          type: string
        details:
          type: object
      required:
        - errorId
        - timestamp
        - message

  messages:
    ErrorMessage:
      payload:
        $ref: "#/components/schemas/ErrorMessage"
```

## Complete AsyncAPI Example

```yaml
asyncapi: 3.0.0
info:
  title: Notification Service
  version: 1.0.0
  description: Service for handling notification events

servers:
  development:
    url: amqps://localhost:5671/
    protocol: amqp
    description: Development RabbitMQ server
    security:
      - tls: []

channels:
  notifications.email.sent:
    address: notifications.email.sent
    messages:
      EmailSentEvent:
        $ref: "#/components/messages/EmailSentEvent"
    description: Queue for email sent notifications

  notifications.sms.sent:
    address: notifications.sms.sent
    messages:
      SMSSentEvent:
        $ref: "#/components/messages/SMSSentEvent"
    description: Queue for SMS sent notifications

  notifications.events:
    address: notifications.events
    messages:
      NotificationSentEvent:
        $ref: "#/components/messages/NotificationSentEvent"
      NotificationFailedEvent:
        $ref: "#/components/messages/NotificationFailedEvent"
    description: Exchange for notification events

operations:
  EmailSent:
    action: receive
    channel:
      $ref: "#/channels/notifications.email.sent"
    messages:
      - $ref: "#/channels/notifications.email.sent/messages/EmailSentEvent"

  SMSSent:
    action: receive
    channel:
      $ref: "#/channels/notifications.sms.sent"
    messages:
      - $ref: "#/channels/notifications.sms.sent/messages/SMSSentEvent"

  NotificationSent:
    action: send
    channel:
      $ref: "#/channels/notifications.events"
    messages:
      - $ref: "#/channels/notifications.events/messages/NotificationSentEvent"

  NotificationFailed:
    action: send
    channel:
      $ref: "#/channels/notifications.events"
    messages:
      - $ref: "#/channels/notifications.events/messages/NotificationFailedEvent"

components:
  schemas:
    NotificationId:
      type: string
      description: Unique identifier for notification
    
    NotificationStatus:
      type: string
      enum:
        - sent
        - failed
        - delivered
        - read
      description: Current status of notification

    EmailAddress:
      type: string
      format: email
      description: Email address

    NotificationPayload:
      type: object
      properties:
        id:
          $ref: "#/components/schemas/NotificationId"
        recipient:
          $ref: "#/components/schemas/EmailAddress"
        subject:
          type: string
        content:
          type: string
        status:
          $ref: "#/components/schemas/NotificationStatus"
        timestamp:
          type: string
          format: date-time
      required:
        - id
        - recipient
        - subject
        - content
        - status
        - timestamp

  messages:
    EmailSentEvent:
      summary: Email sent notification
      payload:
        $ref: "#/components/schemas/NotificationPayload"
      description: Event sent when email has been sent successfully

    SMSSentEvent:
      summary: SMS sent notification
      payload:
        $ref: "#/components/schemas/NotificationPayload"
      description: Event sent when SMS has been sent successfully

    NotificationSentEvent:
      summary: Notification sent event
      payload:
        $ref: "#/components/schemas/NotificationPayload"
      description: Generic event sent when notification has been processed

    NotificationFailedEvent:
      summary: Notification failed event
      payload:
        $ref: "#/components/schemas/NotificationPayload"
      description: Event sent when notification processing failed
```

## AsyncAPI Tooling Integration

### 2. Validation

Validate your AsyncAPI specification:
```bash
asyncapi validate asyncapi.yaml
```

### 3. Integration with Build Pipeline

Note: While AsyncAPI validation is recommended, it's often missing from standard CI/CD pipelines. The validation should be integrated into the build process as a pre-commit hook or CI step to maintain contract consistency.

### 4. Runtime Considerations

Validation occurs separately from runtime execution. The AsyncAPI specification serves as a contract between the messaging layer and application code, but runtime behavior depends on the RabbitMQ client implementation rather than the spec itself.

## Key AsyncAPI Elements for RabbitMQ

### 1. Servers Section
```yaml
servers:
  production:
    url: amqps://rabbitmq.company.com:5671/
    protocol: amqp
    protocolVersion: 0.9.1
    description: Production RabbitMQ cluster
    security:
      - tls: []
      - basic_auth: []
```

### 2. Channel Properties
```yaml
channels:
  my.queue:
    address: my.queue
    description: My message queue
    bindings:
      amqp:
        is: queue
        durable: true
        exclusive: false
        autoDelete: false
        bindingVersion: 0.2.0
```

### 3. Message Bindings
```yaml
channels:
  my.exchange:
    address: my.exchange
    messages:
      MyMessage:
        $ref: "#/components/messages/MyMessage"
    bindings:
      amqp:
        is: exchange
        type: topic
        durable: true
        autoDelete: false
        bindingVersion: 0.2.0
```

### 4. Schemas
**MUST** use schemas **ONLY** when appropriate field appears more than 1 time
```yaml
    step_id:
      $ref: '#/components/schemas/StepId'
```

## Best Practices Summary

### ✅ Do's
1. **Use clear, descriptive names** for channels, operations, and messages
2. **Define consistent schemas** for message payloads
3. **Document all important elements** with summaries and descriptions
4. **Use versioning** for your AsyncAPI documents
5. **Implement proper error handling** documentation
6. **Validate your schemas** against real data structures
7. **Integrate validation into CI/CD** pipeline for contract enforcement

### ❌ Don'ts
1. **Don't use generic names** like "event" or "message" without context
2. **Don't omit required fields** from message schemas
3. **Don't neglect documentation** - all elements should be described
4. **Don't forget security** - document authentication requirements
5. **Don't ignore bindings** - RabbitMQ specific configurations matter
6. **Don't skip validation** - always validate your AsyncAPI documents
7. **Don't assume sync between spec and code** - manual synchronization required for constants

This skill provides the foundation for creating comprehensive AsyncAPI specifications for RabbitMQ-based systems, enabling better service discovery, automated tooling integration, and improved developer experience.
