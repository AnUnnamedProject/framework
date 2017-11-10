// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	_ "net/http/pprof" // pprof

	"reflect"

	"flag"

	"github.com/AnUnnamedProject/framework/cache"
	"github.com/AnUnnamedProject/i18n"
)

// VERSION of the Framework.
const VERSION = "0.5.0"

// App contains a pointer to the framework instance.
// It's initialized automatically when application starts.
// To start HTTP listening use framework.Run()
var App *Engine

type (
	// Engine is the framework struct.
	Engine struct {
		cache       cache.Cache
		Controllers []Controller
		pool        *ContextPool
		Router      Router
		View        Renderer

		GRPCServer             *GRPCServer
		grpcServices           []GRPCServices
		grpcUnaryInterceptors  []grpc.UnaryServerInterceptor
		grpcStreamInterceptors []grpc.StreamServerInterceptor

		Path        string
		staticDir   string
		sharedData  map[string]string
		middlewares []HandlerFunc
	}

	// ContextPool contains the framework Context pool.
	ContextPool struct {
		c chan *Context
	}
)

// NewContextPool create a new pool.
func NewContextPool() *ContextPool {
	return &ContextPool{
		c: make(chan *Context),
	}
}

// Get a context from the pool.
func (cp *ContextPool) Get(rw ResponseWriter, req *Request) (c *Context) {
	select {
	case c = <-cp.c:
	// reuse existing context
	default:
		// create new context
		c = NewContext(rw, req)
	}
	return
}

// Put a context inside the pool.
func (cp *ContextPool) Put(c *Context) {
	c.Reset()
	select {
	case cp.c <- c:
	default: // Discard the buffer if the pool is full.
	}
}

// Use appends a new global middleware.
// Global middlewares are called on each request (including static files)
func Use(filter HandlerFunc) {
	App.middlewares = append(App.middlewares, filter)
}

// New initialize the engine and return an Engine struct.
func New() *Engine {
	engine := &Engine{}

	Log = NewLogger(os.Stdout)

	engine.pool = NewContextPool()
	engine.Router = NewRouter()

	engine.sharedData = make(map[string]string)

	// Try to determine engine path
	wd, _ := osext.ExecutableFolder()
	engine.Path = filepath.Clean(wd) + "/"

	if strings.Contains(filepath.Dir(engine.Path), "go-build") {
		wd, _ = os.Getwd()
		engine.Path = filepath.Clean(wd) + "/"
	}

	if filepath.Base(engine.Path) == "tests" {
		engine.Path = engine.Path + "../"
	}

	if flag.Lookup("test.v") != nil {
		SetMode("test")
		engine.Path = engine.Path + "../example"
	}

	// Load configuration file.
	Config = LoadConfig(engine.Path + "config/app.json")

	return engine
}

// Init framework.
func (engine *Engine) Init() {
	var err error

	// Parse views from views directory.
	engine.View, err = NewView(engine.Path + "views")
	if err != nil {
		Log.Error(err)
	}

	// get static directory for public static files (css/js/img)
	engine.staticDir, err = getAbsolutePath(engine.Path + "public")
	if err != nil {
		Log.Error(err)
	}

	// Fancy banner and information about running application
	if !Config.Bool("quiet") {
		Log.Info(fmt.Sprintf("%s", strings.Repeat("=", 80)))
		Log.Info(fmt.Sprintf("%-15s: v%s", "Framework", VERSION))
		Log.Info(fmt.Sprintf("%s", strings.Repeat("=", 80)))
		Log.Info(fmt.Sprintf("%-15s: %s", "Name", Config.String("name")))
		Log.Info(fmt.Sprintf("%-15s: %s", "Author", Config.String("author")))
		Log.Info(fmt.Sprintf("%-15s: %s", "Version", Config.String("version")))
		Log.Info(fmt.Sprintf("%-15s: %s", "Mode", Config.String("mode")))
		Log.Info(fmt.Sprintf("%s", strings.Repeat("=", 80)))
	}

	// Check if caching is enabled: register and assign it to the framework instance
	if Config.Get("cache") != nil {
		if Config.String("mode") == DebugMode {
			Log.Debug(fmt.Sprintf("registering cache: %s", Config.String("cache")))
		}

		switch Config.String("cache") {
		case "file":
			_ = cache.Register("file", cache.NewFileCache)
		case "memory":
			_ = cache.Register("memory", cache.NewMemoryCache)
		}

		engine.cache, err = cache.NewCache(Config.String("cache"), Config.String("cache_config"))
		if err != nil {
			Log.Error(err)
		}
	}

	// Load translations
	_, err = os.Stat("i18n")
	if err == nil {
		if Config.String("mode") == DebugMode {
			Log.Debug("Importing I18N translations.")
		}
		if err := i18n.Load("i18n"); err != nil {
			Log.Error(fmt.Errorf("error loading i18n: %s", err.Error()))
		}
	}

	// Initialize Controllers
	for _, c := range engine.Controllers {
		method := reflect.ValueOf(c).MethodByName("Init")
		if method.IsValid() {
			method.Call(nil)
		}
	}

	// Enable pprof
	if Config.Get("pprof") != nil {
		go func() {
			Log.Info(fmt.Sprintf("pprof enabled and listening on %s", Config.String("pprof")))
			Log.Error(http.ListenAndServe(Config.String("pprof"), nil))
		}()
	}
}

// Run start listening on configured HTTP port.
func Run() {
	App.Init()

	if Config.String("address") != "" {
		Log.Info(fmt.Sprintf("listening on %s", Config.String("address")))
		Log.Error(http.ListenAndServe(Config.String("address"), App.Router))
	} else {
		Log.Info(fmt.Sprintf("listening on port :%d", Config.Int("port")))
		Log.Error(http.ListenAndServe(fmt.Sprintf(":%d", Config.Int("port")), App.Router))
	}
}

// RunTLS start HTTPS listening on configured port.
func RunTLS() {
	App.Init()

	if Config.String("cert") == "" || Config.String("cert_key") == "" {
		panic("Invalid cert files or key. Please review your configuration.")
	}

	if Config.String("address") != "" {
		Log.Info(fmt.Sprintf("listening TLS on %s", Config.String("address")))
		Log.Error(http.ListenAndServeTLS(Config.String("address"), Config.String("cert"), Config.String("cert_key"), App.Router))
	} else {
		Log.Info(fmt.Sprintf("listening TLS on port :%d", Config.Int("port")))
		Log.Error(http.ListenAndServeTLS(fmt.Sprintf(":%d", Config.Int("port")), Config.String("cert"), Config.String("cert_key"), App.Router))
	}
}

// RunGRPC start gRPC listening on configured port.
func RunGRPC() {
	App.Init()

	Log.Info(fmt.Sprintf("listening gRPC on port :%d", Config.Int("grpc_port")))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", Config.Int("grpc_port")))
	if err != nil {
		Log.Fatalf("failed: %v", err)
	}

	var options []grpc.ServerOption

	if Config.String("grpc_cert") != "" && Config.String("grpc_cert_key") != "" {
		config := &tls.Config{}
		cert, err := tls.LoadX509KeyPair(Config.String("grpc_cert"), Config.String("grpc_cert_key"))
		if err != nil {
			Log.Fatal(err)
		}
		config.Certificates = []tls.Certificate{cert}
		options = append(options, grpc.Creds(credentials.NewTLS(config)))
	}

	if len(App.grpcUnaryInterceptors) > 0 {
		options = append(options, grpc.UnaryInterceptor(ServerUnaryInterceptor()))
	}

	if len(App.grpcStreamInterceptors) > 0 {
		options = append(options, grpc.StreamInterceptor(ServerStreamInterceptor()))
	}

	App.GRPCServer = NewGRPC(options...)

	// Initialize gRPC services
	Log.Info("Registering gRPC services")

	var args []reflect.Value
	args = append(args, reflect.ValueOf(App.GRPCServer.Server))

	for _, c := range App.grpcServices {
		method := reflect.ValueOf(c).MethodByName("Init")
		if method.IsValid() {
			method.Call(args)
		}
	}

	Log.Error(App.GRPCServer.Serve(l))
}

// Create a new framework instance on application init.
func init() {
	App = New()
}
