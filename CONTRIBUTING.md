# 贡献指南

感谢您对节点管理测速系统的关注！我们欢迎各种形式的贡献，无论是新功能、bug修复、文档改进，还是其他任何形式的帮助。

## 贡献流程

1. **Fork 本仓库**：点击 GitHub 页面右上角的 "Fork" 按钮，将本仓库复制到您的 GitHub 账户下。

2. **克隆您的 Fork**：
   ```bash
   git clone https://github.com/您的用户名/node-speedtest.git
   cd node-speedtest
   ```

3. **创建特性分支**：
   ```bash
   git checkout -b feature/amazing-feature
   ```

4. **进行您的更改**：实现您的功能或修复 bug。

5. **提交您的更改**：
   ```bash
   git commit -m 'Add some amazing feature'
   ```

6. **推送到您的 Fork**：
   ```bash
   git push origin feature/amazing-feature
   ```

7. **创建 Pull Request**：前往 GitHub 上的原始仓库，您会看到一个创建 Pull Request 的提示。

## 开发环境设置

### 前提条件

- Go 1.18 或更高版本
- Docker 和 Docker Compose（可选，用于容器化开发）
- Git

### 本地开发

1. **设置开发环境**：
   ```bash
   # 安装依赖
   go mod download
   
   # 构建面板
   cd panel
   go build -o panel
   
   # 构建节点
   cd ../node
   go build -o node
   ```

2. **运行测试**：
   ```bash
   go test ./...
   ```

### Docker 开发

1. **使用 Docker Compose 启动开发环境**：
   ```bash
   docker-compose up -d
   ```

2. **查看日志**：
   ```bash
   docker-compose logs -f
   ```

## 代码风格指南

- 遵循 Go 的官方代码风格指南
- 使用 `gofmt` 格式化代码
- 添加适当的注释和文档
- 确保所有测试通过

## Pull Request 指南

- 确保 PR 描述清楚地说明了更改的内容和原因
- 包含任何必要的文档更新
- 确保所有测试通过
- 如果添加了新功能，请添加相应的测试
- 如果修复了 bug，请添加一个测试用例来防止回归

## 问题报告

如果您发现了 bug 或有功能请求，请创建一个 issue。请尽可能详细地描述问题或请求，包括：

- 问题的详细描述
- 重现步骤（如适用）
- 预期行为和实际行为
- 环境信息（操作系统、Go 版本等）

## 联系方式

如果您有任何问题，可以通过 GitHub issues 联系我们。

感谢您的贡献！ 