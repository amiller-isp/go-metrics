package metrics

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// InfoLogger provides a method for logging info messages and is implemented
// by the standard `log` package as well as various other packages.
type InfoLogger interface {
	Printf(format string, args ...interface{})
}

// LoggerClient simple dumps metrics into the log. Useful when running
// locally for testing. Can be used with multiple different logging systems.
type LoggerClient struct {
	logger InfoLogger
	rate   float64
	tagMap map[string]string
}

// NewLoggerClient creates a new logging client. If `logger` is `nil` then it
// defaults to stdout using the built-in `log` package. It is equivalent to:
//
//   metrics.NewLoggerClient(log.New(os.Stdout, "", 0))
func NewLoggerClient(logger InfoLogger) *LoggerClient {
	if logger == nil {
		logger = log.New(os.Stdout, "", 0)
	}

	return &LoggerClient{
		logger: logger,
		rate:   1.0,
	}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *LoggerClient) WithTags(tags map[string]string) Client {
	return &LoggerClient{
		logger: c.logger,
		rate:   c.rate,
		tagMap: combine(c.tagMap, tags),
	}
}

// WithRate clones this client with a given sample rate. Subsequent calls
// will be limited to logging metrics at this rate.
func (c *LoggerClient) WithRate(rate float64) Client {
	return &LoggerClient{
		logger: c.logger,
		rate:   rate,
		tagMap: combine(map[string]string{}, c.tagMap),
	}
}

// print out the metric call, taking into account sample rate.
func (c *LoggerClient) print(t string, name string, value interface{}, sampled interface{}) {
	if c.rate == 1.0 {
		c.logger.Printf("%s %s:%v %v", t, name, value, c.tagMap)
		return
	}

	if rand.Float64() < c.rate {
		if value == sampled {
			c.logger.Printf("%s %s:%v (%v) %v", t, name, value, c.rate, c.tagMap)
		} else {
			c.logger.Printf("%s %s:%v (%v * %v) %v", t, name, sampled, value, c.rate, c.tagMap)
		}
	}
}

// Count adds some value to a metric.
func (c *LoggerClient) Count(name string, value int64) {
	c.print("Count", name, value, float64(value)*c.rate)
}

// Incr adds one to a metric.
func (c *LoggerClient) Incr(name string) {
	c.Count(name, 1)
}

// Decr subtracts one from a metric.
func (c *LoggerClient) Decr(name string) {
	c.Count(name, -1)
}

// Gauge sets a numeric value.
func (c *LoggerClient) Gauge(name string, value float64) {
	c.print("Gauge", name, value, value)
}

// Event tracks an event that may be relevant to other metrics.
func (c *LoggerClient) Event(e *statsd.Event) {
	c.logger.Printf("Event %s\n%s %v", e.Title, e.Text, c.tagMap)
}

// Timing tracks a duration.
func (c *LoggerClient) Timing(name string, value time.Duration) {
	c.print("Timing", name, value, value)
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *LoggerClient) Histogram(name string, value float64) {
	c.print("Histogram", name, value, value)
}
