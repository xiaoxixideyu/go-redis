# Golang 版本 Redis

本项目基本基于该教程完成：[Go手写Redis](https://www.bilibili.com/video/BV1Zd4y1d7LY)

## 项目特点：

* 尽可能地模仿 Redis 源码
* **单线程（单协程）** 结合 **epoll（IO多路复用）**，避免使用并发 goroutinue 和 channel
* 使用引用计数管理堆对象（其实只有代码层面的意义，作用只是让代码看起来更像 Redis 源码。因为 Go 的内存管理是由三色标记的 GC 自动管理，所以引用计数并不会导致堆对象被直接回收）

## 项目进展

目前和教程一样，只实现了以下三个命令：

* set
* get
* expire

接下来的计划是逐渐补充其它数据结构的相关命令，例如 lpush、rpush 等等。
