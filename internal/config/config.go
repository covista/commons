package config

type Config struct {
	GRPC     GRPC
	HTTP     HTTP
	Database Database
}

type Database struct {
	Host     string
	Database string
	User     string
	Password string
	Port     string
}

type GRPC struct {
	ListenAddress string
	Port          string
}

type HTTP struct {
	ListenAddress string
	Port          string
}
