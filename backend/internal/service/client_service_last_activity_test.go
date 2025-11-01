package service_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lazycatapps/gosmee/backend/internal/models"
	"github.com/lazycatapps/gosmee/backend/internal/pkg/logger"
	"github.com/lazycatapps/gosmee/backend/internal/repository"
	"github.com/lazycatapps/gosmee/backend/internal/service"

	"gopkg.in/yaml.v3"
)

// lastActivityExpectation defines expected last activity time for a client
type lastActivityExpectation struct {
	ClientID     string `yaml:"clientId"`
	LastActivity string `yaml:"lastActivity"`
}

var _ = Describe("ClientService last activity enrichment", func() {
	type testCase struct {
		description     string
		clientsFile     string
		eventsFile      string
		listGoldenFile  string
		statsGoldenFile string
		listRequestFile string
		defaultUserID   string
	}

	type clientFixture struct {
		ID          string `yaml:"id"`
		UserID      string `yaml:"userId"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Status      string `yaml:"status"`
		SmeeURL     string `yaml:"smeeUrl"`
		TargetURL   string `yaml:"targetUrl"`
		CreatedAt   string `yaml:"createdAt"`
		UpdatedAt   string `yaml:"updatedAt"`
	}

	type clientsSpec struct {
		Description string          `yaml:"description"`
		UserID      string          `yaml:"userId"`
		Clients     []clientFixture `yaml:"clients"`
	}

	type eventFixture struct {
		ClientID   string `yaml:"clientId"`
		DateDir    string `yaml:"dateDir"`
		ID         string `yaml:"id"`
		Timestamp  string `yaml:"timestamp"`
		EventType  string `yaml:"eventType"`
		Source     string `yaml:"source"`
		Status     string `yaml:"status"`
		StatusCode int    `yaml:"statusCode"`
	}

	type eventsSpec struct {
		Description string         `yaml:"description"`
		Events      []eventFixture `yaml:"events"`
	}

	type expectationsSpec struct {
		Description  string                    `yaml:"description"`
		Expectations []lastActivityExpectation `yaml:"expectations"`
	}

	type listRequestSpec struct {
		Description string `yaml:"description"`
		Page        int    `yaml:"page"`
		PageSize    int    `yaml:"pageSize"`
	}

	buildService := func(baseDir string) (*service.ClientService, repository.ClientRepository, *repository.FileQuotaRepository) {
		clientRepo, err := repository.NewFileClientRepository(baseDir)
		Expect(err).NotTo(HaveOccurred())

		eventRepo := repository.NewFileEventRepository(baseDir)
		quotaRepo := repository.NewFileQuotaRepository(baseDir, 10*1024*1024, 1000)
		log := logger.New()
		processService := service.NewProcessService(false, 0, log)

		clientService := service.NewClientService(clientRepo, quotaRepo, eventRepo, processService, baseDir, log)

		return clientService, clientRepo, quotaRepo
	}

	setupClients := func(baseDir string, spec clientsSpec, repo repository.ClientRepository) {
		for _, fixture := range spec.Clients {
			userID := fixture.UserID
			if userID == "" {
				userID = spec.UserID
			}

			createdAt := parseTimeOrNow(fixture.CreatedAt)
			updatedAt := parseTimeOrNow(fixture.UpdatedAt)

			client := &models.Client{
				ID:            fixture.ID,
				UserID:        userID,
				Name:          fixture.Name,
				Description:   fixture.Description,
				Status:        models.ClientStatus(fixture.Status),
				SmeeURL:       fixture.SmeeURL,
				TargetURL:     fixture.TargetURL,
				TargetTimeout: 60,
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}

			err := repo.Create(client)
			Expect(err).NotTo(HaveOccurred())
		}
	}

	setupEvents := func(baseDir string, spec eventsSpec) {
		for _, fixture := range spec.Events {
			userID := findUserDir(baseDir, fixture.ClientID)
			Expect(userID).NotTo(BeEmpty(), "user directory not found for client %s", fixture.ClientID)

			eventDir := filepath.Join(baseDir, "users", userID, "clients", fixture.ClientID, "events", fixture.DateDir)
			Expect(os.MkdirAll(eventDir, 0755)).To(Succeed())

			event := &models.Event{
				ID:         fixture.ID,
				ClientID:   fixture.ClientID,
				Timestamp:  parseTimeOrNow(fixture.Timestamp),
				EventType:  fixture.EventType,
				Source:     fixture.Source,
				Status:     models.EventStatus(fixture.Status),
				StatusCode: fixture.StatusCode,
				LatencyMs:  123,
				Headers:    map[string]string{"X-Test": "true"},
				Payload:    "{}",
			}

			data, err := json.MarshalIndent(event, "", "  ")
			Expect(err).NotTo(HaveOccurred())

			eventPath := filepath.Join(eventDir, fixture.ID+".json")
			Expect(os.WriteFile(eventPath, data, 0644)).To(Succeed())
		}
	}

	DescribeTable("populates the latest event timestamp for clients",
		func(tc testCase) {
			baseDir := GinkgoT().TempDir()

			clientService, clientRepo, quotaRepo := buildService(baseDir)

			clients := MustLoadYaml[clientsSpec](tc.clientsFile)
			setupClients(baseDir, clients, clientRepo)

			if tc.eventsFile != "" {
				events := MustLoadYaml[eventsSpec](tc.eventsFile)
				setupEvents(baseDir, events)
			}

			quotaRepo.InvalidateCache(clients.UserID)

			listReq := models.ClientListRequest{
				Page:     1,
				PageSize: len(clients.Clients),
			}

			if tc.listRequestFile != "" {
				listSpec := MustLoadYaml[listRequestSpec](tc.listRequestFile)
				if listSpec.Page > 0 {
					listReq.Page = listSpec.Page
				}
				if listSpec.PageSize > 0 {
					listReq.PageSize = listSpec.PageSize
				}
			}

			listResponse, err := clientService.List(clients.UserID, &listReq)
			Expect(err).NotTo(HaveOccurred())

			listGolden := MustLoadYaml[expectationsSpec](tc.listGoldenFile)
			verifyLastActivityFromSummaries(listResponse.Clients, listGolden.Expectations)

			statsGolden := MustLoadYaml[expectationsSpec](tc.statsGoldenFile)
			verifyLastActivityFromService(clientService, statsGolden.Expectations)
		},
		Entry("returns latest timestamps when events exist", testCase{
			description:     "should expose last activity for clients with events and nil for others",
			clientsFile:     "testdata/client_last_activity/basic/clients.yaml",
			eventsFile:      "testdata/client_last_activity/basic/events.yaml",
			listGoldenFile:  "testdata/client_last_activity/basic/expected_list.yaml",
			statsGoldenFile: "testdata/client_last_activity/basic/expected_stats.yaml",
		}),
		Entry("handles clients without any events gracefully", testCase{
			description:     "should keep last activity nil when no events recorded",
			clientsFile:     "testdata/client_last_activity/no_events/clients.yaml",
			eventsFile:      "testdata/client_last_activity/no_events/events.yaml",
			listGoldenFile:  "testdata/client_last_activity/no_events/expected_list.yaml",
			statsGoldenFile: "testdata/client_last_activity/no_events/expected_stats.yaml",
		}),
		Entry("supports flat event file layout without date directories", testCase{
			description:     "should detect latest activity from root-level event files",
			clientsFile:     "testdata/client_last_activity/flat/clients.yaml",
			eventsFile:      "testdata/client_last_activity/flat/events.yaml",
			listGoldenFile:  "testdata/client_last_activity/flat/expected_list.yaml",
			statsGoldenFile: "testdata/client_last_activity/flat/expected_stats.yaml",
		}),
	)
})

func parseTimeOrNow(value string) time.Time {
	if value == "" {
		return time.Now().UTC()
	}

	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Now().UTC()
	}
	return ts
}

func findUserDir(baseDir, clientID string) string {
	usersDir := filepath.Join(baseDir, "users")
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return ""
	}

	for _, userDir := range entries {
		if !userDir.IsDir() {
			continue
		}
		clientPath := filepath.Join(usersDir, userDir.Name(), "clients", clientID)
		if info, err := os.Stat(clientPath); err == nil && info.IsDir() {
			return userDir.Name()
		}
	}

	return ""
}

func verifyLastActivityFromSummaries(summaries []*models.ClientSummary, expectations []lastActivityExpectation) {
	activityByClient := make(map[string]*time.Time, len(summaries))
	for _, summary := range summaries {
		if summary == nil {
			continue
		}
		activityByClient[summary.ID] = summary.LastActivity
	}

	for _, expectation := range expectations {
		actual, exists := activityByClient[expectation.ClientID]
		ExpectWithOffset(1, exists).To(BeTrue(), "client summary missing for %s", expectation.ClientID)
		ExpectWithOffset(1, formatTime(actual)).To(Equal(expectation.LastActivity))
	}
}

func verifyLastActivityFromService(clientService *service.ClientService, expectations []lastActivityExpectation) {
	for _, expectation := range expectations {
		client, err := clientService.Get(expectation.ClientID)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, formatTime(client.LastActivity)).To(Equal(expectation.LastActivity))

		stats, err := clientService.GetStats(expectation.ClientID)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		ExpectWithOffset(1, formatTime(stats.LastEventTime)).To(Equal(expectation.LastActivity))
	}
}

func formatTime(ts *time.Time) string {
	if ts == nil {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

func MustLoadYaml[T any](path string) T {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred(), "failed to read yaml file %s", path)

	var out T
	Expect(yaml.Unmarshal(data, &out)).To(Succeed(), "failed to unmarshal yaml file %s", path)

	return out
}
