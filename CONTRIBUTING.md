# 贡献指南

## 提交代码到 Fork 仓库

由于系统配置了 URL 重写规则，将 `https://github.com/` 重定向到代理服务器，直接推送到 fork 仓库会遇到认证问题。

### 问题现象

执行 `git push fork master` 时出现：
```
fatal: could not read Username for 'https://zetyun.jiasu-1.xyz': No such device or address
```

### 解决方法

**方法一：临时禁用系统配置（推荐）**

```bash
# 1. 临时禁用系统级 git 配置
GIT_CONFIG_NOSYSTEM=1 git push fork master

# 或者先取消 URL 重写（项目级）
git config --unset url.https://zetyun.jiasu-1.xyz/gh/.insteadOf
git push fork master
```

**方法二：修改 .git/config**

在 `.git/config` 中为 fork remote 添加 `pushurl`：

```ini
[remote "fork"]
    url = https://github.com/YOUR_USERNAME/engine.git
    fetch = +refs/heads/*:refs/remotes/fork/*
    pushurl = https://github.com/YOUR_USERNAME/engine.git
```

**方法三：使用 SSH 方式**

```bash
# 修改 remote URL 为 SSH
git remote set-url fork git@github.com:chenjie222/engine.git

# 确保已配置 SSH key
git push fork master
```

### 完整提交流程示例

```bash
# 1. 添加修改的文件
git add README.md command/command_5min.go tools/download_5min_kline.go

# 2. 提交更改
git commit -m "feat: add new feature"

# 3. 禁用系统配置并推送到 fork
GIT_CONFIG_NOSYSTEM=1 git push fork master

# 或者
git config --unset url.https://zetyun.jiasu-1.xyz/gh/.insteadOf
git push fork master
```

### 从原始仓库同步更新

```bash
# 拉取原始仓库的最新代码
git pull origin master

# 解决冲突后推送到 fork
GIT_CONFIG_NOSYSTEM=1 git push fork master
```

## 代码规范

1. **提交信息格式**：`type: description`
   - `feat`: 新功能
   - `fix`: 修复 bug
   - `docs`: 文档更新
   - `refactor`: 代码重构

2. **文件组织**：
   - 新功能放在 `tools/` 目录
   - 命令行接口放在 `command/` 目录
   - 更新 `README.md` 添加使用说明

## 联系方式

如有问题，请通过 GitHub Issues 联系。
