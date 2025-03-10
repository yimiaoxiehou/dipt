package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall/js"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/schollz/progressbar/v3"
)

// Config 定义 JSON 配置文件的结构
type Config struct {
	Registry struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"registry"`
}

// loadConfig 读取和解析 config.json 文件
func loadConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return config, nil
}

// progressRoundTripper 自定义 RoundTripper，用于更新进度条
type progressRoundTripper struct {
	rt  http.RoundTripper
	bar *progressbar.ProgressBar
}

func (p *progressRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := p.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if strings.Contains(req.URL.Path, "/blobs/") {
		resp.Body = &progressReader{reader: resp.Body, bar: p.bar, closer: resp.Body}
	}
	return resp, nil
}

// progressReader 包装 io.Reader，读取数据时更新进度条
type progressReader struct {
	reader io.Reader
	bar    *progressbar.ProgressBar
	closer io.Closer // 保存原始的 io.Closer
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.bar.Add(n)
	}
	return
}

func (pr *progressReader) Close() error {
	if pr.closer != nil {
		return pr.closer.Close()
	}
	return nil
}

// pullAndSaveImage 拉取镜像并保存为 tar 文件，带进度显示
func pullAndSaveImage(imageName, outputFile string, config Config) error {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return fmt.Errorf("解析镜像名称失败: %v", err)
	}

	var auth authn.Authenticator
	if config.Registry.Username != "" && config.Registry.Password != "" {
		auth = authn.FromConfig(authn.AuthConfig{
			Username: config.Registry.Username,
			Password: config.Registry.Password,
		})
	} else {
		auth = authn.Anonymous
	}

	desc, err := remote.Get(ref, remote.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("获取镜像描述失败: %v", err)
	}

	img, err := desc.Image()
	if err != nil {
		return fmt.Errorf("获取 v1.Image 失败: %v", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("获取层信息失败: %v", err)
	}

	var totalSize int64
	for _, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return fmt.Errorf("获取层大小失败: %v", err)
		}
		totalSize += size
	}

	bar := progressbar.NewOptions64(totalSize,
		progressbar.OptionSetDescription("拉取镜像中"),
		progressbar.OptionShowBytes(true),
	)

	rt := &progressRoundTripper{rt: http.DefaultTransport, bar: bar}
	img, err = remote.Image(ref, remote.WithAuth(auth), remote.WithTransport(rt))
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %v", err)
	}

	tarFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建 tar 文件失败: %v", err)
	}
	defer tarFile.Close()

	err = tarball.WriteToFile(outputFile, ref, img)
	if err != nil {
		return fmt.Errorf("保存镜像到 tar 文件失败: %v", err)
	}

	return nil
}

func MyGoFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		imageName := args[0].String()
		outputFile := "image.tar"

		config := Config{}

		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]
			go func() {
				// 拉取镜像并保存
				resp, err := http.DefaultClient.Get("https://example.com")
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(resp)
				fmt.Printf("正在拉取镜像 %s...\n", imageName)

				err = pullAndSaveImage(imageName, outputFile, config)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
				}
				fmt.Printf("镜像已保存到 %s\n", outputFile)

				arrayConstructor := js.Global().Get("Uint8Array")
				data := []byte("ok")
				dataJS := arrayConstructor.New(len(data))
				js.CopyBytesToJS(dataJS, data)

				responseConstructor := js.Global().Get("Response")
				response := responseConstructor.New(dataJS)

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func main() {
	js.Global().Set("MyGoFunc", MyGoFunc())
	<-make(chan bool)
}
