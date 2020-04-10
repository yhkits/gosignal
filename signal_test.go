package gosignal

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSignal_Register(t *testing.T) {
	s := NewSignal()
	t.Log(s)
	s.RegisterSignalHandler(syscall.SIGINT, func(sig os.Signal, args ...interface{}) {
		t.Log("got:", sig)
	})
	t.Log(s)
}

func TestSignal_Listen(t *testing.T) {
	NewSignal().Listen().Run()

	sig := syscall.SIGINT

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval * 10)
}

func TestSignal_SetDefaultHandle(t *testing.T) {
	s := NewSignal()
	s.SetDefaultHandler(func(sig os.Signal, args ...interface{}) {
		s.verboseln("last default handler by", sig)
	})
	s.Listen().Run()
	sig := syscall.SIGINT
	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)
}

var interval = time.Millisecond * 1

func TestSignal_demo(t *testing.T) {

	sig := syscall.SIGINT

	s := NewSignal()
	s.RegisterSignalHandler(sig, nil)

	//s.SetVerbose(false)

	s.Listen().Run()

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)
	s.RegisterSignalHandler(sig, func(sig os.Signal, args ...interface{}) {
		t.Log("signal handle:", sig)
	})

	time.Sleep(interval)
	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)
	t.Log("unregister signal:", sig)
	s.UnregisterSignalHandler(sig)

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	sig = syscall.SIGTERM

	time.Sleep(interval)
	t.Log("register signal:", sig)
	s.RegisterSignalHandler(sig, func(sig os.Signal, args ...interface{}) {
		t.Log("handle:", sig)
	})

	time.Sleep(interval)
	t.Log("disable all handler")
	s.Disable()

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)
	t.Log("enable all handler")
	s.Enable()

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)
	t.Log("set ignore signal:", sig)
	s.SetIgnoreSignal(sig)

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)
	t.Log("del ignore signal:", sig)
	s.DelIgnoreSignal(sig)

	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	//s.Stop()
	s.Run()
	t.Log("send simulation signal:", sig)
	syscall.Kill(os.Getpid(), sig)

	time.Sleep(interval)

	time.Sleep(time.Second)
}

func TestSignal_Stop(t *testing.T) {
	s := NewSignal()
	s.Run()
	s.Run()
	s.Stop()
	s.Stop()
	s.Stop()
}
