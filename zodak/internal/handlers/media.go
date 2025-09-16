package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type MediaItem struct {
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

func ListMedia(c *gin.Context) {
	root := "/var/www/zodak/public/media"
	ents, err := os.ReadDir(root)
	if err != nil { c.JSON(500, gin.H{"error":"fs"}); return }

	var items []MediaItem
	allow := map[string]bool{".mp4":true,".webm":true,".mov":true,".m3u8":true}
	for _, e := range ents {
		if e.IsDir() {
			if st, err := os.Stat(filepath.Join(root, e.Name(), "index.m3u8")); err==nil {
				items = append(items, MediaItem{URL: "/media/"+e.Name()+"/index.m3u8", Title: e.Name(), Size: st.Size(), Modified: st.ModTime()})
			}
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !allow[ext] { continue }
		if info, _ := e.Info(); info != nil {
			name := strings.TrimSuffix(e.Name(), ext)
			items = append(items, MediaItem{
				URL: "/media/"+e.Name(), Title: name, Size: info.Size(), Modified: info.ModTime(),
			})
		}
	}
	sort.Slice(items, func(i,j int)bool{ return items[i].Modified.After(items[j].Modified) })
	c.JSON(http.StatusOK, gin.H{"items": items})
}

