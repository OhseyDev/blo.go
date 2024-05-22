package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/dimfeld/httptreemux"
	"github.com/OhseyDev/gospirit/configuration"
	"github.com/OhseyDev/gospirit/database"
	"github.com/OhseyDev/gospirit/filenames"
	"github.com/OhseyDev/gospirit/flags"
	"github.com/OhseyDev/gospirit/https"
	"github.com/OhseyDev/gospirit/plugins"
	"github.com/OhseyDev/gospirit/server"
	"github.com/OhseyDev/gospirit/structure/methods"
	"github.com/OhseyDev/gospirit/templates"
	"github.com/justinas/alice"
)

func httpsRedirect(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	http.Redirect(w, r, configuration.Config.HttpsUrl+r.RequestURI, http.StatusMovedPermanently)
	return
}

func main() {
	// Setup
	var err error
	// GOMAXPROCS - Maybe not needed
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Write log to file if the log flag was provided
	if flags.Log != "" {
		logFile, err := os.OpenFile(flags.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error: Couldn't open log file: " + err.Error())
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	// Configuration is read from config.json by loading the configuration package
	// Database
	if err = database.Initialize(); err != nil {
		log.Fatal("Error: Couldn't initialize database:", err)
		return
	}

	// Global blog data
	if err = methods.GenerateBlog(); err != nil {
		log.Fatal("Error: Couldn't generate blog data:", err)
		return
	}

	// Templates
	if err = templates.Generate(); err != nil {
		log.Fatal("Error: Couldn't compile templates:", err)
		return
	}

	// Plugins
	if err = plugins.Load(); err == nil {
		// Close LuaPool at the end
		defer plugins.LuaPool.Shutdown()
		log.Println("Plugins loaded.")
	}

	// HTTP(S) Server
	httpPort := configuration.Config.HttpHostAndPort
	httpsPort := configuration.Config.HttpsHostAndPort
	// Check if HTTP/HTTPS flags were provided
	if flags.HttpPort != "" {
		components := strings.SplitAfterN(httpPort, ":", 2)
		httpPort = components[0] + flags.HttpPort
	}
	if flags.HttpsPort != "" {
		components := strings.SplitAfterN(httpsPort, ":", 2)
		httpsPort = components[0] + flags.HttpsPort
	}
	httpRouter := httptreemux.New()
	httpsRouter := httptreemux.New()
	
	server.InitializeBlog(httpRouter)
	server.InitializePages(httpRouter)
	
	switch configuration.Config.HttpsUsage {
	case "AdminOnly":
		httpsRouter = httptreemux.New()
		server.InitializeBlog(httpsRouter)
		server.InitializePages(httpsRouter)
		httpRouter.GET("/admin/", httpsRedirect)
		httpRouter.GET("/admin/*path", httpsRedirect)
		server.InitializeAdmin(httpsRouter)
		log.Println("Starting https server on port " + httpsPort + "...")
		go func() {
			chain := alice.New(server.CheckHost).Then(httpsRouter)
			err := http.ListenAndServeTLS(httpsPort, filenames.HttpsCertFilename, filenames.HttpsKeyFilename, chain)
			if err != nil {
				log.Fatal("Error: Couldn't start the HTTPS server:", err)
			}
		}()
		log.Println("Starting http server on port " + httpPort + "...")
		chain := alice.New(server.CheckHost).Then(httpRouter)
		err := http.ListenAndServe(httpPort, chain)
		if err != nil {
			log.Fatal("Error: Couldn't start the HTTP server:", err)
		}
	case "All":
		httpsRouter = httptreemux.New()
		server.InitializeBlog(httpsRouter)
		server.InitializePages(httpsRouter)
		server.InitializeAdmin(httpsRouter)
		httpRouter.GET("/", httpsRedirect)
		httpRouter.GET("/*path", httpsRedirect)
		log.Println("Starting https server on port " + httpsPort + "...")
		go func() {
			chain := alice.New(server.CheckHost).Then(httpsRouter)
			err := http.ListenAndServeTLS(httpsPort, filenames.HttpsCertFilename, filenames.HttpsKeyFilename, chain)
			if err != nil {
				log.Fatal("Error: Couldn't start the HTTPS server:", err)
			}
		}()
		log.Println("Starting http server on port " + httpPort + "...")
		chain := alice.New(server.CheckHost).Then(httpRouter)
		err := http.ListenAndServe(httpPort, chain)
		if err != nil {
			log.Fatal("Error: Couldn't start the HTTP server:", err)
		}
	default:
		server.InitializeAdmin(httpRouter)
		log.Println("Starting server without HTTPS support. Please enable HTTPS in " + filenames.ConfigFilename + " to improve security.")
		log.Println("Starting http server on port " + httpPort + "...")
		chain := alice.New(server.CheckHost).Then(httpRouter)
		err := http.ListenAndServe(httpPort, chain)
		if err != nil {
			log.Fatal("Error: Couldn't start the HTTP server:", err)
		}
	}
}
