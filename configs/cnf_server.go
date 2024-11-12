package configs

import (
	"crypto/tls"
	"github.com/sagoo-cloud/nexframe/os/file"
	"time"
)

type ServerConfig struct {
	Name string

	Address       string
	Host          string
	HTTPSAddress  string
	HTTPSCertPath string
	HTTPSKeyPath  string

	TLSConfig *tls.Config `json:"tlsConfig"`

	// ReadTimeout 是读取整个请求(包括请求体)的最大持续时间。
	//
	// 由于 ReadTimeout 不允许处理程序对每个请求体的可接受截止时间或
	// 上传速率做出决定,大多数用户会更喜欢使用 ReadHeaderTimeout。
	// 同时使用两者是有效的。
	ReadTimeout time.Duration

	// WriteTimeout 是在超时之前写入响应的最大持续时间。
	// 每当读取新请求的头部时,它都会被重置。
	// 与 ReadTimeout 一样,它不允许处理程序在每个请求的基础上做出决定。
	WriteTimeout time.Duration

	// IdleTimeout 是在启用保持活动连接时等待下一个请求的最大时间。
	// 如果 IdleTimeout 为零,则使用 ReadTimeout 的值。
	// 如果两者都为零,则没有超时。
	IdleTimeout time.Duration

	// MaxHeaderBytes 控制服务器在解析请求头的键和值(包括请求行)时
	// 将读取的最大字节数。它不限制请求体的大小。
	//
	// 它可以在配置文件中使用如下字符串进行配置: 1m, 10m, 500kb 等。
	// 默认为 10240 字节。
	MaxHeaderBytes int

	// KeepAlive 启用 HTTP keep-alive。
	KeepAlive bool

	// ServerAgent 指定服务器代理信息,将写入 HTTP 响应头的 "Server" 字段。
	ServerAgent string

	// ======================================================================================================
	// 静态文件服务
	// ======================================================================================================

	// Rewrites 指定 URI 重写规则映射。
	Rewrites map[string]string

	// IndexFiles 指定静态文件夹的索引文件。
	IndexFiles []string

	// IndexFolder 指定在请求文件夹时是否列出子文件。
	// 如果为 false,服务器响应 HTTP 状态码 403。
	IndexFolder bool

	// ServerRoot 指定静态服务的根目录。
	ServerRoot string

	// SearchPaths 指定静态服务的额外搜索目录。
	SearchPaths []string

	// StaticPaths 指定 URI 到目录映射数组。
	StaticPaths []staticPathItem

	// FileServerEnabled 是静态服务的全局开关。
	// 如果设置了任何静态路径,它会自动启用。
	FileServerEnabled bool

	// ======================================================================================================
	// Cookie
	// ======================================================================================================

	// CookieMaxAge 指定 cookie 项的最大 TTL。
	CookieMaxAge time.Duration

	// CookiePath 指定 cookie 路径。
	// 它也会影响会话 ID 的默认存储。
	CookiePath string

	// CookieDomain 指定 cookie 域。
	// 它也会影响会话 ID 的默认存储。
	CookieDomain string

	// CookieSameSite 指定 cookie 的 SameSite 属性。
	// 它也会影响会话 ID 的默认存储。
	CookieSameSite string

	// CookieSecure 指定 cookie 的 Secure 属性。
	// 它也会影响会话 ID 的默认存储。
	CookieSecure bool

	// CookieHttpOnly 指定 cookie 的 HttpOnly 属性。
	// 它也会影响会话 ID 的默认存储。
	CookieHttpOnly bool

	// ======================================================================================================
	// 会话
	// ======================================================================================================

	// SessionIdName 指定会话 ID 的名称。
	SessionIdName string

	// SessionMaxAge 指定会话项的最大 TTL。
	SessionMaxAge time.Duration

	// SessionPath 指定用于存储会话文件的会话存储目录路径。
	// 它仅在会话存储类型为文件存储时有意义。
	SessionPath string

	// SessionCookieMaxAge 指定会话 ID 的 cookie TTL。
	// 如果设置为 0,则表示它随浏览器会话一起过期。
	SessionCookieMaxAge time.Duration

	// SessionCookieOutput 指定是否自动将会话 ID 输出到 cookie。
	SessionCookieOutput bool

	// ======================================================================================================
	// PProf
	// ======================================================================================================

	PProfEnabled bool   // PProfEnabled 启用 PProf 功能。
	PProfPattern string // PProfPattern 指定路由器的 PProf 服务模式。

	StatsVizEnabled bool   // StatsVizEnabled 启用 StatsViz 功能。
	StatsVizPort    string // StatsVizPort 指定 StatsViz 服务端口。
	// ======================================================================================================
	// API & Swagger.
	// ======================================================================================================

	OpenApiPath       string `json:"openapiPath"`       // OpenApiPath specifies the OpenApi specification file path.
	SwaggerPath       string `json:"swaggerPath"`       // SwaggerPath specifies the swagger UI path for route registering.
	SwaggerUITemplate string `json:"swaggerUITemplate"` // SwaggerUITemplate specifies the swagger UI custom template
	MaxUploadSize     int    `json:"maxUploadSize"`
	RouteOverWrite    bool   `json:"routeOverWrite"`
}

// staticPathItem 是静态路径配置的项目结构。
type staticPathItem struct {
	Prefix string // 路由器 URI。
	Path   string // 静态路径。
}

func LoadServerConfig() *ServerConfig {

	return &ServerConfig{
		Name:              EnvString(ServerName, "server"),
		Host:              EnvString(ServerHost, ":8081"),
		Address:           EnvString(ServerAddress, ":8081"),
		HTTPSAddress:      EnvString(ServerHTTPSAddress, ":43"),
		HTTPSKeyPath:      EnvString(ServerHTTPSKeyPath, ""),
		HTTPSCertPath:     EnvString(ServerHTTPSCertPath, ""),
		ReadTimeout:       EnvDuration(ServerReadTimeout, 60*time.Second),
		WriteTimeout:      EnvDuration(ServerWriteTimeout, 60*time.Second),
		IdleTimeout:       EnvDuration(ServerIdleTimeout, 60*time.Second),
		MaxHeaderBytes:    EnvInt(ServerMaxHeaderBytes, 1<<20),
		KeepAlive:         EnvBool(ServerKeepAlive, true),
		Rewrites:          make(map[string]string),
		StaticPaths:       make([]staticPathItem, 0),
		ServerAgent:       EnvString(ServerServerAgent, "NexFrame-http-server/1.1"),
		IndexFiles:        EnvStringSlice(ServerIndexFiles, []string{"index.html", "index.htm"}),
		IndexFolder:       EnvBool(ServerIndexFolder, false),
		ServerRoot:        EnvString(ServerServerRoot, ""),
		SearchPaths:       EnvStringSlice(ServerSearchPaths, []string{}),
		FileServerEnabled: EnvBool(ServerFileServerEnabled, false),
		PProfEnabled:      EnvBool(ServerPProfEnabled, false),
		PProfPattern:      EnvString(ServerPProfPattern, "/debug/pprof/"),
		StatsVizEnabled:   EnvBool(ServerStatsVizEnabled, false),
		StatsVizPort:      EnvString(ServerStatsVizPort, ":8088"),

		CookieMaxAge: EnvDuration(ServerCookieMaxAge, time.Hour*24*365),
		CookiePath:   EnvString(ServerCookiePath, "/"),
		CookieDomain: EnvString(ServerCookieDomain, ""),

		SessionIdName:       EnvString(ServerSessionIdName, "NexFrameSessionId"),
		SessionMaxAge:       EnvDuration(ServerSessionMaxAge, time.Hour*24),
		SessionPath:         file.Temp("NexFrameSessions"),
		SessionCookieMaxAge: EnvDuration(ServerSessionCookieMaxAge, time.Hour*24),
		SessionCookieOutput: EnvBool(ServerSessionCookieOutput, true),
		MaxUploadSize:       EnvInt(ServerMaxUploadSize, 32),
	}

}
