package example

import (
	"context"
	"fmt"
	"github.com/ServiceWeaver/weaver"
	"github.com/sagoo-cloud/nexframe/nf"

	"net/http"
	"os"
	"strconv"
	"strings"
)

//go:generate weaver generate ./...

type server struct {
	weaver.Implements[weaver.Main]
	Config        model.ModSwaConfig
	ServiceDict   weaver.Ref[dictSrv.T]
	ServiceConfig weaver.Ref[configSrv.T]
	address       weaver.Listener
}

func RunServe(ctx context.Context, s *server) error {

	apiObj := nf.NewAPIFramework()
	// 自动添加 ServiceWeaver 服务
	err := apiObj.AddWeaverService(s)
	if err != nil {
		return fmt.Errorf("failed to add ServiceWeaver services: %v", err)
	}

	router.Iot("/notice", apiObj)
	router.System("/system", apiObj)

	// 设置静态目录和Web根目录
	apiObj.SetStaticDir("./resource").SetWebRoot("./wwwroot")

	// 添加跨域，日志中间件
	apiObj.WithMiddleware(middleware.CORSMiddleware)
	apiObj.WithMiddleware(middleware.LoggingMiddleware)

	// 添加其它中间件
	apiObj.WithMiddleware(
		middleware.Metric,
		middleware.PanicRecover,
	)

	// 设置全局上下文值
	apiObj.SetContextValue("userID", "aa11223344")

	// 设置文件系统
	fs := os.DirFS("./assets")
	apiObj.SetFileSystem(fs)

	if s.Config.System.Debug {
		//apiObj.EnableDebug() // 启用框架调试模式
	}

	//apiObj.PrintAPIRoutes() // 打印所有注册的API路由

	//开启系统性能分析
	s.InitPProf(ctx, "58089")
	swaggerUrl := ""
	if s.Config.System.Debug {
		addr := ":" + strconv.Itoa(s.Config.System.Addr)
		s.Logger(ctx).Debug("Sagooiot HTTP Service ", "address", addr)
		swaggerUrl = fmt.Sprintf("http://localhost%s/swagger/index.html", addr)
		s.Logger(ctx).Debug(swaggerUrl)
		return http.ListenAndServe(addr, apiObj.GetServer())

	} else {

		addr := strings.Split(s.address.Addr().String(), ":")
		n := len(addr) - 1
		addPort := ":" + addr[n]
		swaggerUrl = fmt.Sprintf("http://localhost%s/swagger/index.html", addPort)

		s.Logger(ctx).Debug(swaggerUrl)
		return http.Serve(s.address, apiObj.GetServer())
	}
}
func (s *server) Init(ctx context.Context) (err error) {
	s.Config, err = s.ServiceConfig.Get().GetCfgSwaConfig(ctx)
	if err != nil {
		fmt.Printf("系统配置失败:%+v\n", err)
	}
	return err
}
