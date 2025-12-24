package echoswg

type PathSchemaOption func(*SwaggerPathDefine)

func WithDescription(description string) PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.Description = description
	}
}

func WithOperationId(operationId string) PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.OperationId = operationId
	}
}
func WithSummary(summary string) PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.Summary = summary
	}
}

func WithInternalHttpTraceEnabled() PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.InternalHttpTraceEnabled = true
	}
}
func WithInternalHttpTraceDisabled() PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.InternalHttpTraceEnabled = false
	}
}

// WithAliyunBackend sets the Aliyun API Gateway backend configuration
func WithAliyunBackend(serviceType, serviceAddress, servicePath, serviceMethod string) PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.AliyunBackend = &AliyunBackendConfig{
			ServiceType:    serviceType,
			ServiceAddress: serviceAddress,
			ServicePath:    servicePath,
			ServiceMethod:  serviceMethod,
		}
	}
}

// WithAliyunHttpBackend sets HTTP backend for Aliyun API Gateway
func WithAliyunHttpBackend(serviceAddress, servicePath string) PathSchemaOption {
	return func(pathDefine *SwaggerPathDefine) {
		pathDefine.AliyunBackend = &AliyunBackendConfig{
			ServiceType:    "HTTP",
			ServiceAddress: serviceAddress,
			ServicePath:    servicePath,
		}
	}
}
