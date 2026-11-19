// Добавить в cmd/<appname>/swagger.go после первого `make swag`.
// В run() после config.Load() вызвать configureSwagger(cfg).
package main

import (
	appconfig "<module>/internal/config"
	docs "<module>/internal/transport/<appname>/http/docs"
)

func configureSwagger(cfg *appconfig.Config) {
	docs.SwaggerInfo.Host = cfg.HTTP.SwaggerHost
	docs.SwaggerInfo.BasePath = cfg.HTTP.SwaggerBasePath
	docs.SwaggerInfo.Version = cfg.Logger.Version
}
