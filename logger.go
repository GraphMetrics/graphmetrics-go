package graphmetrics

import "log"

type Logger interface {
	Debug(msg string, metadata map[string]interface{})
	Info(msg string, metadata map[string]interface{})
	Warn(msg string, metadata map[string]interface{})
	Error(msg string, metadata map[string]interface{})
}

type defaultLogger struct {
}

func (*defaultLogger) Debug(msg string, metadata map[string]interface{}) {
	log.Printf("[DEBUG] %s %#v", msg, metadata)
}

func (*defaultLogger) Info(msg string, metadata map[string]interface{}) {
	log.Printf("[INFO] %s %#v", msg, metadata)

}

func (*defaultLogger) Warn(msg string, metadata map[string]interface{}) {
	log.Printf("[WARN] %s %#v", msg, metadata)
}

func (*defaultLogger) Error(msg string, metadata map[string]interface{}) {
	log.Printf("[ERROR] %s %#v", msg, metadata)
}
