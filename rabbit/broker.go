package rabbit

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/epos-eu/converter-service/handler"
	"github.com/epos-eu/converter-service/loggers"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	maxMessages          = 0
	maxReconnectAttempts = 0
	logger               = loggers.BROKER_LOGGER
)

const (
	// exchanges
	ExchangeExternalAccess  = "externalAccess"
	ExchangeMetadataService = "metadataService"

	// queues
	QueueMap       = "map"
	QueueResources = "resources"

	// bindings
	BindingKeyMap = "#.map"

	// routing‑key suffixes
	RkAccessReturn = "access_return"
	RkMapReturn    = "map_return"
)

func init() {
	maxMessages = envInt("MAX_MESSAGES", 1)
	maxReconnectAttempts = envInt("MAX_RECONNECT_ATTEMPTS", 10)
}

type BrokerConfig struct {
	host, user, password, vhost        string
	Conn                               *amqp.Connection
	publishChan, consumeChan           *amqp.Channel
	externalAccessQ, resourcesServiceQ *amqp.Queue
}

func (b *BrokerConfig) dial() error {
	logger.Debug("attempting to connect to RabbitMQ", "host", b.host, "vhost", b.vhost, "user", b.user)
	uri := fmt.Sprintf("amqp://%s:%s@%s/%s", b.user, b.password, b.host, b.vhost)
	var err error
	b.Conn, err = amqp.Dial(uri)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err)
		return fmt.Errorf("error during dial AMQP: %w", err)
	}
	logger.Info("successfully connected to RabbitMQ", "host", b.host, "vhost", b.vhost)
	return nil
}

func NewBroker() *BrokerConfig {
	logger.Debug("initializing new broker with environment variables")
	host := env("BROKER_HOST", "rabbitmq")
	user := env("BROKER_USERNAME", "changeme")
	password := env("BROKER_PASSWORD", "changeme")
	vhost := env("BROKER_VHOST", "changeme")

	logger.Info("broker configuration created", "host", host, "user", user, "vhost", vhost)
	return &BrokerConfig{
		host:     host,
		user:     user,
		password: password,
		vhost:    vhost,
	}
}

func (b *BrokerConfig) Restart() error {
	logger.Info("restarting broker connection")

	// close old stuff if open
	if b.consumeChan != nil {
		logger.Debug("closing consumer channel")
		err := b.consumeChan.Close()
		if err != nil {
			logger.Warn("error closing consumer channel", "error", err)
		}
	}

	if b.publishChan != nil {
		logger.Debug("closing publisher channel")
		err := b.publishChan.Close()
		if err != nil {
			logger.Warn("error closing publisher channel", "error", err)
		}
	}

	if b.Conn != nil {
		logger.Debug("closing connection")
		err := b.Conn.Close()
		if err != nil {
			logger.Warn("error closing connection", "error", err)
		}
	}

	logger.Debug("cleared old connection/channels, starting new connection")

	return b.Start()
}

// Start starts the broker connection to the server and starts the message listening/handling
func (b *BrokerConfig) Start() error {
	logger.Info("starting broker connection")
	err := b.dial()
	if err != nil {
		logger.Error("failed to dial AMQP", "error", err)
		return fmt.Errorf("error while dialing AMQP: %w", err)
	}

	// channels
	logger.Debug("creating publish channel")
	b.publishChan, err = b.Conn.Channel()
	if err != nil {
		logger.Error("failed to create publish channel", "error", err)
		return fmt.Errorf("error on opening the publish channel: %w", err)
	}

	logger.Debug("creating consume channel")
	b.consumeChan, err = b.Conn.Channel()
	if err != nil {
		logger.Error("failed to create consume channel", "error", err)
		return fmt.Errorf("error on opeing the consume channel: %w", err)
	}

	// qos on consumer channel only
	logger.Debug("setting QoS parameters", "prefetch", maxMessages)
	err = b.consumeChan.Qos(maxMessages, 0, false)
	if err != nil {
		logger.Error("failed to set QoS", "error", err)
		return fmt.Errorf("error setting the Qos: %w", err)
	}

	// topology
	logger.Info("initializing external access queue", "exchange", ExchangeExternalAccess, "queue", QueueMap)
	b.externalAccessQ, err = b.initQueue(ExchangeExternalAccess, QueueMap, BindingKeyMap)
	if err != nil {
		logger.Error("failed to initialize external access queue", "error", err)
		return err
	}

	logger.Info("initializing resources service queue", "exchange", ExchangeMetadataService, "queue", QueueResources)
	b.resourcesServiceQ, err = b.initQueue(ExchangeMetadataService, QueueResources, BindingKeyMap)
	if err != nil {
		logger.Error("failed to initialize resources service queue", "error", err)
		return err
	}

	// start consumers
	logger.Info("starting message handlers")
	go b.handleMessages(
		b.externalAccessQ,
		ExchangeExternalAccess,
		RkAccessReturn,
		handler.ExternalAccessHandler,
	)
	go b.handleMessages(
		b.resourcesServiceQ,
		ExchangeMetadataService,
		RkMapReturn,
		handler.ResourcesServiceHandler,
	)
	logger.Info("broker successfully started")
	return nil
}

func (b *BrokerConfig) initQueue(exchange, queue, bindingKey string) (*amqp.Queue, error) {
	logger.Debug("initializing queue", "exchange", exchange, "queue", queue, "bindingKey", bindingKey)
	ch, err := b.Conn.Channel()
	if err != nil {
		logger.Error("failed to create channel for queue initialization", "error", err)
		return nil, err
	}

	logger.Debug("declaring exchange", "name", exchange, "type", "topic")
	err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		logger.Error("failed to declare exchange", "exchange", exchange, "error", err)
		return nil, fmt.Errorf("error declaring exchange: %w", err)
	}

	logger.Debug("declaring queue", "name", queue)
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		logger.Error("failed to declare queue", "queue", queue, "error", err)
		return nil, fmt.Errorf("error declaring queue: %w", err)
	}

	logger.Debug("binding queue", "queue", queue, "bindingKey", bindingKey, "exchange", exchange)
	err = ch.QueueBind(q.Name, bindingKey, exchange, false, nil)
	if err != nil {
		logger.Error("failed to bind queue", "queue", queue, "bindingKey", bindingKey, "exchange", exchange, "error", err)
		return nil, fmt.Errorf("error binding queue: %w", err)
	}

	logger.Debug("closing temporary channel used for queue initialization")
	err = ch.Close()
	if err != nil {
		logger.Warn("error closing temp channel when initializing queue (possible leak)", "queue", queue, "bindingKey", bindingKey, "exchange", exchange, "error", err)
	}
	logger.Info("queue initialized successfully", "exchange", exchange, "queue", queue, "bindingKey", bindingKey)
	return &q, nil
}

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		logger.Debug("environment variable found", "name", k, "value", v)
		return v
	}
	logger.Info("env variable not found, using default", "name", k, "default", def)
	return def
}

func envInt(key string, defaultVal int) int {
	strVal := env(key, strconv.FormatInt(int64(defaultVal), 10))
	if strVal == "" {
		logger.Debug("environment variable not set, using default", "name", key, "default", defaultVal)
		return defaultVal
	}

	val, err := strconv.Atoi(strVal)
	if err != nil {
		logger.Warn("invalid integer value, using default", "name", key, "value", strVal, "error", err, "default", defaultVal)
		return defaultVal
	}

	return val
}

func failOnError(err error, msg string) {
	if err != nil {
		logger.Error(msg, "error", err)
		panic(fmt.Sprintf("%s: %v", msg, err))
	}
}

// Monitor start monitoring a broker config for connection closing. If it happens, it will create a new connection and start it.
func (b *BrokerConfig) Monitor(ctx context.Context) {
	logger.Info("starting connection monitor")
	for {
		// listen to the current connection
		logger.Debug("setting up connection close notification channel")
		closeC := b.Conn.NotifyClose(make(chan *amqp.Error, 1))

		// wait for either a close event or a shutdown signal
		logger.Debug("waiting for close events or shutdown signal")
		select {
		case err, ok := <-closeC:
			// conn closed intentionally (Conn.Close)
			if !ok {
				logger.Info("connection closed intentionally, monitor shutting down")
				return
			}
			logger.Error("connection closed unexpectedly", "error", err)

			// reconnect reusing the same broker pointer
			backoff := time.Second
			logger.Info("attempting to reconnect", "max_attempts", maxReconnectAttempts)
			for attempt := 1; attempt <= maxReconnectAttempts; attempt++ {
				logger.Debug("reconnection attempt", "attempt", attempt, "backoff", backoff)
				err := b.Restart()
				if err != nil {
					logger.Error("reconnection failed", "attempt", attempt, "error", err)
					// exponential back‑off
					time.Sleep(backoff)
					backoff *= 2
					continue
				}
				logger.Info("reconnection successful", "attempt", attempt)
				break
			}

		case <-ctx.Done():
			// propagate graceful shutdown
			logger.Info("received shutdown signal, closing connection")
			_ = b.Conn.Close()
			logger.Info("monitor shutting down")
			return
		}
	}
}
