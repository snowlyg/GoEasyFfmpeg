# GoEasyFfmpeg 管理 ffmpeg 命令

根据开源项目[https://github.com/EasyDarwin/EasyDarwin](https://github.com/EasyDarwin/EasyDarwin) 修改而来

### 主要功能特点
- 管理 ffmpeg 推流的项目，支持 rtmp ,rtsp ,flv ,hls 等方式推流
- 推荐使用 [livego](https://github.com/gwuhaolin/livego/blob/master/README_cn.md) 项目做转发服务器
- 也可以自行使用 nginx + nginx-rtmp 搭建服务器

### 目前存在问题 
- windows 下无法停止 

// 停止代码 windows 下不兼容
```go
cmd := pusher2FfmpegMap[pusher]
proc := cmd.Process
if proc != nil {
    logger.Printf("prepare to SIGTERM to process:%v", proc)
    proc.Signal(syscall.SIGTERM)
    proc.Wait()

    logger.Printf("process:%v terminate.", proc)
}
delete(pusher2FfmpegMap, pusher)
logger.Printf("delete ffmpeg from pull stream from pusher[%v]", pusher)
```