package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log2 "github.com/labstack/gommon/log"
	"github.com/nuts-foundation/nuts-demo-ehr/api"
	"github.com/nuts-foundation/nuts-demo-ehr/client"
	"github.com/nuts-foundation/nuts-demo-ehr/domain/customers"
	bolt "go.etcd.io/bbolt"
)

const assetPath = "web/dist"

//go:embed web/dist/*
var embeddedFiles embed.FS

const apiTimeout = 10 * time.Second

func getFileSystem(useFS bool) http.FileSystem {
	if useFS {
		log.Print("using live mode")
		return http.FS(os.DirFS(assetPath))
	}

	log.Print("using embed mode")
	fsys, err := fs.Sub(embeddedFiles, assetPath)
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}

func main() {
	// config stuff
	config := loadConfig()
	config.Print(log.Writer())
	// load bbolt db
	db, err := bolt.Open(config.DBFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// init node API client
	nodeClient := client.HTTPClient{NutsNodeAddress: config.NutsNodeAddress}

	// auth stuff
	auth := api.NewAuth(nodeClient)

	// Initialize wrapper
	apiWrapper := api.Wrapper{
		Auth:       auth,
		Client:     nodeClient,
		Repository: customers.NewJsonFileRepository(config.CustomersFile),
	}
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Logger.SetLevel(log2.DEBUG)
	e.HTTPErrorHandler = func(err error, ctx echo.Context) {
		if !ctx.Response().Committed {
			ctx.Response().Write([]byte(err.Error()))
			ctx.Echo().Logger.Error(err)
		}
	}
	e.Use(auth.VPHandler)
	e.HTTPErrorHandler = httpErrorHandler



	api.RegisterHandlers(e, apiWrapper)

	// Setup asset serving:
	// Check if we use live mode from the file system or using embedded files
	useFS := len(os.Args) > 1 && os.Args[1] == "live"
	assetHandler := http.FileServer(getFileSystem(useFS))
	e.GET("/*", echo.WrapHandler(assetHandler))

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.HTTPPort)))
}

// httpErrorHandler includes the err.Err() string in a { "error": "msg" } json hash
func httpErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)
	type Map map[string]interface{}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
		if he.Internal != nil {
			err = fmt.Errorf("%v, %v", err, he.Internal)
		}
	} else {
		msg = err.Error()
	}

	if _, ok := msg.(string); ok {
		msg = Map{"error": msg}
	}

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			c.Logger().Error(err)
		}
	}
}
