package timesource

import (
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/valyala/fasthttp"
)

type GlobalTimeSource struct {
	Url string
}

var _ TimeSourcer = (*GlobalTimeSource)(nil)

func NewGlobalTimeSource(url string) *GlobalTimeSource {
	return &GlobalTimeSource{
		Url: url,
	}
}

func (g *GlobalTimeSource) GetTime(mode string) (int64, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 设置请求 URL
	req.SetRequestURI(g.Url + "/timestamp/common")

	// 发起 GET 请求
	err := fasthttp.Do(req, resp)
	if err != nil {
		return 0, err
	}

	// 检查状态码
	if resp.StatusCode() != fasthttp.StatusOK {
		return 0, err
	}

	// 读取响应体
	body := resp.Body()
	timeValue := util.ToInt(string(body))

	return timeValue, nil
}
