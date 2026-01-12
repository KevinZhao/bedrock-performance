# AWS Bedrock Performance Benchmark Tool

一个用于测试AWS Bedrock模型性能的Go语言工具，支持并发测试和详细的性能指标报告。

## 功能特性

- ✅ 支持流式和非流式两种调用模式
- ✅ 可配置的并发梯度测试（逐步增加并发数）
- ✅ 支持自定义prompt大小和模板
- ✅ 实时控制台输出测试进度
- ✅ 详细的Markdown格式测试报告
- ✅ 完整的性能指标统计

## 性能指标

工具会收集并分析以下关键指标：

- **TTFT (Time To First Token)**: 首个token的响应时间（仅流式模式）
- **完成时间**: 每个请求的总耗时
- **延迟统计**: 平均延迟、最小/最大延迟、P50/P95/P99百分位延迟
- **成功率/失败率**: 请求的成功和失败比例
- **Token吞吐量**: 每秒处理的token数量
- **Token消耗**: 每次请求的input和output token统计
- **错误类型分布**: 按错误类型分类的统计信息

## 安装

### 前置要求

- Go 1.21 或更高版本
- AWS账号和有效的Bedrock访问权限

### 编译

```bash
# 克隆或进入项目目录
cd bedrock-performance

# 安装依赖
go mod download

# 编译
go build -o bedrock-bench ./cmd/bedrock-bench
```

## 配置

### 创建配置文件

复制示例配置文件并修改：

```bash
cp config.example.json config.json
```

### 配置文件说明

```json
{
  "aws": {
    "region": "us-east-1",                    // AWS区域
    "access_key_id": "YOUR_ACCESS_KEY",       // AWS访问密钥ID
    "secret_access_key": "YOUR_SECRET_KEY"    // AWS访问密钥
  },
  "model": {
    "id": "anthropic.claude-3-sonnet-20240229-v1:0",  // Bedrock模型ID
    "quota": 1000                             // 配额限制
  },
  "test": {
    "prompt_size": 1000,                      // Prompt大小（字符数）
    "prompt_template": "Your prompt template with {size} placeholder",
    "streaming": true,                        // 是否测试流式模式
    "non_streaming": true,                    // 是否测试非流式模式
    "max_tokens": 2048,                       // 最大生成token数
    "temperature": 0.7                        // 生成温度参数
  },
  "concurrency": {
    "start": 1,                               // 起始并发数
    "end": 10,                                // 结束并发数
    "step": 2,                                // 并发数递增步长
    "duration_seconds": 60                    // 每个并发级别的测试时长（秒）
  },
  "output": {
    "report_file": "benchmark_report.md"      // 输出报告文件名
  }
}
```

### 支持的模型

工具支持以下类型的Bedrock模型：

- **Claude系列**: `anthropic.claude-3-*`, `anthropic.claude-*`
- **Llama系列**: `meta.llama-*`

## 使用方法

### 基本使用

```bash
# 使用默认配置文件（config.json）
./bedrock-bench

# 指定配置文件路径
./bedrock-bench -config /path/to/your/config.json
```

### 运行示例

```bash
# 编译并运行
go run ./cmd/bedrock-bench -config config.json
```

## 输出说明

### 控制台输出

运行期间，控制台会实时显示：
- 测试配置信息
- 当前并发级别
- 实时进度（请求数、成功/失败数、TPS、Token吞吐量）
- 每个并发级别的详细统计结果

### Markdown报告

测试完成后，会生成详细的Markdown格式报告，包含：

1. **测试配置**: 所有测试参数的汇总
2. **总体概览**: 所有测试的汇总统计
3. **按并发级别的详细结果**: 每个并发级别的完整指标
4. **延迟分析**: 延迟分布的详细表格
5. **TTFT分析**: 首token时间的统计（仅流式模式）
6. **错误分析**: 错误类型和分布统计

## 项目结构

```
bedrock-performance/
├── cmd/
│   └── bedrock-bench/
│       └── main.go              # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── bedrock/
│   │   ├── client.go            # Bedrock API客户端
│   │   └── types.go             # 数据类型定义
│   ├── benchmark/
│   │   ├── runner.go            # 测试编排器
│   │   ├── worker.go            # 并发工作器
│   │   └── metrics.go           # 指标收集器
│   └── report/
│       ├── console.go           # 控制台输出
│       └── markdown.go          # Markdown报告生成
├── config.example.json          # 配置文件示例
├── go.mod                       # Go模块定义
└── README.md                    # 项目文档
```

## 注意事项

1. **AWS凭证安全**: 配置文件包含敏感的AWS凭证，请勿提交到版本控制系统
2. **Quota限制**: 确保AWS账号有足够的Bedrock配额，避免被限流
3. **成本控制**: 大规模测试会产生API调用费用，请注意控制测试规模
4. **网络环境**: 确保网络连接稳定，避免超时导致测试结果不准确
5. **并发控制**: 根据实际quota合理设置并发数，避免触发限流

## 故障排查

### 常见错误

1. **ThrottlingError**: 请求速率超过配额限制
   - 解决方案：降低并发数或增加AWS配额

2. **ValidationError**: 请求参数验证失败
   - 解决方案：检查模型ID和参数配置是否正确

3. **AccessDeniedError**: 权限不足
   - 解决方案：检查AWS凭证和IAM权限配置

4. **ModelNotFoundError**: 模型不存在
   - 解决方案：确认模型ID正确且在指定区域可用

## 示例报告

运行测试后，您将得到类似以下的报告：

```
================================================================================
AWS Bedrock Performance Benchmark Tool
================================================================================
Model: anthropic.claude-3-sonnet-20240229-v1:0
Region: us-east-1
Prompt Size: 1000 characters
Max Tokens: 2048
Temperature: 0.70
Concurrency Range: 1 -> 10 (step: 2)
Duration per Level: 60 seconds
================================================================================

[Streaming Mode Test]
...
[Concurrency Level: 1]
Starting test...
  Results:
    Total Requests:     150
    Successful:         150 (100.00%)
    Throughput:
      Requests/sec:     2.50
      Tokens/sec:       5234.56
    ...
```

## 许可证

本项目仅供内部测试使用。

## 贡献

如有问题或建议，请联系项目维护者。
