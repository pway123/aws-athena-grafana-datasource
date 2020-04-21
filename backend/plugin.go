package main

import (
	"github.com/grafana/grafana-plugin-model/go/datasource"
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
)

var pluginLogger = hclog.New(&hclog.LoggerOptions{
	Name:  PluginName,
	Level: hclog.LevelFromString("DEBUG"),
})

func main() {
	pluginLogger.Debug("Running AWS Athena backend datasource")

	plugin.Serve(&plugin.ServeConfig{

		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			PluginName: &datasource.DatasourcePluginImpl{
				Plugin: &AwsAthenaDatasource{
					logger: pluginLogger,
					athena: &AwsAthenaQueryHandler{
						logger: pluginLogger,
						cache:  make(map[string]*QueryCacheInfo),
					},
				},
			},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
