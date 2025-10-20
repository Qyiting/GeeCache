# GeeCache JMeter 性能测试指南

## 🚀 快速开始

### 1. 安装 JMeter
```bash
# macOS 使用 Homebrew
brew install jmeter

# 或从官网下载：https://jmeter.apache.org/download_jmeter.cgi
```

### 2. 启动 GeeCache 服务
```bash
# 启动单节点服务
go run main.go -port=8001

# 或启动多节点（推荐）
go run main.go -port=8001 &
go run main.go -port=8002 &
go run main.go -port=8003 &
```

### 3. 预热缓存（可选）
```bash
# 预热几个常用键，确保后续测试命中缓存
curl -s 'http://localhost:8001/_geecache/scores/Tom'
curl -s 'http://localhost:8001/_geecache/scores/Jack'
curl -s 'http://localhost:8001/_geecache/scores/Sam'
```

### 4. 启动 JMeter GUI 测试
```bash
# 启动 JMeter 并加载测试计划
jmeter -t geecache_performance_test.jmx
```

## 📊 测试场景说明

### 场景一：缓存命中测试（默认启用）
- **目标**：测试缓存命中时的最大 QPS
- **配置**：200 并发线程，持续 30 秒
- **请求**：固定访问 `/_geecache/scores/Tom`
- **预期**：QPS 应达到 10K+

### 场景二：随机键测试（默认禁用）
- **目标**：测试缓存穿透和 SingleFlight 效果
- **配置**：100 并发线程，持续 30 秒
- **请求**：从 CSV 文件随机选择键
- **预期**：观察 SingleFlight 合并效果

## 🎯 关键指标解读

### Summary Report 关键指标：
- **Throughput**：吞吐量（请求/秒），目标 >10,000/sec
- **Average**：平均响应时间，目标 <10ms
- **Error %**：错误率，目标 0%
- **90% Line**：90% 请求的响应时间

### Graph Results：
- 实时查看响应时间趋势
- 观察系统稳定性

### Aggregate Report：
- 详细的统计信息
- 包含最小值、最大值、标准差等

## ⚙️ 调优建议

### 提升 QPS 的配置调整：

1. **增加并发数**：
   - 右键 "Cache Hit Test" → Edit
   - 调整 "Number of Threads" 到 400-800

2. **调整测试时长**：
   - 修改 "Duration (seconds)" 到 60 或更长

3. **系统优化**：
   ```bash
   # 增加文件描述符限制
   ulimit -n 65536
   
   # 查看系统 CPU 核数
   sysctl -n hw.ncpu
   ```

### 测试不同场景：

1. **启用随机键测试**：
   - 禁用 "Cache Hit Test"
   - 启用 "Random Keys Test"
   - 观察 SingleFlight 效果

2. **混合测试**：
   - 同时启用两个测试组
   - 观察系统在混合负载下的表现

## 📈 结果分析

### 达到 10K+ QPS 的标志：
- Summary Report 中 Throughput > 10000/sec
- Error % = 0%
- Average response time < 20ms

### 性能瓶颈排查：
1. **CPU 使用率**：`top` 命令查看 Go 进程 CPU 占用
2. **内存使用**：观察内存是否充足
3. **网络连接**：`netstat -an | grep 8001` 查看连接数

## 🔧 故障排除

### 常见问题：

1. **连接被拒绝**：
   - 确认 GeeCache 服务已启动
   - 检查端口 8001 是否被占用

2. **QPS 较低**：
   - 增加并发线程数
   - 检查系统资源限制
   - 确保使用缓存命中场景

3. **错误率高**：
   - 检查服务日志
   - 确认请求路径正确
   - 验证数据源配置

## 📝 生成测试报告

### 命令行模式（推荐生产环境）：
```bash
# 非 GUI 模式运行，生成 HTML 报告
jmeter -n -t geecache_performance_test.jmx -l results.jtl -e -o report

# 查看报告
open report/index.html
```

### 保存测试结果：
1. 在 GUI 中右键 Summary Report → Save Table Data
2. 保存为 CSV 格式，便于后续分析

---

**提示**：首次测试建议先用较小的并发数（如 50），确认系统正常后再逐步增加到目标并发数。