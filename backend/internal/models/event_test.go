package models_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lazycatapps/gosmee/backend/internal/models"

	"gopkg.in/yaml.v3"
)

var _ = Describe("Event decoding compatibility", func() {
	type eventSpec struct {
		Description string                 `yaml:"description"`
		Event       map[string]interface{} `yaml:"event"`
	}

	type expectation struct {
		ID           string            `yaml:"id"`
		ClientID     string            `yaml:"clientId"`
		Timestamp    string            `yaml:"timestamp"`
		EventType    string            `yaml:"eventType"`
		Source       string            `yaml:"source"`
		Status       string            `yaml:"status"`
		StatusCode   int               `yaml:"statusCode"`
		LatencyMs    int               `yaml:"latencyMs"`
		Payload      string            `yaml:"payload"`
		Response     string            `yaml:"response"`
		ErrorMessage string            `yaml:"errorMessage"`
		Headers      map[string]string `yaml:"headers"`
	}

	type expectationSpec struct {
		Description string      `yaml:"description"`
		Expected    expectation `yaml:"expected"`
	}

	type testCase struct {
		description string
		sourceFile  string
		expectFile  string
	}

	loadEvent := func(path string) eventSpec {
		return mustLoadYAML[eventSpec](path)
	}

	loadExpectation := func(path string) expectationSpec {
		return mustLoadYAML[expectationSpec](path)
	}

	DescribeTable(
		"decodes events from different gosmee formats",
		func(tc testCase) {
			source := loadEvent(tc.sourceFile)
			expectation := loadExpectation(tc.expectFile)

			eventJSON, err := json.Marshal(source.Event)
			Expect(err).NotTo(HaveOccurred())

			var evt models.Event
			Expect(json.Unmarshal(eventJSON, &evt)).To(Succeed())

			expected := expectation.Expected

			Expect(evt.ID).To(Equal(expected.ID))
			Expect(evt.ClientID).To(Equal(expected.ClientID))
			Expect(evt.EventType).To(Equal(expected.EventType))
			Expect(evt.Source).To(Equal(expected.Source))
			Expect(string(evt.Status)).To(Equal(expected.Status))
			Expect(evt.StatusCode).To(Equal(expected.StatusCode))
			Expect(evt.LatencyMs).To(Equal(expected.LatencyMs))
			Expect(evt.Payload).To(Equal(expected.Payload))
			Expect(evt.Response).To(Equal(expected.Response))
			Expect(evt.ErrorMessage).To(Equal(expected.ErrorMessage))
			Expect(evt.Headers).To(Equal(expected.Headers))

			if expected.Timestamp != "" {
				ts, parseErr := time.Parse(time.RFC3339, expected.Timestamp)
				Expect(parseErr).NotTo(HaveOccurred())
				Expect(evt.Timestamp.UTC()).To(Equal(ts.UTC()))
			}

			summary := evt.ToSummary()
			Expect(summary.ID).To(Equal(expected.ID))
			Expect(summary.EventType).To(Equal(expected.EventType))
			Expect(summary.Source).To(Equal(expected.Source))
			Expect(string(summary.Status)).To(Equal(expected.Status))
			Expect(summary.StatusCode).To(Equal(expected.StatusCode))
			Expect(summary.LatencyMs).To(Equal(expected.LatencyMs))
		},
		Entry("gosmee snake_case event with payload object", testCase{
			description: "gosmee snake_case event with payload object",
			sourceFile:  filepath.Join("testdata", "event_unmarshal", "gosmee_object", "source.yaml"),
			expectFile:  filepath.Join("testdata", "event_unmarshal", "gosmee_object", "expected.yaml"),
		}),
		Entry("legacy camelCase event with string payload", testCase{
			description: "legacy camelCase event with string payload",
			sourceFile:  filepath.Join("testdata", "event_unmarshal", "legacy_string", "source.yaml"),
			expectFile:  filepath.Join("testdata", "event_unmarshal", "legacy_string", "expected.yaml"),
		}),
	)
})

func mustLoadYAML[T any](path string) T {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred(), "failed to read yaml file %s", path)

	var out T
	Expect(yaml.Unmarshal(data, &out)).To(Succeed(), "failed to unmarshal yaml file %s", path)

	return out
}
