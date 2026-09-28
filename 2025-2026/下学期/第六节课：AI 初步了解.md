# 第六节课：AI 初步了解

---

## 一、LLM 基础

### 什么是 LLM？

大型语言模型（Large Language Model）本质上是对一句话下一个字的概率预测

举个例子，我说

```
我喜欢
```

这时候ai就会开始预测下一个字的概率，例如

```
你 20%
哈基米 9.9%
我
花
其他xxx 0.1%
```

那么此时LLM就会输出概率最大的`你`(通俗来说，实际会根据temperature的设置而改变, 一般默认为0时去概率最大，如果设置为其他值则可能取其他概率的值)

```
我喜欢你
```

然后再继续往下预测，直到结束输出EOF

但是只能预测下一个字，显然只能用于写文章等等，无法形成一问一答的对话模式

于是我们想到用特定格式来规范ai的输出

```
system: 你是一只猫娘
user: 你好
ai: 你好👋喵~
```

这样ai就知道他接下来续写的是哪个角色的内容

其中这里的前缀角色大致分为三类, system、user、assistant

| Role | 含义  |
|------|-----|
| `system` | 系统提示词，设定 AI 的行为规则 |
| `user` | 用户输入 |
| `assistant` | AI 的回复 |

但是各厂商的格式各不相同，例如OpenAI有一个单独的role叫做Tool, 用来处理ai调用Tool后返回的结果

### Token：LLM 的最小单位

前面说"预测下一个字"，其实不太准确。LLM 处理的最小单位不是"字"，而是 **Token**。

Token 是通过 Tokenizer（分词器）把文本切成的小块，可能是一个词、半个词、一个字、甚至一句话。

```
英文："Hello world"     → ["Hello", " world"]           → 2 个 token
英文："unbelievable"    → ["un", "believ", "able"]      → 3 个 token
中文："我喜欢你"         → ["我", "喜欢", "你"]           → 3 个 token
中文："量子纠缠"         → ["量", "子", "纠", "缠"]       → 4 个 token（生僻词会被拆得更碎）
```

**为什么要知道这个？** 因为：


1. **API 按 Token 收费**，不是按字数。中文通常比英文消耗更多 token（同样的语义）
2. **上下文窗口长度的限制**(context)也是按 token 算的(你所能看到的200K, 1M上下文)

> 想直观感受 token 怎么切的？打开 [OpenAI Tokenizer](https://platform.openai.com/tokenizer) 输入任意文本试试(我喜欢你 给主人留下些什么吧)

### 上下文窗口（Context Window）

LLM 能"看到"的内容有长度限制，这个限制叫**上下文窗口**，单位是 token。

超过这个窗口的内容，LLM 就"看不见"了——这也是为什么后面需要 Memory（记忆）和 RAG（知识库检索）。

各家模型的上下文窗口大小差异很大：

| 模型  | 上下文窗口 |
|-----|-------|
| Claude Opus 4.7 / Sonnet 4.6 | **1M** |
| Gemini 3.1 pro | **1M** |
| DeepSeek-V4 | 1M    |
| GLM 5.1 | 200K  |

### 常见参数

调用 LLM 时，除了传入消息，还可以调一些"旋钮"来控制输出行为：

| 参数  | 作用  | 通俗理解 |
|-----|-----|------|
| `temperature` | 控制随机性 | 越低越确定，越高越有创意。写代码用 0\~0.3，闲聊用 0.7\~1.0 |
| `top_p` | 按概率过滤候选词（0.0 \~ 1.0） | 只从概率前 p% 的 token 里选。和 temperature 选一个调就行 |
| `top_k` | 按数量过滤候选词 | 只从概率最高的 K 个 token 里选 |
| `max_tokens` | 限制输出长度 | 最多生成多少 token，到了就截断 |
| `stop` | 停止标记 | 遇到指定字符串就停止生成 |

> ⚠️ `temperature` 范围因厂商而异：OpenAI / DeepSeek 是 0\~2（默认 1），Anthropic 是 0\~**1**（默认 1），Gemini 是 0\~2（3.x 推荐 1.0）。参数命名也有差异，详见后面各格式的参数表。

#### 参数之间的关系

`temperature`、`top_p`、`top_k` 都是控制"从哪些候选词里选"的，只是过滤方式不同：

```
所有候选 token（按概率从高到低排列）
  │
  ├─ top_k=5        → 只保留概率最高的 5 个
  │
  ├─ top_p=0.9      → 从最高概率开始累加，累加到 90% 就截断
  │
  └─ temperature    → 对概率分布做缩放
       temp=0       → 概率最高的 token 概率接近 100%（几乎必定选它）
       temp=1       → 保持原始概率分布
       temp=2       → 概率分布被拉平，低概率 token 也有机会被选中
```

**实际使用建议**：

* 一般只调 `temperature` 就够了，`top_p` 设成 1.0（不过滤）
* `top_k` 在 OpenAI API 里不直接支持（DeepSeek 也不支持），如果一定要传，需要通过原始 HTTP 请求手动加字段
* 写代码/做推理：`temperature=0`，确保结果稳定可复现
* 闲聊/创意写作：`temperature=0.7~1.0`

### 输出方式：Generate vs Stream

LLM 有两种输出方式：

* **Generate（非流式）**：等模型把所有内容生成完，一次性返回。适合后台任务
* **Stream（流式）**：模型每生成一个 token 就立刻推送给你。用户体验好，像打字机一样一个字一个字蹦出来

两种方式的**总生成时间一样**，只是传输方式不同。日常对话类应用基本都用 Stream。

### 推荐：不写代码也能体验 LLM

在开始写代码之前，推荐先用现成的客户端体验一下 LLM 的能力，对上面的概念会有更直观的感受：

* [**Cherry Studio**](https://github.com/CherryHQ/cherry-studio)：桌面端 AI 客户端（Windows / Mac / Linux），支持接入几乎所有主流模型（OpenAI、Claude、Gemini、DeepSeek、千问...），自带 300+ 预设助手、多模型同时对话、MCP 工具等，开箱即用
* [**RikkaHub**](https://github.com/rikkahub/rikkahub)：Android 端原生 AI 客户端，同样支持多厂商模型切换，支持多模态输入（图片、PDF、文档），适合手机上随时用

两个工具都支持BYOK(Bring your own key)， 即使用你自己的API 地址和 Key


---

## 二、API 格式与调用

### 四种主流 API 格式

目前业界有四种常见的 LLM API 格式，了解它们的区别很重要，因为你选什么格式决定了你用什么 SDK、怎么传参数。

#### 1. OpenAI Chat Completions（最通用）

`POST /v1/chat/completions`

这是目前**事实上的行业标准**。绝大多数厂商都兼容这个格式（比如DeepSeek、qwen、豆包等等国内厂商），学会这一个基本走遍天下。

```json
{
  "model": "deepseek-v4-flash",
  "messages": [
    {"role": "system", "content": "你是一只可爱的猫娘"},
    {"role": "user", "content": "今天天气怎么样？"},
    {"role": "assistant", "content": null, "tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "get_weather", "arguments": "{\"city\":\"北京\"}"}}]},
    {"role": "tool", "tool_call_id": "call_1", "content": "{\"temp\": 25, \"weather\": \"晴\"}"},
    {"role": "assistant", "content": "主人早上好(つω`*)～☆ 北京今天25度，晴天 天气很适合出去呢~"}
  ],
  "temperature": 0.7
}
```

四种角色各司其职：

| role | 说明  |
|------|-----|
| `system` | 系统提示词，设定 AI 行为 |
| `user` | 用户的输入 |
| `assistant` | AI 的回复，也可能包含 `tool_calls`（告诉你它想调什么工具） |
| `tool` | 工具执行后的结果，通过 `tool_call_id` 关联到对应的调用 |

##### 请求参数

| 参数  | 类型  | 必填  | 说明  |
|-----|-----|:---:|-----|
| `model` | string | 是   | 模型名称，如 `deepseek-v4-flash`、`gpt-4o` |
| `messages` | array | 是   | 消息列表，每条包含 `role` 和 `content` |
| `temperature` | number | 否   | 随机性（0\~2），默认 1 |
| `top_p` | number | 否   | 核采样（0\~1），默认 1 |
| `max_tokens` | integer | 否   | 最大输出 token 数 |
| `stop` | string/array | 否   | 遇到这些字符串停止生成 |
| `stream` | boolean | 否   | 是否流式输出，默认 false |
| `tools` | array | 否   | 可用工具列表（下面详解） |
| `tool_choice` | string/object | 否   | 控制是否调工具：`auto`（默认）/ `none` / `required` / 指定某个工具 |
| `response_format` | object | 否   | 设 `{"type": "json_object"}` 可强制输出 JSON |

##### messages 中的字段

```
// user / system 消息：最简单
{"role": "user", "content": "你好"}

// assistant 消息：可能带 tool_calls
{
  "role": "assistant",
  "content": null,                    // 调工具时 content 为 null
  "tool_calls": [{                    // 工具调用列表（可同时调多个）
    "id": "call_abc123",              // 调用 ID，后面 tool 消息要用这个来匹配
    "type": "function",               // 目前只有 function 类型
    "function": {
      "name": "get_weather",          // 工具名
      "arguments": "{\"city\":\"北京\"}" // 参数，JSON 字符串
    }
  }]
}

// tool 消息：返回工具执行结果
{
  "role": "tool",
  "tool_call_id": "call_abc123",      // 对应哪个 tool_calls 的 id
  "content": "{\"temp\": 25}"         // 工具返回的结果，字符串
}
```

`**id**` **和** `**type**` **的作用：**

* `**id**`：配对用的。LLM 可以一次调多个工具（比如用户问"北京和上海的天气"），每个调用有唯一 `id`，你返回结果时通过 `tool_call_id` 告诉 LLM "这个结果对应哪个调用"。如果没有 `id`，两个都叫 `get_weather`，LLM 就分不清哪个结果是哪个城市的

##### tools 的定义格式

```json
{
  "tools": [{
    "type": "function",
    "function": {
      "name": "get_weather",
      "description": "获取指定城市的天气",
      "parameters": {
        "type": "object",
        "properties": {
          "city": {"type": "string", "description": "城市名"}
        },
        "required": ["city"]
      }
    }
  }]
}
```

##### 响应结构

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "model": "deepseek-v4-flash",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "你好！"},
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30
  }
}
```

| 字段  | 说明  |
|-----|-----|
| `choices[].message` | AI 的回复消息 |
| `choices[].finish_reason` | 停止原因：`stop`（正常结束）/ `length`（到达 max_tokens）/ `tool_calls`（要调工具） |
| `usage` | token 用量统计，用来算费用 |

#### 2. OpenAI Responses（OpenAI 新一代接口标准）

`POST /v1/responses`

OpenAI 在 2025 年推出的新格式，作为 Chat Completions 的继任者。结构完全不同：用 `input` 代替 `messages`，支持内建工具（网页搜索、代码执行等），原生支持多轮对话链。

```json
{
  "model": "deepseek-v4-flash",
  "input": "写一首关于编程的俳句",
  "tools": [{"type": "web_search_preview"}]
}
```

特点：更简洁，`input` 可以直接传字符串；用 `previous_response_id` 串联多轮对话，不用自己拼历史消息。**目前只有 OpenAI 自己支持这个格式**，其他厂商暂不兼容。

**为什么 OpenAI 要从 Chat 转向 Responses？**

Chat Completions 最初是为简单对话设计的，但随着 Agent 时代到来，它的局限越来越明显：

* 多步工具调用需要开发者自己拼消息、管状态，写一大堆粘合代码
* 没有内建工具，想让 AI 搜网页、读文件都得自己实现
* OpenAI 曾推出过 **Assistants API** 来解决这些问题，但用起来太重、开发者反馈不好

**Assistants API 是什么？** 它是 OpenAI 在 2023 年推出的一套"托管式 Agent 平台"。和 Chat Completions 最大的区别是：Chat 是无状态的（你每次都要把完整对话历史传过去），而 Assistants 帮你在服务端管理状态。它提供了：

* **Thread**（对话线程）：服务端帮你存对话历史，不用自己拼 messages
* **内建工具**：Code Interpreter（在沙箱里跑 Python）、File Search（从上传的文件里检索）
* **Run**（运行）：提交一个 Run，模型会自动循环调用工具直到完成

听起来很美好，但问题是：API 设计太复杂（Assistant → Thread → Message → Run → Step，层层嵌套），调试困难，延迟高，而且是 OpenAI 独有的，没有其他厂商兼容。

Responses API 的思路是：**把 Chat Completions 的简洁 + Assistants API 的工具能力合二为一**。保留内建工具（网页搜索、文件搜索、代码执行），但用一个扁平的 API 调用就能完成，不需要那套复杂的嵌套结构。Assistants API 计划在 2026 年中弃用，Responses 是官方推荐的未来方向。

> Chat Completions 不会被弃用，会持续支持新模型。但 OpenAI 建议新项目优先用 Responses。

##### 请求参数

| 参数  | 类型  | 必填  | 说明  |
|-----|-----|:---:|-----|
| `model` | string | 是   | 模型名称 |
| `input` | string/array | 是   | 用户输入，可以是字符串或消息数组 |
| `instructions` | string | 否   | 系统指令（替代 Chat 的 system 消息） |
| `tools` | array | 否   | 工具列表，支持内建工具 |
| `temperature` | number | 否   | 随机性（0\~2），默认 1 |
| `top_p` | number | 否   | 核采样（0\~1） |
| `max_output_tokens` | integer | 否   | 最大输出 token 数 |
| `previous_response_id` | string | 否   | 上一轮响应 ID，用于多轮对话 |
| `store` | boolean | 否   | 是否持久化对话状态，默认 true |
| `truncation` | string | 否   | `auto`（自动截断）/ `disabled`（超限报错，默认） |
| `stream` | boolean | 否   | 是否流式输出 |

##### input 的两种格式

```
// 简单字符串（单轮对话最方便）
"input": "写一首关于编程的俳句"

// 消息数组（多角色对话）
"input": [
  {"role": "user", "content": "你好"},
  {"role": "assistant", "content": "你好！"},
  {"role": "user", "content": "今天天气怎么样？"}
]
```

##### 内建工具

Responses API 最大的卖点——不用自己实现，开箱即用：

| 工具  | 说明  |
|-----|-----|
| `web_search_preview` | 网页搜索 |
| `code_interpreter` | 沙箱代码执行（Python） |
| `file_search` | 文件语义搜索 |
| `computer_use` | 计算机操作 |

```json
{"tools": [{"type": "web_search_preview"}]}
```

##### 响应结构

```json
{
  "id": "resp_xxx",
  "output": [
    {
      "type": "message",
      "role": "assistant",
      "content": [{"type": "output_text", "text": "你好！"}]
    }
  ],
  "usage": {"input_tokens": 20, "output_tokens": 10}
}
```

##### 多轮对话：两种方式

**方式一：**`**previous_response_id**`**（推荐，省事）**

```json
// 第一轮
{"model": "gpt-4o", "input": "你好"}
// → 返回 {"id": "resp_001", ...}

// 第二轮：只传新消息，历史由服务端管理
{"model": "gpt-4o", "input": "你叫什么名字", "previous_response_id": "resp_001"}
```

不用自己拼历史，OpenAI 服务端帮你存着（默认保留 30 天）。

**方式二：自己拼** `**input**` **消息数组（和 Chat 一样）**

```json
{
  "model": "gpt-4o",
  "input": [
    {"role": "user", "content": "你好"},
    {"role": "assistant", "content": "你好！有什么可以帮你的？"},
    {"role": "user", "content": "你叫什么名字"}
  ]
}
```

手动管理完整历史，灵活性更高（比如可以修改/删除中间某条消息再重新生成）。

#### 3. Anthropic Messages（Claude 专用 大部分coding plan兼容 例如GLM）

`POST /v1/messages`

Claude 的 API 格式。和 OpenAI 最大的区别：

* `system` 不在消息数组里，而是作为**顶层独立参数**
* 工具调用结果不是单独的 `tool` 角色，而是放在 `user` 消息的 **content block** 里（类型为 `tool_result`）

```json
{
  "model": "claude-sonnet-4-6-20250514",
  "max_tokens": 1024,
  "system": "你是一个助手",
  "messages": [
    {"role": "user", "content": "你好"}
  ]
}
```

特点：messages 里只有 `user` 和 `assistant` 两种角色交替出现。DeepSeek 也提供了 Anthropic 兼容端点（`https://api.deepseek.com/anthropic`），主要是为了兼容 Claude Code 等工具。

##### 请求参数

| 参数  | 类型  | 必填  | 说明  |
|-----|-----|:---:|-----|
| `model` | string | 是   | 模型名称，如 `claude-sonnet-4-6-20250514` |
| `messages` | array | 是   | 消息列表，只有 `user` 和 `assistant` 两种角色 |
| `max_tokens` | integer | **是** | 最大输出 token 数（**Anthropic 强制必填**，OpenAI 可选） |
| `system` | string/array | 否   | 系统提示词，独立于 messages（不是消息角色） |
| `temperature` | number | 否   | 随机性（**0\~1**，不是 0\~2），默认 1 |
| `top_p` | number | 否   | 核采样（0\~1） |
| `top_k` | integer | 否   | 候选词数量过滤 |
| `stop_sequences` | array | 否   | 停止字符串列表（注意不叫 `stop`） |
| `tools` | array | 否   | 工具定义 |
| `metadata` | object | 否   | 请求元数据，如 `{"user_id": "xxx"}` |
| `stream` | boolean | 否   | 是否流式输出 |

> ⚠️ 和 OpenAI 的关键区别：`max_tokens` 必填、`temperature` 最大值只到 1、停止词叫 `stop_sequences` 不叫 `stop`

##### 工具调用格式

Anthropic 的 Tool 调用和 OpenAI 很不一样——没有单独的 `tool` 角色，工具结果放在 `user` 消息的 content block 里：

```
// 1. AI 回复中包含 tool_use block（注意 input 直接是对象，不是 JSON 字符串）
{
  "role": "assistant",
  "content": [
    {"type": "text", "text": "让我查一下天气"},
    {
      "type": "tool_use",
      "id": "toolu_abc123",
      "name": "get_weather",
      "input": {"city": "北京"}
    }
  ]
}

// 2. 工具结果放在 user 消息的 tool_result block 里
{
  "role": "user",
  "content": [
    {
      "type": "tool_result",
      "tool_use_id": "toolu_abc123",
      "content": "北京今天25°C，晴"
    }
  ]
}
```

##### 响应结构

```json
{
  "id": "msg_xxx",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "你好！"}
  ],
  "model": "claude-sonnet-4-6-20250514",
  "stop_reason": "end_turn",
  "usage": {"input_tokens": 20, "output_tokens": 10}
}
```

| 字段  | 说明  |
|-----|-----|
| `content` | 内容块数组，可包含 `text` 和 `tool_use` 两种类型 |
| `stop_reason` | 停止原因：`end_turn`（正常）/ `max_tokens` / `stop_sequence` / `tool_use`（要调工具） |

#### 4. Google Gemini generateContent

`POST /v1beta/models/{model}:generateContent`

Google Gemini 的原生 API，和前面三家都不太一样：

* 消息数组叫 `contents`（不叫 `messages`）
* AI 角色叫 `model`（不叫 `assistant`）
* `system` 指令通过顶层 `system_instruction` 参数传入（和 Anthropic 类似）
* 每条消息的内容用 `parts` 数组包装（不是直接一个 `content` 字符串）

```json
{
  "system_instruction": {
    "parts": [{"text": "你是一个助手"}]
  },
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "你好"}]
    }
  ]
}
```

特点：`parts` 结构天然支持多模态（同一条消息里可以混合文字、图片、文件等）。流式输出用单独的端点 `streamGenerateContent`。Gemini 也提供了 OpenAI 兼容端点，所以同样可以用 OpenAI SDK 调用。

##### 请求参数

Gemini 的生成参数不在顶层，而是包在 `generationConfig` 对象里，且字段命名是**驼峰式**：

| 参数（在 generationConfig 内） | 类型  | 说明  |
|--------------------------|-----|-----|
| `temperature`            | number | 随机性（0\~2），Gemini 3.x 推荐保持默认 1.0 |
| `topP`                   | number | 核采样（注意是驼峰，不是 `top_p`） |
| `topK`                   | integer | 候选词数量过滤 |
| `maxOutputTokens`        | integer | 最大输出 token 数（不是 `max_tokens`） |
| `stopSequences`          | array | 停止字符串列表（驼峰命名） |
| `responseMimeType`       | string | 设 `application/json` 可强制 JSON 输出 |
| `responseSchema`         | object | JSON Schema，配合 `responseMimeType` 使用 |

```json
{
  "generationConfig": {
    "temperature": 1.0,
    "topP": 0.95,
    "topK": 40,
    "maxOutputTokens": 1024
  },
  "system_instruction": {
    "parts": [{"text": "你是一个助手"}]
  },
  "contents": [
    {"role": "user", "parts": [{"text": "你好"}]}
  ]
}
```

##### 工具调用格式

Gemini 的工具定义放在 `function_declarations` 数组里（不是 OpenAI 的 `functions`）：

```json
{
  "tools": [{
    "function_declarations": [{
      "name": "get_weather",
      "description": "获取天气",
      "parameters": {
        "type": "object",
        "properties": {
          "city": {"type": "string", "description": "城市名"}
        },
        "required": ["city"]
      }
    }]
  }]
}
```

AI 返回的工具调用在 `parts` 里以 `functionCall` 形式出现：

```json
{"parts": [{"functionCall": {"name": "get_weather", "args": {"city": "北京"}}}]}
```

##### 响应结构

```json
{
  "candidates": [{
    "content": {
      "role": "model",
      "parts": [{"text": "你好！"}]
    },
    "finishReason": "STOP"
  }],
  "usageMetadata": {
    "promptTokenCount": 10,
    "candidatesTokenCount": 20,
    "totalTokenCount": 30
  }
}
```

| 字段  | 说明  |
|-----|-----|
| `candidates[].content.parts` | 回复内容，可包含 `text` 或 `functionCall` |
| `candidates[].finishReason` | 停止原因：`STOP`（正常）/ `MAX_TOKENS` / `SAFETY`（被安全过滤） |
| `usageMetadata` | token 用量（字段名和 OpenAI 不同） |

#### 格式选择建议

| 场景  | 推荐格式 |
|-----|------|
| 调 DeepSeek / 千问 / 豆包等国内厂商 | OpenAI Chat Completions |
| 调 OpenAI 最新功能（内建工具等） | OpenAI Responses |
| 调 Claude / 接入 Claude Code 生态 | Anthropic Messages |
| 调 Google Gemini 原生能力 | Gemini generateContent |
| 不确定 / 想通用 | **OpenAI Chat Completions**（覆盖面最广） |

#### 四种格式对比速查

同一个概念在四种 API 里命名都不一样，初学者容易混：

| 对比项 | Chat Completions | Responses | Anthropic Messages | Gemini |
|-----|------------------|-----------|--------------------|--------|
| 消息数组 | `messages`       | `input`   | `messages`         | `contents` |
| 系统提示 | `role: "system"` | `instructions` | `system`（顶层参数）     | `system_instruction` |
| AI 角色名 | `assistant`      | `assistant` | `assistant`        | `model` |
| 工具结果角色 | `role: "tool"`   | —         | `role: "user"` + `tool_result` block | `role: "user"` + `functionResponse` |
| 最大输出 | `max_tokens`     | `max_output_tokens` | `max_tokens`（**必填**） | `maxOutputTokens` |
| 停止词 | `stop`           | `stop`    | `stop_sequences`   | `stopSequences` |
| temperature | 0\~2             | 0\~2      | 0\~**1**           | 0\~2   |
| 停止原因 | `finish_reason: "stop"` | —         | `stop_reason: "end_turn"` | `finishReason: "STOP"` |
| 调工具标志 | `finish_reason: "tool_calls"` | —         | `stop_reason: "tool_use"` | `finishReason: "STOP"` + `functionCall` in parts |
| 多轮对话 | 自己拼 messages     | `previous_response_id` 或自己拼 | 自己拼 messages       | 自己拼 contents |
| 参数风格 | snake_case       | snake_case | snake_case         | camelCase |

### 实战：用 OpenAI Go SDK 调用 DeepSeek API

DeepSeek API 兼容 OpenAI Chat格式，所以可以直接用 OpenAI 官方 Go SDK 来调用。大部分国内厂商（千问、豆包等）也兼容这套格式，换个 BaseURL 和模型名就行。

#### 准备工作


1. 去 [DeepSeek 开放平台](https://platform.deepseek.com/api_keys) 注册并创建 API Key
2. 安装依赖：

```bash
go get github.com/openai/openai-go/v3
```

#### 基础调用（非流式）

```go
package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: "deepseek-v4-flash",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个有帮助的助手"),
			openai.UserMessage("用一句话介绍Go语言"),
		},
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(1024),
		TopP:        openai.Float(0.9),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
```

#### 流式调用（Stream）

```go
func streamExample() {
	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
		Model: "deepseek-v4-flash",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("写一首关于编程的短诗"),
		},
		Temperature: openai.Float(1.0),
	})

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
	if stream.Err() != nil {
		panic(stream.Err())
	}
}
```


---

## 三、Context 与 Memory

学会调 API 后，第一个问题是：**怎么让 AI 记住我们之前说了什么？**

| 概念  | 解决什么问题 | 一句话理解 |
|-----|--------|-------|
| **Context** | LLM 单次请求能"看到"什么 | 当前对话的视野范围 |
| **Memory** | LLM 怎么跨对话记住东西 | 给 LLM 装一个硬盘 |

### Context：上下文

LLM 每次请求能看到的 token 有上限（上下文窗口）。Context 就是"这次请求里 LLM 实际能看到的内容"，通常包括：

* 系统提示词（System Prompt）
* 对话历史（当前会话的消息）
* RAG 检索到的参考资料
* Tool 的描述信息
* Memory 中检索出的相关记忆

**Context 越大 ≠ 越好**：塞太多内容会浪费 token、增加延迟，还可能干扰 LLM 的注意力。

#### 示例：无上下文的对话

每次请求只发当前消息，LLM 看不到之前说了什么：

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}

		resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Model: "deepseek-v4-flash",
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("你是一个助手"),
				openai.UserMessage(scanner.Text()),
			},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println("AI:", resp.Choices[0].Message.Content)
	}
}
```

试一下就会发现问题：你说"我叫张三"，再问"我叫什么"，AI 答不上来——因为它每次都是全新的，没有上下文。

#### 示例：有上下文的对话

我们手动对话历史拼进 messages，LLM 就能看到完整的对话了。

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	// ---- 取消下面这行注释，开启上下文 ----
	// history := []openai.ChatCompletionMessageParamUnion{
	// 	openai.SystemMessage("你是一个助手"),
	// }

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		// 无上下文：每次都是全新对话
		messages := []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一个助手"),
			openai.UserMessage(input),
		}

		// ---- 有上下文：取消下面两行注释，并注释掉上面的 messages ----
		// history = append(history, openai.UserMessage(input))
		// messages := history

		resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Model:    "deepseek-v4-flash",
			Messages: messages,
		})
		if err != nil {
			panic(err)
		}

		reply := resp.Choices[0].Message.Content
		fmt.Println("AI:", reply)

		// ---- 有上下文：取消下面这行注释 ----
		// history = append(history, openai.AssistantMessage(reply))
	}
}
```

取消注释后再试：说"我叫张三"，问"我叫什么"，AI 就能回答了——因为它看到了完整的对话历史。

> 这就是 Context 的本质：**把历史消息拼进 messages，LLM 就有了"上下文"**。但对话关掉就没了，不会跨会话保留。

### Memory：记忆

Memory 是**跨对话的持久化存储**。Context 关掉就没了，Memory 存在文件/数据库里，下次对话还能用。

> ⚠️ 注意区分：对话历史属于 **Context**，不是 Memory。Context 是"这次对话能看到什么"，Memory 是"对话结束后还记得什么"。

#### 示例：用文件实现 Memory

把记忆存在 `memory.md` 里，每次启动时读取注入 System Prompt，每轮对话后让 AI 自己决定要不要更新记忆：

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const memoryFile = "memory.md"

func loadMemory() string {
	data, err := os.ReadFile(memoryFile)
	if err != nil {
		return ""
	}
	return string(data)
}

func saveMemory(content string) {
	os.WriteFile(memoryFile, []byte(content), 0644)
}

func main() {
	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	memory := loadMemory()

	systemPrompt := "你是一个助手。"
	if memory != "" {
		systemPrompt += "\n\n以下是你对用户的记忆：\n" + memory
	}

	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		history = append(history, openai.UserMessage(input))

		resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Model:    "deepseek-v4-flash",
			Messages: history,
		})
		if err != nil {
			panic(err)
		}

		reply := resp.Choices[0].Message.Content
		fmt.Println("AI:", reply)
		history = append(history, openai.AssistantMessage(reply))

		// 让 AI 决定是否更新记忆
		memResp, _ := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Model: "deepseek-v4-flash",
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(
					"你是一个记忆管理器。根据对话判断是否需要更新用户记忆。\n" +
						"当前记忆：\n" + memory + "\n\n" +
						"如果需要更新，直接输出新的完整记忆内容（简洁的要点列表）。\n" +
						"如果不需要更新，只输出 NO_UPDATE"),
				openai.UserMessage(fmt.Sprintf("用户说: %s\nAI回复: %s", input, reply)),
			},
			Temperature: openai.Float(0),
		})

		memResult := memResp.Choices[0].Message.Content
		if !strings.Contains(memResult, "NO_UPDATE") {
			memory = memResult
			saveMemory(memory)
			fmt.Println("[记忆已更新]")
		}
	}
}
```

运行后和 AI 聊几句，退出程序，再次启动——AI 还记得你之前说过的事。打开 `memory.md` 可以看到 AI 自己写的记忆内容。


---

## 四、Tool 调用

### 从对话到行动：什么是 Agent？

到目前为止 LLM 只会"说话"。但如果它能**调用工具**（查天气、读数据库、发消息），就能真正做事了。

```
第一层：LLM 会说话
  chatModel.Generate(messages) → 回一条消息

第二层：LLM + Tool，有手了
  LLM 决定调工具 → 你的代码执行工具 → 结果返回 LLM → 产出回答

第三层：Agent = LLM + Tool + 循环
  think → act → observe → think → ...（直到不再需要调工具）
```

这一节先学第二层：**怎么让 LLM 调用工具**。

### 什么是 Tool？

Tool 本质上是"带参数 schema 的可调用能力"。

LLM 不执行代码，只负责生成"调用哪个工具、传什么参数"。你的代码拿到这个决策后，执行对应逻辑，再把结果返回给 LLM。

### OpenAI SDK 的 Tool 定义方式

在 openai-go/v3 里，Tool 通过 JSON Schema 描述参数：

```go
tools := []openai.ChatCompletionToolParam{
    {
        Type: "function",
        Function: openai.ChatCompletionToolFunctionParam{
            Name:        "update_memory",
            Description: openai.String("当从对话中了解到用户的重要信息时，调用此工具更新记忆"),
            Parameters: openai.Object(map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "date": map[string]any{
                        "type":        "string",
                        "description": "当前的日期",
                    },
                },
                "required": []string{"date"},
            }),
        },
    },
}
```

### 带 Tool 的完整调用流程

还记得上面 Memory 部分用**额外的 LLM 调用**来判断"要不要更新记忆"吗？那种方式每轮都需要两次请求。现在有了 Tool，LLM 可以在正常对话中**自己决定**什么时候调用 `update_memory`——不需要单独的判断请求了。

```go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const memoryFile = "memory.md"

func loadMemory() string {
	data, err := os.ReadFile(memoryFile)
	if err != nil {
		return ""
	}
	return string(data)
}

func saveMemory(content string) {
	os.WriteFile(memoryFile, []byte(content), 0644)
}

func main() {
	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	tools := []openai.ChatCompletionToolParam{
		{
			Type: "function",
			Function: openai.ChatCompletionToolFunctionParam{
				Name:        "update_memory",
				Description: openai.String("当从对话中了解到用户的重要信息（姓名、喜好、身份等）时，调用此工具更新记忆"),
				Parameters: openai.Object(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"content": map[string]any{
							"type":        "string",
							"description": "更新后的完整记忆内容（简洁的要点列表）",
						},
					},
					"required": []string{"content"},
				}),
			},
		},
	}

	memory := loadMemory()

	systemPrompt := "你是一个助手。当你从对话中了解到用户的重要信息时，主动调用 update_memory 工具来记住。"
	if memory != "" {
		systemPrompt += "\n\n你对用户的记忆：\n" + memory
	}

	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}
		history = append(history, openai.UserMessage(scanner.Text()))

		// Agent Loop: 循环直到 LLM 不再调用工具
		for {
			resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
				Model:    "deepseek-v4-flash",
				Messages: history,
				Tools:    tools,
			})
			if err != nil {
				panic(err)
			}

			choice := resp.Choices[0]
			history = append(history, choice.Message.ToParam())

			// LLM 没调工具 → 输出回答，结束本轮
			if choice.FinishReason != "tool_calls" {
				fmt.Println("AI:", choice.Message.Content)
				break
			}

			// LLM 调了工具 → 执行，把结果返回，继续循环
			for _, tc := range choice.Message.ToolCalls {
				var args struct {
					Content string `json:"content"`
				}
				json.Unmarshal([]byte(tc.Function.Arguments), &args)

				saveMemory(args.Content)
				memory = args.Content
				fmt.Println("[记忆已更新]")

				history = append(history, openai.ToolMessage("记忆已更新", tc.ID))
			}
		}
	}
}
```

这段代码同时演示了 **Agent Loop**（第三层）的模式：

```
用户提问 → LLM 思考 → 要调工具？
                         ├─ 是 → 执行工具 → 结果返回 LLM → 继续思考
                         └─ 否 → 输出最终回答
```

和上面 Memory 部分的对比：

|     | Memory 部分的做法 | Tool 的做法 |
|-----|--------------|----------|
| 判断方式 | 每轮额外调一次 LLM 问"要不要更新" | LLM 在对话中自己决定调 `update_memory` |
| LLM 调用次数 | 每轮固定 2 次（对话 + 判断） | 需要更新时才多 1 次，不需要时只 1 次 |
| 结构化程度 | 文本匹配 `NO_UPDATE`，容易出错 | 工具调用，参数结构明确 |


---

## 五、RAG：知识库检索

### 为什么需要 RAG？

LLM 的知识有截止日期，也不知道你的私有数据（公司文档、课程资料等）。

RAG（检索增强生成）= 先从知识库里找相关内容，再把内容塞进 Context，让 LLM 参考着回答。

### 流程

```
【数据准备】
文档 → 分块 → embedding model（向量化）→ 存入向量数据库

【查询时】
用户问题 → embedding model→ 向量数据库检索相似内容 → 注入 Context → LLM 回答

苹果 [0.1, 0.2] 橘子[0.1, 0.1] 陈越[0.7, 0.8]
```

### 关键概念

* **Embedding**：把文字转成一串数字（向量），语义相近的文字，向量也相近
* **向量数据库**：存向量、查相似，常用：Milvus、Chroma、ES、Qdrant等等
* **召回率**：检索到的相关内容占全部相关内容的比例，越高越好
* **Rerank**：对召回结果二次排序，提升精度


---

## 六、MCP 与 Skill

### MCP：接入现成工具生态

#### 什么是 MCP？

MCP（Model Context Protocol）是标准化的"远程能力暴露协议"。

简单理解：**别人写好了一堆工具（搜索、数据库、GitHub、Slack...），你通过 MCP Server 直接接入用，不用自己实现。**

MCP 不只是 Tool 协议，还可以承载更多上下文与能力（比如往 Context 里注入资源、提供 Prompt 模板等）。但大家最常用到的还是"通过 MCP 接入工具"这一部分。

#### 和自己定义 Tool 的区别

|     | 自己定义 Tool | MCP |
|-----|-----------|-----|
| 适合  | 自己写的业务逻辑  | 复用第三方工具 |
| Tool 在哪 | 同进程 Go 函数 | 独立进程（MCP Server），通过协议通信 |
| 谁维护 | 你自己       | 社区/厂商 |
| 复杂度 | 低         | 稍高（需要启动 MCP Server） |

#### MCP 的工作方式

MCP 的核心是 Client-Server 架构：

```
你的应用（MCP Client）
    ↕  JSON-RPC over stdio/SSE
MCP Server（独立进程）
    ├── 工具列表（list_tools）
    ├── 调用工具（call_tool）
    └── 资源/Prompt（可选）
```

你的应用作为 Client，启动一个 MCP Server 进程（或连接到已运行的 Server），通过 JSON-RPC 协议：


1. **发现工具**：调用 `list_tools`，拿到 Server 暴露的所有工具及其参数 schema
2. **调用工具**：调用 `call_tool`，传入工具名和参数，拿到执行结果
3. **转成 OpenAI Tool 格式**：把 MCP 的工具描述转成 `ChatCompletionToolParam`，塞给 LLM

#### 代码示例：集成 MCP Server

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type MCPClient struct{}

func (c *MCPClient) ListTools() ([]MCPTool, error) {
	return []MCPTool{
		{
			Name:        "github_search",
			Description: "搜索 GitHub 仓库",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string", "description": "搜索关键词"},
				},
				"required": []string{"query"},
			},
		},
	}, nil
}

func (c *MCPClient) CallTool(name string, args map[string]any) (string, error) {
	return fmt.Sprintf("搜索结果：找到 3 个与 %s 相关的仓库", args["query"]), nil
}

func mcpToOpenAITool(t MCPTool) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: "function",
		Function: openai.ChatCompletionToolFunctionParam{
			Name:        t.Name,
			Description: openai.String(t.Description),
			Parameters:  openai.Object(t.InputSchema),
		},
	}
}

func main() {
	mcpClient := &MCPClient{}
	mcpTools, _ := mcpClient.ListTools()

	var tools []openai.ChatCompletionToolParam
	for _, t := range mcpTools {
		tools = append(tools, mcpToOpenAITool(t))
	}

	client := openai.NewClient(
		option.WithAPIKey("your-deepseek-api-key"),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	resp, _ := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model:    "deepseek-v4-flash",
		Messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage("搜一下 Go 语言的 MCP 库")},
		Tools:    tools,
	})

	choice := resp.Choices[0]
	if choice.FinishReason == "tool_calls" {
		for _, tc := range choice.Message.ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
			result, _ := mcpClient.CallTool(tc.Function.Name, args)
			fmt.Println("MCP 工具返回:", result)
		}
	}
}
```

关键点：**MCP 工具和自定义工具对 LLM 来说没有区别**，都是 `ChatCompletionToolParam`。区别只在执行层——自定义工具调本地函数，MCP 工具调远程 Server。

#### 实际体验

很多 AI 客户端已经支持 MCP，比如 Cherry Studio、Claude Desktop——在设置里配置 MCP Server 地址，就能直接用。Go 开发者可以用 [mcp-go](https://github.com/mark3labs/mcp-go) 库来写 MCP Server 或 Client。

#### MCP 工具资源

* https://github.com/punkpeye/awesome-mcp-servers — GitHub 精选列表
* https://glama.ai/mcp/servers — 精选 MCP 服务集合
* https://mcp.composio.dev/ — 可组合的 MCP 服务平台

### Skill：给 Agent 划定能力边界

#### 什么是 Skill？

Skill 是对一组相关 Tool 和 Prompt 的封装，代表 Agent 的一项"专项能力"。

它解决的核心问题：**Tool 太多，全塞给 LLM 会乱**。

#### 和 Tool 的区别

|     | Tool | Skill |
|-----|------|-------|
| 粒度  | 单个函数 | 一组工具 + Prompt + 流程 |
| 职责  | 执行一个动作 | 完成一类任务 |
| 例子  | `search(query)` | "信息检索"（搜索 + 过滤 + 摘要） |
| 给 LLM 看的 | 一个工具描述 | 一组精选的工具 + 专属 Prompt |

#### 为什么需要 Skill？

当 Agent 的 Tool 越来越多，直接把几十个 Tool 全塞给 LLM 会有两个问题：


1. **Token 浪费**：每次请求都要把所有 Tool 的描述发给 LLM
2. **选择困难**：Tool 太多，LLM 容易选错

Skill 的做法：先判断用户意图属于哪个 Skill，再只把这个 Skill 下的 Tool 暴露给 LLM。

#### 怎么组织 Skill？

```
Skill = 专属 Prompt + 精选 Tools + 执行流程
```

举个例子，一个学习助手 Agent 可以这样拆 Skill：

| Skill | 专属 Prompt | 精选 Tools |
|-------|-----------|----------|
| 课表查询  | "你是课表助手，帮学生查课程安排" | `query_schedule`, `query_teacher` |
| 作业答疑  | "你是答疑助手，根据课程内容回答问题" | `search_material`, `query_knowledge_base` |
| 成绩分析  | "你是成绩分析助手，帮学生分析学习情况" | `query_grades`, `calculate_gpa` |

#### Skill 的常见组织方式


1. **硬编码路由**：用关键词或规则匹配，简单但不够灵活
2. **LLM 分类**：用一次 LLM 调用来判断意图，灵活但有额外延迟和 token 开销
3. **Embedding 相似度**：把 Skill 描述做 embedding，找最相似的 Skill，适合 Skill 数量很多的场景

> Skill 更多是一种架构思想，而非具体 API。核心思路都是"按意图分组，按需加载"。


---

## 七、完整架构总览

```
用户消息
    ↓
Skill 路由（意图分类，选择处理流程）
    ↓
Agent（LLM + Tool + 循环）
    │
    ├── Context（本次请求 LLM 能看到的所有内容）
    │    ├── 系统提示词（System Prompt）
    │    ├── 对话历史（当前会话的消息）
    │    ├── Memory 注入（跨对话的持久化记忆）
    │    └── RAG 检索结果（知识库相关内容）
    │
    ├── Tools（Agent 可调用的工具）
    │    ├── 业务 Tool（自己定义的函数）
    │    └── 第三方 Tool（通过 MCP 接入）
    │
    └── Memory（持久化存储）
         ├── 用户偏好 / 历史事实
         └── 用户画像
    ↓
最终回答
```

各部分的关系：

* **Context** 是 LLM 的"视野"，所有信息最终都要进入 Context 才能被 LLM 看到
* \
* **Memory** 和 **RAG** 是 Context 的"供给方"——Memory 提供跨对话记忆，RAG 提供外部知识
* **Tool** 和 **MCP** 是 Agent 的"手"——Tool 是自己写的，MCP 是别人写好的
* **Skill** 是"调度员"——决定用哪些 Tool、哪套 Prompt 来处理当前请求