package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange     = "app.events"
	EventsExchangeType = "topic"
	EmployeeSyncQueue  = "business.employee.sync"
	EmployeeSyncDLQ    = "business.employee.sync.dlq"
	EmployeeEventsKey  = "employee.event.*"
	EmployeeCreatedKey = "employee.event.created"
	EmployeeUpdatedKey = "employee.event.updated"
	EmployeeDeletedKey = "employee.event.deleted"

	ExtensionDispatchQueue = "business.extensions.dispatch"
	ExtensionDispatchDLQ   = "business.extensions.dispatch.dlq"
	TaskEventsKey          = "task.event.*"
)

type Client struct {
	url      string
	conn     *amqp.Connection
	ch       *amqp.Channel
	mu       sync.RWMutex
	isClosed bool
}

func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	c := &Client{
		url:  url,
		conn: conn,
		ch:   ch,
	}

	if err := c.SetupTopology(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to setup rabbitmq topology: %w", err)
	}

	return c, nil
}

func (c *Client) SetupTopology() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Declare Main Topic Exchange
	err := c.ch.ExchangeDeclare(
		EventsExchange,
		EventsExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", EventsExchange, err)
	}

	// 2. Declare DLQ
	_, err = c.ch.QueueDeclare(
		EmployeeSyncDLQ,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dlq %s: %w", EmployeeSyncDLQ, err)
	}

	// 3. Declare Business Employee Sync Queue with DLQ arguments
	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": EmployeeSyncDLQ,
	}
	_, err = c.ch.QueueDeclare(
		EmployeeSyncQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", EmployeeSyncQueue, err)
	}

	// 4. Bind Queue to Exchange
	err = c.ch.QueueBind(
		EmployeeSyncQueue,
		EmployeeEventsKey,
		EventsExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", EmployeeSyncQueue, EventsExchange, err)
	}

	// 5. Declare Extension Dispatch DLQ
	_, err = c.ch.QueueDeclare(
		ExtensionDispatchDLQ,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dlq %s: %w", ExtensionDispatchDLQ, err)
	}

	// 6. Declare Extension Dispatch Queue with DLQ arguments
	extArgs := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": ExtensionDispatchDLQ,
	}
	_, err = c.ch.QueueDeclare(
		ExtensionDispatchQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		extArgs,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", ExtensionDispatchQueue, err)
	}

	// 7. Bind Extension Dispatch Queue to Exchange (all task.event.* messages)
	err = c.ch.QueueBind(
		ExtensionDispatchQueue,
		TaskEventsKey,
		EventsExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", ExtensionDispatchQueue, EventsExchange, err)
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, routingKey string, body []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed {
		return fmt.Errorf("rabbitmq client is closed")
	}

	return c.ch.PublishWithContext(
		ctx,
		EventsExchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (c *Client) Consume(ctx context.Context, queueName string, concurrency int, handler func(ctx context.Context, msg amqp.Delivery) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open consumer channel: %w", err)
	}

	// Prefetch count to distribute load evenly across workers
	if err := ch.Qos(concurrency*2, 0, false); err != nil {
		ch.Close()
		return fmt.Errorf("failed to set qos: %w", err)
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
		queueName,
		"",    // consumer tag
		false, // auto-ack (we use explicit Ack/Nack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to start consuming %s: %w", queueName, err)
	}

	msgChan := make(chan amqp.Delivery, concurrency)
	var wg sync.WaitGroup

	// Start Worker Pool
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for msg := range msgChan {
				workerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				err := handler(workerCtx, msg)
				cancel()

				if err != nil {
					log.Printf("[Worker %d] Error processing message (%s): %v", workerID, msg.RoutingKey, err)
					// Reject and don't requeue -> routes to DLQ
					_ = msg.Nack(false, false)
				} else {
					_ = msg.Ack(false)
				}
			}
		}(i + 1)
	}

	// Dispatcher goroutine
	go func() {
		defer func() {
			close(msgChan)
			wg.Wait()
			ch.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					_ = msg.Nack(false, true) // Requeue if shutting down
					return
				case msgChan <- msg:
				}
			}
		}
	}()

	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isClosed = true
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
