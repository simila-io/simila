package persistence

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"time"
)

type (
	DbOption func(cfg *DbConfig)

	DbConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		DbName   string
		SslMode  string
	}

	ContainerDb interface {
		io.Closer
		DbConfig() DbConfig
	}

	containerDb struct {
		c     testcontainers.Container
		dbCfg DbConfig
	}
)

func WithHost(host string) DbOption {
	return func(c *DbConfig) {
		c.Host = host
	}
}

func WithPort(port string) DbOption {
	return func(c *DbConfig) {
		c.Port = port
	}
}

func WithDbName(dbName string) DbOption {
	return func(c *DbConfig) {
		c.DbName = dbName
	}
}

func (ds DbConfig) DataSourceNoDb() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s",
		ds.Host, ds.Port, ds.User, ds.Password, ds.SslMode)
}

func (ds DbConfig) DataSourceFull() string {
	return fmt.Sprintf("%s dbname=%s", ds.DataSourceNoDb(), ds.DbName)
}

func NewNilContainerDb(opts ...DbOption) (ContainerDb, error) {
	dbCfg := DbConfig{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		SslMode:  "disable",
	}
	for _, opt := range opts {
		opt(&dbCfg)
	}
	return &containerDb{dbCfg: dbCfg}, nil
}

func (cd containerDb) Close() error {
	if cd.c == nil {
		return nil
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()
	return cd.c.Terminate(ctx)
}

func (cd containerDb) DbConfig() DbConfig {
	return cd.dbCfg
}

// NewPgContainerDb runs pg database in a docker container.
func NewPgContainerDb(image string, opts ...DbOption) (ContainerDb, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dbCfg := DbConfig{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		SslMode:  "disable",
	}
	for _, opt := range opts {
		opt(&dbCfg)
	}

	natPort, _ := nat.NewPort("tcp", dbCfg.Port)
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{string(natPort)},
		HostConfigModifier: func(config *container.HostConfig) {
			config.AutoRemove = true
		},
		Env: map[string]string{
			"POSTGRES_USER":     dbCfg.User,
			"POSTGRES_PASSWORD": dbCfg.Password,
			"POSTGRES_DB":       dbCfg.DbName,
		},
		WaitingFor: wait.ForListeningPort(natPort),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	mappedPort, err := c.MappedPort(ctx, natPort)
	if err != nil {
		return nil, err
	}
	dbCfg.Port = mappedPort.Port()
	return &containerDb{c: c, dbCfg: dbCfg}, nil
}
