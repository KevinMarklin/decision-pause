package handler

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// staticHandler — раздача собранного Mini App (SPA) из dir.
//
// Правила:
//   - существующий файл (index.html, JS/CSS с хешем) отдаётся как есть;
//   - запрос без расширения, которого нет на диске (клиентский маршрут) → index.html;
//   - отсутствующий файл с расширением → 404, а не index.html (иначе битый
//     скрипт получил бы HTML и упал бы с syntax error).
//
// Путь нормализуется через path.Clean — выход за пределы dir невозможен.
func staticHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		rel := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if st, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err == nil && !st.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		if path.Ext(rel) == "" {
			index := filepath.Join(dir, "index.html")
			if _, err := os.Stat(index); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.ServeFile(w, r, index)
				return
			}
		}
		http.NotFound(w, r)
	})
}
