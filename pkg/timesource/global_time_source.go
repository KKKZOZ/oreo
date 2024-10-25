package timesource

import (
	"io"
	"net/http"

	"github.com/oreo-dtx-lab/oreo/internal/util"
)

type GlobalTimeSource struct {
	Url string
}

var _ TimeSourcer = (*GlobalTimeSource)(nil)

var HttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        6000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     1000,
	},
}

func NewGlobalTimeSource(url string) *GlobalTimeSource {
	return &GlobalTimeSource{
		Url: url,
	}
}

func (g *GlobalTimeSource) GetTime(mode string) (int64, error) {
	resp, err := HttpClient.Get(g.Url + "/timestamp/common")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	timeValue := util.ToInt(string(body))
	return timeValue, nil
}
