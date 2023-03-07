package server

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/google/pprof/driver"
	"github.com/sjp00556/pprof-browser/pkg/consts"
	"github.com/sjp00556/pprof-browser/pkg/util"
)

func newPProfMuxMgr(profileDir string) *pprofMuxMgr {
	mgr := &pprofMuxMgr{pp: make(map[string]*pprofMux, 0), profileDir: profileDir}
	go mgr.release()
	return mgr
}

type pprofMuxMgr struct {
	sync.Mutex
	pp         map[string]*pprofMux
	viewPath   string
	profileDir string
}

func (pm *pprofMuxMgr) GetHandler(name string) (http.Handler, error) {
	pm.Lock()
	p := pm.pp[name]
	if p != nil {
		p.accessTime = time.Now().Unix()
		pm.Unlock()
		return p.handler, nil
	}
	pm.Unlock()

	//check profile file exist
	filePath := getFilePath(pm.profileDir, name)
	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}

	// start the pprof web handler: pass -http and -no_browser so it starts the
	// handler but does not try to launch a browser
	// our startHTTP will do the appropriate interception
	flags := util.NewPProfFlags([]string{"-http=localhost:0", "-no_browser", filePath})

	p = newPProfMux(name)
	options := &driver.Options{
		Flagset:    flags,
		HTTPServer: p.startHTTP,
	}
	if err := driver.PProf(options); err != nil {
		return nil, err
	}

	if p.handler == nil {
		return nil, fmt.Errorf("no handler")
	}

	pm.Lock()
	pm.pp[name] = p
	pm.Unlock()
	return p.handler, nil
}

func (pm *pprofMuxMgr) release() {
	t := time.NewTicker(time.Minute * 5)
	defer t.Stop()
	for {
		<-t.C
		pm.Lock()
		for _, p := range pm.pp {
			if p.accessTime < time.Now().Unix()-3600 {
				delete(pm.pp, p.name)
			}
		}
		pm.Unlock()
	}
}

func newPProfMux(name string) *pprofMux {
	return &pprofMux{name: name, accessTime: time.Now().Unix()}
}

type pprofMux struct {
	handler    http.Handler
	name       string
	accessTime int64
}

func (p *pprofMux) startHTTP(args *driver.HTTPServerArgs) error {
	mux := http.NewServeMux()
	viewPath := getViewPath(p.name)
	for pattern, handler := range args.Handlers {
		var joinedPattern string
		if pattern == "/" {
			joinedPattern = viewPath
		} else {
			joinedPattern = path.Join(viewPath, pattern)
		}
		mux.Handle(joinedPattern, handler)
	}

	// enable gzip compression: flamegraphs can be big!
	p.handler = gziphandler.GzipHandler(mux)
	return nil
}

func getViewPath(fname string) string {
	return path.Join(consts.View, fname) + "/"
}

func getFilePath(dir, fname string) string {
	arrName := strings.Split(fname, "-")
	prefix := "file"
	if len(arrName) > 1 {
		prefix = arrName[0]
	}
	return path.Join(dir, prefix, fname)
}
