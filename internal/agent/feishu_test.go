package agent

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/weirwei/ikit/iutil"
)

func TestFeishu(t *testing.T) {
	feishuWebhookURL := ""
	jsonStr := `{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "test ProductHunt Daily",
                "content": [
                    [
						{
							"tag": "text",
							"text": "\n"
						},
                        {
                            "tag": "a",
                            "text": "Talo",
                            "href": "https://www.producthunt.com/posts/talo-ai?utm_campaign=producthunt-api\u0026amp;utm_medium=api-v2\u0026amp;utm_source=Application%3A+decohack+%28ID%3A+131684%29"
                        },
                        {
                            "tag": "text",
                            "text": "："
                        },
						{
							"tag": "text",
							"text": "视频通话实时AI语音翻译器\n\n"
						},
						{
							"tag": "text",
							"text": "使用Talo提升您的视频通话体验，这是一款领先的实时人工智能翻译工具。轻松打破语言障碍，全球连接，享受即时、精准的翻译。非常适合商务交流。\n\n"
						}
                    ]
                ]
            }
        }
    }
}`
	resp, err := http.Post(feishuWebhookURL, "application/json", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		t.Errorf("发送消息到飞书失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("飞书API返回状态码错误: %d", resp.StatusCode)
	}
	t.Log(iutil.ToJson(resp))
}
