module github.com/ToshihitoKon/slack-quickpost

go 1.26

require (
	github.com/slack-go/slack v0.30.1
	github.com/spf13/pflag v1.0.10
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/gorilla/websocket v1.5.3 // indirect

// slack-quickpost-embed --source module fails outside a module directory.
retract v0.9.0
