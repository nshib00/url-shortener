package config

import "fmt"

const (
	// Пути к config-файлам относительно корня проекта
	pathLocalStr = "config/local.yaml"
	pathProdStr  = "config/prod.yaml"
)

type EnvType int

const (
	EnvUnknown = iota - 1
	EnvLocal
	EnvProd
)

func (et EnvType) String() string {
	switch et {
	case EnvLocal:
		return pathLocalStr
	case EnvProd:
		return pathProdStr
	default:
		return ""
	}
}

func ParseEnvType(s string) (EnvType, error) {
	switch s {
	case "local":
		return EnvLocal, nil
	case "prod":
		return EnvProd, nil
	default:
		return EnvUnknown, fmt.Errorf("config: unknown env type")
	}
}
