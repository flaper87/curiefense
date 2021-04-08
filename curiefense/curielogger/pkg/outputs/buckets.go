package outputs

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pierrec/lz4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"io"
	"sync"
	"time"
)

type Bucket struct {
	bucket, prefix string
	storageClient  *blob.Bucket
	w              *io.PipeWriter
	writeCancel    context.CancelFunc
	size           *atomic.Int64

	closed *atomic.Bool
	wg     *sync.WaitGroup
	lock   *sync.Mutex
}

func NewBucket(v *viper.Viper) *Bucket {
	log.Info(`initialized bucket export`)
	g := &Bucket{
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
		log.Error(err)
		return nil
	}
	g.rotateUploader()
	go g.flusher(v.GetDuration(`BUCKET_FLUSH_SECONDS`) * time.Second)
	return g
}

func (g *Bucket) rotateUploader() {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.closed.Load() {
		return
	}
	if g.size.Load() == 0 {
		if g.writeCancel != nil {
			g.writeCancel()
		}
	} else {
		if g.w != nil {
			g.w.Close()
		}
	}
	t := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	g.writeCancel = cancel
	w, err := g.storageClient.NewWriter(ctx, fmt.Sprintf(`%s/created_date=%s/hour=%s/%s.json.lz4`, g.prefix, t.Format(`2006-01-02`), t.Format(`15`), uuid.New().String()), &blob.WriterOptions{
		ContentEncoding: "lz4",
		ContentType:     "application/json",
	})
	if err != nil {
		log.Error(err)
		return
	}
	pr, pw := io.Pipe()
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer w.Close()
		gzw := lz4.NewWriter(w)
		defer gzw.Close()
		io.Copy(gzw, pr)
	}()
	g.w = pw
}

func (g *Bucket) flusher(duration time.Duration) {
	if duration.Seconds() < 1 {
		duration = time.Second
	}
	t := time.NewTicker(duration)
	for range t.C {
		g.rotateUploader()
	}
}

func (g *Bucket) Write(p []byte) (n int, err error) {
	g.size.Inc()
	return g.w.Write(p)
}

func (g *Bucket) Close() error {
	g.closed.Store(true)
	defer g.wg.Wait()
	if g.size.Load() == 0 {
		g.writeCancel()
		return nil
	}
	g.w.Close()
	return nil
}
