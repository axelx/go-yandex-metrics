package config

import (
	"flag"
	"net/http"
	"os"
	"strconv"
)

type ConfigAgentFlag struct {
	FlagServerAddr      string
	FlagReportFrequency int
	FlagPollFrequency   int
}

type ConfigAgent struct {
	Client          *http.Client
	BaseURL         string
	ReportFrequency int
	PollFrequency   int
}

// ////// middleware
//type LoggingRoundTripper struct {
//	next http.RoundTripper
//}
//
//func NewLoggingRoundTripper(next http.RoundTripper) *LoggingRoundTripper {
//	return &LoggingRoundTripper{
//		next: next,
//	}
//}
//
//func (rt *LoggingRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
//	defer func(begin time.Time) {
//
//		//bodyReq, _ := io.ReadAll(req.Body)
//		//req.Body.Close()
//		////todo как прочитть боди реквест ?
//		//
//		//bodyResp, _ := io.ReadAll(resp.Body)
//		fmt.Printf("LoggingRoundTripper method=%s host=%s  bodyReq= status_code=%d  bodyResp= err=%v took=%s\n\n",
//			req.Method, req.URL.Host, resp.StatusCode, err, time.Since(begin),
//		)
//	}(time.Now())
//
//	return rt.next.RoundTrip(req)
//}

func NewConfigAgent() ConfigAgent {

	cf := ConfigAgentFlag{
		FlagServerAddr:      "",
		FlagReportFrequency: 1,
		FlagPollFrequency:   1,
	}
	parseFlagsAgent(&cf)

	var rt = http.DefaultTransport
	//rt = NewLoggingRoundTripper(rt)

	confDefault := ConfigAgent{
		Client: &http.Client{
			Transport: rt,
		},
		BaseURL:         "http://localhost:8080/update/",
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

func parseFlagsAgent(c *ConfigAgentFlag) {
	flag.StringVar(&c.FlagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&c.FlagReportFrequency, "r", 10, "report frequency to run server")
	flag.IntVar(&c.FlagPollFrequency, "p", 2, "poll frequency")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.FlagServerAddr = envRunAddr

	}
	if envReportFrequency := os.Getenv("REPORT_INTERVAL"); envReportFrequency != "" {
		if v, err := strconv.Atoi(envReportFrequency); err == nil {
			c.FlagReportFrequency = v
		}
	}
	if envPollFrequency := os.Getenv("POLL_INTERVAL"); envPollFrequency != "" {
		if v, err := strconv.Atoi(envPollFrequency); err == nil {
			c.FlagPollFrequency = v
		}
	}
}
