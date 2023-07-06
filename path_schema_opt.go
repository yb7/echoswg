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
