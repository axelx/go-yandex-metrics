package config

import "net/http"

type ConfigAgentFlag struct {
	FlagServerAddr      string
	FlagReportFrequency int
	FlagPollFrequency   int
}
type ConfigServerFlag struct {
	FlagRunAddr string
}

type ConfigAgent struct {
	Client          *http.Client
	BaseURL         string
	ReportFrequency int
	PollFrequency   int
}

func NewConfigServerFlag() ConfigServerFlag {
	return ConfigServerFlag{
		FlagRunAddr: "",
	}
}

func NewConfigAgentFlag() ConfigAgentFlag {
	return ConfigAgentFlag{
		FlagServerAddr:      "",
		FlagReportFrequency: 1,
		FlagPollFrequency:   1,
	}
}

func NewConfigAgent(cf ConfigAgentFlag) ConfigAgent {
	confDefault := ConfigAgent{
		Client:          &http.Client{},
		BaseURL:         "",
		ReportFrequency: 10,
		PollFrequency:   2,
	}

	if cf.FlagServerAddr != "" {
		confDefault.BaseURL = "http://" + cf.FlagServerAddr + "/update/"
	}
	if cf.FlagReportFrequency != 0 {
		confDefault.ReportFrequency = cf.FlagReportFrequency
	}
	if cf.FlagPollFrequency != 0 {
		confDefault.PollFrequency = cf.FlagPollFrequency
	}
	return confDefault
}
