// Package queuedemo contains runnable examples for the public queue interface.
package queuedemo

// Scenario describes one documented queue behavior exposed by demo:queue.
type Scenario struct {
	Name        string   `json:"name"`
	Section     string   `json:"section"`
	Status      string   `json:"status"`
	Connections []string `json:"connections"`
	Requires    []string `json:"requires,omitempty"`
}

var scenarios = []Scenario{
	{Name: "basic", Section: "Creating Jobs", Status: "implemented", Connections: []string{"sync", "redis", "rabbitmq"}},
	{Name: "strategies", Section: "Job Strategy Interfaces", Status: "implemented", Connections: []string{"sync"}},
	{Name: "unique", Section: "Unique Jobs", Status: "implemented", Connections: []string{"sync", "redis"}, Requires: []string{"redis"}},
	{Name: "debounce", Section: "Debounced Jobs", Status: "implemented", Connections: []string{"redis", "rabbitmq"}, Requires: []string{"redis", "rabbitmq"}},
	{Name: "middleware", Section: "Job Middleware", Status: "implemented", Connections: []string{"sync"}},
	{Name: "dispatch", Section: "Dispatching Jobs", Status: "implemented", Connections: []string{"sync", "redis", "rabbitmq"}},
	{Name: "chain", Section: "Job Chaining", Status: "implemented", Connections: []string{"sync", "redis", "rabbitmq"}},
	{Name: "batch", Section: "Job Batching", Status: "implemented", Connections: []string{"sync", "redis", "rabbitmq"}},
	{Name: "worker", Section: "Running The Queue Worker", Status: "implemented", Connections: []string{"redis", "rabbitmq"}, Requires: []string{"redis", "rabbitmq"}},
	{Name: "failure", Section: "Dealing With Failed Jobs", Status: "implemented", Connections: []string{"redis", "rabbitmq"}, Requires: []string{"redis", "rabbitmq"}},
	{Name: "failed-commands", Section: "Dealing With Failed Jobs", Status: "implemented", Connections: []string{"sync"}},
	{Name: "restart", Section: "Queue Workers and Deployment", Status: "implemented", Connections: []string{"redis", "rabbitmq"}, Requires: []string{"redis", "rabbitmq"}},
	{Name: "events", Section: "Lifecycle Events", Status: "implemented", Connections: []string{"redis", "rabbitmq"}, Requires: []string{"redis", "rabbitmq"}},
	{Name: "encryption", Section: "Encrypted Payloads", Connections: []string{"redis", "rabbitmq"}},
	{Name: "custom-driver", Section: "Custom Drivers", Connections: []string{"custom"}},
	{Name: "errors", Section: "Error Constants", Connections: []string{"sync", "redis", "rabbitmq"}},
	{Name: "redis", Section: "Redis Connection", Connections: []string{"redis"}, Requires: []string{"redis"}},
	{Name: "rabbitmq", Section: "RabbitMQ Configuration", Connections: []string{"rabbitmq"}, Requires: []string{"rabbitmq"}},
}

// Scenarios returns a copy of the queue demo catalog in presentation order.
func Scenarios() []Scenario {
	result := make([]Scenario, len(scenarios))
	for index, scenario := range scenarios {
		result[index] = scenario
		if result[index].Status == "" {
			result[index].Status = "planned"
		}
		result[index].Connections = append([]string(nil), scenario.Connections...)
		result[index].Requires = append([]string(nil), scenario.Requires...)
	}
	return result
}
