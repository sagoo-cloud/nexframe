package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kardianos/service"
)

// ServiceManager 结构体封装了服务管理的核心功能
type ServiceManager struct {
	config  *service.Config
	program *program
	service service.Service
}

type program struct {
	exit    chan struct{}
	exePath string
	args    []string
}

// NewServiceManager 创建一个新的 ServiceManager 实例
func NewServiceManager(name, displayName, description string) (*ServiceManager, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	prg := &program{
		exit:    make(chan struct{}),
		exePath: exePath,
		args:    os.Args[1:],
	}

	svcConfig := &service.Config{
		Name:        name,
		DisplayName: displayName,
		Description: description,
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		config:  svcConfig,
		program: prg,
		service: s,
	}, nil
}

// Run 运行服务
func (sm *ServiceManager) Run() error {
	return sm.service.Run()
}

// Control 控制服务（启动、停止、重启等）
func (sm *ServiceManager) Control(action string) error {
	return service.Control(sm.service, action)
}

// Start 实现 service.Interface
func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

// Stop 实现 service.Interface
func (p *program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}

func (p *program) run() {
	cmd := exec.Command(p.exePath, append([]string{"run"}, p.args...)...)
	cmd.Dir = filepath.Dir(p.exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running service: %v\n", err)
	}
}

// Daemonize 尝试将当前程序作为守护进程运行
func Daemonize() error {
	if len(os.Args) > 1 && os.Args[1] == "run" {
		// 已经作为守护进程运行
		return nil
	}

	cmd := exec.Command(os.Args[0], append([]string{"run"}, os.Args[1:]...)...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动守护进程失败: %v", err)
	}

	fmt.Println("守护进程启动成功")
	os.Exit(0)
	return nil // 这行代码永远不会被执行
}

// IsDaemon 检查当前进程是否作为守护进程运行
func IsDaemon() bool {
	return len(os.Args) > 1 && os.Args[1] == "run"
}
