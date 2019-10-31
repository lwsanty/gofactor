module github.com/lwsanty/gofactor

go 1.13

require (
	github.com/bblfsh/go-driver/v2 v2.7.3
	github.com/bblfsh/sdk/v3 v3.3.1
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.4 // indirect
)

// FIXME https://github.com/bblfsh/go-driver/issues/67
replace github.com/bblfsh/go-driver/v2 v2.7.3 => /home/lwsanty/goproj/bblfsh/go-driver

replace github.com/bblfsh/sdk/v3 => github.com/lwsanty/sdk/v3 v3.2.1-0.20191101155937-e335ce2434f4
