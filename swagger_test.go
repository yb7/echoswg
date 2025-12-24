package echoswg

import (
  "testing"
  "reflect"
  "fmt"
  "encoding/json"
)

func TestBuildSwaggerPath(t *testing.T) {
  type Resp struct {
    ID int64
  }
  path := BuildSwaggerPath(&SwaggerPathDefine{
    Tag: "pet",
    Method: "PUT",
    Summary: "summary",
    Description: "description",
    Path: "/pets/:ID",
    Handlers: []interface{} {
      func(*PetPutA) error {
        return nil
      },
      func(*PetPutB) (*Resp, error) {
        return nil, nil
      },
    },
  })
  if path.Path != "/pets/{ID}" {
    t.Fatalf("bad request path %s", path.Path)
  }
  content, _ := json.Marshal(path.JSON)
  fmt.Println(string(content))
  reflect.DeepEqual(path.JSON, map[string]interface{} {
    "put": map[string]interface{} {
      "tags": []string{"pet"},
      "summary": "summary",
      "description": "description",
      "produces":    []string{"application/json"},
      "consumes":    []string{"application/json"},
      "parameters": []map[string]interface{}{
        {
          "description": "the id",
          "format": "int64",
          "in": "path",
          "name": "ID",
          "required": true,
          "type": "integer",
        }, {
        "description": "",
        "format": "string",
        "in": "query",
        "name": "Category",
        "required": false,
        "type": "string",
      }, {
        "description": "",
        "format": "string",
        "in": "query",
        "name": "Color",
        "required": true,
        "type": "string",
      }, {
        "in": "body",
        "name": "body",
        "required": true,
        "schema": map[string]interface{} {
          "$ref": "#/definitions/anonymous00",
        },
      }},
      "responses": map[string]interface{} {
        "200": map[string]interface{} {
          "description": "successful operation",
          "schema": map[string]interface{} {
            "$ref": "#/definitions/Resp",
          },
        },
        "500": map[string]interface{} {
          "description": "Interal Server Error",
        },
      },
    },
  })
}

func TestGetRootOfPtr(t *testing.T) {
  var str = "a"
  var ptr = &str
  ptr2 := &ptr
  ptr3 := &ptr2
  rootType := getRootOfPtr(reflect.TypeOf(ptr3))
  if rootType != reflect.TypeOf(str) {
    t.Fatalf("root type is %s", rootType)
  }
}

func TestAliyunApiGatewayExtensions(t *testing.T) {
  // Enable Aliyun mode
  originalValue := AliyunApiGatewayEnabled
  AliyunApiGatewayEnabled = true
  defer func() { AliyunApiGatewayEnabled = originalValue }()

  type Resp struct {
    ID int64
  }
  
  path := BuildSwaggerPath(&SwaggerPathDefine{
    Tag: "pet",
    Method: "GET",
    Summary: "Get pet",
    Description: "Get a pet by ID",
    Path: "/pets/:id",
    Handlers: []interface{}{
      func(*PetPutA) (*Resp, error) {
        return nil, nil
      },
    },
    AliyunBackend: &AliyunBackendConfig{
      ServiceType:    "HTTP",
      ServiceAddress: "http://backend.example.com",
      ServicePath:    "/api/pets",
      ServiceMethod:  "GET",
    },
  })

  if path.Path != "/pets/{id}" {
    t.Fatalf("bad request path %s", path.Path)
  }

  content, _ := json.Marshal(path.JSON)
  fmt.Println(string(content))

  getMethod, ok := path.JSON["get"].(map[string]interface{})
  if !ok {
    t.Fatal("get method not found")
  }

  // Check for Aliyun extensions
  if _, hasRequestConfig := getMethod["x-aliyun-apigateway-request-config"]; !hasRequestConfig {
    t.Fatal("missing x-aliyun-apigateway-request-config")
  }

  if _, hasBackend := getMethod["x-aliyun-apigateway-backend"]; !hasBackend {
    t.Fatal("missing x-aliyun-apigateway-backend")
  }

  // Verify backend configuration
  backend := getMethod["x-aliyun-apigateway-backend"].(map[string]interface{})
  if backend["serviceType"] != "HTTP" {
    t.Fatalf("expected serviceType HTTP, got %v", backend["serviceType"])
  }
  if backend["serviceAddress"] != "http://backend.example.com" {
    t.Fatalf("expected serviceAddress http://backend.example.com, got %v", backend["serviceAddress"])
  }
  if backend["servicePath"] != "/api/pets" {
    t.Fatalf("expected servicePath /api/pets, got %v", backend["servicePath"])
  }
}
