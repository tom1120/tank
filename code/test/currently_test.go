package test

// 网络请求并发测试
import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestConcurrentRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d; want %d", resp.StatusCode, http.StatusOK)
			}
			fmt.Println(resp.StatusCode)
		}()
	}

	select {}

}

func TestConcurrentRequestsVideoEnqueue(t *testing.T) {
	urlstr := "http://127.0.0.1:6010/api/asynq/asynqVideoTask"
	taskarr := []string{"1920", "1280", "640", "1536", "480"}
	for i := 0; i < 5; i++ {
		go func(j int) {
			postformdata := url.Values{}
			postformdata.Add("uuid", "c23e0c6a-7cdd-4089-687b-d097d782426c")
			fmt.Println(j)
			postformdata.Add("w", taskarr[j])
			req, err := http.NewRequest("POST", urlstr, strings.NewReader(postformdata.Encode()))
			if err != nil {
				t.Fatal(err)
			}

			req.SetBasicAuth("tank", "123456")
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			fmt.Println(resp.StatusCode)
			// 读取响应Body数据:
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(string(b))

		}(i)
	}

	select {}
}
