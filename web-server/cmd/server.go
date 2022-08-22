package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/danielpenchev98/UShare/web-server/api/rest"
	"github.com/danielpenchev98/UShare/web-server/internal/auth"
	cronJob "github.com/danielpenchev98/UShare/web-server/internal/cron"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dao"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dbconn"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"github.com/danielpenchev98/UShare/web-server/internal/middleware"
	val "github.com/danielpenchev98/UShare/web-server/internal/validator"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

const (
	hostParamName     = "HOST"
	portParamName     = "PORT"
	groupDirParamName = "GROUP_DIR"
)

type ServerConfig struct {
	Host string
	Port int
}

var groupDirPath string

func main() {
	serverCfg, err := getServerConfig()
	if err != nil {
		log.Fatalf("Proble with the server config. Reason %s", err)
	}

	if err = createGroupsDir(); err != nil {
		log.Fatal(err)
	}

	httpServer := createHttpServer(serverCfg.Host, serverCfg.Port)
	asyncJob := createCronJob()
	asyncJob.Start()
	defer asyncJob.Stop()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(errors.Wrapf(err, "server listen-and-serve failed"))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	log.Println("shutting down http server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		panic(errors.Wrapf(err, "failed to shutdown server"))
	}
	<-ctx.Done()
}

func getServerConfig() (ServerConfig, error) {
	portStr := os.Getenv(portParamName)
	if portStr == "" {
		return ServerConfig{}, errors.Errorf("Please set %s env variable", portParamName)
	}

	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		return ServerConfig{}, errors.Errorf("The env variable %s has illegal port number", portParamName)
	}

	return ServerConfig{
		Host: os.Getenv(hostParamName),
		Port: portNum,
	}, nil
}

func createGroupsDir() error {
	currDir := os.Getenv("GROUP_DIR")
	if currDir == "" {
		return errors.New("Please set GROUP_DIR is not set")
	}

	groupDirPath = currDir + "/groups"
	if _, err := os.Stat(groupDirPath); err == nil {
		return nil
	}

	if err := os.Mkdir(groupDirPath, 0755); err != nil {
		return myerr.NewServerErrorWrap(err, "Couldnt create directory for the groups")
	}

	return nil
}

func createUamDAO() dao.UamDAO {
	dbConn, err := dbconn.GetDBConn(dbconn.PostgresDialectorCreator)
	if err != nil {
		log.Fatal(myerr.NewServerErrorWrap(err, "Couldnt create a connection to the database"))
	}

	uamDAO := dao.NewUamDAOImpl(dbConn)
	if err = uamDAO.Migrate(); err != nil {
		log.Fatal(myerr.NewServerErrorWrap(err, "Couldnt migrate the database schemas"))
	}

	return uamDAO
}

func createFmDAO() dao.FmDAO {
	dbConn, err := dbconn.GetDBConn(dbconn.PostgresDialectorCreator)
	if err != nil {
		log.Fatal(myerr.NewServerErrorWrap(err, "Couldnt create a connection to the database"))
	}

	fmDAO := dao.NewFmDAOImpl(dbConn)
	if err = fmDAO.Migrate(); err != nil {
		log.Fatal(myerr.NewServerErrorWrap(err, "Couldnt migrate the database schemas"))
	}

	return fmDAO
}

func createHttpServer(host string, port int) *http.Server {
	var router = gin.Default()

	jwtCreator, err := auth.NewJwtCreatorImpl()
	if err != nil {
		log.Fatal(myerr.NewServerErrorWrap(err, "Couldnt create a new Jwt Creator"))
	}

	filter := middleware.NewAuthzFilterImpl(jwtCreator)
	uamEndpoint := rest.NewUamEndPointImpl(createUamDAO(), jwtCreator, val.NewBasicValidator(), groupDirPath)
	fmEndpoint := rest.NewFileManagementEndpointImpl(createUamDAO(), createFmDAO(), groupDirPath)

	v1 := router.Group("/v1")
	{
		public := v1.Group("/public")
		{
			public.GET("/healthcheck", rest.CheckHealth)
			public.POST("/user/registration", uamEndpoint.CreateUser)
			public.POST("/user/login", uamEndpoint.Login)
		}

		protected := v1.Group("/protected").Use(filter.Authz)
		{
			protected.DELETE("/group/membership/revocation", uamEndpoint.RevokeMembership)
			protected.POST("/group/creation", uamEndpoint.CreateGroup)
			protected.POST("/group/invitation", uamEndpoint.AddMember)
			protected.DELETE("/group/user/deletion", uamEndpoint.DeleteUser)
			protected.DELETE("/group/deletion", uamEndpoint.DeleteGroup)
			protected.POST("/group/file/upload", fmEndpoint.UploadFile)
			protected.GET("/group/file/download", fmEndpoint.DownloadFile)
			protected.DELETE("/group/file/deletion", fmEndpoint.DeleteFile)
			protected.GET("/group/files", fmEndpoint.RetrieveAllFilesInfo)
			protected.GET("/groups", uamEndpoint.GetAllGroupsInfo)
			protected.GET("/users", uamEndpoint.GetAllUsersInfo)
			protected.GET("/group/users", uamEndpoint.GetAllUsersInGroup)
		}
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: router,
	}

	return httpServer
}

func createCronJob() *cron.Cron {
	groupDeleter := cronJob.NewGroupEraserJobImpl(createUamDAO(), groupDirPath)
	asyncJob := cron.New()
	asyncJob.AddFunc("@every 1m", groupDeleter.DeleteGroups)
	return asyncJob
}
