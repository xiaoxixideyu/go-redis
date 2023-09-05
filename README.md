# Golang 版本 Redis

本项目基本基于该教程完成：[Go手写Redis](https://www.bilibili.com/video/BV1Zd4y1d7LY)

## 项目特点：

* 尽可能地模仿 Redis 源码
* **单线程（单协程）** 结合 **epoll（IO多路复用）**，避免使用并发 goroutinue 和 channel
* 使用引用计数管理KV对象（其实只有代码层面的意义，作用只是让代码看起来更像 Redis 源码。因为 Go 的内存管理是由三色标记的 GC 自动管理，所以引用计数并不会导致KV对象被直接回收）

## 项目进展

已实现 STRING、LIST、HASH 这几个数据结构的部分命令，剩余的数据结构、命令正在逐步完善。