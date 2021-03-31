package drivers

import (
	"cloud.google.com/go/storage"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io"
	"log"
	"sync"
	"time"
)

type GCS struct {
	bucket, prefix, serviceAccount string
	storageClient                  *storage.Client
	w                              *io.PipeWriter

	closed *atomic.Bool
	wg     *sync.WaitGroup
	lock   *sync.Mutex
}

func NewGCS(v *viper.Viper) *GCS {
	g := &GCS{
		bucket:         v.GetString(`GCS_EXPORT_BUCKET`),
		prefix:         v.GetString(`GCS_EXPORT_BUCKET_PREFIX`),
		serviceAccount: v.GetString(`GCS_EXPORT_SERVICE_ACCOUNT`),
		closed:         atomic.NewBool(false),
		wg:             &sync.WaitGroup{},
		lock:           &sync.Mutex{},
	}
	var err error
	g.storageClient, err = storage.NewClient(context.Background(), option.WithTokenSource(google.ComputeTokenSource(g.serviceAccount)))
	if err != nil {
		log.Print(err)
		return nil
	}
	g.rotateUploader()
	go g.flusher(v.GetDuration(`GCS_FLUSH_SECONDS`) * time.Second)
	return g
}

func (g *GCS) rotateUploader() {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.closed.Load() {
		return
	}
	if g.w != nil {
		g.w.Close()
	}
	t := time.Now()
	w := g.storageClient.Bucket(g.bucket).Object(fmt.Sprintf(`%s/created_date=%s/hour=%s/%s.json.gz`, g.prefix, t.Format(`2006-01-02`), t.Format(`15`), uuid.New().String())).NewWriter(context.Background())
	pr, pw := io.Pipe()
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer w.Close()
		gzw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			log.Print(err)
			return
		}
		defer gzw.Close()
		io.Copy(w, pr)
	}()
	g.w = pw
}

func (g *GCS) flusher(duration time.Duration) {
	if duration.Seconds() < 1 {
		duration = time.Second
	}
	t := time.NewTicker(duration)
	for range t.C {
		g.rotateUploader()
	}
}

func (g *GCS) Write(p []byte) (n int, err error) {
	return g.Write(p)
}

func (g *GCS) Close() error {
	g.closed.Store(true)
	g.w.Close()
	g.wg.Wait()
	return nil
}