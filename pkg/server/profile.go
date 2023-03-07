package server

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

func fetchProfile(url string, filePath string) (err error) {
	cli := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:  2 * time.Second,
				Deadline: time.Now().Add(3 * time.Second),
			}).Dial,
			TLSHandshakeTimeout: 2 * time.Second,
		},
		Timeout: 120 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := cli.Do(req)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return err
	}

	//save to file
	return writeFile(filePath, data)
}

func writeFile(fpath string, data []byte) error {
	dir := path.Dir(fpath)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if e := os.MkdirAll(dir, 0755); e != nil {
				return err
			}
		} else {
			return err
		}
	}

	return ioutil.WriteFile(fpath, data, 0644)
}
