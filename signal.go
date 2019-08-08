package gosignal

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
)

type SignalHandler func(sig os.Signal, args ...interface{}) // 定义处理函数类型

type Signal struct {
	mtx            sync.Mutex
	defaultHandler SignalHandler               // 默认处理器
	set            map[os.Signal]SignalHandler // 注册的信号集
	ignore         map[os.Signal]bool          // 忽略的信号集
	sig            chan os.Signal              // 信号通道
	enable         bool                        // 是否启用
	tips           bool                        // 提示
	goroutine      bool                        // 是否以协程方式执行信号处理程序
	stop           chan struct{}               // 停止
}

func NewSignal() *Signal {
	return &Signal{
		set:    make(map[os.Signal]SignalHandler),
		ignore: make(map[os.Signal]bool),
		sig:    make(chan os.Signal),
		stop:   make(chan struct{}),
		enable: true,
		tips:   true,
	}
}

// 设置是否显示处理信号的详细信息
func (s *Signal) SetVerbose(b bool) {
	s.tips = b
}

// 设置是否并发方式执行信号处理程序
func (s *Signal) SetConcurrent(b bool) {
	s.goroutine = b
}

// 设置默认的信号处理程序
//  如果收到没有注册处理程序的信号，默认信号处理程序将被调用(假如这个信号没有被设置忽略)
func (s *Signal) SetDefaultHandler(handler SignalHandler) {
	s.defaultHandler = handler
}

// 注册信号处理程序
func (s *Signal) RegisterSignalHandler(sig os.Signal, h SignalHandler) {
	s.mtx.Lock()
	s.set[sig] = h
	s.mtx.Unlock()
}

// 注销信号处理程序
func (s *Signal) UnregisterSignalHandler(sig os.Signal) {
	s.mtx.Lock()
	delete(s.set, sig)
	s.mtx.Unlock()
}

// 设置忽略信号
func (s *Signal) SetIgnoreSignal(sig os.Signal) {
	s.mtx.Lock()
	s.ignore[sig] = true
	s.mtx.Unlock()
}

// 恢复被忽略的信号
func (s *Signal) DelIgnoreSignal(sig os.Signal) {
	s.mtx.Lock()
	delete(s.ignore, sig)
	s.mtx.Unlock()
}

// 调用后将全局禁用信号处理程序
// 跳过收到的所有信号
// 直到调 Enable 后恢复处理之后收到的信号
func (s *Signal) Disable() {
	s.mtx.Lock()
	s.enable = false
	s.mtx.Unlock()
}

// 重新启用信号处理程序
func (s *Signal) Enable() {
	s.mtx.Lock()
	s.enable = true
	s.mtx.Unlock()
}

func (s *Signal) Listen(sig ...os.Signal) *Signal {
	signal.Notify(s.sig, sig...)
	return s
}

func (s *Signal) Run() {
	go func() {
		for {
			select {
			case <-s.stop:
				break
			case sig := <-s.sig:
				s.handler(sig, nil)
			}
		}
	}()
}

func (s *Signal) Stop() {
	s.stop <- struct{}{}
	signal.Stop(s.sig) // 调用后将不再处理任何信号, 系统恢复为针对中断信号的默认动作是关闭进程, 即当前进程被终止
}

func (s *Signal) handler(sig os.Signal, args ...interface{}) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if !s.enable {
		s.verbosef("got signal: %s (state: %s)\n", sig, "disable handlers")
		return
	}

	if _, ok := s.ignore[sig]; ok {
		s.verbosef("got signal: %s (state: %s)\n", sig, "ignore signal")
		return
	}

	h, ok := s.set[sig]
	if ok {
		if h != nil {
			s.verbosef("got signal: %s (state: %s %p)\n", sig, "call handler", h)
			if s.goroutine {
				go h(sig, args...)
			} else {
				h(sig, args...)
			}
			return
		} else {
			s.verbosef("got signal: %s (state: %s)\n", sig, "registered but not set handler")
			return
		}
	}
	if s.defaultHandler != nil {
		s.verbosef("got signal: %s (state: %s %p)\n", sig, "call default handler", h)
		if s.goroutine {
			go s.defaultHandler(sig, args...)
		} else {
			s.defaultHandler(sig, args...)
		}
		return
	}
	s.verbosef("got signal: %s (state: %s)\n", sig, "unregistered handler and unset default handler")
}

func (s *Signal) verbosef(format string, a ...interface{}) {
	if !s.tips {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		short := path.Base(file)
		args := []interface{}{short, line}
		args = append(args, a...)
		fmt.Printf("%s:%d "+format, args...)
	} else {
		fmt.Printf(format, a...)
	}
}

func (s *Signal) verboseln(a ...interface{}) {
	if !s.tips {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		short := path.Base(file)
		args := []interface{}{fmt.Sprintf("%s:%d", short, line)}
		args = append(args, a...)
		fmt.Println(args...)
	} else {
		fmt.Println(a...)
	}
}
