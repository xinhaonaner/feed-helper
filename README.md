# feed-helper

XML 转 CSV 的 Web 工具，支持大文件（500MB+）流式转换，适合部署在 Windows 上通过浏览器使用。

## 功能

- 上传 XML 文件，按「行元素」拆成多行，直接子元素作为 CSV 列
- 流式解析与输出，不将整份文件读入内存
- 可配置行元素 tag（默认 `item`，常见于 RSS/Atom）

## 本地运行

```bash
go run .
```

默认监听 `http://localhost:8080`。在浏览器打开该地址，选择 XML 文件并（可选）填写「行元素 tag」后提交，即可下载 CSV。

### 启动参数

- `-port=8080`：监听端口，默认 8080
- `-open-browser`：启动时自动在默认浏览器中打开页面

示例：

```bash
go run . -port=9000 -open-browser
```

## Windows 部署（生成 exe）

在任意系统上交叉编译：

```bash
GOOS=windows GOARCH=amd64 go build -o feed-helper.exe .
```

将 `feed-helper.exe` 拷贝到 Windows 电脑后，双击运行即可。程序会启动本地 HTTP 服务，在浏览器中访问 **http://localhost:8080** 使用上传与下载功能。

若希望启动时自动打开浏览器，可将 exe 放在某目录，然后创建快捷方式，在「目标」后追加参数，例如：

```
C:\path\to\feed-helper.exe -open-browser
```

## 行元素 tag 说明

XML 中**每个**名为「行元素 tag」的节点会对应 CSV 的**一行**；该节点的**直接子元素**名作为列名，其文本内容作为该行对应列的值。

例如行元素 tag 为 `item` 时：

```xml
<root>
  <item><title>标题一</title><link>https://a.com</link></item>
  <item><title>标题二</title><link>https://b.com</link></item>
</root>
```

会得到 CSV：

| title | link |
|-------|------|
| 标题一 | https://a.com |
| 标题二 | https://b.com |

若你的 XML 用其他标签包裹每一行（如 `record`、`row`），在页面的「行元素 tag」输入框中填写即可。
