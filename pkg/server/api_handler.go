package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sjp00556/pprof-browser/pkg/consts"
)

func (m *APIServer) registerAPIRoutes(router *mux.Router) {
	router.NewRoute().
		Methods(http.MethodGet).
		Path(consts.Ping).
		HandlerFunc(m.ping)
	router.NewRoute().
		Methods(http.MethodGet).
		Path(consts.Root).
		Handler(http.FileServer(http.Dir(m.dir + consts.Static)))
	router.NewRoute().
		Methods(http.MethodGet).
		PathPrefix(consts.View + "/{name}").
		HandlerFunc(m.view)
	router.NewRoute().
		Methods(http.MethodPost).
		PathPrefix(consts.View).
		HandlerFunc(m.viewPost)
	router.NewRoute().
		Methods(http.MethodGet).
		Path(consts.Fetch).
		HandlerFunc(m.getFetch)
	router.NewRoute().
		Methods(http.MethodPost).
		Path(consts.Fetch).
		HandlerFunc(m.postFetch)
	router.NewRoute().
		Methods(http.MethodPost).
		Path(consts.Upload).
		HandlerFunc(m.upload)
	router.NewRoute().
		Methods(http.MethodGet).
		PathPrefix(consts.Static).
		Handler(http.FileServer(http.Dir(m.dir)))
}

func (m *APIServer) ping(w http.ResponseWriter, r *http.Request) {
	sendOkReply(w, r, newSuccessHTTPReply("pong"))
}

func (m *APIServer) view(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["name"] == "" {
		sendErrReply(w, r, newErrHTTPReply(fmt.Errorf("no profile name")))
		return
	}

	h, err := m.pprofMuxMgr.GetHandler(vars["name"])
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}

	h.ServeHTTP(w, r)
}

func (m *APIServer) viewPost(w http.ResponseWriter, r *http.Request) {
	fname, err := extractString(r, "fname")
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}
	viewPath := getViewPath(fname)
	http.Redirect(w, r, viewPath, http.StatusSeeOther)
}

func (m *APIServer) getFetch(w http.ResponseWriter, r *http.Request) {
	fname, err := m.doFetch(w, r)
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}
	sendOkReply(w, r, newSuccessHTTPReply(fname))
}

func (m *APIServer) postFetch(w http.ResponseWriter, r *http.Request) {
	fname, err := m.doFetch(w, r)
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}

	viewPath := getViewPath(fname)
	http.Redirect(w, r, viewPath, http.StatusSeeOther)
}

func (m *APIServer) doFetch(w http.ResponseWriter, r *http.Request) (fname string, err error) {
	hostPort, err := extractString(r, "hostPort")
	if err != nil {
		return
	}
	profileType, err := extractString(r, "profileType")
	if err != nil {
		return
	}
	if !consts.ProfileTypes[profileType] {
		return
	}

	//for profile
	seconds, _ := extractUint64(r, "seconds")

	url := fmt.Sprintf("http://%v/debug/pprof/%v", hostPort, profileType)
	if profileType == "profile" && seconds > 0 {
		url += "?seconds=" + fmt.Sprint(seconds)
	}

	fname = time.Now().Format("20060102-150405") + "-" + hostPort + "-" + profileType
	fpath := getFilePath(m.profileDir, fname)
	err = fetchProfile(url, fpath)
	if err != nil {
		return
	}
	return
}

func (m *APIServer) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(consts.MaxUploadSize); err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}
	uploadedFile, fileHandler, err := r.FormFile(consts.FileFormID)
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}
	defer uploadedFile.Close()
	data, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}

	fname := time.Now().Format("20060102-150405") + "-" + fileHandler.Filename
	fpath := getFilePath(m.profileDir, fname)
	err = writeFile(fpath, data)
	if err != nil {
		sendErrReply(w, r, newErrHTTPReply(err))
		return
	}

	viewPath := getViewPath(fname)
	http.Redirect(w, r, viewPath, http.StatusSeeOther)
}
