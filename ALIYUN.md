# Aliyun API Gateway Support

This library now supports generating OpenAPI specifications compatible with Aliyun API Gateway.

## Overview

Aliyun API Gateway requires specific OpenAPI extensions to properly configure API routes. This library automatically adds the necessary `x-aliyun-apigateway-*` extensions when Aliyun mode is enabled.

## Enabling Aliyun Mode

Set the `ECHOSWG_ALIYUN_GATEWAY` environment variable to `on`:

```bash
export ECHOSWG_ALIYUN_GATEWAY=on
```

## Features

When Aliyun mode is enabled, the following OpenAPI extensions are automatically added to each operation:

### 1. Request Configuration (`x-aliyun-apigateway-request-config`)

This extension configures how the API Gateway handles incoming requests:
- `requestProtocol`: Set to "HTTP"
- `requestHttpMethod`: The HTTP method (GET, POST, etc.)
- `requestPath`: The path pattern
- `requestMode`: Set to "PASSTHROUGH" for direct forwarding

### 2. Backend Configuration (`x-aliyun-apigateway-backend`)

This extension configures the backend service that handles the request:
- `serviceType`: Type of backend service (HTTP, FUNCTION, MOCK, etc.)
- `serviceAddress`: Backend service address
- `servicePath`: Backend service path
- `serviceMethod`: HTTP method for backend request

## Usage Examples

### Basic Usage

```go
package main

import (
    "os"
    "github.com/labstack/echo/v4"
    "github.com/yb7/echoswg"
)

func main() {
    // Enable Aliyun mode
    os.Setenv("ECHOSWG_ALIYUN_GATEWAY", "on")
    
    e := echo.New()
    
    // Create an API group
    petGroup := echoswg.NewApiGroup(e, "pet", "/api/v1/pets")
    
    // Define a simple GET endpoint
    // The Aliyun extensions will be added automatically
    petGroup.GET("/:id", 
        echoswg.WithSummary("Get pet by ID"),
        echoswg.WithDescription("Retrieve a pet by its ID"),
        func(req *GetPetRequest) (*PetResponse, error) {
            // Your handler logic
            return &PetResponse{ID: req.ID}, nil
        },
    )
    
    e.Start(":8080")
}

type GetPetRequest struct {
    ID int64 `param:"id"`
}

type PetResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

### Custom Backend Configuration

You can specify custom backend configurations for each endpoint:

```go
// Configure a specific backend service
petGroup.GET("/:id",
    echoswg.WithSummary("Get pet by ID"),
    echoswg.WithAliyunHttpBackend(
        "http://backend.example.com",  // Backend service address
        "/internal/pets",               // Backend path
    ),
    func(req *GetPetRequest) (*PetResponse, error) {
        return &PetResponse{ID: req.ID}, nil
    },
)
```

### Advanced Backend Configuration

For more control, use `WithAliyunBackend`:

```go
petGroup.POST("/",
    echoswg.WithSummary("Create pet"),
    echoswg.WithAliyunBackend(
        "HTTP",                        // Service type
        "http://backend.example.com",  // Service address
        "/internal/pets",              // Service path
        "POST",                        // Service method
    ),
    func(req *CreatePetRequest) (*PetResponse, error) {
        return &PetResponse{ID: 123}, nil
    },
)
```

### Function Compute Backend

To use Aliyun Function Compute as the backend:

```go
petGroup.GET("/:id",
    echoswg.WithSummary("Get pet by ID"),
    echoswg.WithAliyunBackend(
        "FUNCTION",                    // Service type
        "acs:fc:cn-hangzhou:123456:services/myService.LATEST/functions/myFunction",
        "",                            // Service path (not needed for Function Compute)
        "POST",                        // Function Compute typically uses POST
    ),
    func(req *GetPetRequest) (*PetResponse, error) {
        return &PetResponse{ID: req.ID}, nil
    },
)
```

## Generated OpenAPI Extensions

When you enable Aliyun mode, the generated OpenAPI spec will include extensions like:

```json
{
  "paths": {
    "/api/v1/pets/{id}": {
      "get": {
        "summary": "Get pet by ID",
        "operationId": "pet.GetPet",
        "x-aliyun-apigateway-request-config": {
          "requestProtocol": "HTTP",
          "requestHttpMethod": "GET",
          "requestPath": "/api/v1/pets/{id}",
          "requestMode": "PASSTHROUGH"
        },
        "x-aliyun-apigateway-backend": {
          "serviceType": "HTTP",
          "serviceAddress": "http://backend.example.com",
          "servicePath": "/internal/pets",
          "serviceMethod": "GET"
        }
      }
    }
  }
}
```

## Disabling Aliyun Mode

To generate standard OpenAPI specs without Aliyun extensions, simply don't set the environment variable, or set it to something other than "on":

```bash
unset ECHOSWG_ALIYUN_GATEWAY
# or
export ECHOSWG_ALIYUN_GATEWAY=off
```

## Notes

- The Aliyun extensions are only added when `ECHOSWG_ALIYUN_GATEWAY=on`
- If no custom backend is specified, a default HTTP backend configuration is used
- The generated OpenAPI spec is compatible with both Aliyun API Gateway and standard OpenAPI tools
- All standard echoswg features continue to work when Aliyun mode is enabled
