package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
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

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，返回空的配置
			return config, nil
		}
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析 JSON
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// pullAndSaveImage 拉取镜像并保存为 tar 文件
func pullAndSaveImage(imageName, outputFile string, config Config) error {
	// 解析镜像名称
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return fmt.Errorf("解析镜像名称失败: %v", err)
	}

	// 设置认证信息
	var auth authn.Authenticator
	if config.Registry.Username != "" && config.Registry.Password != "" {
		auth = authn.FromConfig(authn.AuthConfig{
			Username: config.Registry.Username,
			Password: config.Registry.Password,
		})
	} else {
		auth = authn.Anonymous
	}

	// 拉取镜像
	img, err := remote.Image(ref, remote.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %v", err)
	}

	// 创建输出文件
	tarFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建 tar 文件失败: %v", err)
	}
	defer tarFile.Close()

	// 将镜像保存为 tar 文件
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
