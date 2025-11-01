package repository_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/repository"

	"gopkg.in/yaml.v3"
)

var _ = Describe("FileEventRepository", func() {
	type expectedSummary struct {
		ID              string `yaml:"id"`
		Status          string `yaml:"status"`
		StatusCode      int    `yaml:"statusCode"`
		LatencyMs       int    `yaml:"latencyMs"`
		Timestamp       string `yaml:"timestamp"`
		PayloadContains string `yaml:"payloadContains"`
	}

	type testCase struct {
		Description string          `yaml:"description"`
		ClientID    string          `yaml:"clientId"`
		FileName    string          `yaml:"fileName"`
		EventFile   string          `yaml:"eventFile"`
		Expected    expectedSummary `yaml:"expected"`
	}

	createTestRepository := func(dir string) repository.EventRepository {
		return repository.NewFileEventRepository(dir)
	}

	setupEventFile := func(baseDir string, tc testCase) string {
		clientDir := filepath.Join(baseDir, "users", "test-user", "clients", tc.ClientID, "events")
		Expect(os.MkdirAll(clientDir, 0o755)).To(Succeed())

		eventData, err := os.ReadFile(tc.EventFile)
		Expect(err).NotTo(HaveOccurred())

		targetPath := filepath.Join(clientDir, tc.FileName)
		Expect(os.WriteFile(targetPath, eventData, 0o644)).To(Succeed())
		return targetPath
	}

	DescribeTable(
		"reads events from filesystem and populates metadata",
		func(tcPath string) {
			tc := MustLoadYaml[testCase](tcPath)
			baseDir := GinkgoT().TempDir()

			repo := createTestRepository(baseDir)

			eventSourcePath := filepath.Join(filepath.Dir(tcPath), tc.EventFile)
			_ = setupEventFile(baseDir, testCase{
				ClientID:  tc.ClientID,
				FileName:  tc.FileName,
				EventFile: eventSourcePath,
			})

			request := &models.EventListRequest{
				Page:      1,
				PageSize:  10,
				SortBy:    "timestamp",
				SortOrder: "desc",
			}

			response, err := repo.GetByClientID(tc.ClientID, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Total).To(Equal(1))
			Expect(response.Events).To(HaveLen(1))

			eventSummary := response.Events[0]
			Expect(eventSummary.ID).To(Equal(tc.Expected.ID))
			Expect(string(eventSummary.Status)).To(Equal(tc.Expected.Status))
			Expect(eventSummary.StatusCode).To(Equal(tc.Expected.StatusCode))
			Expect(eventSummary.LatencyMs).To(Equal(tc.Expected.LatencyMs))

			expectedTimestamp, err := time.Parse(time.RFC3339Nano, tc.Expected.Timestamp)
			Expect(err).NotTo(HaveOccurred())
			Expect(eventSummary.Timestamp.UTC()).To(Equal(expectedTimestamp.UTC()))

			event, err := repo.Get(tc.ClientID, eventSummary.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(event.Payload).To(ContainSubstring(tc.Expected.PayloadContains))
			Expect(string(event.Status)).To(Equal(tc.Expected.Status))
			Expect(event.ClientID).To(Equal(tc.ClientID))

			if tc.Expected.StatusCode != 0 {
				Expect(event.StatusCode).To(Equal(tc.Expected.StatusCode))
			}

			if tc.Expected.LatencyMs != 0 {
				Expect(event.LatencyMs).To(Equal(tc.Expected.LatencyMs))
			}

		},
		Entry("fallback to raw payload when gosmee stores JSON payload only",
			filepath.Join("testdata", "event_repository", "raw_payload", "case.yaml")),
		Entry("parses structured gosmee event metadata",
			filepath.Join("testdata", "event_repository", "structured_event", "case.yaml")),
	)
})

func MustLoadYaml[T any](path string) T {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred(), "failed to read yaml file %s", path)

	var out T
	Expect(yaml.Unmarshal(data, &out)).To(Succeed(), "failed to unmarshal yaml file %s", path)

	return out
}
