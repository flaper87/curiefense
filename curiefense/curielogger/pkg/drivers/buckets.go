package drivers

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"io"
	"log"
	"sync"
	"time"
)

type GCS struct {
	bucket, prefix string
	storageClient  *blob.Bucket
	w              *io.PipeWriter
	writeCancel    context.CancelFunc
	size           *atomic.Int64

	closed *atomic.Bool
	wg     *sync.WaitGroup
	lock   *sync.Mutex
}

func NewGCS(v *viper.Viper) *GCS {
	log.Print(`initialized bucket export`)
	g := &GCS{
		bucket: v.GetString(`EXPORT_BUCKET_URL`),
		prefix: v.GetString(`EXPORT_BUCKET_PREFIX`),
		closed: atomic.NewBool(false),
		wg:     &sync.WaitGroup{},
		lock:   &sync.Mutex{},
		size:   atomic.NewInt64(0),
	}
	var err error
	g.storageClient, err = blob.OpenBucket(context.Background(), g.bucket)
	if err != nil {
		log.Print(err)
		return nil
	}
	g.rotateUploader()
	go g.flusher(v.GetDuration(`BUCKET_FLUSH_SECONDS`) * time.Second)
	return g
}

func (g *GCS) rotateUploader() {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.closed.Load() {
		return
	}
	if g.size.Load() == 0 {
		g.writeCancel()
	} else {
		if g.w != nil {
			g.w.Close()
		}
	}
	t := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	g.writeCancel = cancel
	w, err := g.storageClient.NewWriter(ctx, fmt.Sprintf(`%s/created_date=%s/hour=%s/%s.json.gz`, g.prefix, t.Format(`2006-01-02`), t.Format(`15`), uuid.New().String()), &blob.WriterOptions{
		ContentEncoding: "gzip",
		ContentType:     "application/json",
	})
	if err != nil {
		log.Println(err)
		return
	}
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
	g.size.Inc()
	return g.w.Write(p)
}

func (g *GCS) Close() error {
	g.closed.Store(true)
	defer g.wg.Wait()
	if g.size.Load() == 0 {
		g.writeCancel()
		return nil
	}
	g.w.Close()
	return nil
}
