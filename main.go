package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"

	// handler
	_accountHttpDelivery "main/atm/delivery/http"
	_authenticationHttpDelivery "main/atm/delivery/http"
	_transactionHttpDelivery "main/atm/delivery/http"
	_userHttpDelivery "main/atm/delivery/http"
	_httpDeliveryMiddleware "main/atm/delivery/http/middleware"

	// service
	_accountUcase "main/atm/usecase"
	_authenticationUcase "main/atm/usecase"
	_externalUcase "main/atm/usecase"
	_notificationUcase "main/atm/usecase"
	_pollingUcase "main/atm/usecase"
	_userUcase "main/atm/usecase"

	// repository
	_accountRepo "main/atm/repository/mysql"
	_authenticationRepo "main/atm/repository/mysql"
	_transactionRepo "main/atm/repository/mysql"
	_userRepo "main/atm/repository/mysql"

	// logging
	"main/logger"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

func main() {
	logger.Info("start program...")

	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	dbconnection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "true")
	val.Add("loc", "Asia/Bangkok")
	dsn := fmt.Sprintf("%s?%s", dbconnection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)

	if err != nil {
		log.Fatal(err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	redisHost := viper.GetString(`redis.host`)
	redisdbPort := viper.GetString(`redis.port`)
	redisdbPass := viper.GetString(`redis.pass`)

	addr := fmt.Sprintf("%s:%s", redisHost, redisdbPort)
	password := redisdbPass

	redis := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// elasticHost := viper.GetString(`elastic.host`)
	// elasticPort := viper.GetString(`elastic.port`)
	// elasticAddr := fmt.Sprintf("%s:%s", elasticHost, elasticPort)

	// client, err := elastic.NewClient(elastic.SetURL(elasticAddr), elastic.SetSniff(false))
	// if err != nil {
	// 	panic(err)
	// }
	// client.Ping("Pong")

	// pong, err := redis.Ping().Result()

	// kafkaHost := viper.GetString(`kafka.host`)
	// kafkaPort := viper.GetString(`kafka.port`)

	config := sarama.NewConfig()
	config.ClientID = "my-kafka-client"
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	// config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Replace with your Kafka brokers' addresses
	brokers := []string{viper.GetString("kafka.broker_address")}
	kafkaClient, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaClient.Close()

	// producer, err := producer.CreateProducer("localhost:9092")
	// if err != nil {
	// 	log.Fatal("Error creating producer:", err)
	// }
	// defer producer.Close()

	// consumer, err := consumer.CreateConsumer("localhost:9092", "test-topic")
	// if err != nil {
	// 	log.Fatal("Error creating consumer:", err)
	// }
	// defer consumer.Close()

	// fmt.Println("Kafka ping successful")

	e := echo.New()
	middL := _httpDeliveryMiddleware.InitMiddleware()
	e.Use(middL.CORS)
	e.Use(middL.RateLimitMiddleware)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3001", "http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	ar := _accountRepo.NewMysqlAccountRepository(dbConn, redis)
	authr := _authenticationRepo.NewMysqlAuthenticationRepository(dbConn, redis)
	ur := _userRepo.NewMysqlUserRepository(dbConn)
	tr := _transactionRepo.NewMysqlTransactionRepository(dbConn, redis)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	au := _accountUcase.NewAccountUsecase(ar, tr, redis, timeoutContext)
	auth := _authenticationUcase.NewAuthenticationUsecase(authr, timeoutContext)
	uu := _userUcase.NewUserUsecase(ur, timeoutContext)
	tu := _accountUcase.NewTransactionUsecase(tr, au, timeoutContext, redis, kafkaClient)
	nu := _notificationUcase.NewNotificationUsecase(tu, timeoutContext, kafkaClient)
	xu := _externalUcase.NewExternalUsecase(timeoutContext, kafkaClient)

	_accountHttpDelivery.NewAccountHandler(e, au)
	_authenticationHttpDelivery.NewAuthenticationHandler(e, auth)
	_userHttpDelivery.NewUserHandler(e, uu, auth)
	_transactionHttpDelivery.NewTransactionHandler(e, tu, redis)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go nu.SendTransactionNotification(ctx)
	go xu.SendSms(ctx)
	go tu.ConsumeScheduledTransaction(ctx)

	//polling service init
	pollingInterval := 15 * time.Second

	pu := _pollingUcase.NewPollingUsecase(tu, pollingInterval)
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go pu.Polling(ctx, &wg, pollingInterval, stopChan)

	log.Fatal(e.Start(viper.GetString("server.address"))) //nolint

	sigchan := make(chan os.Signal, 1) // Wait for OS signals (e.g., Ctrl+C) to gracefully stop the consumer
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan
	time.Sleep(1 * time.Minute) // Let the polling routine run for a specified duration

	close(stopChan) // Signal the polling routine to stop

	wg.Wait() // Wait for the polling routine to finish before exiting
}
