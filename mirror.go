package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/goproxyio/goproxy/module"
)

func copyFile(from, to string) error {
	var err error
	if _, err = os.Stat(from); err != nil {
		return err
	}
	var buf []byte
	if buf, err = ioutil.ReadFile(from); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(to, buf, 0644)
}

func removeFile(name string) {
	if _, err := os.Stat(name); err == nil {
		os.Remove(name)
	}
}

func zipReplace(rZip, wZip, old, new string) error {
	rf, err := zip.OpenReader(rZip)
	if err != nil {
		return err
	}
	defer rf.Close()
	wf, err := os.Create(wZip)
	if err != nil {
		return err
	}
	defer wf.Close()
	w := zip.NewWriter(wf)
	for _, r := range rf.File {
		rz, err := r.Open()
		if err != nil {
			return err
		}
		wz, err := w.Create(strings.Replace(r.Name, old, new, 1))
		if err != nil {
			return err
		}
		if _, err = io.Copy(wz, rz); err != nil {
			return err
		}
		rz.Close()
	}
	w.Close()
	return nil
}

func mirrorHandler(orig, mirror string, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "goproxy: %s %s download %s\n", r.RemoteAddr, time.Now().Format("2006-01-02 15:04:05"), r.URL.Path)
		if _, err := os.Stat(filepath.Join(cacheDir, r.URL.Path)); err != nil {
			mirrorPath := mirror + strings.TrimPrefix(r.URL.Path, orig)

			if strings.HasSuffix(mirrorPath, "/@v/list") {
				to := filepath.Join(cacheDir, r.URL.Path)
				err = copyFile(filepath.Join(cacheDir, mirrorPath), to)
				if err == nil {
					inner.ServeHTTP(w, r)
				} else {
					removeFile(to)
					w.Write([]byte{})
				}
				return
			}

			if strings.HasSuffix(mirrorPath, "/@latest") {
				path := strings.TrimSuffix(mirrorPath, "/@latest")
				path = strings.TrimPrefix(path, "/")
				path, err = module.DecodePath(path)
				if err != nil {
					ReturnServerError(w, err)
					return
				}
				goGet(path, "latest", "", w, r)
			}

			suffix := path.Ext(mirrorPath)
			if suffix == ".info" || suffix == ".mod" || suffix == ".zip" {
				mod := strings.Split(mirrorPath, "/@v/")
				if len(mod) != 2 {
					ReturnBadRequest(w, fmt.Errorf("bad module path:%s", mirrorPath))
					return
				}
				path := strings.TrimPrefix(mod[0], "/")
				path, err = module.DecodePath(path)
				if err != nil {
					ReturnServerError(w, err)
					return
				}
				version := strings.TrimSuffix(mod[1], suffix)
				version, err = module.DecodeVersion(version)
				if err != nil {
					ReturnServerError(w, err)
					return
				}
				goGet(path, version, suffix, w, r)

				modFunc := func() error {
					from := filepath.Join(cacheDir, mirrorPath)
					var err error
					if _, err = os.Stat(from); err != nil {
						return err
					}
					var buf []byte
					if buf, err = ioutil.ReadFile(from); err != nil {
						return err
					}
					to := filepath.Join(cacheDir, r.URL.Path)
					if err = os.MkdirAll(filepath.Dir(to), 0755); err != nil {
						return err
					}
					ss := strings.SplitN(string(buf), "\n", 2)
					// Maybe is a pseudo 'go.mod' from modfetch.ImportRepoRev(path, rev)
					if len(ss) == 1 || ss[1] == "" {
						ss[0] = strings.Replace(ss[0], strings.TrimPrefix(mirror, "/"), strings.TrimPrefix(orig, "/"), 1)
						buf = []byte(ss[0] + "\n")
					}
					err = ioutil.WriteFile(to, buf, 0644)
					if err != nil {
						removeFile(to)
					}
					return err
				}
				zipFunc := func() error {
					from := filepath.Join(cacheDir, mirrorPath)
					if _, err := os.Stat(from); err != nil {
						return err
					}
					to := filepath.Join(cacheDir, r.URL.Path)
					if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
						return err
					}
					err := zipReplace(from, to, strings.TrimPrefix(mirror, "/"), strings.TrimPrefix(orig, "/"))
					if err != nil {
						removeFile(to)
					}
					return err
				}
				switch suffix {
				case ".info":
					to := filepath.Join(cacheDir, r.URL.Path)
					err = copyFile(filepath.Join(cacheDir, mirrorPath), to)
					if err != nil {
						removeFile(to)
					}
				case ".mod":
					modFunc()
				case ".zip":
					zipFunc()
				}
			}
		}
		inner.ServeHTTP(w, r)
	})
}
