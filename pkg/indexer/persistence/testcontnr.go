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

	DbContainer interface {
		io.Closer
		DbConfig() DbConfig
	}

	dbContainer struct {
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

func NewNilDbContainer(opts ...DbOption) (DbContainer, error) {
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
	return &dbContainer{dbCfg: dbCfg}, nil
}

func (dc dbContainer) Close() error {
	if dc.c == nil {
		return nil
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFn()
	return dc.c.Terminate(ctx)
}

func (dc dbContainer) DbConfig() DbConfig {
	return dc.dbCfg
}

// NewPgDbContainer runs pg database in a docker container.
func NewPgDbContainer(ctx context.Context, image string, opts ...DbOption) (DbContainer, error) {
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
	return &dbContainer{c: c, dbCfg: dbCfg}, nil
}
