package main

import (
	"context"
	"flag"
	"log"
	"order/src/application/commands"
	"order/src/application/events"
	"order/src/controllers"
	order_nats "order/src/nats"
	"order/src/repositories"
	"order/src/routers"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JohnSalazar/microservices-go-common/config"
	common_grpc_client "github.com/JohnSalazar/microservices-go-common/grpc/email/client"
	helpers "github.com/JohnSalazar/microservices-go-common/helpers"
	"github.com/JohnSalazar/microservices-go-common/httputil"
	common_log "github.com/JohnSalazar/microservices-go-common/logs"
	"github.com/JohnSalazar/microservices-go-common/middlewares"
	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_repositories "github.com/JohnSalazar/microservices-go-common/repositories"
	common_security "github.com/JohnSalazar/microservices-go-common/security"
	common_services "github.com/JohnSalazar/microservices-go-common/services"
	common_tasks "github.com/JohnSalazar/microservices-go-common/tasks"
	common_validator "github.com/JohnSalazar/microservices-go-common/validators"
	"github.com/nats-io/nats.go"

	provider "github.com/JohnSalazar/microservices-go-common/trace/otel/jaeger"

	"go.mongodb.org/mongo-driver/mongo"

	common_consul "github.com/JohnSalazar/microservices-go-common/consul"
	consul "github.com/hashicorp/consul/api"
)

type Main struct {
	config              *config.Config
	client              *mongo.Client
	natsConn            *nats.Conn
	securityKeysService common_services.SecurityKeysService
	managerCertificates common_security.ManagerCertificates
	adminMongoDbService *common_services.AdminMongoDbService
	httpServer          httputil.HttpServer
	consulClient        *consul.Client
	serviceID           string
}

func NewMain(
	config *config.Config,
	client *mongo.Client,
	natsConn *nats.Conn,
	securityKeysService common_services.SecurityKeysService,
	managerCertificates common_security.ManagerCertificates,
	adminMongoDbService *common_services.AdminMongoDbService,
	httpServer httputil.HttpServer,
	consulClient *consul.Client,
	serviceID string,
) *Main {
	return &Main{
		config:              config,
		client:              client,
		natsConn:            natsConn,
		securityKeysService: securityKeysService,
		managerCertificates: managerCertificates,
		adminMongoDbService: adminMongoDbService,
		httpServer:          httpServer,
		consulClient:        consulClient,
		serviceID:           serviceID,
	}
}

var production *bool
var disableTrace *bool

func main() {
	production = flag.Bool("prod", false, "use -prod=true to run in production mode")
	disableTrace = flag.Bool("disable-trace", false, "use disable-trace=true if you want to disable tracing completly")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	app, err := startup(ctx)
	if err != nil {
		panic(err)
	}

	err = app.client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer app.client.Disconnect(ctx)

	err = app.client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	defer app.natsConn.Close()

	providerTracer, err := provider.NewProvider(provider.ProviderConfig{
		JaegerEndpoint: app.config.Jaeger.JaegerEndpoint,
		ServiceName:    app.config.Jaeger.ServiceName,
		ServiceVersion: app.config.Jaeger.ServiceVersion,
		Production:     *production,
		Disabled:       *disableTrace,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer providerTracer.Close(ctx)
	log.Println("Connected to Jaegger")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	userMongoExporter, err := app.adminMongoDbService.VerifyMongoDBExporterUser()
	if err != nil {
		log.Fatal(err)
	}

	if !userMongoExporter {
		log.Fatal("MongoDB Exporter user not found!")
	}

	app.httpServer.RunTLSServer()

	<-done
	err = app.consulClient.Agent().ServiceDeregister(app.serviceID)
	if err != nil {
		log.Printf("consul deregister error: %s", err)
	}

	log.Print("Server Stopped")
	os.Exit(0)
}

func startup(ctx context.Context) (*Main, error) {
	logger := common_log.NewLogger()
	config := config.LoadConfig(*production, "./config/")
	helpers.CreateFolder(config.Folders)
	common_validator.NewValidator("en")

	consulClient, serviceID, err := common_consul.NewConsulClient(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	checkServiceName := common_tasks.NewCheckServiceNameTask()

	certificateServiceNameDone := make(chan bool)
	go checkServiceName.ReloadServiceName(
		ctx,
		config,
		consulClient,
		config.Certificates.ServiceName,
		common_consul.CertificatesAndSecurityKeys,
		certificateServiceNameDone)
	<-certificateServiceNameDone

	emailsServiceNameDone := make(chan bool)
	go checkServiceName.ReloadServiceName(
		ctx,
		config,
		consulClient,
		config.EmailService.ServiceName,
		common_consul.EmailService,
		emailsServiceNameDone)
	<-emailsServiceNameDone

	metricService, err := common_services.NewMetricsService(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	client, err := repositories.NewMongoClient(config)
	if err != nil {
		return nil, err
	}

	certificatesService := common_services.NewCertificatesService(config)
	managerCertificates := common_security.NewManagerCertificates(config, certificatesService)
	emailService := common_grpc_client.NewEmailServiceClientGrpc(config, certificatesService)

	checkCertificates := common_tasks.NewCheckCertificatesTask(config, managerCertificates, emailService)
	certsDone := make(chan bool)
	go checkCertificates.Start(ctx, certsDone)
	<-certsDone

	nc, err := common_nats.NewNats(config, certificatesService)
	if err != nil {
		log.Fatalf("Nats connect error: %+v", err)
	}
	log.Printf("Nats Connected Status: %+v	", nc.Status().String())

	orderSubjects := []string{
		string(common_nats.OrderCreate),
		string(common_nats.OrderStatus),
	}

	_, err = common_nats.NewJetStream(nc, "order", orderSubjects)
	if err != nil {
		log.Fatalf("Nats JetStream create error: %+v", err)
	}

	storeSubjects := []string{string(common_nats.StoreBooked)}
	js, err := common_nats.NewJetStream(nc, "store2", storeSubjects)
	if err != nil {
		log.Fatalf("Nats JetStream create error: %+v", err)
	}

	natsPublisher := common_nats.NewPublisher(js)

	database := repositories.NewMongoDatabase(config, client)
	adminMongoDbRepository := common_repositories.NewAdminMongoDbRepository(database)
	adminMongoDbService := common_services.NewAdminMongoDbService(config, adminMongoDbRepository)

	orderRepository := repositories.NewOrderRepository(database)

	securityKeysService := common_services.NewSecurityKeysService(config, certificatesService)
	managerSecurityKeys := common_security.NewManagerSecurityKeys(config, securityKeysService)
	managerTokens := common_security.NewManagerTokens(config, managerSecurityKeys)

	orderEventHandler := events.NewOrderEventHandler(emailService, natsPublisher)
	orderCommandHandler := commands.NewOrderCommandHandler(orderRepository, orderEventHandler)

	listens := order_nats.NewListen(
		config,
		js,
		orderCommandHandler,
		emailService)

	listens.Listen()

	authentication := middlewares.NewAuthentication(logger, managerTokens)
	orderController := controllers.NewOrderController(orderRepository)
	router := routers.NewRouter(config, metricService, authentication, orderController)
	httpServer := httputil.NewHttpServer(config, router.RouterSetup(), certificatesService)
	app := NewMain(
		config,
		client,
		nc,
		securityKeysService,
		managerCertificates,
		adminMongoDbService,
		httpServer,
		consulClient,
		serviceID,
	)

	return app, nil
}
