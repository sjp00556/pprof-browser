package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sjp00556/pprof-browser/pkg/consts"
	"github.com/sjp00556/pprof-browser/pkg/model"
	"github.com/sjp00556/pprof-browser/pkg/util/config"
)

// Server represents the server in a cluster
type APIServer struct {
	port        string
	dir         string
	profileDir  string
	staticDir   string
	wg          sync.WaitGroup
	apiServer   *http.Server
	pprofMuxMgr *pprofMuxMgr
}

// NewServer creates a new server
func NewAPIServer() *APIServer {
	return &APIServer{}
}

// Start starts a server
func (m *APIServer) Start(cfg *config.Config) (err error) {
	m.port = cfg.GetString(consts.CfgKeyPort)
	m.dir = cfg.GetString(consts.CfgKeyDir)
	m.profileDir = path.Join(m.dir, "profile")
	m.staticDir = path.Join(m.dir, "static")
	m.pprofMuxMgr = newPProfMuxMgr(m.profileDir)
	m.startHTTPService()
	m.wg.Add(1)
	return nil
}

// Shutdown closes the server
func (m *APIServer) Shutdown() {
	var err error
	if m.apiServer != nil {
		if err = m.apiServer.Shutdown(context.Background()); err != nil {
			log.Printf("action[Shutdown] failed, err: %v", err)
		}
	}
	m.wg.Done()
}

// Sync waits for the execution termination of the server
func (m *APIServer) Sync() {
	m.wg.Wait()
}

func (m *APIServer) startHTTPService() {
	router := mux.NewRouter().SkipClean(true)
	m.registerAPIRoutes(router)
	var s = &http.Server{
		Addr:    ":" + m.port,
		Handler: router,
	}
	var serveAPI = func() {
		if err := s.ListenAndServe(); err != nil {
			log.Printf("serveAPI: serve http server failed: err(%v)", err)
			return
		}
	}
	go serveAPI()
	m.apiServer = s
	return
}

func newSuccessHTTPReply(data interface{}) *model.HTTPReply {
	return &model.HTTPReply{Code: http.StatusOK, Msg: "", Data: data}
}

func newErrHTTPReply(err error) *model.HTTPReply {
	return newErrHTTPReplyWithData(nil, err)
}

func newErrHTTPReplyWithData(data interface{}, err error) *model.HTTPReply {
	if err == nil {
		return newSuccessHTTPReply(data)
	}
	return &model.HTTPReply{Code: 400, Msg: err.Error(), Data: data}
}

func sendOkReply(w http.ResponseWriter, r *http.Request, httpReply *model.HTTPReply) {
	reply, err := json.Marshal(httpReply)
	if err != nil {
		log.Printf("fail to marshal http reply[%v]. URL[%v], remoteAddr[%v] err:[%v]", httpReply, r.URL, r.RemoteAddr, err)
		http.Error(w, "fail to marshal http reply", http.StatusBadRequest)
		return
	}
	send(w, r, reply)
	return
}

func send(w http.ResponseWriter, r *http.Request, reply []byte) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(reply)))
	if _, err := w.Write(reply); err != nil {
		log.Printf("fail to write http reply[%s] len[%d].URL[%v],remoteAddr[%v] err:[%v]", string(reply), len(reply), r.URL, r.RemoteAddr, err)
		return
	}
	log.Printf("URL[%v],remoteAddr[%v],response ok", r.URL, r.RemoteAddr)
	return
}

func sendErrReply(w http.ResponseWriter, r *http.Request, httpReply *model.HTTPReply) {
	log.Printf("URL[%v],remoteAddr[%v],response err[%v]", r.URL, r.RemoteAddr, httpReply)
	reply, err := json.Marshal(httpReply)
	if err != nil {
		log.Printf("fail to marshal http reply[%v]. URL[%v],remoteAddr[%v] err:[%v]", httpReply, r.URL, r.RemoteAddr, err)
		http.Error(w, "fail to marshal http reply", http.StatusBadRequest)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(reply)))
	if _, err = w.Write(reply); err != nil {
		log.Printf("fail to write http reply[%s] len[%d].URL[%v],remoteAddr[%v] err:[%v]", string(reply), len(reply), r.URL, r.RemoteAddr, err)
	}
	return
}

func sendReply(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	if err != nil {
		sendErrReply(w, r, newErrHTTPReplyWithData(data, err))
		return
	}
	sendOkReply(w, r, newSuccessHTTPReply(data))
}
