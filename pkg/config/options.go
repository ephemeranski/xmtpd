package config

import (
	"time"
)

type ApiOptions struct {
	Port int `short:"p" long:"port" description:"Port to listen on" env:"XMTPD_API_PORT" default:"5050"`
}

type ContractsOptions struct {
	RpcUrl                         string        `long:"rpc-url"                  env:"XMTPD_CONTRACTS_RPC_URL"                  description:"Blockchain RPC URL"`
	NodesContractAddress           string        `long:"nodes-address"            env:"XMTPD_CONTRACTS_NODES_ADDRESS"            description:"Node contract address"`
	MessagesContractAddress        string        `long:"messages-address"         env:"XMTPD_CONTRACTS_MESSAGES_ADDRESS"         description:"Message contract address"`
	IdentityUpdatesContractAddress string        `long:"identity-updates-address" env:"XMTPD_CONTRACTS_IDENTITY_UPDATES_ADDRESS" description:"Identity updates contract address"`
	ChainID                        int           `long:"chain-id"                 env:"XMTPD_CONTRACTS_CHAIN_ID"                 description:"Chain ID for the appchain"               default:"31337"`
	RefreshInterval                time.Duration `long:"refresh-interval"         env:"XMTPD_CONTRACTS_REFRESH_INTERVAL"         description:"Refresh interval for the nodes registry" default:"60s"`
}

type DbOptions struct {
	ReaderConnectionString string        `long:"reader-connection-string" env:"XMTPD_DB_READER_CONNECTION_STRING" description:"Reader connection string"`
	WriterConnectionString string        `long:"writer-connection-string" env:"XMTPD_DB_WRITER_CONNECTION_STRING" description:"Writer connection string"                       required:"true"`
	ReadTimeout            time.Duration `long:"read-timeout"             env:"XMTPD_DB_READ_TIMEOUT"             description:"Timeout for reading from the database"                          default:"10s"`
	WriteTimeout           time.Duration `long:"write-timeout"            env:"XMTPD_DB_WRITE_TIMEOUT"            description:"Timeout for writing to the database"                            default:"10s"`
	MaxOpenConns           int           `long:"max-open-conns"           env:"XMTPD_DB_MAX_OPEN_CONNS"           description:"Maximum number of open connections"                             default:"80"`
	WaitForDB              time.Duration `long:"wait-for"                 env:"XMTPD_DB_WAIT_FOR"                 description:"wait for DB on start, up to specified duration"`
}

// MetricsOptions are settings used to start a prometheus server
type MetricsOptions struct {
	Enable  bool   `long:"enable"          env:"XMTPD_METRICS_ENABLE"          description:"Enable the metrics server"`
	Address string `long:"metrics-address" env:"XMTPD_METRICS_METRICS_ADDRESS" description:"Listening address of the metrics server"   default:"127.0.0.1"`
	Port    int    `long:"metrics-port"    env:"XMTPD_METRICS_METRICS_PORT"    description:"Listening HTTP port of the metrics server" default:"8008"`
}

type PayerOptions struct {
	PrivateKey string `long:"private-key" env:"XMTPD_PAYER_PRIVATE_KEY" description:"Private key used to sign blockchain transactions"`
}

type MlsValidationOptions struct {
	GrpcAddress string `long:"grpc-address" env:"XMTPD_MLS_VALIDATION_GRPC_ADDRESS" description:"Address of the MLS validation service"`
}

// TracingOptions are settings controlling collection of DD APM traces and error tracking.
type TracingOptions struct {
	Enable bool `long:"enable" env:"XMTPD_TRACING_ENABLE" description:"Enable DD APM trace collection"`
}

// ReflectionOptions are settings controlling collection of GRPC reflection settings.
type ReflectionOptions struct {
	Enable bool `long:"enable" env:"XMTPD_REFLECTION_ENABLE" description:"Enable GRPC reflection"`
}

type LogOptions struct {
	LogLevel    string `short:"l" long:"log-level"    env:"XMTPD_LOG_LEVEL"    description:"Define the logging level, supported strings are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and their lower-case forms." default:"INFO"`
	LogEncoding string `          long:"log-encoding" env:"XMTPD_LOG_ENCODING" description:"Log encoding format. Either console or json"                                                                                  default:"console"`
}

type SignerOptions struct {
	PrivateKey string `long:"private-key" env:"XMTPD_SIGNER_PRIVATE_KEY" description:"Private key used to sign messages" required:"true"`
}

type GenerateKeyOptions struct{}

type RegisterNodeOptions struct {
	HttpAddress     string `long:"http-address"      description:"HTTP address to register for the node"                            required:"true"`
	OwnerAddress    string `long:"owner-address"     description:"Blockchain address of the intended owner of the registration NFT" required:"true"`
	AdminPrivateKey string `long:"admin-private-key" description:"Private key of the admin to register the node"                    required:"true"`
	SigningKey      string `long:"signing-key"       description:"Signing key of the node to register"                              required:"true"`
}

type ServerOptions struct {
	API           ApiOptions           `group:"API Options"            namespace:"api"`
	DB            DbOptions            `group:"Database Options"       namespace:"db"`
	Contracts     ContractsOptions     `group:"Contracts Options"      namespace:"contracts"`
	Metrics       MetricsOptions       `group:"Metrics Options"        namespace:"metrics"`
	Payer         PayerOptions         `group:"Payer Options"          namespace:"payer"`
	Reflection    ReflectionOptions    `group:"Reflection Options"     namespace:"reflection"`
	Tracing       TracingOptions       `group:"DD APM Tracing Options" namespace:"tracing"`
	MlsValidation MlsValidationOptions `group:"MLS Validation Options" namespace:"mls-validation"`
	Log           LogOptions           `group:"Log Options"            namespace:"log"`
	Signer        SignerOptions        `group:"Signer Options"         namespace:"signer"`
}
