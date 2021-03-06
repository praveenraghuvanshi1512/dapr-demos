package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
	"github.com/pkg/errors"
)

var (
	logger = log.New(os.Stdout, "", 0)
	client dapr.Client

	serviceAddress    = getEnvVar("ADDRESS", ":60012")
	sourceBindingName = getEnvVar("SOURCE_BINDING", "fanout-service-source-event-binding")
	targetServiceID   = getEnvVar("TARGET_SERVICE", "")
	targetMethodName  = getEnvVar("TARGET_METHOD", "")
	targetFormat      = getEnvVar("TARGET_FORMAT", "json")
)

func main() {
	// create Dapr service
	s, err := daprd.NewService(serviceAddress)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	c, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("failed to create Dapr client: %v", err)
	}
	client = c
	defer client.Close()

	s.AddBindingInvocationHandler(sourceBindingName, eventHandler)

	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// SourceEvent represents the input event
type SourceEvent struct {
	ID          string  `json:"id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Time        int64   `json:"time"`
}

func eventHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	logger.Printf("Source: %s", in.Data)

	var e SourceEvent
	if err := json.Unmarshal(in.Data, &e); err != nil {
		return nil, errors.Errorf("error parsing input content: %v", err)
	}

	var (
		me error
		b  []byte
		ct string
	)

	switch strings.ToLower(targetFormat) {
	case "json":
		b = in.Data
		ct = "application/json"
	case "xml":
		if b, me = xml.Marshal(&e); me != nil {
			return nil, errors.Errorf("error while converting content: %v", me)
		}
		ct = "application/xml"
	case "csv":
		b = []byte(fmt.Sprintf(`"%s",%f,%f,"%s"`,
			e.ID, e.Temperature, e.Humidity, time.Unix(e.Time, 0).Format(time.RFC3339)))
		ct = "text/csv"
	default:
		return nil, errors.Errorf("invalid target format: %s", targetFormat)
	}

	content := &dapr.DataContent{Data: b, ContentType: ct}
	logger.Printf("Target: %+v", content)

	if out, err = client.InvokeServiceWithContent(ctx, targetServiceID, targetMethodName, content); err != nil {
		return nil, errors.Wrap(err, "error invoking target binding")
	}

	return
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
