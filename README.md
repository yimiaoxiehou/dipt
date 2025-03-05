# 🐳 DIPT (Docker Image Pull Tar)

一个无需 Docker 环境即可拉取 Docker 镜像并保存为 tar 文件的 Go 工具。

## ✨ 功能特点

- 🚀 无需安装 Docker 即可拉取镜像
- 🔄 支持从公共或私有 Docker Registry 拉取镜像
- 💾 将镜像保存为标准 tar 文件，可用于离线环境
- 🔐 支持认证信息配置，可访问私有仓库
- 🛠️ 轻量级命令行工具，易于使用

## 📥 安装

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/iwen-conf/dipt.git
cd dip

# 编译
go build -o dip

# 可选：将编译好的二进制文件移动到PATH路径
mv dip /usr/local/bin/
```

## 📚 使用方法

### 基本用法

```bash
# 拉取镜像并保存为默认的image.tar文件
dip ubuntu:latest

# 拉取镜像并指定输出文件名
dip nginx:alpine my-nginx.tar
```

### 使用私有仓库

创建`config.json`文件，包含私有仓库的认证信息：

```json
{
  "registry": {
    "username": "your-username",
    "password": "your-password"
  }
}
```

然后拉取私有仓库中的镜像：

```bash
dip private-registry.example.com/myapp:1.0 myapp.tar
```

## ⚙️ 配置文件

配置文件`config.json`应放在与程序相同的目录下，格式如下：

```json
{
  "registry": {
    "username": "your-username",
    "password": "your-password"
  }
}
```

如果不提供配置文件或配置文件中没有认证信息，程序将使用匿名访问拉取公共镜像。

## 🌟 优势

- **🔥 无需 Docker 环境**：在无法安装 Docker 的环境中也能拉取镜像
- **📦 离线部署支持**：可以在有网络的环境中拉取镜像，然后将 tar 文件传输到离线环境使用
- **⚡ 轻量级**：只依赖 Go 标准库和容器注册表交互库
- **🔒 安全**：不需要 Docker daemon 权限

## 📄 许可证

本项目采用 MIT 许可证。详见[LICENSE](LICENSE)文件。

## 👥 贡献

欢迎提交问题和贡献代码！

- GitHub: [https://github.com/iwen-conf/dipt.git](https://github.com/iwen-conf/dipt.git)
- 联系邮箱: iluwenconf@163.com

让我们一起改进这个工具！🚀
