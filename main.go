package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
func main() {
	// 检查命令行参数
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("用法: go run main.go <镜像名称> [输出文件]")
		fmt.Println("示例: go run main.go ubuntu:latest image.tar")
		os.Exit(1)
	}

	// 获取镜像名称和输出文件路径
	imageName := os.Args[1]
	outputFile := "image.tar"
	if len(os.Args) == 3 {
		outputFile = os.Args[2]
	}

	// 加载配置文件
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}

	// 拉取镜像并保存
	fmt.Printf("正在拉取镜像 %s...\n", imageName)
	err = pullAndSaveImage(imageName, outputFile, config)
	if err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}
	fmt.Printf("镜像已保存到 %s\n", outputFile)
}
