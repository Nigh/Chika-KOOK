module Chika

go 1.18

require (
	github.com/jpillora/overseer v1.1.6
	github.com/lonelyevil/khl v0.0.20
	github.com/lonelyevil/khl/log_adapter/plog v0.0.20
	github.com/phuslu/log v1.0.71
	github.com/spf13/viper v1.10.1
	local/khlcard v0.0.0-00010101000000-000000000000
)

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.0.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jpillora/s3 v1.1.4 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace local/khlcard => ./kcard
