package logs

import (
	"fmt"
	"time"
	"net/http"
)

type Manager struct {
	api *Api
	collector FileCollector
	parser LineParser
	id IdFeed
	sphinx SphinxFeed
}

type StateHandler struct {
	m *Manager
}

func NewManager() *Manager {
	m := Manager{}
	m.api = InitApi(StateHandler{&m})
	m.Init()
	return &m
}

func (m *Manager) WaitUntilCompletion() {
	<-m.collector.Done
	fmt.Println("collector complete. waiting for queues to empty.")
	for {
		if len(m.parser.Out) > 0 ||
			 len(m.id.Out) > 0 {
				 time.Sleep(3*time.Second)
		} else {
			fmt.Println("Queues complete.")
		}
	}
}

func (m *Manager) Daily() {
	go m.collector.DailyLogsForever(m.parser.Out, m.id.Out)
	m.Process()
}

func (m *Manager) Historical() {
	go m.collector.GetLogsBackwards()
	m.Process()
}

func (m *Manager) Process() {
	go m.parser.ParseLinesForever()
	go m.id.QueryIdsForever()
	go m.sphinx.InsertSphinxForever()
}

func (m *Manager) Init() {
	m.collector = FileCollector{}
	m.collector.InitChan()

	m.parser = LineParser{In: m.collector.Out}
	m.parser.InitChan()

	m.id = IdFeed{In: m.parser.Out}
	m.id.InitChan()
	m.id.Connect()

	m.sphinx = SphinxFeed{In: m.id.Out}
	m.sphinx.Connect()

	m.collector.InitDb(&m.sphinx, &m.id)
}

func (s StateHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	// Ignore the request, return a string as the response
	writer.Header()["Content-Type"] = []string {"application/json"}
	writer.Write([]byte(s.m.getState()))
}

func (m *Manager) getState() string {
	// Queue lengths, messages processed.
	return fmt.Sprintf("%d", 1000)
}
